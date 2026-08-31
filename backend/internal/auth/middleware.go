package auth

import (
	"strings"

	"ipw/internal/config"
	"ipw/internal/httpx"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	localsUserID = "auth.user_id"
	localsClaims = "auth.claims"
)

// Middleware provides route guards backed by access-token verification.
type Middleware struct {
	tokens *tokenIssuer
}

func NewMiddleware(cfg config.AuthConfig) *Middleware {
	return &Middleware{tokens: newTokenIssuer(cfg)}
}

// RequireAuth rejects the request unless it carries a valid access token, in the
// Authorization header ("Bearer <jwt>") or the "access_token" cookie.
func (m *Middleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := bearerToken(c)
		if raw == "" {
			return httpx.ErrUnauthorized("missing access token")
		}
		claims, err := m.tokens.parseAccess(raw)
		if err != nil {
			return httpx.ErrUnauthorized("invalid or expired access token")
		}
		uid, err := uuid.Parse(claims.Subject)
		if err != nil {
			return httpx.ErrUnauthorized("malformed token subject")
		}
		c.Locals(localsUserID, uid)
		c.Locals(localsClaims, claims)
		return c.Next()
	}
}

// OptionalAuth attaches the caller's identity when a valid token is present but
// never rejects the request. Use for endpoints whose response varies by viewer
// (e.g. a project draft visible only to its owner).
func (m *Middleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := bearerToken(c)
		if raw == "" {
			return c.Next()
		}
		claims, err := m.tokens.parseAccess(raw)
		if err != nil {
			return c.Next()
		}
		if uid, err := uuid.Parse(claims.Subject); err == nil {
			c.Locals(localsUserID, uid)
			c.Locals(localsClaims, claims)
		}
		return c.Next()
	}
}

// RequireCapability ensures the caller has a given capability. Must run after
// RequireAuth.
func (m *Middleware) RequireCapability(cap Capability) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals(localsClaims).(*Claims)
		if !ok {
			return httpx.ErrUnauthorized("authentication required")
		}
		if !claims.has(cap) {
			return httpx.ErrForbidden("missing required capability: " + string(cap))
		}
		return c.Next()
	}
}

// Capability is a coarse permission carried in the access token.
type Capability string

const (
	CapFreelancer Capability = "freelancer"
	CapClient     Capability = "client"
	CapAdmin      Capability = "admin"
)

func (cl *Claims) has(cap Capability) bool {
	switch cap {
	case CapFreelancer:
		return cl.IsFreelancer || cl.IsAdmin
	case CapClient:
		return cl.IsClient || cl.IsAdmin
	case CapAdmin:
		return cl.IsAdmin
	default:
		return false
	}
}

// UserID returns the authenticated user's ID. Only valid after RequireAuth.
func UserID(c *fiber.Ctx) uuid.UUID {
	id, _ := c.Locals(localsUserID).(uuid.UUID)
	return id
}

// IsAdmin reports whether the caller's token carries the admin capability. Only
// meaningful after RequireAuth or OptionalAuth.
func IsAdmin(c *fiber.Ctx) bool {
	claims, ok := c.Locals(localsClaims).(*Claims)
	return ok && claims.IsAdmin
}

func bearerToken(c *fiber.Ctx) string {
	if h := c.Get(fiber.HeaderAuthorization); h != "" {
		if after, ok := strings.CutPrefix(h, "Bearer "); ok {
			return strings.TrimSpace(after)
		}
	}
	return c.Cookies("access_token")
}
