package project

import (
	"context"
	"errors"
	"strings"
	"time"

	"ipw/internal/catalog"
	"ipw/internal/httpx"

	"github.com/google/uuid"
)

type Service struct {
	store   Store
	catalog catalog.Store
}

func NewService(store Store, cat catalog.Store) *Service {
	return &Service{store: store, catalog: cat}
}

// ---- projects ------------------------------------------------------------

type ProjectInput struct {
	Title              string
	Description        string
	CategoryID         *uuid.UUID
	BudgetType         string
	FixedAmountCents   *int64
	HourlyRateMinCents *int64
	HourlyRateMaxCents *int64
	Currency           string
	ExperienceLevel    string
	SkillIDs           []uuid.UUID
}

func (s *Service) CreateProject(ctx context.Context, clientID uuid.UUID, in ProjectInput) (*Project, error) {
	p, err := s.validateProjectInput(ctx, in)
	if err != nil {
		return nil, err
	}
	p.ID = uuid.New()
	p.ClientID = clientID
	p.Status = StatusDraft
	if err := s.store.CreateProject(ctx, p); err != nil {
		return nil, err
	}
	if err := s.store.SetProjectSkills(ctx, p.ID, in.SkillIDs); err != nil {
		return nil, err
	}
	return s.store.GetProject(ctx, p.ID)
}

func (s *Service) UpdateProject(ctx context.Context, clientID, projectID uuid.UUID, in ProjectInput) (*Project, error) {
	existing, err := s.ownedProject(ctx, clientID, projectID)
	if err != nil {
		return nil, err
	}
	if existing.Status != StatusDraft && existing.Status != StatusOpen {
		return nil, httpx.ErrConflict("only draft or open projects can be edited")
	}
	p, err := s.validateProjectInput(ctx, in)
	if err != nil {
		return nil, err
	}
	p.ID = projectID
	p.ClientID = clientID
	if err := s.store.UpdateProject(ctx, p); err != nil {
		return nil, err
	}
	if err := s.store.SetProjectSkills(ctx, projectID, in.SkillIDs); err != nil {
		return nil, err
	}
	return s.store.GetProject(ctx, projectID)
}

func (s *Service) PublishProject(ctx context.Context, clientID, projectID uuid.UUID) (*Project, error) {
	p, err := s.ownedProject(ctx, clientID, projectID)
	if err != nil {
		return nil, err
	}
	if p.Status != StatusDraft {
		return nil, httpx.ErrConflict("project is not a draft")
	}
	if strings.TrimSpace(p.Title) == "" || strings.TrimSpace(p.Description) == "" {
		return nil, httpx.ErrBadRequest("title and description are required to publish")
	}
	now := time.Now()
	if err := s.store.SetProjectStatus(ctx, projectID, StatusOpen, &now); err != nil {
		return nil, err
	}
	return s.store.GetProject(ctx, projectID)
}

func (s *Service) CloseProject(ctx context.Context, clientID, projectID uuid.UUID) (*Project, error) {
	p, err := s.ownedProject(ctx, clientID, projectID)
	if err != nil {
		return nil, err
	}
	if p.Status == StatusCompleted || p.Status == StatusCancelled {
		return nil, httpx.ErrConflict("project is already closed")
	}
	if err := s.store.SetProjectStatus(ctx, projectID, StatusCancelled, nil); err != nil {
		return nil, err
	}
	return s.store.GetProject(ctx, projectID)
}

func (s *Service) GetProjectForViewer(ctx context.Context, projectID uuid.UUID, viewerID uuid.UUID) (*Project, error) {
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	// Drafts are visible only to their owner.
	if p.Status == StatusDraft && p.ClientID != viewerID {
		return nil, httpx.ErrNotFound("project not found")
	}
	return p, nil
}

// ListPublic returns open projects. An unknown category slug yields an empty
// page rather than an error.
func (s *Service) ListPublic(ctx context.Context, f Filter) ([]Project, int64, error) {
	return s.store.ListPublicProjects(ctx, f)
}

func (s *Service) ListMine(ctx context.Context, clientID uuid.UUID) ([]Project, error) {
	return s.store.ListProjectsByClient(ctx, clientID)
}

// ---- proposals ----------------------------------------------------------

type ProposalInput struct {
	CoverLetter    string
	BidAmountCents int64
	EstimatedDays  *int
}

func (s *Service) SubmitProposal(ctx context.Context, freelancerID, projectID uuid.UUID, in ProposalInput) (*Proposal, error) {
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if p.Status != StatusOpen {
		return nil, httpx.ErrConflict("project is not accepting proposals")
	}
	if p.ClientID == freelancerID {
		return nil, httpx.ErrBadRequest("you cannot bid on your own project")
	}
	if in.BidAmountCents <= 0 {
		return nil, httpx.ErrBadRequest("bid amount must be positive")
	}
	if in.EstimatedDays != nil && *in.EstimatedDays <= 0 {
		return nil, httpx.ErrBadRequest("estimated days must be positive")
	}
	exists, err := s.store.HasProposal(ctx, projectID, freelancerID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, httpx.ErrConflict("you have already submitted a proposal for this project")
	}

	prop := &Proposal{
		ID:             uuid.New(),
		ProjectID:      projectID,
		FreelancerID:   freelancerID,
		CoverLetter:    strings.TrimSpace(in.CoverLetter),
		BidAmountCents: in.BidAmountCents,
		EstimatedDays:  in.EstimatedDays,
		Status:         ProposalPending,
	}
	if err := s.store.CreateProposal(ctx, prop); err != nil {
		return nil, err
	}
	return prop, nil
}

func (s *Service) ListProjectProposals(ctx context.Context, clientID, projectID uuid.UUID) ([]Proposal, error) {
	if _, err := s.ownedProject(ctx, clientID, projectID); err != nil {
		return nil, err
	}
	return s.store.ListProposalsByProject(ctx, projectID)
}

func (s *Service) ListMyProposals(ctx context.Context, freelancerID uuid.UUID) ([]Proposal, error) {
	return s.store.ListProposalsByFreelancer(ctx, freelancerID)
}

func (s *Service) WithdrawProposal(ctx context.Context, freelancerID, proposalID uuid.UUID) error {
	prop, err := s.store.GetProposal(ctx, proposalID)
	if err != nil {
		return mapNotFound(err)
	}
	if prop.FreelancerID != freelancerID {
		return httpx.ErrForbidden("not your proposal")
	}
	if prop.Status != ProposalPending && prop.Status != ProposalShortlisted {
		return httpx.ErrConflict("proposal can no longer be withdrawn")
	}
	return s.store.SetProposalStatus(ctx, proposalID, ProposalWithdrawn)
}

// SetProposalDecision lets the project owner shortlist or decline a proposal.
func (s *Service) SetProposalDecision(ctx context.Context, clientID, proposalID uuid.UUID, decision ProposalStatus) (*Proposal, error) {
	if decision != ProposalShortlisted && decision != ProposalDeclined {
		return nil, httpx.ErrBadRequest("unsupported decision")
	}
	prop, err := s.store.GetProposal(ctx, proposalID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if _, err := s.ownedProject(ctx, clientID, prop.ProjectID); err != nil {
		return nil, err
	}
	if prop.Status != ProposalPending && prop.Status != ProposalShortlisted {
		return nil, httpx.ErrConflict("proposal is not open for a decision")
	}
	if err := s.store.SetProposalStatus(ctx, proposalID, decision); err != nil {
		return nil, err
	}
	return s.store.GetProposal(ctx, proposalID)
}

// ---- shared helpers ---------------------------------------------------

func (s *Service) ownedProject(ctx context.Context, clientID, projectID uuid.UUID) (*Project, error) {
	p, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if p.ClientID != clientID {
		return nil, httpx.ErrForbidden("not your project")
	}
	return p, nil
}

func (s *Service) validateProjectInput(ctx context.Context, in ProjectInput) (*Project, error) {
	title := strings.TrimSpace(in.Title)
	if len(title) < 5 {
		return nil, httpx.ErrBadRequest("title must be at least 5 characters")
	}
	bt := BudgetType(strings.TrimSpace(in.BudgetType))
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "RUB"
	}
	level := ExperienceLevel(strings.TrimSpace(in.ExperienceLevel))
	if level == "" {
		level = ExperienceAny
	}
	if !validExperience(level) {
		return nil, httpx.ErrBadRequest("invalid experience level")
	}

	p := &Project{
		Title:           title,
		Description:     strings.TrimSpace(in.Description),
		CategoryID:      in.CategoryID,
		Currency:        currency,
		ExperienceLevel: level,
	}

	switch bt {
	case BudgetFixed:
		if in.FixedAmountCents == nil || *in.FixedAmountCents <= 0 {
			return nil, httpx.ErrBadRequest("fixed budget requires a positive fixedAmountCents")
		}
		p.BudgetType = BudgetFixed
		p.FixedAmountCents = in.FixedAmountCents
	case BudgetHourly:
		lo, hi := in.HourlyRateMinCents, in.HourlyRateMaxCents
		if lo == nil || hi == nil || *lo <= 0 || *hi < *lo {
			return nil, httpx.ErrBadRequest("hourly budget requires hourlyRateMinCents <= hourlyRateMaxCents, both positive")
		}
		p.BudgetType = BudgetHourly
		p.HourlyRateMinCents = lo
		p.HourlyRateMaxCents = hi
	default:
		return nil, httpx.ErrBadRequest(`budgetType must be "fixed" or "hourly"`)
	}

	if in.CategoryID != nil {
		ok, err := s.catalog.CategoryExists(ctx, *in.CategoryID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpx.ErrBadRequest("unknown category")
		}
	}
	if len(in.SkillIDs) > 20 {
		return nil, httpx.ErrBadRequest("at most 20 skills")
	}
	if len(in.SkillIDs) > 0 {
		ok, err := s.catalog.SkillIDsExist(ctx, dedupeUUIDs(in.SkillIDs))
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpx.ErrBadRequest("one or more skills are unknown")
		}
	}
	return p, nil
}

func validExperience(l ExperienceLevel) bool {
	switch l {
	case ExperienceAny, ExperienceEntry, ExperienceIntermediate, ExperienceExpert:
		return true
	}
	return false
}

func mapNotFound(err error) error {
	if errors.Is(err, ErrNotFound) {
		return httpx.ErrNotFound("project not found")
	}
	return err
}

func dedupeUUIDs(in []uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(in))
	seen := map[uuid.UUID]bool{}
	for _, id := range in {
		if id == uuid.Nil || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
