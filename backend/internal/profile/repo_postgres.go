package profile

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned when a profile row does not exist yet.
var ErrNotFound = errors.New("profile not found")

type freelancerRow struct {
	UserID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Headline          string
	Bio               string
	HourlyRateCents   int64
	Currency          string
	Availability      string
	PrimaryCategoryID *uuid.UUID `gorm:"type:uuid"`
	Languages         []byte     `gorm:"type:jsonb"`
	Location          string
	RatingAvg         float64
	RatingCount       int
	JobsCompleted     int
	TotalEarnedCents  int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (freelancerRow) TableName() string { return "freelancer_profiles" }

type clientRow struct {
	UserID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyName     string
	About           string
	Website         string
	Location        string
	PaymentVerified bool
	RatingAvg       float64
	RatingCount     int
	HiresCount      int
	TotalSpentCents int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (clientRow) TableName() string { return "client_profiles" }

type portfolioRow struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	Title       string
	Description string
	URL         string
	ImageKey    string
	Position    int
	CreatedAt   time.Time
}

func (portfolioRow) TableName() string { return "portfolio_items" }

type freelancerSkillRow struct {
	UserID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	SkillID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (freelancerSkillRow) TableName() string { return "freelancer_skills" }

type PostgresStore struct{ db *gorm.DB }

func NewPostgresStore(db *gorm.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) GetFreelancer(ctx context.Context, userID uuid.UUID) (*Freelancer, error) {
	var row freelancerRow
	err := s.db.WithContext(ctx).First(&row, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var skillIDs []uuid.UUID
	if err := s.db.WithContext(ctx).Model(&freelancerSkillRow{}).
		Where("user_id = ?", userID).Pluck("skill_id", &skillIDs).Error; err != nil {
		return nil, err
	}

	langs := []string{}
	if len(row.Languages) > 0 {
		_ = json.Unmarshal(row.Languages, &langs)
	}

	return &Freelancer{
		UserID:            row.UserID,
		Headline:          row.Headline,
		Bio:               row.Bio,
		HourlyRateCents:   row.HourlyRateCents,
		Currency:          row.Currency,
		Availability:      Availability(row.Availability),
		PrimaryCategoryID: row.PrimaryCategoryID,
		Languages:         langs,
		Location:          row.Location,
		SkillIDs:          skillIDs,
		RatingAvg:         row.RatingAvg,
		RatingCount:       row.RatingCount,
		JobsCompleted:     row.JobsCompleted,
		TotalEarnedCents:  row.TotalEarnedCents,
	}, nil
}

// UpsertFreelancer writes the editable fields only; reputation aggregates are
// owned by other features and never touched here.
func (s *PostgresStore) UpsertFreelancer(ctx context.Context, f *Freelancer) error {
	langs, err := json.Marshal(f.Languages)
	if err != nil {
		return err
	}
	now := time.Now()
	row := freelancerRow{
		UserID:            f.UserID,
		Headline:          f.Headline,
		Bio:               f.Bio,
		HourlyRateCents:   f.HourlyRateCents,
		Currency:          f.Currency,
		Availability:      string(f.Availability),
		PrimaryCategoryID: f.PrimaryCategoryID,
		Languages:         langs,
		Location:          f.Location,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"headline", "bio", "hourly_rate_cents", "currency", "availability",
			"primary_category_id", "languages", "location", "updated_at",
		}),
	}).Create(&row).Error
}

func (s *PostgresStore) SetFreelancerSkills(ctx context.Context, userID uuid.UUID, skillIDs []uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&freelancerSkillRow{}).Error; err != nil {
			return err
		}
		if len(skillIDs) == 0 {
			return nil
		}
		rows := make([]freelancerSkillRow, len(skillIDs))
		for i, id := range skillIDs {
			rows[i] = freelancerSkillRow{UserID: userID, SkillID: id}
		}
		return tx.Create(&rows).Error
	})
}

func (s *PostgresStore) GetClient(ctx context.Context, userID uuid.UUID) (*Client, error) {
	var row clientRow
	err := s.db.WithContext(ctx).First(&row, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &Client{
		UserID:          row.UserID,
		CompanyName:     row.CompanyName,
		About:           row.About,
		Website:         row.Website,
		Location:        row.Location,
		PaymentVerified: row.PaymentVerified,
		RatingAvg:       row.RatingAvg,
		RatingCount:     row.RatingCount,
		HiresCount:      row.HiresCount,
		TotalSpentCents: row.TotalSpentCents,
	}, nil
}

func (s *PostgresStore) UpsertClient(ctx context.Context, c *Client) error {
	now := time.Now()
	row := clientRow{
		UserID:      c.UserID,
		CompanyName: c.CompanyName,
		About:       c.About,
		Website:     c.Website,
		Location:    c.Location,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"company_name", "about", "website", "location", "updated_at",
		}),
	}).Create(&row).Error
}

func (s *PostgresStore) AddPortfolioItem(ctx context.Context, item *PortfolioItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	row := portfolioRow{
		ID:          item.ID,
		UserID:      item.UserID,
		Title:       item.Title,
		Description: item.Description,
		URL:         item.URL,
		ImageKey:    item.ImageKey,
		Position:    item.Position,
		CreatedAt:   time.Now(),
	}
	return s.db.WithContext(ctx).Create(&row).Error
}

func (s *PostgresStore) ListPortfolio(ctx context.Context, userID uuid.UUID) ([]PortfolioItem, error) {
	var rows []portfolioRow
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("position, created_at").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]PortfolioItem, len(rows))
	for i, r := range rows {
		out[i] = PortfolioItem{
			ID: r.ID, UserID: r.UserID, Title: r.Title, Description: r.Description,
			URL: r.URL, ImageKey: r.ImageKey, Position: r.Position,
		}
	}
	return out, nil
}

func (s *PostgresStore) DeletePortfolioItem(ctx context.Context, userID, itemID uuid.UUID) error {
	res := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", itemID, userID).Delete(&portfolioRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
