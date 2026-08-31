package contract

import (
	"context"
	"testing"
	"time"

	"ipw/internal/project"

	"github.com/google/uuid"
)

// --- in-memory fakes ---------------------------------------------------

type memStore struct {
	contracts  map[uuid.UUID]*Contract
	milestones map[uuid.UUID]*Milestone
}

func newMemStore() *memStore {
	return &memStore{contracts: map[uuid.UUID]*Contract{}, milestones: map[uuid.UUID]*Milestone{}}
}

func (m *memStore) CreateContract(_ context.Context, c *Contract) error {
	m.contracts[c.ID] = c
	return nil
}
func (m *memStore) GetContract(_ context.Context, id uuid.UUID) (*Contract, error) {
	if c, ok := m.contracts[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, ErrNotFound
}
func (m *memStore) ContractByProposal(context.Context, uuid.UUID) (*Contract, error) {
	return nil, ErrNotFound
}
func (m *memStore) ListContractsForUser(context.Context, uuid.UUID) ([]Contract, error) {
	return nil, nil
}
func (m *memStore) SetContractStatus(_ context.Context, id uuid.UUID, s Status, ended *time.Time) error {
	m.contracts[id].Status = s
	m.contracts[id].EndedAt = ended
	return nil
}
func (m *memStore) CreateMilestone(_ context.Context, ms *Milestone) error {
	m.milestones[ms.ID] = ms
	return nil
}
func (m *memStore) GetMilestone(_ context.Context, id uuid.UUID) (*Milestone, error) {
	if ms, ok := m.milestones[id]; ok {
		cp := *ms
		return &cp, nil
	}
	return nil, ErrNotFound
}
func (m *memStore) ListMilestones(_ context.Context, cid uuid.UUID) ([]Milestone, error) {
	var out []Milestone
	for _, ms := range m.milestones {
		if ms.ContractID == cid {
			out = append(out, *ms)
		}
	}
	return out, nil
}
func (m *memStore) NextMilestoneSequence(_ context.Context, cid uuid.UUID) (int, error) {
	n := 0
	for _, ms := range m.milestones {
		if ms.ContractID == cid && ms.Sequence > n {
			n = ms.Sequence
		}
	}
	return n + 1, nil
}
func (m *memStore) UpdateMilestoneStatus(_ context.Context, id uuid.UUID, s MilestoneStatus, _ map[string]time.Time, _ string) error {
	m.milestones[id].Status = s
	return nil
}
func (m *memStore) SumApprovedOrReleased(context.Context, uuid.UUID) (int64, error) { return 0, nil }

type nopProjects struct{}

func (nopProjects) GetProject(context.Context, uuid.UUID) (*project.Project, error) { return nil, nil }
func (nopProjects) GetProposal(context.Context, uuid.UUID) (*project.Proposal, error) {
	return nil, nil
}
func (nopProjects) ListProposalsByProject(context.Context, uuid.UUID) ([]project.Proposal, error) {
	return nil, nil
}
func (nopProjects) SetProposalStatus(context.Context, uuid.UUID, project.ProposalStatus) error {
	return nil
}
func (nopProjects) SetProjectStatus(context.Context, uuid.UUID, project.Status, *time.Time) error {
	return nil
}

// --- tests -----------------------------------------------------------

func seedContract(store *memStore) (*Service, *Contract, uuid.UUID) {
	svc := NewService(store, nopProjects{})
	clientID := uuid.New()
	c := &Contract{
		ID: uuid.New(), ClientID: clientID, FreelancerID: uuid.New(),
		Type: TypeFixed, AgreedAmountCents: 100000, Status: StatusActive,
	}
	store.contracts[c.ID] = c
	return svc, c, clientID
}

func TestAddMilestoneRejectsOverAllocation(t *testing.T) {
	store := newMemStore()
	svc, c, clientID := seedContract(store)
	ctx := context.Background()

	if _, err := svc.AddMilestone(ctx, clientID, c.ID, MilestoneInput{Title: "A", AmountCents: 60000}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddMilestone(ctx, clientID, c.ID, MilestoneInput{Title: "B", AmountCents: 50000}); err == nil {
		t.Fatal("expected over-allocation to be rejected (60000+50000 > 100000)")
	}
	if _, err := svc.AddMilestone(ctx, clientID, c.ID, MilestoneInput{Title: "B", AmountCents: 40000}); err != nil {
		t.Fatalf("40000 should fit: %v", err)
	}
}

func TestSubmitAndRequestChanges(t *testing.T) {
	store := newMemStore()
	svc, c, clientID := seedContract(store)
	ctx := context.Background()
	freelancerID := c.FreelancerID

	m, err := svc.AddMilestone(ctx, clientID, c.ID, MilestoneInput{Title: "Stage 1", AmountCents: 50000})
	if err != nil {
		t.Fatal(err)
	}

	// Cannot submit a pending (unfunded) milestone.
	if _, err := svc.SubmitMilestone(ctx, freelancerID, m.ID, "wip"); err == nil {
		t.Fatal("submit before fund should fail")
	}

	// Billing would fund it; simulate that here.
	store.milestones[m.ID].Status = MilestoneFunded

	if _, err := svc.SubmitMilestone(ctx, uuid.New(), m.ID, "x"); err == nil {
		t.Fatal("stranger submit should fail")
	}
	if _, err := svc.SubmitMilestone(ctx, freelancerID, m.ID, "done"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RequestMilestoneChanges(ctx, clientID, m.ID); err != nil {
		t.Fatal(err)
	}
	if got := store.milestones[m.ID].Status; got != MilestoneFunded {
		t.Fatalf("request-changes should return milestone to funded, got %s", got)
	}
	// Cancelling is only allowed while pending.
	if _, err := svc.CancelMilestone(ctx, clientID, m.ID); err == nil {
		t.Fatal("cancel of a funded milestone should fail")
	}
}
