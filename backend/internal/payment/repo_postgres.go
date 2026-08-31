package payment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type row struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	MilestoneID       uuid.UUID `gorm:"type:uuid"`
	ContractID        uuid.UUID `gorm:"type:uuid"`
	PayerID           uuid.UUID `gorm:"type:uuid"`
	AmountCents       int64
	Currency          string
	Provider          string
	ProviderInvoiceID string
	PaymentURL        string
	Status            string
	CreatedAt         time.Time
	PaidAt            *time.Time
}

func (row) TableName() string { return "payments" }

func (r row) toDomain() *Payment {
	return &Payment{
		ID: r.ID, MilestoneID: r.MilestoneID, ContractID: r.ContractID, PayerID: r.PayerID,
		AmountCents: r.AmountCents, Currency: r.Currency, Provider: r.Provider,
		ProviderInvoiceID: r.ProviderInvoiceID, PaymentURL: r.PaymentURL,
		Status: Status(r.Status), CreatedAt: r.CreatedAt, PaidAt: r.PaidAt,
	}
}

type PostgresStore struct{ db *gorm.DB }

func NewPostgresStore(db *gorm.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) Create(ctx context.Context, p *Payment) error {
	rec := row{
		ID: p.ID, MilestoneID: p.MilestoneID, ContractID: p.ContractID, PayerID: p.PayerID,
		AmountCents: p.AmountCents, Currency: p.Currency, Provider: p.Provider,
		Status: string(p.Status), CreatedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	*p = *rec.toDomain()
	return nil
}

func (s *PostgresStore) SetInvoice(ctx context.Context, id uuid.UUID, providerInvoiceID, url string) error {
	return s.db.WithContext(ctx).Model(&row{}).Where("id = ?", id).
		Updates(map[string]any{"provider_invoice_id": providerInvoiceID, "payment_url": url}).Error
}

func (s *PostgresStore) MarkStatus(ctx context.Context, id uuid.UUID, status Status, paidAt *time.Time) error {
	fields := map[string]any{"status": string(status)}
	if paidAt != nil {
		fields["paid_at"] = *paidAt
	}
	return s.db.WithContext(ctx).Model(&row{}).Where("id = ?", id).Updates(fields).Error
}

func (s *PostgresStore) Get(ctx context.Context, id uuid.UUID) (*Payment, error) {
	var rec row
	err := s.db.WithContext(ctx).First(&rec, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec.toDomain(), nil
}

func (s *PostgresStore) ByProviderInvoice(ctx context.Context, provider, invoiceID string) (*Payment, error) {
	var rec row
	err := s.db.WithContext(ctx).First(&rec, "provider = ? AND provider_invoice_id = ?", provider, invoiceID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec.toDomain(), nil
}

func (s *PostgresStore) LiveForMilestone(ctx context.Context, milestoneID uuid.UUID) (*Payment, error) {
	var rec row
	err := s.db.WithContext(ctx).
		Where("milestone_id = ? AND status IN ?", milestoneID, []string{string(StatusPending), string(StatusPaid)}).
		First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec.toDomain(), nil
}
