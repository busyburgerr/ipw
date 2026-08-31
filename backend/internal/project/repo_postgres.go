package project

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type projectRow struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	ClientID           uuid.UUID `gorm:"type:uuid"`
	Title              string
	Description        string
	CategoryID         *uuid.UUID `gorm:"type:uuid"`
	BudgetType         string
	FixedAmountCents   *int64
	HourlyRateMinCents *int64
	HourlyRateMaxCents *int64
	Currency           string
	ExperienceLevel    string
	Status             string
	ProposalsCount     int
	PublishedAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (projectRow) TableName() string { return "projects" }

type projectSkillRow struct {
	ProjectID uuid.UUID `gorm:"type:uuid;primaryKey"`
	SkillID   uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (projectSkillRow) TableName() string { return "project_skills" }

type proposalRow struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID      uuid.UUID `gorm:"type:uuid"`
	FreelancerID   uuid.UUID `gorm:"type:uuid"`
	CoverLetter    string
	BidAmountCents int64
	EstimatedDays  *int
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (proposalRow) TableName() string { return "proposals" }

type PostgresStore struct{ db *gorm.DB }

func NewPostgresStore(db *gorm.DB) *PostgresStore { return &PostgresStore{db: db} }

// ---- projects ------------------------------------------------------------

func (s *PostgresStore) CreateProject(ctx context.Context, p *Project) error {
	row := toProjectRow(p)
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	*p = *fromProjectRow(row, nil)
	return nil
}

func (s *PostgresStore) GetProject(ctx context.Context, id uuid.UUID) (*Project, error) {
	var row projectRow
	err := s.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	skills, err := s.projectSkillIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	return fromProjectRow(row, skills), nil
}

func (s *PostgresStore) UpdateProject(ctx context.Context, p *Project) error {
	p.UpdatedAt = time.Now()
	row := toProjectRow(p)
	return s.db.WithContext(ctx).Model(&projectRow{}).
		Where("id = ?", p.ID).
		Select("title", "description", "category_id", "budget_type", "fixed_amount_cents",
			"hourly_rate_min_cents", "hourly_rate_max_cents", "currency", "experience_level", "updated_at").
		Updates(&row).Error
}

func (s *PostgresStore) SetProjectSkills(ctx context.Context, projectID uuid.UUID, skillIDs []uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", projectID).Delete(&projectSkillRow{}).Error; err != nil {
			return err
		}
		if len(skillIDs) == 0 {
			return nil
		}
		rows := make([]projectSkillRow, len(skillIDs))
		for i, id := range skillIDs {
			rows[i] = projectSkillRow{ProjectID: projectID, SkillID: id}
		}
		return tx.Create(&rows).Error
	})
}

func (s *PostgresStore) SetProjectStatus(ctx context.Context, id uuid.UUID, status Status, publishedAt *time.Time) error {
	fields := map[string]any{"status": string(status), "updated_at": time.Now()}
	if publishedAt != nil {
		fields["published_at"] = *publishedAt
	}
	return s.db.WithContext(ctx).Model(&projectRow{}).Where("id = ?", id).Updates(fields).Error
}

func (s *PostgresStore) ListPublicProjects(ctx context.Context, f Filter) ([]Project, int64, error) {
	q := s.db.WithContext(ctx).Model(&projectRow{}).Where("status = ?", string(StatusOpen))

	if f.Query != "" {
		like := "%" + f.Query + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}
	if f.CategorySlug != "" {
		q = q.Where("category_id = (SELECT id FROM categories WHERE slug = ?)", f.CategorySlug)
	}
	if f.BudgetType != "" {
		q = q.Where("budget_type = ?", string(f.BudgetType))
	}
	if len(f.SkillIDs) > 0 {
		q = q.Where(`id IN (
			SELECT project_id FROM project_skills
			WHERE skill_id IN ?
			GROUP BY project_id
			HAVING COUNT(DISTINCT skill_id) = ?)`, f.SkillIDs, len(f.SkillIDs))
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows []projectRow
	err := q.Order("published_at DESC NULLS LAST").
		Limit(limit).Offset(f.Offset).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return s.hydrateProjects(ctx, rows)
}

func (s *PostgresStore) ListProjectsByClient(ctx context.Context, clientID uuid.UUID) ([]Project, error) {
	var rows []projectRow
	err := s.db.WithContext(ctx).Where("client_id = ?", clientID).
		Order("created_at DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	list, _, err := s.hydrateProjects(ctx, rows)
	return list, err
}

func (s *PostgresStore) hydrateProjects(ctx context.Context, rows []projectRow) ([]Project, int64, error) {
	if len(rows) == 0 {
		return []Project{}, 0, nil
	}
	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	var links []projectSkillRow
	if err := s.db.WithContext(ctx).Where("project_id IN ?", ids).Find(&links).Error; err != nil {
		return nil, 0, err
	}
	bySkill := map[uuid.UUID][]uuid.UUID{}
	for _, l := range links {
		bySkill[l.ProjectID] = append(bySkill[l.ProjectID], l.SkillID)
	}
	out := make([]Project, len(rows))
	for i, r := range rows {
		out[i] = *fromProjectRow(r, bySkill[r.ID])
	}
	return out, int64(len(out)), nil
}

func (s *PostgresStore) projectSkillIDs(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := s.db.WithContext(ctx).Model(&projectSkillRow{}).
		Where("project_id = ?", projectID).Pluck("skill_id", &ids).Error
	return ids, err
}

// ---- proposals ----------------------------------------------------------

func (s *PostgresStore) CreateProposal(ctx context.Context, p *Proposal) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := toProposalRow(p)
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		*p = *fromProposalRow(row)
		return tx.Model(&projectRow{}).Where("id = ?", p.ProjectID).
			UpdateColumn("proposals_count", gorm.Expr("proposals_count + 1")).Error
	})
}

func (s *PostgresStore) GetProposal(ctx context.Context, id uuid.UUID) (*Proposal, error) {
	var row proposalRow
	err := s.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromProposalRow(row), nil
}

func (s *PostgresStore) ListProposalsByProject(ctx context.Context, projectID uuid.UUID) ([]Proposal, error) {
	var rows []proposalRow
	err := s.db.WithContext(ctx).Where("project_id = ?", projectID).
		Order("created_at DESC").Find(&rows).Error
	return mapProposals(rows), err
}

func (s *PostgresStore) ListProposalsByFreelancer(ctx context.Context, freelancerID uuid.UUID) ([]Proposal, error) {
	var rows []proposalRow
	err := s.db.WithContext(ctx).Where("freelancer_id = ?", freelancerID).
		Order("created_at DESC").Find(&rows).Error
	return mapProposals(rows), err
}

func (s *PostgresStore) SetProposalStatus(ctx context.Context, id uuid.UUID, status ProposalStatus) error {
	return s.db.WithContext(ctx).Model(&proposalRow{}).Where("id = ?", id).
		Updates(map[string]any{"status": string(status), "updated_at": time.Now()}).Error
}

func (s *PostgresStore) HasProposal(ctx context.Context, projectID, freelancerID uuid.UUID) (bool, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&proposalRow{}).
		Where("project_id = ? AND freelancer_id = ?", projectID, freelancerID).Count(&n).Error
	return n > 0, err
}

// ---- mapping ------------------------------------------------------------

func toProjectRow(p *Project) projectRow {
	return projectRow{
		ID: p.ID, ClientID: p.ClientID, Title: p.Title, Description: p.Description,
		CategoryID: p.CategoryID, BudgetType: string(p.BudgetType),
		FixedAmountCents: p.FixedAmountCents, HourlyRateMinCents: p.HourlyRateMinCents,
		HourlyRateMaxCents: p.HourlyRateMaxCents, Currency: p.Currency,
		ExperienceLevel: string(p.ExperienceLevel), Status: string(p.Status),
		ProposalsCount: p.ProposalsCount, PublishedAt: p.PublishedAt,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func fromProjectRow(r projectRow, skillIDs []uuid.UUID) *Project {
	return &Project{
		ID: r.ID, ClientID: r.ClientID, Title: r.Title, Description: r.Description,
		CategoryID: r.CategoryID, BudgetType: BudgetType(r.BudgetType),
		FixedAmountCents: r.FixedAmountCents, HourlyRateMinCents: r.HourlyRateMinCents,
		HourlyRateMaxCents: r.HourlyRateMaxCents, Currency: r.Currency,
		ExperienceLevel: ExperienceLevel(r.ExperienceLevel), Status: Status(r.Status),
		SkillIDs: skillIDs, ProposalsCount: r.ProposalsCount, PublishedAt: r.PublishedAt,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func toProposalRow(p *Proposal) proposalRow {
	return proposalRow{
		ID: p.ID, ProjectID: p.ProjectID, FreelancerID: p.FreelancerID,
		CoverLetter: p.CoverLetter, BidAmountCents: p.BidAmountCents,
		EstimatedDays: p.EstimatedDays, Status: string(p.Status),
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func fromProposalRow(r proposalRow) *Proposal {
	return &Proposal{
		ID: r.ID, ProjectID: r.ProjectID, FreelancerID: r.FreelancerID,
		CoverLetter: r.CoverLetter, BidAmountCents: r.BidAmountCents,
		EstimatedDays: r.EstimatedDays, Status: ProposalStatus(r.Status),
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func mapProposals(rows []proposalRow) []Proposal {
	out := make([]Proposal, len(rows))
	for i, r := range rows {
		out[i] = *fromProposalRow(r)
	}
	return out
}
