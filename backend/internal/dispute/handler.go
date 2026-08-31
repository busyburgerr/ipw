package dispute

import (
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

func NewHandler(svc *Service, mw *auth.Middleware) *Handler { return &Handler{svc: svc, mw: mw} }

func (h *Handler) Register(app *fiber.App) {
	g := app.Group("/api/v1")

	g.Post("/contracts/:id/disputes", h.mw.RequireAuth(), h.raise)
	g.Get("/me/disputes", h.mw.RequireAuth(), h.listMine)
	g.Get("/disputes/:id", h.mw.RequireAuth(), h.get)
	g.Post("/disputes/:id/withdraw", h.mw.RequireAuth(), h.withdraw)

	admin := []fiber.Handler{h.mw.RequireAuth(), h.mw.RequireCapability(auth.CapAdmin)}
	g.Get("/admin/disputes", append(cloneHandlers(admin), h.listForArbiter)...)
	g.Post("/admin/disputes/:id/claim", append(cloneHandlers(admin), h.claim)...)
	g.Post("/admin/disputes/:id/resolve", append(cloneHandlers(admin), h.resolve)...)
}

func cloneHandlers(s []fiber.Handler) []fiber.Handler {
	out := make([]fiber.Handler, len(s))
	copy(out, s)
	return out
}

type disputeDTO struct {
	ID             string     `json:"id"`
	ContractID     string     `json:"contractId"`
	MilestoneID    *string    `json:"milestoneId"`
	RaisedBy       string     `json:"raisedBy"`
	Reason         string     `json:"reason"`
	Status         string     `json:"status"`
	ResolutionNote string     `json:"resolutionNote"`
	CreatedAt      time.Time  `json:"createdAt"`
	ResolvedAt     *time.Time `json:"resolvedAt"`
}

func toDTO(d *Dispute) disputeDTO {
	var ms *string
	if d.MilestoneID != nil {
		s := d.MilestoneID.String()
		ms = &s
	}
	return disputeDTO{
		ID: d.ID.String(), ContractID: d.ContractID.String(), MilestoneID: ms,
		RaisedBy: d.RaisedBy.String(), Reason: d.Reason, Status: string(d.Status),
		ResolutionNote: d.ResolutionNote, CreatedAt: d.CreatedAt, ResolvedAt: d.ResolvedAt,
	}
}

func mapDTOs(ds []Dispute) []disputeDTO {
	out := make([]disputeDTO, len(ds))
	for i := range ds {
		out[i] = toDTO(&ds[i])
	}
	return out
}

type raiseRequest struct {
	MilestoneID string `json:"milestoneId"`
	Reason      string `json:"reason"`
}

func (h *Handler) raise(c *fiber.Ctx) error {
	cid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid contract id")
	}
	var req raiseRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	var msID *uuid.UUID
	if req.MilestoneID != "" {
		id, err := uuid.Parse(req.MilestoneID)
		if err != nil {
			return httpx.ErrBadRequest("invalid milestoneId")
		}
		msID = &id
	}
	d, err := h.svc.Raise(c.Context(), auth.UserID(c), cid, RaiseInput{MilestoneID: msID, Reason: req.Reason})
	if err != nil {
		return err
	}
	return httpx.Created(c, toDTO(d))
}

func (h *Handler) listMine(c *fiber.Ctx) error {
	ds, err := h.svc.ListMine(c.Context(), auth.UserID(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"disputes": mapDTOs(ds)})
}

func (h *Handler) get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid dispute id")
	}
	d, err := h.svc.Get(c.Context(), auth.UserID(c), id, auth.IsAdmin(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, toDTO(d))
}

func (h *Handler) withdraw(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid dispute id")
	}
	d, err := h.svc.Withdraw(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, toDTO(d))
}

func (h *Handler) listForArbiter(c *fiber.Ctx) error {
	ds, err := h.svc.ListForArbiter(c.Context())
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"disputes": mapDTOs(ds)})
}

func (h *Handler) claim(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid dispute id")
	}
	d, err := h.svc.Claim(c.Context(), auth.UserID(c), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, toDTO(d))
}

type resolveRequest struct {
	Outcome string `json:"outcome"` // "client" | "freelancer"
	Note    string `json:"note"`
}

func (h *Handler) resolve(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid dispute id")
	}
	var req resolveRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	d, err := h.svc.Resolve(c.Context(), auth.UserID(c), id, Resolution(req.Outcome), req.Note)
	if err != nil {
		return err
	}
	return httpx.OK(c, toDTO(d))
}
