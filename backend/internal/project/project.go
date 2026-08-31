// Package project owns the job marketplace: a client's Project posting and the
// freelancers' Proposals against it. Accepting a proposal is handled by the
// contract feature, which transitions the project to in_progress.
package project

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type BudgetType string

const (
	BudgetFixed  BudgetType = "fixed"
	BudgetHourly BudgetType = "hourly"
)

type ExperienceLevel string

const (
	ExperienceAny          ExperienceLevel = "any"
	ExperienceEntry        ExperienceLevel = "entry"
	ExperienceIntermediate ExperienceLevel = "intermediate"
	ExperienceExpert       ExperienceLevel = "expert"
)

type Status string

const (
	StatusDraft      Status = "draft"
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type ProposalStatus string

const (
	ProposalPending     ProposalStatus = "pending"
	ProposalShortlisted ProposalStatus = "shortlisted"
	ProposalAccepted    ProposalStatus = "accepted"
	ProposalDeclined    ProposalStatus = "declined"
	ProposalWithdrawn   ProposalStatus = "withdrawn"
)

type Project struct {
	ID                 uuid.UUID
	ClientID           uuid.UUID
	Title              string
	Description        string
	CategoryID         *uuid.UUID
	BudgetType         BudgetType
	FixedAmountCents   *int64
	HourlyRateMinCents *int64
	HourlyRateMaxCents *int64
	Currency           string
	ExperienceLevel    ExperienceLevel
	Status             Status
	SkillIDs           []uuid.UUID
	ProposalsCount     int
	PublishedAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Proposal struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	FreelancerID   uuid.UUID
	CoverLetter    string
	BidAmountCents int64
	EstimatedDays  *int
	Status         ProposalStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Filter narrows a public project listing.
type Filter struct {
	Query        string
	CategorySlug string
	SkillIDs     []uuid.UUID
	BudgetType   BudgetType
	Limit        int
	Offset       int
}

type Store interface {
	CreateProject(ctx context.Context, p *Project) error
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	UpdateProject(ctx context.Context, p *Project) error
	SetProjectSkills(ctx context.Context, projectID uuid.UUID, skillIDs []uuid.UUID) error
	ListPublicProjects(ctx context.Context, f Filter) ([]Project, int64, error)
	ListProjectsByClient(ctx context.Context, clientID uuid.UUID) ([]Project, error)
	SetProjectStatus(ctx context.Context, id uuid.UUID, status Status, publishedAt *time.Time) error

	CreateProposal(ctx context.Context, p *Proposal) error
	GetProposal(ctx context.Context, id uuid.UUID) (*Proposal, error)
	ListProposalsByProject(ctx context.Context, projectID uuid.UUID) ([]Proposal, error)
	ListProposalsByFreelancer(ctx context.Context, freelancerID uuid.UUID) ([]Proposal, error)
	SetProposalStatus(ctx context.Context, id uuid.UUID, status ProposalStatus) error
	HasProposal(ctx context.Context, projectID, freelancerID uuid.UUID) (bool, error)
}
