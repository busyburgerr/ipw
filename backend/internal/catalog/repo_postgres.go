package catalog

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type categoryRow struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ParentID *uuid.UUID `gorm:"type:uuid"`
	Slug     string
	Name     string
	Position int
}

func (categoryRow) TableName() string { return "categories" }

type skillRow struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CategoryID *uuid.UUID `gorm:"type:uuid"`
	Slug       string
	Name       string
}

func (skillRow) TableName() string { return "skills" }

type PostgresStore struct{ db *gorm.DB }

func NewPostgresStore(db *gorm.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) ListCategories(ctx context.Context) ([]Category, error) {
	var rows []categoryRow
	err := s.db.WithContext(ctx).Order("position, name").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Category, len(rows))
	for i, r := range rows {
		out[i] = Category{ID: r.ID, ParentID: r.ParentID, Slug: r.Slug, Name: r.Name, Position: r.Position}
	}
	return out, nil
}

func (s *PostgresStore) ListSkills(ctx context.Context, categorySlug string) ([]Skill, error) {
	q := s.db.WithContext(ctx).Model(&skillRow{}).Order("skills.name")
	if categorySlug != "" {
		q = q.Joins("JOIN categories ON categories.id = skills.category_id").
			Where("categories.slug = ?", categorySlug)
	}
	var rows []skillRow
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return toSkills(rows), nil
}

func (s *PostgresStore) SkillsByIDs(ctx context.Context, ids []uuid.UUID) ([]Skill, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []skillRow
	err := s.db.WithContext(ctx).Where("id IN ?", ids).Order("name").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toSkills(rows), nil
}

func (s *PostgresStore) CategoryExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&categoryRow{}).Where("id = ?", id).Count(&n).Error
	return n > 0, err
}

func (s *PostgresStore) SkillIDsExist(ctx context.Context, ids []uuid.UUID) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	var n int64
	err := s.db.WithContext(ctx).Model(&skillRow{}).Where("id IN ?", ids).Count(&n).Error
	return int(n) == len(ids), err
}

// UpsertCategory inserts or updates a category by slug and returns its ID. Used
// by Seed.
func (s *PostgresStore) UpsertCategory(ctx context.Context, c Category) (uuid.UUID, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	row := categoryRow{ID: c.ID, ParentID: c.ParentID, Slug: c.Slug, Name: c.Name, Position: c.Position}
	err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "parent_id", "position"}),
	}).Create(&row).Error
	if err != nil {
		return uuid.Nil, err
	}
	var saved categoryRow
	if err := s.db.WithContext(ctx).Select("id").First(&saved, "slug = ?", c.Slug).Error; err != nil {
		return uuid.Nil, err
	}
	return saved.ID, nil
}

// UpsertSkill inserts or updates a skill by slug. Used by Seed.
func (s *PostgresStore) UpsertSkill(ctx context.Context, sk Skill) error {
	if sk.ID == uuid.Nil {
		sk.ID = uuid.New()
	}
	row := skillRow{ID: sk.ID, CategoryID: sk.CategoryID, Slug: sk.Slug, Name: sk.Name}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "category_id"}),
	}).Create(&row).Error
}

func toSkills(rows []skillRow) []Skill {
	out := make([]Skill, len(rows))
	for i, r := range rows {
		out[i] = Skill{ID: r.ID, CategoryID: r.CategoryID, Slug: r.Slug, Name: r.Name}
	}
	return out
}
