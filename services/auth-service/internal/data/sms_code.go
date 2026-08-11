package data

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ride-hailing/services/auth-service/internal/biz"
)

type SMSCodeRepo struct {
	rdb *redis.Client
	log *zap.Logger
}

func NewSMSCodeRepo(rdb *redis.Client, log *zap.Logger) *SMSCodeRepo {
	return &SMSCodeRepo{rdb: rdb, log: log}
}

func (r *SMSCodeRepo) ReserveSend(ctx context.Context, mobile string, role biz.Role, _ time.Time, ttl time.Duration) (bool, error) {
	return r.rdb.SetNX(ctx, smsLimitKey(mobile, role), "1", ttl).Result()
}

func (r *SMSCodeRepo) Store(ctx context.Context, item *biz.SMSCode, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Until(item.ExpiresAt)
	}
	if ttl <= 0 {
		return biz.ErrSMSCodeNotFound
	}
	item.Code = strings.TrimSpace(item.Code)
	if !isSixDigitSMSCode(item.Code) {
		return biz.ErrInvalidSMSCode
	}
	return r.rdb.Set(ctx, smsCodeKey(item.Mobile, item.Role), item.Code, ttl).Err()
}

func (r *SMSCodeRepo) GetActive(ctx context.Context, mobile string, role biz.Role, _ time.Time) (*biz.SMSCode, error) {
	key := smsCodeKey(mobile, role)
	code, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, biz.ErrSMSCodeNotFound
		}
		return nil, err
	}
	code = strings.TrimSpace(code)
	if !isSixDigitSMSCode(code) {
		_ = r.rdb.Del(ctx, key).Err()
		r.log.Warn("deleted invalid sms code from redis", zap.String("key", key))
		return nil, biz.ErrSMSCodeNotFound
	}
	ttl, err := r.rdb.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return &biz.SMSCode{
		Mobile:    mobile,
		Role:      role,
		Code:      code,
		Status:    biz.SMSCodeStatusActive,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

func (r *SMSCodeRepo) DeleteActive(ctx context.Context, mobile string, role biz.Role) error {
	return r.rdb.Del(ctx, smsCodeKey(mobile, role), smsLimitKey(mobile, role)).Err()
}

func (r *SMSCodeRepo) IsLocked(ctx context.Context, mobile string, role biz.Role, _ time.Time) (bool, error) {
	count, err := r.rdb.Exists(ctx, smsLockKey(mobile, role)).Result()
	return count > 0, err
}

func (r *SMSCodeRepo) RecordFailure(ctx context.Context, mobile string, role biz.Role, _ time.Time, maxAttempts int, lockTTL time.Duration) (bool, error) {
	key := smsFailKey(mobile, role)
	failures, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if failures == 1 {
		if err := r.rdb.Expire(ctx, key, lockTTL).Err(); err != nil {
			return false, err
		}
	}
	if int(failures) < maxAttempts {
		return false, nil
	}
	if err := r.rdb.Set(ctx, smsLockKey(mobile, role), "1", lockTTL).Err(); err != nil {
		return false, err
	}
	_ = r.rdb.Expire(ctx, key, lockTTL).Err()
	return true, nil
}

func (r *SMSCodeRepo) ClearFailures(ctx context.Context, mobile string, role biz.Role) error {
	return r.rdb.Del(ctx, smsFailKey(mobile, role), smsLockKey(mobile, role)).Err()
}

func smsLimitKey(mobile string, role biz.Role) string {
	return fmt.Sprintf("sms:limit:%s:%s", role, mobile)
}

func smsCodeKey(mobile string, role biz.Role) string {
	return fmt.Sprintf("sms:code:%s:%s", role, mobile)
}

func smsFailKey(mobile string, role biz.Role) string {
	return fmt.Sprintf("sms:fail:%s:%s", role, mobile)
}

func smsLockKey(mobile string, role biz.Role) string {
	return fmt.Sprintf("sms:lock:%s:%s", role, mobile)
}

func isSixDigitSMSCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
