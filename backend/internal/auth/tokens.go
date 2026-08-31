package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"ipw/internal/config"
	"ipw/internal/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is the payload of the short-lived access token.
type Claims struct {
	jwt.RegisteredClaims
	IsFreelancer bool `json:"frl"`
	IsClient     bool `json:"cln"`
	IsAdmin      bool `json:"adm"`
}

type tokenIssuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func newTokenIssuer(cfg config.AuthConfig) *tokenIssuer {
	return &tokenIssuer{
		secret:     []byte(cfg.JWTSecret),
		accessTTL:  cfg.AccessTokenTTL,
		refreshTTL: cfg.RefreshTokenTTL,
	}
}

func (t *tokenIssuer) issueAccess(u *user.User) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.accessTTL)),
		},
		IsFreelancer: u.IsFreelancer,
		IsClient:     u.IsClient,
		IsAdmin:      u.IsAdmin,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
}

func (t *tokenIssuer) parseAccess(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(tk *jwt.Token) (any, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return t.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// newRefreshToken returns the opaque token to hand to the client and the hash to
// persist. The plaintext is never stored.
func newRefreshToken() (plaintext string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	plaintext = base64.RawURLEncoding.EncodeToString(buf)
	return plaintext, hashRefreshToken(plaintext), nil
}

func hashRefreshToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

func newSessionID() uuid.UUID { return uuid.New() }
