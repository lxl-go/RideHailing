package biz

import (
	"context"
	"time"
)

type AccountRepo interface {
	GetByPrincipalAndRole(ctx context.Context, principal string, role Role) (*Account, error)
	GetByID(ctx context.Context, id int64) (*Account, error)
	Create(ctx context.Context, account *Account) error
	UpdateLastLogin(ctx context.Context, id int64) error
	EnsureUserRole(ctx context.Context, userID int64, role Role) error
	UserHasRole(ctx context.Context, userID int64, role Role) (bool, error)
}

type SMSCodeRepo interface {
	ReserveSend(ctx context.Context, mobile string, role Role, now time.Time, ttl time.Duration) (bool, error)
	Store(ctx context.Context, item *SMSCode, ttl time.Duration) error
	GetActive(ctx context.Context, mobile string, role Role, now time.Time) (*SMSCode, error)
	DeleteActive(ctx context.Context, mobile string, role Role) error
	IsLocked(ctx context.Context, mobile string, role Role, now time.Time) (bool, error)
	RecordFailure(ctx context.Context, mobile string, role Role, now time.Time, maxAttempts int, lockTTL time.Duration) (bool, error)
	ClearFailures(ctx context.Context, mobile string, role Role) error
}

type SessionRepo interface {
	Create(ctx context.Context, session *AuthSession) error
	GetByRefreshTokenHash(ctx context.Context, hash string) (*AuthSession, error)
	Revoke(ctx context.Context, id int64) error
}

type PermissionRepo interface {
	CheckUserPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
}

type SMSSender interface {
	SendVerificationCode(ctx context.Context, mobile string, code string) error
}
