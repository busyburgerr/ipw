// Package contract owns the working relationship created when a client accepts a
// proposal: the Contract and, for fixed-price work, its Milestones.
//
// Money movement (funding a milestone into escrow, releasing it on approval) is
// added by the payments feature; this package models the structure and the
// non-monetary transitions (submit work, approve, request changes).
package contract

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeFixed  Type = "fixed"
	TypeHourly Type = "hourly"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
	StatusDisputed  Status = "disputed"
)

type MilestoneStatus string

const (
	MilestonePending   MilestoneStatus = "pending"
	MilestoneFunded    MilestoneStatus = "funded"
	MilestoneSubmitted MilestoneStatus = "submitted"
	MilestoneApproved  MilestoneStatus = "approved"
	MilestoneReleased  MilestoneStatus = "released"
	MilestoneCancelled MilestoneStatus = "cancelled"
)

type Contract struct {
	ID                uuid.UUID
	ProjectID         uuid.UUID
	ProposalID        uuid.UUID
	ClientID          uuid.UUID
	FreelancerID      uuid.UUID
	Type              Type
	AgreedAmountCents int64
	Currency          string
	Status            Status
	StartedAt         time.Time
	EndedAt           *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Milestone struct {
	ID              uuid.UUID
	ContractID      uuid.UUID
	Sequence        int
	Title           string
	Description     string
	AmountCents     int64
	Status          MilestoneStatus
	DeliverableNote string
	DueDate         *time.Time
	FundedAt        *time.Time
	SubmittedAt     *time.Time
	ApprovedAt      *time.Time
	ReleasedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Store interface {
	CreateContract(ctx context.Context, c *Contract) error
	GetContract(ctx context.Context, id uuid.UUID) (*Contract, error)
	ContractByProposal(ctx context.Context, proposalID uuid.UUID) (*Contract, error)
	ListContractsForUser(ctx context.Context, userID uuid.UUID) ([]Contract, error)
	SetContractStatus(ctx context.Context, id uuid.UUID, status Status, endedAt *time.Time) error

	CreateMilestone(ctx context.Context, m *Milestone) error
	GetMilestone(ctx context.Context, id uuid.UUID) (*Milestone, error)
	ListMilestones(ctx context.Context, contractID uuid.UUID) ([]Milestone, error)
	NextMilestoneSequence(ctx context.Context, contractID uuid.UUID) (int, error)
	UpdateMilestoneStatus(ctx context.Context, id uuid.UUID, status MilestoneStatus, ts map[string]time.Time, note string) error
	SumApprovedOrReleased(ctx context.Context, contractID uuid.UUID) (int64, error)
}
