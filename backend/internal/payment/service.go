package payment

import (
	"context"
	"time"

	"ipw/internal/httpx"

	"github.com/google/uuid"
)

type Service struct {
	store    Store
	provider Provider
}

func NewService(store Store, provider Provider) *Service {
	return &Service{store: store, provider: provider}
}

func (s *Service) ProviderName() string { return s.provider.Name() }

type NewInvoiceParams struct {
	MilestoneID uuid.UUID
	ContractID  uuid.UUID
	PayerID     uuid.UUID
	PayerEmail  string
	AmountCents int64
	Currency    string
	Description string
}

// CreateInvoice creates a payment row and asks the provider for a payment link.
func (s *Service) CreateInvoice(ctx context.Context, in NewInvoiceParams) (*Payment, error) {
	if existing, err := s.store.LiveForMilestone(ctx, in.MilestoneID); err == nil && existing != nil {
		return existing, nil // idempotent: reuse the open invoice
	}

	p := &Payment{
		ID:          uuid.New(),
		MilestoneID: in.MilestoneID,
		ContractID:  in.ContractID,
		PayerID:     in.PayerID,
		AmountCents: in.AmountCents,
		Currency:    in.Currency,
		Provider:    s.provider.Name(),
		Status:      StatusPending,
	}
	if err := s.store.Create(ctx, p); err != nil {
		return nil, err
	}

	inv, err := s.provider.CreateInvoice(ctx, CreateInvoiceParams{
		PaymentID:   p.ID,
		AmountCents: in.AmountCents,
		Currency:    in.Currency,
		PayerEmail:  in.PayerEmail,
		Description: in.Description,
	})
	if err != nil {
		_ = s.store.MarkStatus(ctx, p.ID, StatusFailed, nil)
		return nil, httpx.NewDomainError(502, "payment_provider_error", "could not create a payment: "+err.Error())
	}
	if err := s.store.SetInvoice(ctx, p.ID, inv.ProviderInvoiceID, inv.PaymentURL); err != nil {
		return nil, err
	}
	p.ProviderInvoiceID = inv.ProviderInvoiceID
	p.PaymentURL = inv.PaymentURL
	return p, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Payment, error) {
	return s.store.Get(ctx, id)
}

// HandleWebhook parses a provider callback and, when it reports a successful
// payment that is still pending on our side, marks it paid and returns it so the
// caller (billing) can fund escrow. A nil Payment with nil error means "nothing
// to do" (duplicate or non-terminal event).
func (s *Service) HandleWebhook(ctx context.Context, headers map[string]string, body []byte) (*Payment, error) {
	ev, err := s.provider.ParseWebhook(headers, body)
	if err != nil {
		return nil, httpx.ErrBadRequest("invalid webhook: " + err.Error())
	}
	p, err := s.store.ByProviderInvoice(ctx, s.provider.Name(), ev.ProviderInvoiceID)
	if err != nil {
		return nil, httpx.ErrNotFound("unknown invoice")
	}

	switch ev.Status {
	case InvoicePaid:
		if p.Status != StatusPending {
			return nil, nil // already handled
		}
		now := time.Now()
		if err := s.store.MarkStatus(ctx, p.ID, StatusPaid, &now); err != nil {
			return nil, err
		}
		p.Status = StatusPaid
		p.PaidAt = &now
		return p, nil
	case InvoiceExpired, InvoiceFailed:
		if p.Status == StatusPending {
			_ = s.store.MarkStatus(ctx, p.ID, Status(ev.Status), nil)
		}
		return nil, nil
	default:
		return nil, nil
	}
}

// MarkPaidByID is the stub-provider dev shortcut: it fabricates a webhook-style
// confirmation for a payment. Never wired in production.
func (s *Service) MarkPaidByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	p, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, httpx.ErrNotFound("payment not found")
	}
	if p.Status != StatusPending {
		return nil, nil
	}
	now := time.Now()
	if err := s.store.MarkStatus(ctx, p.ID, StatusPaid, &now); err != nil {
		return nil, err
	}
	p.Status = StatusPaid
	p.PaidAt = &now
	return p, nil
}
