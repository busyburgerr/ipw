package contract

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type contractRow struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID         uuid.UUID `gorm:"type:uuid"`
	ProposalID        uuid.UUID `gorm:"type:uuid"`
	ClientID          uuid.UUID `gorm:"type:uuid"`
	FreelancerID      uuid.UUID `gorm:"type:uuid"`
	Type              string
	AgreedAmountCents int64
	Currency          string
	Status            string
	StartedAt         time.Time
	EndedAt           *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (contractRow) TableName() string { return "contracts" }

type milestoneRow struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	ContractID      uuid.UUID `gorm:"type:uuid"`
	Sequence        int
	Title           string
	Description     string
	AmountCents     int64
	Status          string
	DeliverableNote string
	DueDate         *time.Time
	FundedAt        *time.Time
	SubmittedAt     *time.Time
	ApprovedAt      *time.Time
	ReleasedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (milestoneRow) TableName() string { return "milestones" }

type PostgresStore struct{ db *gorm.DB }

func NewPostgresStore(db *gorm.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) CreateContract(ctx context.Context, c *Contract) error {
	row := toContractRow(c)
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	*c = *fromContractRow(row)
	return nil
}

func (s *PostgresStore) GetContract(ctx context.Context, id uuid.UUID) (*Contract, error) {
	var row contractRow
	err := s.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromContractRow(row), nil
}

func (s *PostgresStore) ContractByProposal(ctx context.Context, proposalID uuid.UUID) (*Contract, error) {
	var row contractRow
	err := s.db.WithContext(ctx).First(&row, "proposal_id = ?", proposalID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromContractRow(row), nil
}

func (s *PostgresStore) ListContractsForUser(ctx context.Context, userID uuid.UUID) ([]Contract, error) {
	var rows []contractRow
	err := s.db.WithContext(ctx).
		Where("client_id = ? OR freelancer_id = ?", userID, userID).
		Order("created_at DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Contract, len(rows))
	for i, r := range rows {
		out[i] = *fromContractRow(r)
	}
	return out, nil
}

func (s *PostgresStore) SetContractStatus(ctx context.Context, id uuid.UUID, status Status, endedAt *time.Time) error {
	fields := map[string]any{"status": string(status), "updated_at": time.Now()}
	if endedAt != nil {
		fields["ended_at"] = *endedAt
	}
	return s.db.WithContext(ctx).Model(&contractRow{}).Where("id = ?", id).Updates(fields).Error
}

func (s *PostgresStore) CreateMilestone(ctx context.Context, m *Milestone) error {
	row := toMilestoneRow(m)
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	*m = *fromMilestoneRow(row)
	return nil
}

func (s *PostgresStore) GetMilestone(ctx context.Context, id uuid.UUID) (*Milestone, error) {
	var row milestoneRow
	err := s.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromMilestoneRow(row), nil
}

func (s *PostgresStore) ListMilestones(ctx context.Context, contractID uuid.UUID) ([]Milestone, error) {
	var rows []milestoneRow
	err := s.db.WithContext(ctx).Where("contract_id = ?", contractID).
		Order("sequence").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Milestone, len(rows))
	for i, r := range rows {
		out[i] = *fromMilestoneRow(r)
	}
	return out, nil
}

func (s *PostgresStore) NextMilestoneSequence(ctx context.Context, contractID uuid.UUID) (int, error) {
	var max *int
	err := s.db.WithContext(ctx).Model(&milestoneRow{}).
		Where("contract_id = ?", contractID).
		Select("MAX(sequence)").Scan(&max).Error
	if err != nil {
		return 0, err
	}
	if max == nil {
		return 1, nil
	}
	return *max + 1, nil
}

// UpdateMilestoneStatus sets the status, any of the timestamp columns given in
// ts, and (when non-empty) the deliverable note, in one update.
func (s *PostgresStore) UpdateMilestoneStatus(ctx context.Context, id uuid.UUID, status MilestoneStatus, ts map[string]time.Time, note string) error {
	fields := map[string]any{"status": string(status), "updated_at": time.Now()}
	for col, when := range ts {
		fields[col] = when
	}
	if note != "" {
		fields["deliverable_note"] = note
	}
	return s.db.WithContext(ctx).Model(&milestoneRow{}).Where("id = ?", id).Updates(fields).Error
}

func (s *PostgresStore) SumApprovedOrReleased(ctx context.Context, contractID uuid.UUID) (int64, error) {
	var sum *int64
	err := s.db.WithContext(ctx).Model(&milestoneRow{}).
		Where("contract_id = ? AND status IN ?", contractID, []string{string(MilestoneApproved), string(MilestoneReleased)}).
		Select("SUM(amount_cents)").Scan(&sum).Error
	if err != nil {
		return 0, err
	}
	if sum == nil {
		return 0, nil
	}
	return *sum, nil
}

// ---- mapping ----------------------------------------------------------

func toContractRow(c *Contract) contractRow {
	return contractRow{
		ID: c.ID, ProjectID: c.ProjectID, ProposalID: c.ProposalID,
		ClientID: c.ClientID, FreelancerID: c.FreelancerID, Type: string(c.Type),
		AgreedAmountCents: c.AgreedAmountCents, Currency: c.Currency, Status: string(c.Status),
		StartedAt: c.StartedAt, EndedAt: c.EndedAt, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func fromContractRow(r contractRow) *Contract {
	return &Contract{
		ID: r.ID, ProjectID: r.ProjectID, ProposalID: r.ProposalID,
		ClientID: r.ClientID, FreelancerID: r.FreelancerID, Type: Type(r.Type),
		AgreedAmountCents: r.AgreedAmountCents, Currency: r.Currency, Status: Status(r.Status),
		StartedAt: r.StartedAt, EndedAt: r.EndedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func toMilestoneRow(m *Milestone) milestoneRow {
	return milestoneRow{
		ID: m.ID, ContractID: m.ContractID, Sequence: m.Sequence, Title: m.Title,
		Description: m.Description, AmountCents: m.AmountCents, Status: string(m.Status),
		DeliverableNote: m.DeliverableNote, DueDate: m.DueDate,
		FundedAt: m.FundedAt, SubmittedAt: m.SubmittedAt, ApprovedAt: m.ApprovedAt, ReleasedAt: m.ReleasedAt,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func fromMilestoneRow(r milestoneRow) *Milestone {
	return &Milestone{
		ID: r.ID, ContractID: r.ContractID, Sequence: r.Sequence, Title: r.Title,
		Description: r.Description, AmountCents: r.AmountCents, Status: MilestoneStatus(r.Status),
		DeliverableNote: r.DeliverableNote, DueDate: r.DueDate,
		FundedAt: r.FundedAt, SubmittedAt: r.SubmittedAt, ApprovedAt: r.ApprovedAt, ReleasedAt: r.ReleasedAt,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}
