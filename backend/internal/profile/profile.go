// Package profile owns the freelancer and client profiles that hang off a user
// account. A user with the matching capability has exactly one of each.
package profile

import (
	"context"

	"github.com/google/uuid"
)

type Availability string

const (
	AvailabilityAvailable   Availability = "available"
	AvailabilityLimited     Availability = "limited"
	AvailabilityUnavailable Availability = "unavailable"
	AvailabilityUnknown     Availability = "unknown"
)

func (a Availability) valid() bool {
	switch a {
	case AvailabilityAvailable, AvailabilityLimited, AvailabilityUnavailable, AvailabilityUnknown:
		return true
	}
	return false
}

// Freelancer is a freelancer's public-facing profile plus denormalised
// reputation aggregates (maintained by the reviews/contracts features).
type Freelancer struct {
	UserID            uuid.UUID
	Headline          string
	Bio               string
	HourlyRateCents   int64
	Currency          string
	Availability      Availability
	PrimaryCategoryID *uuid.UUID
	Languages         []string
	Location          string
	SkillIDs          []uuid.UUID

	RatingAvg        float64
	RatingCount      int
	JobsCompleted    int
	TotalEarnedCents int64
}

type Client struct {
	UserID          uuid.UUID
	CompanyName     string
	About           string
	Website         string
	Location        string
	PaymentVerified bool

	RatingAvg       float64
	RatingCount     int
	HiresCount      int
	TotalSpentCents int64
}

type PortfolioItem struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	URL         string
	ImageKey    string
	Position    int
}

// FreelancerFilter narrows the public freelancer directory.
type FreelancerFilter struct {
	Query        string
	CategorySlug string
	Limit        int
	Offset       int
}

type Store interface {
	GetFreelancer(ctx context.Context, userID uuid.UUID) (*Freelancer, error)
	ListFreelancers(ctx context.Context, f FreelancerFilter) ([]Freelancer, error)
	UpsertFreelancer(ctx context.Context, f *Freelancer) error
	SetFreelancerSkills(ctx context.Context, userID uuid.UUID, skillIDs []uuid.UUID) error

	GetClient(ctx context.Context, userID uuid.UUID) (*Client, error)
	UpsertClient(ctx context.Context, c *Client) error

	AddPortfolioItem(ctx context.Context, item *PortfolioItem) error
	ListPortfolio(ctx context.Context, userID uuid.UUID) ([]PortfolioItem, error)
	DeletePortfolioItem(ctx context.Context, userID, itemID uuid.UUID) error
}
