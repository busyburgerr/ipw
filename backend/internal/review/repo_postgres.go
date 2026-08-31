package review

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("review not found")

type row struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ContractID  uuid.UUID `gorm:"type:uuid"`
	ReviewerID  uuid.UUID `gorm:"type:uuid"`
	RevieweeID  uuid.UUID `gorm:"type:uuid"`
	Direction   string
	Rating      int
	Comment     string
	PublishedAt *time.Time
	CreatedAt   time.Time
}

func (row) TableName() string { return "reviews" }

func (r row) toDomain() *Review {
	return &Review{
		ID: r.ID, ContractID: r.ContractID, ReviewerID: r.ReviewerID, RevieweeID: r.RevieweeID,
		Direction: Direction(r.Direction), Rating: r.Rating, Comment: r.Comment,
		PublishedAt: r.PublishedAt, CreatedAt: r.CreatedAt,
	}
}

type PostgresStore struct{ db *gorm.DB }

func NewPostgresStore(db *gorm.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) Create(ctx context.Context, r *Review) error {
	rec := row{
		ID: r.ID, ContractID: r.ContractID, ReviewerID: r.ReviewerID, RevieweeID: r.RevieweeID,
		Direction: string(r.Direction), Rating: r.Rating, Comment: r.Comment, CreatedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	*r = *rec.toDomain()
	return nil
}

func (s *PostgresStore) ByContractAndReviewer(ctx context.Context, contractID, reviewerID uuid.UUID) (*Review, error) {
	var rec row
	err := s.db.WithContext(ctx).First(&rec, "contract_id = ? AND reviewer_id = ?", contractID, reviewerID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec.toDomain(), nil
}

func (s *PostgresStore) ListByContract(ctx context.Context, contractID uuid.UUID) ([]Review, error) {
	var rows []row
	if err := s.db.WithContext(ctx).Where("contract_id = ?", contractID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows), nil
}

func (s *PostgresStore) ListPublishedForReviewee(ctx context.Context, revieweeID uuid.UUID) ([]Review, error) {
	var rows []row
	err := s.db.WithContext(ctx).
		Where("reviewee_id = ? AND published_at IS NOT NULL", revieweeID).
		Order("published_at DESC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows), nil
}

func (s *PostgresStore) PublishForContract(ctx context.Context, contractID uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).Model(&row{}).
		Where("contract_id = ? AND published_at IS NULL", contractID).
		Update("published_at", at).Error
}

func (s *PostgresStore) RevieweeRatingSummary(ctx context.Context, revieweeID uuid.UUID) (float64, int, error) {
	var res struct {
		Avg   float64
		Count int
	}
	err := s.db.WithContext(ctx).Model(&row{}).
		Where("reviewee_id = ? AND published_at IS NOT NULL", revieweeID).
		Select("COALESCE(AVG(rating), 0) AS avg, COUNT(*) AS count").Scan(&res).Error
	return res.Avg, res.Count, err
}

func (s *PostgresStore) UnpublishedOlderThan(ctx context.Context, cutoff time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := s.db.WithContext(ctx).Model(&row{}).
		Where("published_at IS NULL AND created_at < ?", cutoff).
		Distinct().Pluck("contract_id", &ids).Error
	return ids, err
}

func mapRows(rows []row) []Review {
	out := make([]Review, len(rows))
	for i, r := range rows {
		out[i] = *r.toDomain()
	}
	return out
}
