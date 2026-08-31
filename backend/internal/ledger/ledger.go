// Package ledger is a minimal double-entry accounting engine. Every posted
// transaction is a set of signed entries that sum to zero; an account's balance
// is the sum of its entries. Nothing else in the system is allowed to mutate
// balances.
package ledger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Account kinds.
const (
	AccountEscrow          = "escrow"           // owner = milestone id
	AccountUserBalance     = "user_balance"     // owner = user id
	AccountPayoutClearing  = "payout_clearing"  // owner = user id
	AccountPlatformRevenue = "platform_revenue" // singleton
	AccountExternal        = "external"         // singleton — the outside world
)

// Transaction kinds.
const (
	TxnEscrowFund     = "escrow_fund"
	TxnEscrowRelease  = "escrow_release"
	TxnRefund         = "refund"
	TxnPayout         = "payout"
	TxnPayoutReversal = "payout_reversal"
)

// Posting is one leg of a transaction. Amount is signed: positive credits the
// account (balance up), negative debits it.
type Posting struct {
	AccountKind string
	OwnerID     *uuid.UUID
	AmountCents int64
}

type TxnInput struct {
	Kind           string
	IdempotencyKey string
	Reference      string
	ContractID     *uuid.UUID
	MilestoneID    *uuid.UUID
	Currency       string
	Postings       []Posting
}

var (
	ErrUnbalanced = errors.New("ledger: postings do not sum to zero")
	ErrTooFewLegs = errors.New("ledger: a transaction needs at least two postings")
)

type Ledger struct{ db *gorm.DB }

func New(db *gorm.DB) *Ledger { return &Ledger{db: db} }

type accountRow struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Kind      string
	OwnerID   *uuid.UUID `gorm:"type:uuid"`
	Currency  string
	CreatedAt time.Time
}

func (accountRow) TableName() string { return "ledger_accounts" }

type txnRow struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	Kind           string
	IdempotencyKey string
	Reference      string
	ContractID     *uuid.UUID `gorm:"type:uuid"`
	MilestoneID    *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time
}

func (txnRow) TableName() string { return "ledger_transactions" }

type entryRow struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	TransactionID uuid.UUID `gorm:"type:uuid"`
	AccountID     uuid.UUID `gorm:"type:uuid"`
	AmountCents   int64
	CreatedAt     time.Time
}

func (entryRow) TableName() string { return "ledger_entries" }

// Post records a transaction. It is idempotent on IdempotencyKey: a repeated
// call with the same key is a no-op that returns the original transaction id.
func (l *Ledger) Post(ctx context.Context, in TxnInput) (uuid.UUID, error) {
	if len(in.Postings) < 2 {
		return uuid.Nil, ErrTooFewLegs
	}
	var sum int64
	for _, p := range in.Postings {
		sum += p.AmountCents
	}
	if sum != 0 {
		return uuid.Nil, fmt.Errorf("%w (got %d)", ErrUnbalanced, sum)
	}
	currency := in.Currency
	if currency == "" {
		currency = "RUB"
	}

	var txnID uuid.UUID
	err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Idempotency: try to claim the key.
		txn := txnRow{
			ID: uuid.New(), Kind: in.Kind, IdempotencyKey: in.IdempotencyKey,
			Reference: in.Reference, ContractID: in.ContractID, MilestoneID: in.MilestoneID,
		}
		res := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "idempotency_key"}}, DoNothing: true}).Create(&txn)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// Already posted — load and return it.
			var existing txnRow
			if err := tx.First(&existing, "idempotency_key = ?", in.IdempotencyKey).Error; err != nil {
				return err
			}
			txnID = existing.ID
			return nil
		}
		txnID = txn.ID

		for _, p := range in.Postings {
			acc, err := getOrCreateAccount(tx, p.AccountKind, p.OwnerID, currency)
			if err != nil {
				return err
			}
			if err := tx.Create(&entryRow{
				ID: uuid.New(), TransactionID: txn.ID, AccountID: acc.ID, AmountCents: p.AmountCents,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return txnID, err
}

// Balance returns the current balance of one account (0 if it has no entries).
func (l *Ledger) Balance(ctx context.Context, kind string, ownerID *uuid.UUID, currency string) (int64, error) {
	if currency == "" {
		currency = "RUB"
	}
	q := l.db.WithContext(ctx).Model(&entryRow{}).
		Joins("JOIN ledger_accounts a ON a.id = ledger_entries.account_id").
		Where("a.kind = ? AND a.currency = ?", kind, currency)
	if ownerID == nil {
		q = q.Where("a.owner_id IS NULL")
	} else {
		q = q.Where("a.owner_id = ?", *ownerID)
	}
	var sum *int64
	if err := q.Select("SUM(ledger_entries.amount_cents)").Scan(&sum).Error; err != nil {
		return 0, err
	}
	if sum == nil {
		return 0, nil
	}
	return *sum, nil
}

func getOrCreateAccount(tx *gorm.DB, kind string, ownerID *uuid.UUID, currency string) (accountRow, error) {
	lookup := func() (accountRow, bool, error) {
		var found accountRow
		q := tx.Where("kind = ? AND currency = ?", kind, currency)
		if ownerID == nil {
			q = q.Where("owner_id IS NULL")
		} else {
			q = q.Where("owner_id = ?", *ownerID)
		}
		err := q.First(&found).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return accountRow{}, false, nil
		}
		return found, err == nil, err
	}

	if acc, ok, err := lookup(); err != nil || ok {
		return acc, err
	}

	// Singleton accounts are pre-seeded by migration; only owner-scoped
	// accounts are created here. Two concurrent first-uses can race, so the
	// insert is a no-op on conflict and we re-read.
	acc := accountRow{ID: uuid.New(), Kind: kind, OwnerID: ownerID, Currency: currency}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "kind"}, {Name: "owner_id"}, {Name: "currency"}},
		DoNothing: true,
	}).Create(&acc).Error; err != nil {
		return accountRow{}, err
	}

	found, ok, err := lookup()
	if err != nil {
		return accountRow{}, err
	}
	if !ok {
		return accountRow{}, fmt.Errorf("ledger: account %s/%v not found after create", kind, ownerID)
	}
	return found, nil
}
