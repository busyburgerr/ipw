package contract

import (
	"context"
	"time"

	"ipw/internal/auth"
	"ipw/internal/httpx"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
	mw  *auth.Middleware
}

func NewHandler(svc *Service, mw *auth.Middleware) *Handler {
	return &Handler{svc: svc, mw: mw}
}

func (h *Handler) Register(app *fiber.App) {
	g := app.Group("/api/v1")

	authOnly := []fiber.Handler{h.mw.RequireAuth()}
	client := []fiber.Handler{h.mw.RequireAuth(), h.mw.RequireCapability(auth.CapClient)}
	freelancer := []fiber.Handler{h.mw.RequireAuth(), h.mw.RequireCapability(auth.CapFreelancer)}

	g.Post("/proposals/:id/accept", append(clone(client), h.acceptProposal)...)

	g.Get("/me/contracts", append(clone(authOnly), h.listMyContracts)...)
	g.Get("/contracts/:id", append(clone(authOnly), h.getContract)...)
	g.Post("/contracts/:id/milestones", append(clone(client), h.addMilestone)...)
	g.Post("/contracts/:id/complete", append(clone(client), h.completeContract)...)

	g.Post("/milestones/:id/fund", append(clone(client), h.fundMilestone)...)
	g.Post("/milestones/:id/approve", append(clone(client), h.approveMilestone)...)
	g.Post("/milestones/:id/request-changes", append(clone(client), h.requestChanges)...)
	g.Post("/milestones/:id/cancel", append(clone(client), h.cancelMilestone)...)
	g.Post("/milestones/:id/submit", append(clone(freelancer), h.submitMilestone)...)
}

func clone(s []fiber.Handler) []fiber.Handler {
	out := make([]fiber.Handler, len(s))
	copy(out, s)
	return out
}

// ---- DTOs -----------------------------------------------------------

type contractDTO struct {
	ID                string     `json:"id"`
	ProjectID         string     `json:"projectId"`
	ProposalID        string     `json:"proposalId"`
	ClientID          string     `json:"clientId"`
	FreelancerID      string     `json:"freelancerId"`
	Type              string     `json:"type"`
	AgreedAmountCents int64      `json:"agreedAmountCents"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	StartedAt         time.Time  `json:"startedAt"`
	EndedAt           *time.Time `json:"endedAt"`
}

type milestoneDTO struct {
	ID              string     `json:"id"`
	ContractID      string     `json:"contractId"`
	Sequence        int        `json:"sequence"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	AmountCents     int64      `json:"amountCents"`
	Status          string     `json:"status"`
	DeliverableNote string     `json:"deliverableNote"`
	DueDate         *time.Time `json:"dueDate"`
	FundedAt        *time.Time `json:"fundedAt"`
	SubmittedAt     *time.Time `json:"submittedAt"`
	ApprovedAt      *time.Time `json:"approvedAt"`
	ReleasedAt      *time.Time `json:"releasedAt"`
}

func toContractDTO(c *Contract) contractDTO {
	return contractDTO{
		ID: c.ID.String(), ProjectID: c.ProjectID.String(), ProposalID: c.ProposalID.String(),
		ClientID: c.ClientID.String(), FreelancerID: c.FreelancerID.String(), Type: string(c.Type),
		AgreedAmountCents: c.AgreedAmountCents, Currency: c.Currency, Status: string(c.Status),
		StartedAt: c.StartedAt, EndedAt: c.EndedAt,
	}
}

func toMilestoneDTO(m *Milestone) milestoneDTO {
	return milestoneDTO{
		ID: m.ID.String(), ContractID: m.ContractID.String(), Sequence: m.Sequence,
		Title: m.Title, Description: m.Description, AmountCents: m.AmountCents,
		Status: string(m.Status), DeliverableNote: m.DeliverableNote, DueDate: m.DueDate,
		FundedAt: m.FundedAt, SubmittedAt: m.SubmittedAt, ApprovedAt: m.ApprovedAt, ReleasedAt: m.ReleasedAt,
	}
}

func mapMilestoneDTOs(ms []Milestone) []milestoneDTO {
	out := make([]milestoneDTO, len(ms))
	for i := range ms {
		out[i] = toMilestoneDTO(&ms[i])
	}
	return out
}

// ---- handlers -----------------------------------------------------

func (h *Handler) acceptProposal(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid proposal id")
	}
	ct, err := h.svc.AcceptProposal(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.Created(c, toContractDTO(ct))
}

func (h *Handler) listMyContracts(c *fiber.Ctx) error {
	list, err := h.svc.ListMyContracts(c.Context(), auth.UserID(c))
	if err != nil {
		return err
	}
	out := make([]contractDTO, len(list))
	for i := range list {
		out[i] = toContractDTO(&list[i])
	}
	return httpx.OK(c, fiber.Map{"contracts": out})
}

func (h *Handler) getContract(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid contract id")
	}
	ct, ms, err := h.svc.GetContractForParty(c.Context(), id, auth.UserID(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"contract": toContractDTO(ct), "milestones": mapMilestoneDTOs(ms)})
}

type milestoneRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	AmountCents int64   `json:"amountCents"`
	DueDate     *string `json:"dueDate"` // YYYY-MM-DD
}

func (h *Handler) addMilestone(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid contract id")
	}
	var req milestoneRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	due, err := parseDate(req.DueDate)
	if err != nil {
		return httpx.ErrBadRequest("dueDate must be YYYY-MM-DD")
	}
	m, err := h.svc.AddMilestone(c.Context(), auth.UserID(c), id, MilestoneInput{
		Title: req.Title, Description: req.Description, AmountCents: req.AmountCents, DueDate: due,
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toMilestoneDTO(m))
}

func (h *Handler) completeContract(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid contract id")
	}
	ct, err := h.svc.CompleteContract(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, toContractDTO(ct))
}

func (h *Handler) fundMilestone(c *fiber.Ctx) error { return h.milestoneAction(c, h.svc.FundMilestone) }
func (h *Handler) approveMilestone(c *fiber.Ctx) error {
	return h.milestoneAction(c, h.svc.ApproveMilestone)
}
func (h *Handler) cancelMilestone(c *fiber.Ctx) error {
	return h.milestoneAction(c, h.svc.CancelMilestone)
}
func (h *Handler) requestChanges(c *fiber.Ctx) error {
	return h.milestoneAction(c, h.svc.RequestMilestoneChanges)
}

func (h *Handler) milestoneAction(c *fiber.Ctx, fn func(ctx context.Context, actor, milestoneID uuid.UUID) (*Milestone, error)) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid milestone id")
	}
	m, err := fn(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, toMilestoneDTO(m))
}

type submitRequest struct {
	DeliverableNote string `json:"deliverableNote"`
}

func (h *Handler) submitMilestone(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid milestone id")
	}
	var req submitRequest
	_ = c.BodyParser(&req)
	m, err := h.svc.SubmitMilestone(c.Context(), auth.UserID(c), id, req.DeliverableNote)
	if err != nil {
		return err
	}
	return httpx.OK(c, toMilestoneDTO(m))
}

func parseDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
