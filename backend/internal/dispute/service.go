package dispute

import (
	"context"
	"errors"
	"strings"
	"time"

	"ipw/internal/contract"
	"ipw/internal/httpx"

	"github.com/google/uuid"
)

type Service struct {
	store     Store
	contracts contract.Store
	billing   BillingPort
}

func NewService(store Store, contracts contract.Store, billing BillingPort) *Service {
	return &Service{store: store, contracts: contracts, billing: billing}
}

type RaiseInput struct {
	MilestoneID *uuid.UUID
	Reason      string
}

func (s *Service) Raise(ctx context.Context, userID, contractID uuid.UUID, in RaiseInput) (*Dispute, error) {
	if strings.TrimSpace(in.Reason) == "" {
		return nil, httpx.ErrBadRequest("reason is required")
	}
	c, err := s.contracts.GetContract(ctx, contractID)
	if err != nil {
		if errors.Is(err, contract.ErrNotFound) {
			return nil, httpx.ErrNotFound("contract not found")
		}
		return nil, err
	}
	if userID != c.ClientID && userID != c.FreelancerID {
		return nil, httpx.ErrForbidden("not a party to this contract")
	}
	if c.Status != contract.StatusActive {
		return nil, httpx.ErrConflict("only an active contract can be disputed")
	}
	if in.MilestoneID != nil {
		m, err := s.contracts.GetMilestone(ctx, *in.MilestoneID)
		if err != nil || m.ContractID != contractID {
			return nil, httpx.ErrBadRequest("milestone does not belong to this contract")
		}
	}

	d := &Dispute{
		ID: uuid.New(), ContractID: contractID, MilestoneID: in.MilestoneID,
		RaisedBy: userID, Reason: strings.TrimSpace(in.Reason), Status: StatusOpen,
	}
	if err := s.store.Create(ctx, d); err != nil {
		return nil, err
	}
	if err := s.contracts.SetContractStatus(ctx, contractID, contract.StatusDisputed, nil); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) ListMine(ctx context.Context, userID uuid.UUID) ([]Dispute, error) {
	return s.store.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID uuid.UUID, id uuid.UUID, isAdmin bool) (*Dispute, error) {
	d, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if isAdmin {
		return d, nil
	}
	c, err := s.contracts.GetContract(ctx, d.ContractID)
	if err != nil {
		return nil, err
	}
	if userID != c.ClientID && userID != c.FreelancerID {
		return nil, httpx.ErrForbidden("not a party to this dispute")
	}
	return d, nil
}

func (s *Service) Withdraw(ctx context.Context, userID, id uuid.UUID) (*Dispute, error) {
	d, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if d.RaisedBy != userID {
		return nil, httpx.ErrForbidden("only the party who raised the dispute can withdraw it")
	}
	if d.Status != StatusOpen {
		return nil, httpx.ErrConflict("dispute is no longer open")
	}
	d.Status = StatusWithdrawn
	now := time.Now()
	d.ResolvedAt = &now
	if err := s.store.Update(ctx, d); err != nil {
		return nil, err
	}
	if err := s.contracts.SetContractStatus(ctx, d.ContractID, contract.StatusActive, nil); err != nil {
		return nil, err
	}
	return d, nil
}

// ---- arbiter (admin) ------------------------------------------------

func (s *Service) ListForArbiter(ctx context.Context) ([]Dispute, error) {
	return s.store.ListByStatus(ctx, StatusOpen, StatusUnderReview)
}

func (s *Service) Claim(ctx context.Context, arbiterID, id uuid.UUID) (*Dispute, error) {
	d, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if d.Status != StatusOpen {
		return nil, httpx.ErrConflict("dispute is not open")
	}
	d.Status = StatusUnderReview
	d.ArbiterID = &arbiterID
	if err := s.store.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

type Resolution string

const (
	ForClient     Resolution = "client"
	ForFreelancer Resolution = "freelancer"
)

func (s *Service) Resolve(ctx context.Context, arbiterID, id uuid.UUID, outcome Resolution, note string) (*Dispute, error) {
	d, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if d.Status != StatusOpen && d.Status != StatusUnderReview {
		return nil, httpx.ErrConflict("dispute is already resolved")
	}

	switch outcome {
	case ForFreelancer:
		if d.MilestoneID != nil {
			if err := s.billing.AdminReleaseMilestone(ctx, *d.MilestoneID); err != nil {
				return nil, err
			}
		}
		d.Status = StatusResolvedFreelancer
	case ForClient:
		if d.MilestoneID != nil {
			if err := s.billing.AdminRefundMilestone(ctx, *d.MilestoneID); err != nil {
				return nil, err
			}
		}
		d.Status = StatusResolvedClient
	default:
		return nil, httpx.ErrBadRequest("outcome must be 'client' or 'freelancer'")
	}

	now := time.Now()
	d.ArbiterID = &arbiterID
	d.ResolutionNote = strings.TrimSpace(note)
	d.ResolvedAt = &now
	if err := s.store.Update(ctx, d); err != nil {
		return nil, err
	}
	// Return the contract to active so the client can complete or continue it.
	if err := s.contracts.SetContractStatus(ctx, d.ContractID, contract.StatusActive, nil); err != nil {
		return nil, err
	}
	return d, nil
}

func mapNotFound(err error) error {
	if errors.Is(err, ErrNotFound) {
		return httpx.ErrNotFound("dispute not found")
	}
	return err
}
