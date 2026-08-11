package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lock 分布式锁（对标文档 4.6 节 Redis 规范）
type Lock struct {
	client     *redis.Client
	key        string
	token      string
	ttl        time.Duration
	autoExtend bool
	stopCh     chan struct{}
}

// NewLock 创建分布式锁
// key 格式：lock:{domain}:{resource}（对标文档 Key 命名规范）
func NewLock(client *redis.Client, domain, resource string, ttl time.Duration) *Lock {
	return &Lock{
		client: client,
		key:    fmt.Sprintf("lock:%s:%s", domain, resource),
		token:  fmt.Sprintf("%d", time.Now().UnixNano()),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
}

// Acquire 获取锁（对标文档分布式锁实现）
func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	ok, err := l.client.SetNX(ctx, l.key, l.token, l.ttl).Result()
	if err != nil {
		return false, err
	}
	if ok && l.autoExtend {
		go l.autoExtension(ctx)
	}
	return ok, nil
}

// Release 释放锁
func (l *Lock) Release(ctx context.Context) error {
	close(l.stopCh)
	// Lua 脚本：原子性检查并删除
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	end
	return 0
	`
	return l.client.Eval(ctx, script, []string{l.key}, l.token).Err()
}

func (l *Lock) autoExtension(ctx context.Context) {
	ticker := time.NewTicker(l.ttl / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.client.Expire(ctx, l.key, l.ttl).Result()
		case <-l.stopCh:
			return
		}
	}
}

// WithAutoExtend 启用自动续期（对标文档"锁 TTL = 预估耗时 × 2 + 续期"）
func (l *Lock) WithAutoExtend() *Lock {
	l.autoExtend = true
	return l
}

// SmsLimit 短信限流检查（对标文档 4.6 节限流实现）
func SmsLimit(ctx context.Context, client *redis.Client, phone string) (bool, error) {
	key := fmt.Sprintf("sms:limit:%s", phone)
	return client.SetNX(ctx, key, "1", time.Minute).Result()
}

// StoreVerificationCode 存储验证码（对标文档 4.6 节）
func StoreVerificationCode(ctx context.Context, client *redis.Client, phone, code string, ttl time.Duration) error {
	key := fmt.Sprintf("sms:code:%s", phone)
	return client.Set(ctx, key, code, ttl).Err()
}

// VerifyCode 校验验证码（校验后删除，防重复使用）
func VerifyCode(ctx context.Context, client *redis.Client, phone, code string) bool {
	key := fmt.Sprintf("sms:code:%s", phone)
	val, err := client.Get(ctx, key).Result()
	if err != nil || val != code {
		return false
	}
	client.Del(ctx, key)
	return true
}
