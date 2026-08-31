package profile

import (
	"context"
	"errors"
	"strings"

	"ipw/internal/auth"
	"ipw/internal/catalog"
	"ipw/internal/httpx"
	"ipw/internal/platform/storage"
	"ipw/internal/user"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const maxPortfolioImageBytes = 5 << 20 // 5 MiB
const maxAvatarBytes = 3 << 20         // 3 MiB

type Handler struct {
	svc     *Service
	users   user.Store
	catalog catalog.Store
	files   *storage.Client
	mw      *auth.Middleware
}

func NewHandler(svc *Service, users user.Store, cat catalog.Store, files *storage.Client, mw *auth.Middleware) *Handler {
	return &Handler{svc: svc, users: users, catalog: cat, files: files, mw: mw}
}

func (h *Handler) Register(app *fiber.App) {
	// Public profile views.
	app.Get("/api/v1/freelancers", h.listFreelancers)
	pub := app.Group("/api/v1/profiles")
	pub.Get("/freelancers/:userId", h.publicFreelancer)
	pub.Get("/clients/:userId", h.publicClient)

	// "My profile" — authenticated self-service.
	me := app.Group("/api/v1/me", h.mw.RequireAuth())
	me.Post("/avatar", h.uploadAvatar)

	frl := me.Group("/freelancer-profile", h.mw.RequireCapability(auth.CapFreelancer))
	frl.Get("", h.getMyFreelancer)
	frl.Put("", h.saveMyFreelancer)
	frl.Put("/skills", h.setMySkills)
	frl.Get("/portfolio", h.listMyPortfolio)
	frl.Post("/portfolio", h.addPortfolioItem)
	frl.Delete("/portfolio/:itemId", h.deletePortfolioItem)

	cln := me.Group("/client-profile", h.mw.RequireCapability(auth.CapClient))
	cln.Get("", h.getMyClient)
	cln.Put("", h.saveMyClient)
}

// ---- DTOs ------------------------------------------------------------------

type freelancerDTO struct {
	UserID            string     `json:"userId"`
	DisplayName       string     `json:"displayName"`
	AvatarURL         string     `json:"avatarUrl"`
	Headline          string     `json:"headline"`
	Bio               string     `json:"bio"`
	HourlyRate        int64      `json:"hourlyRateCents"`
	Currency          string     `json:"currency"`
	Availability      string     `json:"availability"`
	PrimaryCategoryID *string    `json:"primaryCategoryId"`
	Languages         []string   `json:"languages"`
	Location          string     `json:"location"`
	Skills            []skillDTO `json:"skills"`
	RatingAvg         float64    `json:"ratingAvg"`
	RatingCount       int        `json:"ratingCount"`
	JobsCompleted     int        `json:"jobsCompleted"`
	TotalEarned       int64      `json:"totalEarnedCents"`
}

type skillDTO struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type clientDTO struct {
	UserID          string  `json:"userId"`
	DisplayName     string  `json:"displayName"`
	AvatarURL       string  `json:"avatarUrl"`
	CompanyName     string  `json:"companyName"`
	About           string  `json:"about"`
	Website         string  `json:"website"`
	Location        string  `json:"location"`
	PaymentVerified bool    `json:"paymentVerified"`
	RatingAvg       float64 `json:"ratingAvg"`
	RatingCount     int     `json:"ratingCount"`
	HiresCount      int     `json:"hiresCount"`
	TotalSpent      int64   `json:"totalSpentCents"`
}

type portfolioDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	ImageURL    string `json:"imageUrl"`
	Position    int    `json:"position"`
}

// ---- handlers ------------------------------------------------------------

func (h *Handler) listFreelancers(c *fiber.Ctx) error {
	list, err := h.svc.ListFreelancers(c.Context(), FreelancerFilter{
		Query:        strings.TrimSpace(c.Query("q")),
		CategorySlug: strings.TrimSpace(c.Query("category")),
		Limit:        c.QueryInt("limit", 30),
		Offset:       c.QueryInt("offset", 0),
	})
	if err != nil {
		return err
	}
	out := make([]freelancerDTO, len(list))
	for i := range list {
		dto, err := h.freelancerToDTO(c.Context(), &list[i])
		if err != nil {
			return err
		}
		out[i] = dto
	}
	return httpx.OK(c, fiber.Map{"freelancers": out})
}

func (h *Handler) publicFreelancer(c *fiber.Ctx) error {
	uid, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return httpx.ErrBadRequest("invalid user id")
	}
	f, err := h.svc.GetFreelancer(c.Context(), uid)
	if err != nil {
		return mapNotFound(err, "freelancer profile")
	}
	dto, err := h.freelancerToDTO(c.Context(), f)
	if err != nil {
		return err
	}
	return httpx.OK(c, dto)
}

func (h *Handler) publicClient(c *fiber.Ctx) error {
	uid, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return httpx.ErrBadRequest("invalid user id")
	}
	cl, err := h.svc.GetClient(c.Context(), uid)
	if err != nil {
		return mapNotFound(err, "client profile")
	}
	return httpx.OK(c, h.clientToDTO(c.Context(), cl))
}

func (h *Handler) getMyFreelancer(c *fiber.Ctx) error {
	f, err := h.svc.GetFreelancer(c.Context(), auth.UserID(c))
	if err != nil {
		return mapNotFound(err, "freelancer profile")
	}
	dto, err := h.freelancerToDTO(c.Context(), f)
	if err != nil {
		return err
	}
	return httpx.OK(c, dto)
}

type freelancerRequest struct {
	Headline          string   `json:"headline"`
	Bio               string   `json:"bio"`
	HourlyRateCents   int64    `json:"hourlyRateCents"`
	Currency          string   `json:"currency"`
	Availability      string   `json:"availability"`
	PrimaryCategoryID string   `json:"primaryCategoryId"`
	Languages         []string `json:"languages"`
	Location          string   `json:"location"`
}

func (h *Handler) saveMyFreelancer(c *fiber.Ctx) error {
	var req freelancerRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	catID, err := parseOptionalUUID(req.PrimaryCategoryID)
	if err != nil {
		return httpx.ErrBadRequest("invalid primaryCategoryId")
	}
	f, err := h.svc.SaveFreelancer(c.Context(), auth.UserID(c), FreelancerInput{
		Headline:          req.Headline,
		Bio:               req.Bio,
		HourlyRateCents:   req.HourlyRateCents,
		Currency:          req.Currency,
		Availability:      req.Availability,
		PrimaryCategoryID: catID,
		Languages:         req.Languages,
		Location:          req.Location,
	})
	if err != nil {
		return err
	}
	dto, err := h.freelancerToDTO(c.Context(), f)
	if err != nil {
		return err
	}
	return httpx.OK(c, dto)
}

type skillsRequest struct {
	SkillIDs []string `json:"skillIds"`
}

func (h *Handler) setMySkills(c *fiber.Ctx) error {
	var req skillsRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	ids := make([]uuid.UUID, 0, len(req.SkillIDs))
	for _, s := range req.SkillIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return httpx.ErrBadRequest("invalid skill id: " + s)
		}
		ids = append(ids, id)
	}
	f, err := h.svc.SetFreelancerSkills(c.Context(), auth.UserID(c), ids)
	if err != nil {
		return mapNotFound(err, "freelancer profile")
	}
	dto, err := h.freelancerToDTO(c.Context(), f)
	if err != nil {
		return err
	}
	return httpx.OK(c, dto)
}

func (h *Handler) getMyClient(c *fiber.Ctx) error {
	cl, err := h.svc.GetClient(c.Context(), auth.UserID(c))
	if err != nil {
		return mapNotFound(err, "client profile")
	}
	return httpx.OK(c, h.clientToDTO(c.Context(), cl))
}

type clientRequest struct {
	CompanyName string `json:"companyName"`
	About       string `json:"about"`
	Website     string `json:"website"`
	Location    string `json:"location"`
}

func (h *Handler) saveMyClient(c *fiber.Ctx) error {
	var req clientRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	cl, err := h.svc.SaveClient(c.Context(), auth.UserID(c), ClientInput{
		CompanyName: req.CompanyName,
		About:       req.About,
		Website:     req.Website,
		Location:    req.Location,
	})
	if err != nil {
		return err
	}
	return httpx.OK(c, h.clientToDTO(c.Context(), cl))
}

func (h *Handler) uploadAvatar(c *fiber.Ctx) error {
	img, err := httpx.ReadImage(c, "file", maxAvatarBytes)
	if err != nil {
		return err
	}
	key, err := h.files.Put(c.Context(), "avatars", img.Data, img.ContentType, img.Ext)
	if err != nil {
		return err
	}
	url := h.files.PublicURL(key)
	if err := h.users.SetAvatarURL(c.Context(), auth.UserID(c), url); err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"avatarUrl": url})
}

func (h *Handler) listMyPortfolio(c *fiber.Ctx) error {
	items, err := h.svc.ListPortfolio(c.Context(), auth.UserID(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.Map{"items": h.portfolioToDTO(items)})
}

func (h *Handler) addPortfolioItem(c *fiber.Ctx) error {
	uid := auth.UserID(c)
	in := PortfolioInput{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		URL:         c.FormValue("url"),
	}
	// Image is optional.
	if _, ferr := c.FormFile("file"); ferr == nil {
		img, err := httpx.ReadImage(c, "file", maxPortfolioImageBytes)
		if err != nil {
			return err
		}
		key, err := h.files.Put(c.Context(), "portfolio", img.Data, img.ContentType, img.Ext)
		if err != nil {
			return err
		}
		in.ImageKey = key
	}
	item, err := h.svc.AddPortfolioItem(c.Context(), uid, in)
	if err != nil {
		return err
	}
	return httpx.Created(c, h.portfolioItemToDTO(*item))
}

func (h *Handler) deletePortfolioItem(c *fiber.Ctx) error {
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return httpx.ErrBadRequest("invalid item id")
	}
	if err := h.svc.DeletePortfolioItem(c.Context(), auth.UserID(c), itemID); err != nil {
		return mapNotFound(err, "portfolio item")
	}
	return httpx.OK(c, fiber.Map{"status": "deleted"})
}

// ---- mapping helpers ----------------------------------------------------

func (h *Handler) freelancerToDTO(ctx context.Context, f *Freelancer) (freelancerDTO, error) {
	skills, err := h.catalog.SkillsByIDs(ctx, f.SkillIDs)
	if err != nil {
		return freelancerDTO{}, err
	}
	sd := make([]skillDTO, len(skills))
	for i, s := range skills {
		sd[i] = skillDTO{ID: s.ID.String(), Slug: s.Slug, Name: s.Name}
	}

	var name, avatar string
	if u, err := h.users.GetByID(ctx, f.UserID); err == nil {
		name, avatar = u.DisplayName, u.AvatarURL
	}

	return freelancerDTO{
		UserID:            f.UserID.String(),
		DisplayName:       name,
		AvatarURL:         avatar,
		Headline:          f.Headline,
		Bio:               f.Bio,
		HourlyRate:        f.HourlyRateCents,
		Currency:          f.Currency,
		Availability:      string(f.Availability),
		PrimaryCategoryID: uuidPtrString(f.PrimaryCategoryID),
		Languages:         f.Languages,
		Location:          f.Location,
		Skills:            sd,
		RatingAvg:         f.RatingAvg,
		RatingCount:       f.RatingCount,
		JobsCompleted:     f.JobsCompleted,
		TotalEarned:       f.TotalEarnedCents,
	}, nil
}

func (h *Handler) clientToDTO(ctx context.Context, cl *Client) clientDTO {
	var name, avatar string
	if u, err := h.users.GetByID(ctx, cl.UserID); err == nil {
		name, avatar = u.DisplayName, u.AvatarURL
	}
	return clientDTO{
		UserID:          cl.UserID.String(),
		DisplayName:     name,
		AvatarURL:       avatar,
		CompanyName:     cl.CompanyName,
		About:           cl.About,
		Website:         cl.Website,
		Location:        cl.Location,
		PaymentVerified: cl.PaymentVerified,
		RatingAvg:       cl.RatingAvg,
		RatingCount:     cl.RatingCount,
		HiresCount:      cl.HiresCount,
		TotalSpent:      cl.TotalSpentCents,
	}
}

func (h *Handler) portfolioToDTO(items []PortfolioItem) []portfolioDTO {
	out := make([]portfolioDTO, len(items))
	for i, it := range items {
		out[i] = h.portfolioItemToDTO(it)
	}
	return out
}

func (h *Handler) portfolioItemToDTO(it PortfolioItem) portfolioDTO {
	return portfolioDTO{
		ID:          it.ID.String(),
		Title:       it.Title,
		Description: it.Description,
		URL:         it.URL,
		ImageURL:    h.files.PublicURL(it.ImageKey),
		Position:    it.Position,
	}
}

func mapNotFound(err error, what string) error {
	if errors.Is(err, ErrNotFound) {
		return httpx.ErrNotFound(what + " not found")
	}
	return err
}

func parseOptionalUUID(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func uuidPtrString(p *uuid.UUID) *string {
	if p == nil {
		return nil
	}
	s := p.String()
	return &s
}
