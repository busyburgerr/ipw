// Package catalog owns the service taxonomy: a category tree and the skills
// attached to it. Categories and skills are curated data — clients read them,
// only admins (or seeds/migrations) write them.
package catalog

import (
	"context"

	"github.com/google/uuid"
)

type Category struct {
	ID       uuid.UUID
	ParentID *uuid.UUID
	Slug     string
	Name     string
	Position int
}

type Skill struct {
	ID         uuid.UUID
	CategoryID *uuid.UUID
	Slug       string
	Name       string
}

// Store is the read contract consumed by other features (profiles, projects)
// to validate and resolve category/skill references.
type Store interface {
	ListCategories(ctx context.Context) ([]Category, error)
	ListSkills(ctx context.Context, categorySlug string) ([]Skill, error)
	SkillsByIDs(ctx context.Context, ids []uuid.UUID) ([]Skill, error)
	CategoryExists(ctx context.Context, id uuid.UUID) (bool, error)
	SkillIDsExist(ctx context.Context, ids []uuid.UUID) (bool, error)
}
