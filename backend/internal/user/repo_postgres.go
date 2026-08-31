package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// row is the GORM-mapped table representation, kept separate from the domain
// entity so the domain never carries ORM tags.
type row struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email           string
	PasswordHash    string
	DisplayName     string
	AvatarURL       string
	Country         string
	Timezone        string
	IsFreelancer    bool
	IsClient        bool
	IsAdmin         bool
	Status          string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (row) TableName() string { return "users" }

func (r row) toDomain() *User {
	return &User{
		ID:              r.ID,
		Email:           r.Email,
		PasswordHash:    r.PasswordHash,
		DisplayName:     r.DisplayName,
		AvatarURL:       r.AvatarURL,
		Country:         r.Country,
		Timezone:        r.Timezone,
		IsFreelancer:    r.IsFreelancer,
		IsClient:        r.IsClient,
		IsAdmin:         r.IsAdmin,
		Status:          Status(r.Status),
		EmailVerifiedAt: r.EmailVerifiedAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func fromDomain(u *User) row {
	return row{
		ID:              u.ID,
		Email:           u.Email,
		PasswordHash:    u.PasswordHash,
		DisplayName:     u.DisplayName,
		AvatarURL:       u.AvatarURL,
		Country:         u.Country,
		Timezone:        u.Timezone,
		IsFreelancer:    u.IsFreelancer,
		IsClient:        u.IsClient,
		IsAdmin:         u.IsAdmin,
		Status:          string(u.Status),
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// ErrNotFound is returned when a lookup matches no row.
var ErrNotFound = errors.New("user not found")

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore(db *gorm.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Create(ctx context.Context, u *User) error {
	rec := fromDomain(u)
	if err := s.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	*u = *rec.toDomain()
	return nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
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

func (s *PostgresStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var rec row
	err := s.db.WithContext(ctx).First(&rec, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rec.toDomain(), nil
}

func (s *PostgresStore) Update(ctx context.Context, u *User) error {
	u.UpdatedAt = time.Now()
	rec := fromDomain(u)
	return s.db.WithContext(ctx).Model(&row{}).Where("id = ?", u.ID).Save(&rec).Error
}

func (s *PostgresStore) SetAvatarURL(ctx context.Context, id uuid.UUID, url string) error {
	return s.db.WithContext(ctx).Model(&row{}).
		Where("id = ?", id).
		Updates(map[string]any{"avatar_url": url, "updated_at": time.Now()}).Error
}

func (s *PostgresStore) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&row{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
