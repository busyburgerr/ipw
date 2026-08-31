package review

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
	g.Post("/contracts/:id/reviews", h.mw.RequireAuth(), h.submit)
	g.Get("/contracts/:id/reviews", h.mw.RequireAuth(), h.listForContract)
	g.Get("/users/:id/reviews", h.listForUser)
}

type reviewDTO struct {
	ID          string     `json:"id"`
	ContractID  string     `json:"contractId"`
	ReviewerID  string     `json:"reviewerId"`
	RevieweeID  string     `json:"revieweeId"`
	Direction   string     `json:"direction"`
	Rating      int        `json:"rating"`
	Comment     string     `json:"comment"`
	Published   bool       `json:"published"`
	PublishedAt *time.Time `json:"publishedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func toDTO(r *Review) reviewDTO {
	return reviewDTO{
		ID: r.ID.String(), ContractID: r.ContractID.String(), ReviewerID: r.ReviewerID.String(),
		RevieweeID: r.RevieweeID.String(), Direction: string(r.Direction), Rating: r.Rating,
		Comment: r.Comment, Published: r.PublishedAt != nil, PublishedAt: r.PublishedAt, CreatedAt: r.CreatedAt,
	}
}

func mapDTOs(rs []Review) []reviewDTO {
	out := make([]reviewDTO, len(rs))
	for i := range rs {
		out[i] = toDTO(&rs[i])
	}
	return out
}

type submitRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

func (h *Handler) submit(c *fiber.Ctx) error {
	cid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid contract id")
	}
	var req submitRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	r, err := h.svc.Submit(c.Context(), auth.UserID(c), cid, Input{Rating: req.Rating, Comment: req.Comment})
	if err != nil {
		return err
	}
	return httpx.Created(c, toDTO(r))
}

func (h *Handler) listForContract(c *fiber.Ctx) error {
	cid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid contract id")
	}
	rs, err := h.svc.VisibleForContract(c.Context(), cid, auth.UserID(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"reviews": mapDTOs(rs)})
}

func (h *Handler) listForUser(c *fiber.Ctx) error {
	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.ErrBadRequest("invalid user id")
	}
	rs, err := h.svc.PublishedForUser(c.Context(), uid)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"reviews": mapDTOs(rs)})
}
