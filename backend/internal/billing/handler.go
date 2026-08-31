package billing

import (
	"time"

	"ipw/internal/auth"
	"ipw/internal/httpx"
	"ipw/internal/payment"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svc      *Service
	payments *payment.Service
	mw       *auth.Middleware
	devMode  bool
}

func NewHandler(svc *Service, payments *payment.Service, mw *auth.Middleware, devMode bool) *Handler {
	return &Handler{svc: svc, payments: payments, mw: mw, devMode: devMode}
}

func (h *Handler) Register(app *fiber.App) {
	g := app.Group("/api/v1")

	client := func() []fiber.Handler {
		return []fiber.Handler{h.mw.RequireAuth(), h.mw.RequireCapability(auth.CapClient)}
	}
	freelancer := func() []fiber.Handler {
		return []fiber.Handler{h.mw.RequireAuth(), h.mw.RequireCapability(auth.CapFreelancer)}
	}
	admin := func() []fiber.Handler {
		return []fiber.Handler{h.mw.RequireAuth(), h.mw.RequireCapability(auth.CapAdmin)}
	}

	g.Post("/milestones/:id/fund", append(client(), h.fundMilestone)...)
	g.Post("/milestones/:id/approve", append(client(), h.approveMilestone)...)
	g.Post("/milestones/:id/refund", append(client(), h.refundMilestone)...)

	g.Get("/me/wallet", h.mw.RequireAuth(), h.walletSummary)
	g.Post("/me/payouts", append(freelancer(), h.requestPayout)...)
	g.Get("/me/payouts", append(freelancer(), h.listMyPayouts)...)
	g.Post("/admin/payouts/:id/process", append(admin(), h.processPayout)...)

	// Provider webhook (no auth; verified by signature inside the provider).
	g.Post("/payments/webhook", h.webhook)

	if h.devMode {
		g.Post("/dev/payments/:id/pay", h.devPay)
	}
}

// ---- DTOs -----------------------------------------------------------

type paymentDTO struct {
	ID          string `json:"id"`
	MilestoneID string `json:"milestoneId"`
	AmountCents int64  `json:"amountCents"`
	Currency    string `json:"currency"`
	Provider    string `json:"provider"`
	Status      string `json:"status"`
	PaymentURL  string `json:"paymentUrl"`
}

func toPaymentDTO(p *payment.Payment) paymentDTO {
	return paymentDTO{
		ID: p.ID.String(), MilestoneID: p.MilestoneID.String(), AmountCents: p.AmountCents,
		Currency: p.Currency, Provider: p.Provider, Status: string(p.Status), PaymentURL: p.PaymentURL,
	}
}

type payoutDTO struct {
	ID          string     `json:"id"`
	AmountCents int64      `json:"amountCents"`
	Currency    string     `json:"currency"`
	Method      string     `json:"method"`
	Destination string     `json:"destination"`
	Status      string     `json:"status"`
	Note        string     `json:"note"`
	CreatedAt   time.Time  `json:"createdAt"`
	ProcessedAt *time.Time `json:"processedAt"`
}

func toPayoutDTO(p *Payout) payoutDTO {
	return payoutDTO{
		ID: p.ID.String(), AmountCents: p.AmountCents, Currency: p.Currency, Method: p.Method,
		Destination: p.Destination, Status: string(p.Status), Note: p.Note,
		CreatedAt: p.CreatedAt, ProcessedAt: p.ProcessedAt,
	}
}

// ---- handlers -----------------------------------------------------

func (h *Handler) fundMilestone(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid milestone id")
	}
	p, err := h.svc.FundMilestone(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.Created(c, toPaymentDTO(p))
}

func (h *Handler) approveMilestone(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid milestone id")
	}
	m, err := h.svc.ApproveMilestone(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{
		"id": m.ID.String(), "status": string(m.Status), "releasedAt": m.ReleasedAt,
	})
}

func (h *Handler) refundMilestone(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid milestone id")
	}
	m, err := h.svc.RefundMilestone(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"id": m.ID.String(), "status": string(m.Status)})
}

func (h *Handler) walletSummary(c *fiber.Ctx) error {
	sum, err := h.svc.WalletSummary(c.Context(), auth.UserID(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{
		"availableCents":     sum.AvailableCents,
		"pendingPayoutCents": sum.PendingPayoutCents,
		"currency":           sum.Currency,
	})
}

type payoutRequest struct {
	AmountCents int64  `json:"amountCents"`
	Method      string `json:"method"`
	Destination string `json:"destination"`
}

func (h *Handler) requestPayout(c *fiber.Ctx) error {
	var req payoutRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	p, err := h.svc.RequestPayout(c.Context(), auth.UserID(c), PayoutRequest{
		AmountCents: req.AmountCents, Method: req.Method, Destination: req.Destination,
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toPayoutDTO(p))
}

func (h *Handler) listMyPayouts(c *fiber.Ctx) error {
	list, err := h.svc.ListMyPayouts(c.Context(), auth.UserID(c))
	if err != nil {
		return err
	}
	out := make([]payoutDTO, len(list))
	for i := range list {
		out[i] = toPayoutDTO(&list[i])
	}
	return httpx.OK(c, fiber.Map{"payouts": out})
}

type processPayoutRequest struct {
	Decision string `json:"decision"` // "paid" | "rejected"
	Note     string `json:"note"`
}

func (h *Handler) processPayout(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid payout id")
	}
	var req processPayoutRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	p, err := h.svc.ProcessPayout(c.Context(), id, PayoutStatus(req.Decision), req.Note)
	if err != nil {
		return err
	}
	return httpx.OK(c, toPayoutDTO(p))
}

func (h *Handler) webhook(c *fiber.Ctx) error {
	headers := map[string]string{}
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[lower(string(k))] = string(v)
	})
	paid, err := h.payments.HandleWebhook(c.Context(), headers, c.Body())
	if err != nil {
		return err
	}
	if paid != nil {
		if err := h.svc.ConfirmPaidPayment(c.Context(), paid); err != nil {
			return err
		}
	}
	return httpx.OK(c, fiber.Map{"received": true})
}

func (h *Handler) devPay(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid payment id")
	}
	paid, err := h.payments.MarkPaidByID(c.Context(), id)
	if err != nil {
		return err
	}
	if paid != nil {
		if err := h.svc.ConfirmPaidPayment(c.Context(), paid); err != nil {
			return err
		}
	}
	return httpx.OK(c, fiber.Map{"status": "paid"})
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
