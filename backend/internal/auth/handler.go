package auth

import (
	"time"

	"ipw/internal/config"
	"ipw/internal/httpx"
	"ipw/internal/user"

	"github.com/gofiber/fiber/v2"
)

// Handler exposes the auth HTTP API.
type Handler struct {
	svc *Service
	mw  *Middleware
	cfg config.AuthConfig
}

func NewHandler(svc *Service, mw *Middleware, cfg config.AuthConfig) *Handler {
	return &Handler{svc: svc, mw: mw, cfg: cfg}
}

// Register mounts the auth routes on the app.
func (h *Handler) Register(app *fiber.App) {
	g := app.Group("/api/v1/auth")
	g.Post("/register", h.register)
	g.Post("/login", h.login)
	g.Post("/refresh", h.refresh)
	g.Post("/logout", h.logout)
	g.Get("/me", h.mw.RequireAuth(), h.me)
}

type registerRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	DisplayName  string `json:"displayName"`
	AsFreelancer bool   `json:"asFreelancer"`
	AsClient     bool   `json:"asClient"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type userDTO struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"displayName"`
	AvatarURL     string    `json:"avatarUrl"`
	Country       string    `json:"country"`
	Timezone      string    `json:"timezone"`
	IsFreelancer  bool      `json:"isFreelancer"`
	IsClient      bool      `json:"isClient"`
	IsAdmin       bool      `json:"isAdmin"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
}

func toUserDTO(u *user.User) userDTO {
	return userDTO{
		ID:            u.ID.String(),
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		Country:       u.Country,
		Timezone:      u.Timezone,
		IsFreelancer:  u.IsFreelancer,
		IsClient:      u.IsClient,
		IsAdmin:       u.IsAdmin,
		EmailVerified: u.EmailVerifiedAt != nil,
		CreatedAt:     u.CreatedAt,
	}
}

type authResponse struct {
	User         userDTO `json:"user"`
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
	ExpiresIn    int     `json:"expiresIn"`
}

func (h *Handler) sessionContext(c *fiber.Ctx) sessionContext {
	return sessionContext{UserAgent: c.Get(fiber.HeaderUserAgent), IP: c.IP()}
}

// setRefreshCookie also mirrors the refresh token into an HttpOnly cookie so
// browser clients need not store it in JS-readable storage.
func (h *Handler) setRefreshCookie(c *fiber.Ctx, token string, ttl time.Duration) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(ttl),
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: "Lax",
		Domain:   h.cfg.CookieDomain,
	})
}

func (h *Handler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: "Lax",
		Domain:   h.cfg.CookieDomain,
	})
}

func (h *Handler) register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	u, pair, err := h.svc.Register(c.Context(), RegisterInput{
		Email:        req.Email,
		Password:     req.Password,
		DisplayName:  req.DisplayName,
		AsFreelancer: req.AsFreelancer,
		AsClient:     req.AsClient,
	}, h.sessionContext(c))
	if err != nil {
		return err
	}
	h.setRefreshCookie(c, pair.RefreshToken, h.cfg.RefreshTokenTTL)
	return httpx.Created(c, authResponse{toUserDTO(u), pair.AccessToken, pair.RefreshToken, pair.ExpiresIn})
}

func (h *Handler) login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.ErrBadRequest("invalid request body")
	}
	u, pair, err := h.svc.Login(c.Context(), req.Email, req.Password, h.sessionContext(c))
	if err != nil {
		return err
	}
	h.setRefreshCookie(c, pair.RefreshToken, h.cfg.RefreshTokenTTL)
	return httpx.OK(c, authResponse{toUserDTO(u), pair.AccessToken, pair.RefreshToken, pair.ExpiresIn})
}

func (h *Handler) refresh(c *fiber.Ctx) error {
	token := h.refreshTokenFromRequest(c)
	if token == "" {
		return httpx.ErrUnauthorized("missing refresh token")
	}
	u, pair, err := h.svc.Refresh(c.Context(), token, h.sessionContext(c))
	if err != nil {
		return err
	}
	h.setRefreshCookie(c, pair.RefreshToken, h.cfg.RefreshTokenTTL)
	return httpx.OK(c, authResponse{toUserDTO(u), pair.AccessToken, pair.RefreshToken, pair.ExpiresIn})
}

func (h *Handler) logout(c *fiber.Ctx) error {
	if token := h.refreshTokenFromRequest(c); token != "" {
		if err := h.svc.Logout(c.Context(), token); err != nil {
			return err
		}
	}
	h.clearRefreshCookie(c)
	return httpx.OK(c, fiber.Map{"status": "logged_out"})
}

func (h *Handler) me(c *fiber.Ctx) error {
	u, err := h.svc.UserByID(c.Context(), UserID(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, toUserDTO(u))
}

func (h *Handler) refreshTokenFromRequest(c *fiber.Ctx) string {
	var req refreshRequest
	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}
	return c.Cookies("refresh_token")
}
