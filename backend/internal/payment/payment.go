package payment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
	StatusFailed  Status = "failed"
	StatusExpired Status = "expired"
)

type Payment struct {
	ID                uuid.UUID
	MilestoneID       uuid.UUID
	ContractID        uuid.UUID
	PayerID           uuid.UUID
	AmountCents       int64
	Currency          string
	Provider          string
	ProviderInvoiceID string
	PaymentURL        string
	Status            Status
	CreatedAt         time.Time
	PaidAt            *time.Time
}

var ErrNotFound = errors.New("payment not found")

type Store interface {
	Create(ctx context.Context, p *Payment) error
	SetInvoice(ctx context.Context, id uuid.UUID, providerInvoiceID, url string) error
	MarkStatus(ctx context.Context, id uuid.UUID, status Status, paidAt *time.Time) error
	Get(ctx context.Context, id uuid.UUID) (*Payment, error)
	ByProviderInvoice(ctx context.Context, provider, invoiceID string) (*Payment, error)
	LiveForMilestone(ctx context.Context, milestoneID uuid.UUID) (*Payment, error)
}
