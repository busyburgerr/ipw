// Package dispute handles contested contracts: a party raises a dispute, an
// admin arbiter reviews it and resolves it, which can force-release or refund
// the milestone's escrow through the billing feature.
package dispute

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusOpen               Status = "open"
	StatusUnderReview        Status = "under_review"
	StatusResolvedClient     Status = "resolved_client"
	StatusResolvedFreelancer Status = "resolved_freelancer"
	StatusResolvedSplit      Status = "resolved_split"
	StatusWithdrawn          Status = "withdrawn"
)

type Dispute struct {
	ID             uuid.UUID
	ContractID     uuid.UUID
	MilestoneID    *uuid.UUID
	RaisedBy       uuid.UUID
	Reason         string
	Status         Status
	ResolutionNote string
	ArbiterID      *uuid.UUID
	CreatedAt      time.Time
	ResolvedAt     *time.Time
}

type Store interface {
	Create(ctx context.Context, d *Dispute) error
	Get(ctx context.Context, id uuid.UUID) (*Dispute, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Dispute, error)
	ListByStatus(ctx context.Context, statuses ...Status) ([]Dispute, error)
	Update(ctx context.Context, d *Dispute) error
}

// BillingPort is the slice of the billing feature the arbiter needs.
type BillingPort interface {
	AdminReleaseMilestone(ctx context.Context, milestoneID uuid.UUID) error
	AdminRefundMilestone(ctx context.Context, milestoneID uuid.UUID) error
}
