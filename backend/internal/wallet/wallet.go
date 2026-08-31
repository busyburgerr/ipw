// Package wallet gives the rest of the system a small, domain-meaningful API
// over the double-entry ledger: fund/release/refund escrow, hold/settle payouts,
// and read balances. It never touches ledger tables directly except through
// ledger.Ledger.
package wallet

import (
	"context"
	"fmt"

	"ipw/internal/ledger"

	"github.com/google/uuid"
)

type Service struct {
	l             *ledger.Ledger
	commissionBps int64
	currency      string
}

func NewService(l *ledger.Ledger, commissionBps int64) *Service {
	return &Service{l: l, commissionBps: commissionBps, currency: "RUB"}
}

// AvailableBalance is what a freelancer can withdraw right now.
func (s *Service) AvailableBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.l.Balance(ctx, ledger.AccountUserBalance, &userID, s.currency)
}

// PendingPayout is the amount currently held in payout clearing for a user.
func (s *Service) PendingPayout(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.l.Balance(ctx, ledger.AccountPayoutClearing, &userID, s.currency)
}

// EscrowBalance is what is currently held for a milestone.
func (s *Service) EscrowBalance(ctx context.Context, milestoneID uuid.UUID) (int64, error) {
	return s.l.Balance(ctx, ledger.AccountEscrow, &milestoneID, s.currency)
}

// FundEscrow records that `amount` has entered the system for a milestone.
func (s *Service) FundEscrow(ctx context.Context, contractID, milestoneID uuid.UUID, amount int64, reference string) error {
	_, err := s.l.Post(ctx, ledger.TxnInput{
		Kind:           ledger.TxnEscrowFund,
		IdempotencyKey: "escrow_fund:" + milestoneID.String(),
		Reference:      reference,
		ContractID:     &contractID,
		MilestoneID:    &milestoneID,
		Currency:       s.currency,
		Postings: []ledger.Posting{
			{AccountKind: ledger.AccountExternal, AmountCents: -amount},
			{AccountKind: ledger.AccountEscrow, OwnerID: &milestoneID, AmountCents: amount},
		},
	})
	return err
}

// ReleaseEscrow splits a milestone's escrow between the freelancer's balance and
// platform revenue. Returns the net paid and the commission withheld.
func (s *Service) ReleaseEscrow(ctx context.Context, contractID, milestoneID, freelancerID uuid.UUID, amount int64) (net, commission int64, err error) {
	commission = amount * s.commissionBps / 10000
	net = amount - commission
	_, err = s.l.Post(ctx, ledger.TxnInput{
		Kind:           ledger.TxnEscrowRelease,
		IdempotencyKey: "escrow_release:" + milestoneID.String(),
		ContractID:     &contractID,
		MilestoneID:    &milestoneID,
		Currency:       s.currency,
		Postings: []ledger.Posting{
			{AccountKind: ledger.AccountEscrow, OwnerID: &milestoneID, AmountCents: -amount},
			{AccountKind: ledger.AccountUserBalance, OwnerID: &freelancerID, AmountCents: net},
			{AccountKind: ledger.AccountPlatformRevenue, AmountCents: commission},
		},
	})
	return net, commission, err
}

// RefundEscrow returns a funded milestone's money to the outside world (a real
// refund to the client then happens out of band).
func (s *Service) RefundEscrow(ctx context.Context, contractID, milestoneID uuid.UUID, amount int64) error {
	_, err := s.l.Post(ctx, ledger.TxnInput{
		Kind:           ledger.TxnRefund,
		IdempotencyKey: "refund:" + milestoneID.String(),
		ContractID:     &contractID,
		MilestoneID:    &milestoneID,
		Currency:       s.currency,
		Postings: []ledger.Posting{
			{AccountKind: ledger.AccountEscrow, OwnerID: &milestoneID, AmountCents: -amount},
			{AccountKind: ledger.AccountExternal, AmountCents: amount},
		},
	})
	return err
}

// HoldForPayout moves money from a freelancer's available balance into payout
// clearing while a withdrawal is processed.
func (s *Service) HoldForPayout(ctx context.Context, freelancerID, payoutID uuid.UUID, amount int64) error {
	_, err := s.l.Post(ctx, ledger.TxnInput{
		Kind:           ledger.TxnPayout,
		IdempotencyKey: "payout_hold:" + payoutID.String(),
		Reference:      payoutID.String(),
		Currency:       s.currency,
		Postings: []ledger.Posting{
			{AccountKind: ledger.AccountUserBalance, OwnerID: &freelancerID, AmountCents: -amount},
			{AccountKind: ledger.AccountPayoutClearing, OwnerID: &freelancerID, AmountCents: amount},
		},
	})
	return err
}

// SettlePayout finalises a payout that has actually been sent.
func (s *Service) SettlePayout(ctx context.Context, freelancerID, payoutID uuid.UUID, amount int64) error {
	_, err := s.l.Post(ctx, ledger.TxnInput{
		Kind:           ledger.TxnPayout,
		IdempotencyKey: "payout_settle:" + payoutID.String(),
		Reference:      payoutID.String(),
		Currency:       s.currency,
		Postings: []ledger.Posting{
			{AccountKind: ledger.AccountPayoutClearing, OwnerID: &freelancerID, AmountCents: -amount},
			{AccountKind: ledger.AccountExternal, AmountCents: amount},
		},
	})
	return err
}

// ReversePayout returns held money to a freelancer when a payout is rejected.
func (s *Service) ReversePayout(ctx context.Context, freelancerID, payoutID uuid.UUID, amount int64) error {
	_, err := s.l.Post(ctx, ledger.TxnInput{
		Kind:           ledger.TxnPayoutReversal,
		IdempotencyKey: "payout_reverse:" + payoutID.String(),
		Reference:      payoutID.String(),
		Currency:       s.currency,
		Postings: []ledger.Posting{
			{AccountKind: ledger.AccountPayoutClearing, OwnerID: &freelancerID, AmountCents: -amount},
			{AccountKind: ledger.AccountUserBalance, OwnerID: &freelancerID, AmountCents: amount},
		},
	})
	return err
}

// Commission returns the platform's cut of an amount, in cents.
func (s *Service) Commission(amount int64) int64 { return amount * s.commissionBps / 10000 }

func (s *Service) String() string {
	return fmt.Sprintf("wallet(commission=%d bps, currency=%s)", s.commissionBps, s.currency)
}
