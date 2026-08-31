package payment

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// lavaProvider talks to lava.top's public API (https://gate.lava.top).
//
// NOTE: lava.top's invoice endpoint is oriented around pre-registered "offers".
// The exact request/response shape must be confirmed against the account's
// integration settings and https://gate.lava.top/docs once real credentials
// exist. The mapping below is a best-effort starting point and is deliberately
// isolated behind the Provider interface so only this file changes.
type lavaProvider struct {
	apiKey     string
	webhookKey string
	baseURL    string
	offerID    string
	http       *http.Client
}

func NewLavaProvider(apiKey, webhookKey, baseURL, offerID string) Provider {
	return &lavaProvider{
		apiKey:     apiKey,
		webhookKey: webhookKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		offerID:    offerID,
		http:       &http.Client{Timeout: 15 * time.Second},
	}
}

func (l *lavaProvider) Name() string { return "lava" }

type lavaInvoiceRequest struct {
	Email       string  `json:"email"`
	OfferID     string  `json:"offerId,omitempty"`
	Currency    string  `json:"currency"`
	Sum         float64 `json:"sum"`
	Description string  `json:"description,omitempty"`
}

type lavaInvoiceResponse struct {
	ID         string `json:"id"`
	PaymentURL string `json:"paymentUrl"`
	Status     string `json:"status"`
}

func (l *lavaProvider) CreateInvoice(ctx context.Context, p CreateInvoiceParams) (Invoice, error) {
	body, _ := json.Marshal(lavaInvoiceRequest{
		Email:       p.PayerEmail,
		OfferID:     l.offerID,
		Currency:    p.Currency,
		Sum:         float64(p.AmountCents) / 100,
		Description: p.Description,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.baseURL+"/api/v2/invoice", bytes.NewReader(body))
	if err != nil {
		return Invoice{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", l.apiKey)

	resp, err := l.http.Do(req)
	if err != nil {
		return Invoice{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return Invoice{}, fmt.Errorf("lava invoice failed: %d %s", resp.StatusCode, string(raw))
	}
	var out lavaInvoiceResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return Invoice{}, fmt.Errorf("lava invoice: decode response: %w", err)
	}
	return Invoice{ProviderInvoiceID: out.ID, PaymentURL: out.PaymentURL}, nil
}

type lavaWebhook struct {
	InvoiceID string `json:"invoiceId"`
	OrderID   string `json:"orderId"`
	Status    string `json:"status"`
}

func (l *lavaProvider) ParseWebhook(headers map[string]string, body []byte) (WebhookEvent, error) {
	// lava.top signs webhooks with the integration's API key in a header.
	if subtle.ConstantTimeCompare([]byte(headers["x-api-key"]), []byte(l.webhookKey)) != 1 {
		return WebhookEvent{}, fmt.Errorf("invalid webhook signature")
	}
	var w lavaWebhook
	if err := json.Unmarshal(body, &w); err != nil {
		return WebhookEvent{}, err
	}
	id := w.InvoiceID
	if id == "" {
		id = w.OrderID
	}
	var status InvoiceStatus
	switch strings.ToLower(w.Status) {
	case "paid", "success", "completed":
		status = InvoicePaid
	case "expired":
		status = InvoiceExpired
	default:
		status = InvoiceFailed
	}
	return WebhookEvent{ProviderInvoiceID: id, Status: status}, nil
}
