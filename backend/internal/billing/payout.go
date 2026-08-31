package billing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayoutStatus string

const (
	PayoutRequested  PayoutStatus = "requested"
	PayoutProcessing PayoutStatus = "processing"
	PayoutPaid       PayoutStatus = "paid"
	PayoutRejected   PayoutStatus = "rejected"
)

type Payout struct {
	ID           uuid.UUID
	FreelancerID uuid.UUID
	AmountCents  int64
	Currency     string
	Method       string
	Destination  string
	Status       PayoutStatus
	Note         string
	CreatedAt    time.Time
	ProcessedAt  *time.Time
}

var errPayoutNotFound = errors.New("payout not found")

type payoutRow struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	FreelancerID uuid.UUID `gorm:"type:uuid"`
	AmountCents  int64
	Currency     string
	Method       string
	Destination  string
	Status       string
	Note         string
	CreatedAt    time.Time
	ProcessedAt  *time.Time
}

func (payoutRow) TableName() string { return "payouts" }

func (r payoutRow) toDomain() *Payout {
	return &Payout{
		ID: r.ID, FreelancerID: r.FreelancerID, AmountCents: r.AmountCents, Currency: r.Currency,
		Method: r.Method, Destination: r.Destination, Status: PayoutStatus(r.Status),
		Note: r.Note, CreatedAt: r.CreatedAt, ProcessedAt: r.ProcessedAt,
	}
}

type payoutStore struct{ db *gorm.DB }

func newPayoutStore(db *gorm.DB) *payoutStore { return &payoutStore{db: db} }

func (s *payoutStore) create(ctx context.Context, p *Payout) error {
	rec := payoutRow{
		ID: p.ID, FreelancerID: p.FreelancerID, AmountCents: p.AmountCents, Currency: p.Currency,
		Method: p.Method, Destination: p.Destination, Status: string(p.Status), CreatedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	*p = *rec.toDomain()
	return nil
}

func (s *payoutStore) get(ctx context.Context, id uuid.UUID) (*Payout, error) {
	var rec payoutRow
	err := s.db.WithContext(ctx).First(&rec, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errPayoutNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec.toDomain(), nil
}

func (s *payoutStore) listForFreelancer(ctx context.Context, freelancerID uuid.UUID) ([]Payout, error) {
	var rows []payoutRow
	err := s.db.WithContext(ctx).Where("freelancer_id = ?", freelancerID).
		Order("created_at DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Payout, len(rows))
	for i, r := range rows {
		out[i] = *r.toDomain()
	}
	return out, nil
}

func (s *payoutStore) setStatus(ctx context.Context, id uuid.UUID, status PayoutStatus, note string, processed bool) error {
	fields := map[string]any{"status": string(status)}
	if note != "" {
		fields["note"] = note
	}
	if processed {
		fields["processed_at"] = time.Now()
	}
	return s.db.WithContext(ctx).Model(&payoutRow{}).Where("id = ?", id).Updates(fields).Error
}
