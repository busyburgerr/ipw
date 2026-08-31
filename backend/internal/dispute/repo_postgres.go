package dispute

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("dispute not found")

type row struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ContractID     uuid.UUID  `gorm:"type:uuid"`
	MilestoneID    *uuid.UUID `gorm:"type:uuid"`
	RaisedBy       uuid.UUID  `gorm:"type:uuid"`
	Reason         string
	Status         string
	ResolutionNote string
	ArbiterID      *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time
	ResolvedAt     *time.Time
}

func (row) TableName() string { return "disputes" }

func (r row) toDomain() *Dispute {
	return &Dispute{
		ID: r.ID, ContractID: r.ContractID, MilestoneID: r.MilestoneID, RaisedBy: r.RaisedBy,
		Reason: r.Reason, Status: Status(r.Status), ResolutionNote: r.ResolutionNote,
		ArbiterID: r.ArbiterID, CreatedAt: r.CreatedAt, ResolvedAt: r.ResolvedAt,
	}
}

type PostgresStore struct{ db *gorm.DB }

func NewPostgresStore(db *gorm.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) Create(ctx context.Context, d *Dispute) error {
	rec := row{
		ID: d.ID, ContractID: d.ContractID, MilestoneID: d.MilestoneID, RaisedBy: d.RaisedBy,
		Reason: d.Reason, Status: string(d.Status), CreatedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	*d = *rec.toDomain()
	return nil
}

func (s *PostgresStore) Get(ctx context.Context, id uuid.UUID) (*Dispute, error) {
	var rec row
	err := s.db.WithContext(ctx).First(&rec, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec.toDomain(), nil
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]Dispute, error) {
	var rows []row
	err := s.db.WithContext(ctx).
		Joins("JOIN contracts c ON c.id = disputes.contract_id").
		Where("disputes.raised_by = ? OR c.client_id = ? OR c.freelancer_id = ?", userID, userID, userID).
		Order("disputes.created_at DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows), nil
}

func (s *PostgresStore) ListByStatus(ctx context.Context, statuses ...Status) ([]Dispute, error) {
	strs := make([]string, len(statuses))
	for i, st := range statuses {
		strs[i] = string(st)
	}
	var rows []row
	err := s.db.WithContext(ctx).Where("status IN ?", strs).
		Order("created_at").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows), nil
}

func (s *PostgresStore) Update(ctx context.Context, d *Dispute) error {
	return s.db.WithContext(ctx).Model(&row{}).Where("id = ?", d.ID).Updates(map[string]any{
		"status":          string(d.Status),
		"resolution_note": d.ResolutionNote,
		"arbiter_id":      d.ArbiterID,
		"resolved_at":     d.ResolvedAt,
	}).Error
}

func mapRows(rows []row) []Dispute {
	out := make([]Dispute, len(rows))
	for i, r := range rows {
		out[i] = *r.toDomain()
	}
	return out
}
