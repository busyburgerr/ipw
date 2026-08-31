package project

import (
	"strconv"
	"strings"
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

	// Empty-prefix sub-groups with middleware leak that middleware to every
	// sibling route, so guards are attached per route instead.
	asClient := []fiber.Handler{h.mw.RequireAuth(), h.mw.RequireCapability(auth.CapClient)}
	asFreelancer := []fiber.Handler{h.mw.RequireAuth(), h.mw.RequireCapability(auth.CapFreelancer)}

	g.Get("/projects", h.list)
	g.Get("/projects/:id", h.mw.OptionalAuth(), h.get)

	g.Post("/projects", append(asClient, h.create)...)
	g.Get("/me/projects", append(asClient, h.listMine)...)
	g.Put("/projects/:id", append(asClient, h.update)...)
	g.Post("/projects/:id/publish", append(asClient, h.publish)...)
	g.Post("/projects/:id/close", append(asClient, h.close)...)
	g.Get("/projects/:id/proposals", append(asClient, h.listProjectProposals)...)
	g.Post("/proposals/:id/shortlist", append(asClient, h.shortlist)...)
	g.Post("/proposals/:id/decline", append(asClient, h.decline)...)

	g.Post("/projects/:id/proposals", append(asFreelancer, h.submitProposal)...)
	g.Get("/me/proposals", append(asFreelancer, h.listMyProposals)...)
	g.Post("/proposals/:id/withdraw", append(asFreelancer, h.withdrawProposal)...)
}

// ---- DTOs -------------------------------------------------------------

type projectDTO struct {
	ID                 string     `json:"id"`
	ClientID           string     `json:"clientId"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	CategoryID         *string    `json:"categoryId"`
	BudgetType         string     `json:"budgetType"`
	FixedAmountCents   *int64     `json:"fixedAmountCents"`
	HourlyRateMinCents *int64     `json:"hourlyRateMinCents"`
	HourlyRateMaxCents *int64     `json:"hourlyRateMaxCents"`
	Currency           string     `json:"currency"`
	ExperienceLevel    string     `json:"experienceLevel"`
	Status             string     `json:"status"`
	SkillIDs           []string   `json:"skillIds"`
	ProposalsCount     int        `json:"proposalsCount"`
	PublishedAt        *time.Time `json:"publishedAt"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type proposalDTO struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"projectId"`
	FreelancerID   string    `json:"freelancerId"`
	CoverLetter    string    `json:"coverLetter"`
	BidAmountCents int64     `json:"bidAmountCents"`
	EstimatedDays  *int      `json:"estimatedDays"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

func toProjectDTO(p *Project) projectDTO {
	skills := make([]string, len(p.SkillIDs))
	for i, s := range p.SkillIDs {
		skills[i] = s.String()
	}
	return projectDTO{
		ID: p.ID.String(), ClientID: p.ClientID.String(), Title: p.Title,
		Description: p.Description, CategoryID: uuidPtrString(p.CategoryID),
		BudgetType: string(p.BudgetType), FixedAmountCents: p.FixedAmountCents,
		HourlyRateMinCents: p.HourlyRateMinCents, HourlyRateMaxCents: p.HourlyRateMaxCents,
		Currency: p.Currency, ExperienceLevel: string(p.ExperienceLevel),
		Status: string(p.Status), SkillIDs: skills, ProposalsCount: p.ProposalsCount,
		PublishedAt: p.PublishedAt, CreatedAt: p.CreatedAt,
	}
}

func toProposalDTO(p *Proposal) proposalDTO {
	return proposalDTO{
		ID: p.ID.String(), ProjectID: p.ProjectID.String(), FreelancerID: p.FreelancerID.String(),
		CoverLetter: p.CoverLetter, BidAmountCents: p.BidAmountCents,
		EstimatedDays: p.EstimatedDays, Status: string(p.Status), CreatedAt: p.CreatedAt,
	}
}

// ---- request bodies -------------------------------------------------

type projectRequest struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	CategoryID         string   `json:"categoryId"`
	BudgetType         string   `json:"budgetType"`
	FixedAmountCents   *int64   `json:"fixedAmountCents"`
	HourlyRateMinCents *int64   `json:"hourlyRateMinCents"`
	HourlyRateMaxCents *int64   `json:"hourlyRateMaxCents"`
	Currency           string   `json:"currency"`
	ExperienceLevel    string   `json:"experienceLevel"`
	SkillIDs           []string `json:"skillIds"`
}

func (r projectRequest) toInput() (ProjectInput, error) {
	catID, err := parseOptionalUUID(r.CategoryID)
	if err != nil {
		return ProjectInput{}, httpx.ErrBadRequest("invalid categoryId")
	}
	skillIDs, err := parseUUIDs(r.SkillIDs)
	if err != nil {
		return ProjectInput{}, err
	}
	return ProjectInput{
		Title: r.Title, Description: r.Description, CategoryID: catID,
		BudgetType: r.BudgetType, FixedAmountCents: r.FixedAmountCents,
		HourlyRateMinCents: r.HourlyRateMinCents, HourlyRateMaxCents: r.HourlyRateMaxCents,
		Currency: r.Currency, ExperienceLevel: r.ExperienceLevel, SkillIDs: skillIDs,
	}, nil
}

type proposalRequest struct {
	CoverLetter    string `json:"coverLetter"`
	BidAmountCents int64  `json:"bidAmountCents"`
	EstimatedDays  *int   `json:"estimatedDays"`
}

// ---- handlers -----------------------------------------------------

func (h *Handler) list(c *fiber.Ctx) error {
	skillIDs, err := parseUUIDs(splitCSV(c.Query("skills")))
	if err != nil {
		return err
	}
	f := Filter{
		Query:        strings.TrimSpace(c.Query("q")),
		CategorySlug: strings.TrimSpace(c.Query("category")),
		SkillIDs:     skillIDs,
		BudgetType:   BudgetType(c.Query("budgetType")),
		Limit:        atoiDefault(c.Query("limit"), 20),
		Offset:       atoiDefault(c.Query("offset"), 0),
	}
	items, total, err := h.svc.ListPublic(c.Context(), f)
	if err != nil {
		return err
	}
	out := make([]projectDTO, len(items))
	for i := range items {
		out[i] = toProjectDTO(&items[i])
	}
	return httpx.OK(c, fiber.Map{"projects": out, "total": total})
}

func (h *Handler) get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid project id")
	}
	p, err := h.svc.GetProjectForViewer(c.Context(), id, auth.UserID(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, toProjectDTO(p))
}

func (h *Handler) create(c *fiber.Ctx) error {
	var req projectRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	in, err := req.toInput()
	if err != nil {
		return err
	}
	p, err := h.svc.CreateProject(c.Context(), auth.UserID(c), in)
	if err != nil {
		return err
	}
	return httpx.Created(c, toProjectDTO(p))
}

func (h *Handler) update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid project id")
	}
	var req projectRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	in, err := req.toInput()
	if err != nil {
		return err
	}
	p, err := h.svc.UpdateProject(c.Context(), auth.UserID(c), id, in)
	if err != nil {
		return err
	}
	return httpx.OK(c, toProjectDTO(p))
}

func (h *Handler) publish(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid project id")
	}
	p, err := h.svc.PublishProject(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, toProjectDTO(p))
}

func (h *Handler) close(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid project id")
	}
	p, err := h.svc.CloseProject(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, toProjectDTO(p))
}

func (h *Handler) listMine(c *fiber.Ctx) error {
	items, err := h.svc.ListMine(c.Context(), auth.UserID(c))
	if err != nil {
		return err
	}
	out := make([]projectDTO, len(items))
	for i := range items {
		out[i] = toProjectDTO(&items[i])
	}
	return httpx.OK(c, fiber.Map{"projects": out})
}

func (h *Handler) submitProposal(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid project id")
	}
	var req proposalRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	p, err := h.svc.SubmitProposal(c.Context(), auth.UserID(c), projectID, ProposalInput{
		CoverLetter:    req.CoverLetter,
		BidAmountCents: req.BidAmountCents,
		EstimatedDays:  req.EstimatedDays,
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toProposalDTO(p))
}

func (h *Handler) listProjectProposals(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid project id")
	}
	items, err := h.svc.ListProjectProposals(c.Context(), auth.UserID(c), projectID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"proposals": mapProposalDTOs(items)})
}

func (h *Handler) listMyProposals(c *fiber.Ctx) error {
	items, err := h.svc.ListMyProposals(c.Context(), auth.UserID(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"proposals": mapProposalDTOs(items)})
}

func (h *Handler) withdrawProposal(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid proposal id")
	}
	if err := h.svc.WithdrawProposal(c.Context(), auth.UserID(c), id); err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"status": "withdrawn"})
}

func (h *Handler) shortlist(c *fiber.Ctx) error {
	return h.proposalDecision(c, ProposalShortlisted)
}

func (h *Handler) decline(c *fiber.Ctx) error {
	return h.proposalDecision(c, ProposalDeclined)
}

func (h *Handler) proposalDecision(c *fiber.Ctx, decision ProposalStatus) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid proposal id")
	}
	p, err := h.svc.SetProposalDecision(c.Context(), auth.UserID(c), id, decision)
	if err != nil {
		return err
	}
	return httpx.OK(c, toProposalDTO(p))
}

// ---- helpers -----------------------------------------------------

func mapProposalDTOs(items []Proposal) []proposalDTO {
	out := make([]proposalDTO, len(items))
	for i := range items {
		out[i] = toProposalDTO(&items[i])
	}
	return out
}

func uuidPtrString(p *uuid.UUID) *string {
	if p == nil {
		return nil
	}
	s := p.String()
	return &s
}

func parseOptionalUUID(s string) (*uuid.UUID, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseUUIDs(in []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, httpx.ErrBadRequest("invalid id: " + s)
		}
		out = append(out, id)
	}
	return out, nil
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
