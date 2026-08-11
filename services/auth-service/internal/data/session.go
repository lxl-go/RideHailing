package data

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/auth-service/internal/biz"
)

type sessionModel struct {
	ID                int64      `gorm:"column:id;primaryKey;autoIncrement"`
	AccountID         int64      `gorm:"column:account_id;not null;index"`
	Role              string     `gorm:"column:role;type:varchar(32);not null"`
	RefreshTokenHash  string     `gorm:"column:refresh_token_hash;type:varchar(128);not null;uniqueIndex"`
	Status            int        `gorm:"column:status;type:tinyint;not null;default:1;index"`
	RefreshExpireTime time.Time  `gorm:"column:refresh_expire_time;not null;index"`
	RevokedAt         *time.Time `gorm:"column:revoked_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (sessionModel) TableName() string {
	return "auth_session"
}

type SessionRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSessionRepo(db *gorm.DB, log *zap.Logger) *SessionRepo {
	return &SessionRepo{db: db, log: log}
}

func (r *SessionRepo) Create(ctx context.Context, session *biz.AuthSession) error {
	record := toSessionRecord(session)
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return err
	}
	session.ID = record.ID
	session.CreatedAt = record.CreatedAt
	session.UpdatedAt = record.UpdatedAt
	return nil
}

func (r *SessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*biz.AuthSession, error) {
	var m sessionModel
	if err := r.db.WithContext(ctx).Where("refresh_token_hash = ?", hash).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrSessionNotFound
		}
		return nil, err
	}
	return toSessionDomain(&m), nil
}

func (r *SessionRepo) Revoke(ctx context.Context, id int64) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&sessionModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":     biz.SessionStatusRevoked,
		"revoked_at": now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrSessionNotFound
	}
	return nil
}

func toSessionDomain(m *sessionModel) *biz.AuthSession {
	return &biz.AuthSession{
		ID:                m.ID,
		AccountID:         m.AccountID,
		Role:              biz.Role(m.Role),
		RefreshTokenHash:  m.RefreshTokenHash,
		Status:            m.Status,
		RefreshExpireTime: m.RefreshExpireTime,
		RevokedAt:         m.RevokedAt,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func toSessionRecord(e *biz.AuthSession) *sessionModel {
	return &sessionModel{
		ID:                e.ID,
		AccountID:         e.AccountID,
		Role:              string(e.Role),
		RefreshTokenHash:  e.RefreshTokenHash,
		Status:            e.Status,
		RefreshExpireTime: e.RefreshExpireTime,
		RevokedAt:         e.RevokedAt,
	}
}
