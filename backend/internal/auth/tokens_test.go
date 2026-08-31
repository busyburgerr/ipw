package auth

import (
	"strings"
	"testing"
	"time"

	"ipw/internal/config"
	"ipw/internal/user"

	"github.com/google/uuid"
)

func testIssuer() *tokenIssuer {
	return newTokenIssuer(config.AuthConfig{
		JWTSecret:       strings.Repeat("k", 40),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	})
}

func TestAccessTokenRoundTrip(t *testing.T) {
	iss := testIssuer()
	u := &user.User{ID: uuid.New(), IsFreelancer: true, IsAdmin: false}

	tok, err := iss.issueAccess(u)
	if err != nil {
		t.Fatalf("issueAccess: %v", err)
	}

	claims, err := iss.parseAccess(tok)
	if err != nil {
		t.Fatalf("parseAccess: %v", err)
	}
	if claims.Subject != u.ID.String() {
		t.Errorf("subject = %q, want %q", claims.Subject, u.ID.String())
	}
	if !claims.IsFreelancer || claims.IsClient || claims.IsAdmin {
		t.Errorf("capability flags not preserved: %+v", claims)
	}
	if !claims.has(CapFreelancer) || claims.has(CapClient) {
		t.Errorf("has() disagrees with flags")
	}
}

func TestAccessTokenRejectsWrongSecret(t *testing.T) {
	tok, err := testIssuer().issueAccess(&user.User{ID: uuid.New()})
	if err != nil {
		t.Fatal(err)
	}
	other := newTokenIssuer(config.AuthConfig{JWTSecret: strings.Repeat("z", 40)})
	if _, err := other.parseAccess(tok); err == nil {
		t.Fatal("expected verification to fail with a different secret")
	}
}

func TestAdminImpliesAllCapabilities(t *testing.T) {
	c := &Claims{IsAdmin: true}
	for _, cap := range []Capability{CapFreelancer, CapClient, CapAdmin} {
		if !c.has(cap) {
			t.Errorf("admin should have %q", cap)
		}
	}
}

func TestRefreshTokenHashIsStableAndOpaque(t *testing.T) {
	plaintext, hash, err := newRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if plaintext == hash {
		t.Fatal("hash must not equal plaintext")
	}
	if hashRefreshToken(plaintext) != hash {
		t.Fatal("hash must be deterministic for the same input")
	}
}
