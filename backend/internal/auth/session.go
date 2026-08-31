package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session is a persisted refresh-token session.
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	UserAgent string
	IP        string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

var errSessionNotFound = errors.New("session not found")

type sessionRow struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	TokenHash string
	UserAgent string
	IP        string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (sessionRow) TableName() string { return "auth_sessions" }

type sessionStore struct {
	db *gorm.DB
}

func newSessionStore(db *gorm.DB) *sessionStore { return &sessionStore{db: db} }

func (s *sessionStore) create(ctx context.Context, sess *Session) error {
	rec := sessionRow(*sess)
	return s.db.WithContext(ctx).Create(&rec).Error
}

func (s *sessionStore) getByHash(ctx context.Context, hash string) (*Session, error) {
	var rec sessionRow
	err := s.db.WithContext(ctx).First(&rec, "token_hash = ?", hash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	sess := Session(rec)
	return &sess, nil
}

func (s *sessionStore) revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&sessionRow{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now).Error
}

func (s *sessionStore) revokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&sessionRow{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (sess *Session) active(now time.Time) bool {
	return sess.RevokedAt == nil && now.Before(sess.ExpiresAt)
}
