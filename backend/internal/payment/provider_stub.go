package payment

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
)

// stubProvider is used when no lava.top API key is configured. It "creates" an
// invoice that is paid by calling POST /api/v1/dev/payments/:id/pay, which posts
// a webhook to the normal handler. Never enabled in production.
type stubProvider struct {
	webhookKey string
	publicURL  string
}

func NewStubProvider(webhookKey, publicBaseURL string) Provider {
	return &stubProvider{webhookKey: webhookKey, publicURL: publicBaseURL}
}

func (s *stubProvider) Name() string { return "stub" }

func (s *stubProvider) CreateInvoice(_ context.Context, p CreateInvoiceParams) (Invoice, error) {
	id := "stub_" + p.PaymentID.String()
	return Invoice{
		ProviderInvoiceID: id,
		PaymentURL:        fmt.Sprintf("%s/api/v1/dev/payments/%s/pay", s.publicURL, p.PaymentID),
	}, nil
}

type stubWebhook struct {
	InvoiceID string `json:"invoiceId"`
	Status    string `json:"status"`
}

func (s *stubProvider) ParseWebhook(headers map[string]string, body []byte) (WebhookEvent, error) {
	if subtle.ConstantTimeCompare([]byte(headers["x-webhook-key"]), []byte(s.webhookKey)) != 1 {
		return WebhookEvent{}, errors.New("bad webhook key")
	}
	var w stubWebhook
	if err := json.Unmarshal(body, &w); err != nil {
		return WebhookEvent{}, err
	}
	status := InvoiceStatus(w.Status)
	if status == "" {
		status = InvoicePaid
	}
	return WebhookEvent{ProviderInvoiceID: w.InvoiceID, Status: status}, nil
}
