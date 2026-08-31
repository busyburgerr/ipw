// Package user owns the account entity and its persistence contract. Other
// features (auth, profiles, contracts) depend on this package, not the other way
// around.
package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

// User is an account. A single account may act as both a freelancer and a
// client; capabilities are explicit flags.
type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	DisplayName     string
	AvatarURL       string
	Country         string
	Timezone        string
	IsFreelancer    bool
	IsClient        bool
	IsAdmin         bool
	Status          Status
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Store is the persistence contract for users. Implemented by the postgres
// package; consumed by services.
type Store interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	SetAvatarURL(ctx context.Context, id uuid.UUID, url string) error
	EmailExists(ctx context.Context, email string) (bool, error)
}
