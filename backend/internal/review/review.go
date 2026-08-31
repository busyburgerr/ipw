// Package review holds the two-sided feedback left after a contract completes.
// Reviews are double-blind: neither side's review is visible until both have
// submitted (or the review window closes).
package review

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Direction string

const (
	ClientToFreelancer Direction = "client_to_freelancer"
	FreelancerToClient Direction = "freelancer_to_client"
)

type Review struct {
	ID          uuid.UUID
	ContractID  uuid.UUID
	ReviewerID  uuid.UUID
	RevieweeID  uuid.UUID
	Direction   Direction
	Rating      int
	Comment     string
	PublishedAt *time.Time
	CreatedAt   time.Time
}

type Store interface {
	Create(ctx context.Context, r *Review) error
	ByContractAndReviewer(ctx context.Context, contractID, reviewerID uuid.UUID) (*Review, error)
	ListByContract(ctx context.Context, contractID uuid.UUID) ([]Review, error)
	ListPublishedForReviewee(ctx context.Context, revieweeID uuid.UUID) ([]Review, error)
	PublishForContract(ctx context.Context, contractID uuid.UUID, at time.Time) error
	RevieweeRatingSummary(ctx context.Context, revieweeID uuid.UUID) (avg float64, count int, err error)
	UnpublishedOlderThan(ctx context.Context, cutoff time.Time) ([]uuid.UUID, error)
}

// ProfilePort lets the review service push updated rating aggregates onto the
// relevant profile. Implemented by the profile store.
type ProfilePort interface {
	SetFreelancerRating(ctx context.Context, userID uuid.UUID, avg float64, count int) error
	SetClientRating(ctx context.Context, userID uuid.UUID, avg float64, count int) error
}
