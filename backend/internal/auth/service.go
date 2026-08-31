package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"ipw/internal/config"
	"ipw/internal/httpx"
	"ipw/internal/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const bcryptCost = 12

// Service implements registration and session lifecycle.
type Service struct {
	users    user.Store
	sessions *sessionStore
	tokens   *tokenIssuer
	cfg      config.AuthConfig
}

func NewService(users user.Store, db *gorm.DB, cfg config.AuthConfig) *Service {
	return &Service{
		users:    users,
		sessions: newSessionStore(db),
		tokens:   newTokenIssuer(cfg),
		cfg:      cfg,
	}
}

// TokenPair is what the client receives on register / login / refresh.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // access-token lifetime, seconds
}

type RegisterInput struct {
	Email        string
	Password     string
	DisplayName  string
	AsFreelancer bool
	AsClient     bool
}

type sessionContext struct {
	UserAgent string
	IP        string
}

func (s *Service) Register(ctx context.Context, in RegisterInput, sc sessionContext) (*user.User, *TokenPair, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if !looksLikeEmail(email) {
		return nil, nil, httpx.ErrBadRequest("invalid email")
	}
	if len(in.Password) < 8 {
		return nil, nil, httpx.ErrBadRequest("password must be at least 8 characters")
	}
	if !in.AsFreelancer && !in.AsClient {
		return nil, nil, httpx.ErrBadRequest("choose at least one role: freelancer or client")
	}

	exists, err := s.users.EmailExists(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, httpx.ErrConflict("an account with this email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, nil, err
	}

	u := &user.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  strings.TrimSpace(in.DisplayName),
		Timezone:     "UTC",
		IsFreelancer: in.AsFreelancer,
		IsClient:     in.AsClient,
		Status:       user.StatusActive,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, nil, err
	}

	pair, err := s.startSession(ctx, u, sc)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

func (s *Service) Login(ctx context.Context, email, password string, sc sessionContext) (*user.User, *TokenPair, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	u, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		// Run a dummy hash comparison to keep timing uniform.
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$"+strings.Repeat("x", 53)), []byte(password))
		return nil, nil, httpx.ErrUnauthorized("invalid email or password")
	}
	if err != nil {
		return nil, nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, nil, httpx.ErrUnauthorized("invalid email or password")
	}
	if u.Status != user.StatusActive {
		return nil, nil, httpx.ErrForbidden("account is not active")
	}

	pair, err := s.startSession(ctx, u, sc)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

// Refresh rotates the refresh token: the presented one is revoked and a new
// session is issued. A reused (already-revoked) token revokes the whole family.
func (s *Service) Refresh(ctx context.Context, refreshToken string, sc sessionContext) (*user.User, *TokenPair, error) {
	sess, err := s.sessions.getByHash(ctx, hashRefreshToken(refreshToken))
	if errors.Is(err, errSessionNotFound) {
		return nil, nil, httpx.ErrUnauthorized("invalid refresh token")
	}
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	if !sess.active(now) {
		// Token reuse or expiry — revoke everything for this user defensively.
		_ = s.sessions.revokeAllForUser(ctx, sess.UserID)
		return nil, nil, httpx.ErrUnauthorized("refresh token is no longer valid")
	}

	u, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.sessions.revoke(ctx, sess.ID); err != nil {
		return nil, nil, err
	}

	pair, err := s.startSession(ctx, u, sc)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	sess, err := s.sessions.getByHash(ctx, hashRefreshToken(refreshToken))
	if errors.Is(err, errSessionNotFound) {
		return nil // already gone
	}
	if err != nil {
		return err
	}
	return s.sessions.revoke(ctx, sess.ID)
}

func (s *Service) UserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *Service) startSession(ctx context.Context, u *user.User, sc sessionContext) (*TokenPair, error) {
	access, err := s.tokens.issueAccess(u)
	if err != nil {
		return nil, err
	}
	plaintext, hash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	sess := &Session{
		ID:        newSessionID(),
		UserID:    u.ID,
		TokenHash: hash,
		UserAgent: truncate(sc.UserAgent, 400),
		IP:        sc.IP,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
		CreatedAt: time.Now(),
	}
	if err := s.sessions.create(ctx, sess); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: plaintext,
		ExpiresIn:    int(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

func looksLikeEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	return at > 0 && at < len(s)-1 && strings.IndexByte(s[at+1:], '.') > 0
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
