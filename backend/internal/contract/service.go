package contract

import (
	"context"
	"errors"
	"strings"
	"time"

	"ipw/internal/httpx"
	"ipw/internal/project"

	"github.com/google/uuid"
)

// ProjectPort is the slice of the project feature this package needs. The
// project.PostgresStore satisfies it directly.
type ProjectPort interface {
	GetProject(ctx context.Context, id uuid.UUID) (*project.Project, error)
	GetProposal(ctx context.Context, id uuid.UUID) (*project.Proposal, error)
	ListProposalsByProject(ctx context.Context, projectID uuid.UUID) ([]project.Proposal, error)
	SetProposalStatus(ctx context.Context, id uuid.UUID, status project.ProposalStatus) error
	SetProjectStatus(ctx context.Context, id uuid.UUID, status project.Status, publishedAt *time.Time) error
}

type Service struct {
	store    Store
	projects ProjectPort
}

func NewService(store Store, projects ProjectPort) *Service {
	return &Service{store: store, projects: projects}
}

// AcceptProposal turns a proposal into a contract: the winning proposal is
// marked accepted, the project moves to in_progress, and every other live
// proposal on the project is declined.
func (s *Service) AcceptProposal(ctx context.Context, clientID, proposalID uuid.UUID) (*Contract, error) {
	prop, err := s.projects.GetProposal(ctx, proposalID)
	if err != nil {
		return nil, notFound(err, "proposal")
	}
	if prop.Status != project.ProposalPending && prop.Status != project.ProposalShortlisted {
		return nil, httpx.ErrConflict("proposal is not open for acceptance")
	}

	proj, err := s.projects.GetProject(ctx, prop.ProjectID)
	if err != nil {
		return nil, notFound(err, "project")
	}
	if proj.ClientID != clientID {
		return nil, httpx.ErrForbidden("not your project")
	}
	if proj.Status != project.StatusOpen {
		return nil, httpx.ErrConflict("project is not open")
	}

	if existing, err := s.store.ContractByProposal(ctx, proposalID); err == nil && existing != nil {
		return nil, httpx.ErrConflict("a contract already exists for this proposal")
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	c := &Contract{
		ID:                uuid.New(),
		ProjectID:         proj.ID,
		ProposalID:        prop.ID,
		ClientID:          clientID,
		FreelancerID:      prop.FreelancerID,
		Type:              Type(proj.BudgetType),
		AgreedAmountCents: prop.BidAmountCents,
		Currency:          proj.Currency,
		Status:            StatusActive,
		StartedAt:         time.Now(),
	}
	if err := s.store.CreateContract(ctx, c); err != nil {
		return nil, err
	}

	if err := s.projects.SetProposalStatus(ctx, prop.ID, project.ProposalAccepted); err != nil {
		return nil, err
	}
	if err := s.projects.SetProjectStatus(ctx, proj.ID, project.StatusInProgress, nil); err != nil {
		return nil, err
	}
	s.declineOtherProposals(ctx, proj.ID, prop.ID)

	return c, nil
}

func (s *Service) declineOtherProposals(ctx context.Context, projectID, keepID uuid.UUID) {
	others, err := s.projects.ListProposalsByProject(ctx, projectID)
	if err != nil {
		return
	}
	for _, o := range others {
		if o.ID == keepID {
			continue
		}
		if o.Status == project.ProposalPending || o.Status == project.ProposalShortlisted {
			_ = s.projects.SetProposalStatus(ctx, o.ID, project.ProposalDeclined)
		}
	}
}

func (s *Service) GetContractForParty(ctx context.Context, contractID, userID uuid.UUID) (*Contract, []Milestone, error) {
	c, err := s.store.GetContract(ctx, contractID)
	if err != nil {
		return nil, nil, notFound(err, "contract")
	}
	if c.ClientID != userID && c.FreelancerID != userID {
		return nil, nil, httpx.ErrForbidden("not a party to this contract")
	}
	ms, err := s.store.ListMilestones(ctx, contractID)
	if err != nil {
		return nil, nil, err
	}
	return c, ms, nil
}

func (s *Service) ListMyContracts(ctx context.Context, userID uuid.UUID) ([]Contract, error) {
	return s.store.ListContractsForUser(ctx, userID)
}

// ---- milestones ----------------------------------------------------

type MilestoneInput struct {
	Title       string
	Description string
	AmountCents int64
	DueDate     *time.Time
}

func (s *Service) AddMilestone(ctx context.Context, clientID, contractID uuid.UUID, in MilestoneInput) (*Milestone, error) {
	c, err := s.clientContract(ctx, clientID, contractID)
	if err != nil {
		return nil, err
	}
	if c.Type != TypeFixed {
		return nil, httpx.ErrConflict("milestones apply to fixed-price contracts only")
	}
	if c.Status != StatusActive {
		return nil, httpx.ErrConflict("contract is not active")
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, httpx.ErrBadRequest("title is required")
	}
	if in.AmountCents <= 0 {
		return nil, httpx.ErrBadRequest("amount must be positive")
	}

	existing, err := s.store.ListMilestones(ctx, contractID)
	if err != nil {
		return nil, err
	}
	var allocated int64
	for _, m := range existing {
		if m.Status != MilestoneCancelled {
			allocated += m.AmountCents
		}
	}
	if allocated+in.AmountCents > c.AgreedAmountCents {
		return nil, httpx.ErrBadRequest("milestone amounts would exceed the agreed contract amount")
	}

	seq, err := s.store.NextMilestoneSequence(ctx, contractID)
	if err != nil {
		return nil, err
	}
	m := &Milestone{
		ID:          uuid.New(),
		ContractID:  contractID,
		Sequence:    seq,
		Title:       strings.TrimSpace(in.Title),
		Description: strings.TrimSpace(in.Description),
		AmountCents: in.AmountCents,
		Status:      MilestonePending,
		DueDate:     in.DueDate,
	}
	if err := s.store.CreateMilestone(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// FundMilestone moves a milestone into escrow.
//
// TODO(payments): this is currently a structural transition only. The payments
// feature will replace the body with: create a lava.top invoice, and on webhook
// confirmation debit the client and credit the milestone's escrow account
// before flipping the status.
func (s *Service) FundMilestone(ctx context.Context, clientID, milestoneID uuid.UUID) (*Milestone, error) {
	m, _, err := s.milestoneForClient(ctx, clientID, milestoneID)
	if err != nil {
		return nil, err
	}
	if m.Status != MilestonePending {
		return nil, httpx.ErrConflict("milestone is not pending")
	}
	if err := s.store.UpdateMilestoneStatus(ctx, milestoneID, MilestoneFunded,
		map[string]time.Time{"funded_at": time.Now()}, ""); err != nil {
		return nil, err
	}
	return s.store.GetMilestone(ctx, milestoneID)
}

func (s *Service) SubmitMilestone(ctx context.Context, freelancerID, milestoneID uuid.UUID, note string) (*Milestone, error) {
	m, c, err := s.milestoneWithContract(ctx, milestoneID)
	if err != nil {
		return nil, err
	}
	if c.FreelancerID != freelancerID {
		return nil, httpx.ErrForbidden("not your contract")
	}
	if m.Status != MilestoneFunded {
		return nil, httpx.ErrConflict("milestone must be funded before work can be submitted")
	}
	if err := s.store.UpdateMilestoneStatus(ctx, milestoneID, MilestoneSubmitted,
		map[string]time.Time{"submitted_at": time.Now()}, strings.TrimSpace(note)); err != nil {
		return nil, err
	}
	return s.store.GetMilestone(ctx, milestoneID)
}

func (s *Service) RequestMilestoneChanges(ctx context.Context, clientID, milestoneID uuid.UUID) (*Milestone, error) {
	m, _, err := s.milestoneForClient(ctx, clientID, milestoneID)
	if err != nil {
		return nil, err
	}
	if m.Status != MilestoneSubmitted {
		return nil, httpx.ErrConflict("milestone is not awaiting review")
	}
	if err := s.store.UpdateMilestoneStatus(ctx, milestoneID, MilestoneFunded, nil, ""); err != nil {
		return nil, err
	}
	return s.store.GetMilestone(ctx, milestoneID)
}

// ApproveMilestone accepts the submitted work.
//
// TODO(payments): after approval the payments feature releases the escrowed
// amount to the freelancer's balance (minus platform commission) and moves the
// milestone to "released".
func (s *Service) ApproveMilestone(ctx context.Context, clientID, milestoneID uuid.UUID) (*Milestone, error) {
	m, _, err := s.milestoneForClient(ctx, clientID, milestoneID)
	if err != nil {
		return nil, err
	}
	if m.Status != MilestoneSubmitted {
		return nil, httpx.ErrConflict("milestone is not awaiting review")
	}
	if err := s.store.UpdateMilestoneStatus(ctx, milestoneID, MilestoneApproved,
		map[string]time.Time{"approved_at": time.Now()}, ""); err != nil {
		return nil, err
	}
	return s.store.GetMilestone(ctx, milestoneID)
}

func (s *Service) CancelMilestone(ctx context.Context, clientID, milestoneID uuid.UUID) (*Milestone, error) {
	m, _, err := s.milestoneForClient(ctx, clientID, milestoneID)
	if err != nil {
		return nil, err
	}
	if m.Status != MilestonePending {
		return nil, httpx.ErrConflict("only a pending milestone can be cancelled")
	}
	if err := s.store.UpdateMilestoneStatus(ctx, milestoneID, MilestoneCancelled, nil, ""); err != nil {
		return nil, err
	}
	return s.store.GetMilestone(ctx, milestoneID)
}

// CompleteContract closes a contract once every milestone has been resolved.
func (s *Service) CompleteContract(ctx context.Context, clientID, contractID uuid.UUID) (*Contract, error) {
	c, err := s.clientContract(ctx, clientID, contractID)
	if err != nil {
		return nil, err
	}
	if c.Status != StatusActive {
		return nil, httpx.ErrConflict("contract is not active")
	}
	ms, err := s.store.ListMilestones(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, httpx.ErrConflict("contract has no milestones")
	}
	for _, m := range ms {
		if m.Status != MilestoneApproved && m.Status != MilestoneReleased && m.Status != MilestoneCancelled {
			return nil, httpx.ErrConflict("all milestones must be approved or cancelled first")
		}
	}
	now := time.Now()
	if err := s.store.SetContractStatus(ctx, contractID, StatusCompleted, &now); err != nil {
		return nil, err
	}
	return s.store.GetContract(ctx, contractID)
}

// ---- shared helpers ----------------------------------------------

func (s *Service) clientContract(ctx context.Context, clientID, contractID uuid.UUID) (*Contract, error) {
	c, err := s.store.GetContract(ctx, contractID)
	if err != nil {
		return nil, notFound(err, "contract")
	}
	if c.ClientID != clientID {
		return nil, httpx.ErrForbidden("not your contract")
	}
	return c, nil
}

func (s *Service) milestoneWithContract(ctx context.Context, milestoneID uuid.UUID) (*Milestone, *Contract, error) {
	m, err := s.store.GetMilestone(ctx, milestoneID)
	if err != nil {
		return nil, nil, notFound(err, "milestone")
	}
	c, err := s.store.GetContract(ctx, m.ContractID)
	if err != nil {
		return nil, nil, err
	}
	return m, c, nil
}

func (s *Service) milestoneForClient(ctx context.Context, clientID, milestoneID uuid.UUID) (*Milestone, *Contract, error) {
	m, c, err := s.milestoneWithContract(ctx, milestoneID)
	if err != nil {
		return nil, nil, err
	}
	if c.ClientID != clientID {
		return nil, nil, httpx.ErrForbidden("not your contract")
	}
	return m, c, nil
}

func notFound(err error, what string) error {
	if errors.Is(err, ErrNotFound) || errors.Is(err, project.ErrNotFound) {
		return httpx.ErrNotFound(what + " not found")
	}
	return err
}
