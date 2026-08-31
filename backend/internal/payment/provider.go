// Package payment tracks external payment attempts (invoices) and abstracts the
// payment-acceptance provider. lava.top is the real provider; a stub provider
// backs local development.
//
// Payouts to freelancers are a SEPARATE concern (see the payout feature) —
// lava.top cannot send money to third parties.
package payment

import (
	"context"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoicePaid    InvoiceStatus = "paid"
	InvoiceFailed  InvoiceStatus = "failed"
	InvoiceExpired InvoiceStatus = "expired"
	InvoicePending InvoiceStatus = "pending"
)

type CreateInvoiceParams struct {
	PaymentID   uuid.UUID
	AmountCents int64
	Currency    string
	PayerEmail  string
	Description string
}

type Invoice struct {
	ProviderInvoiceID string
	PaymentURL        string
}

// WebhookEvent is the normalised result of parsing a provider callback.
type WebhookEvent struct {
	ProviderInvoiceID string
	Status            InvoiceStatus
}

// Provider is the payment-acceptance contract. Implementations: lavaProvider,
// stubProvider.
type Provider interface {
	Name() string
	CreateInvoice(ctx context.Context, p CreateInvoiceParams) (Invoice, error)
	ParseWebhook(headers map[string]string, body []byte) (WebhookEvent, error)
}
