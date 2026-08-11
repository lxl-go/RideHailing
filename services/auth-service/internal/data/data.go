package data

import (
	"context"
	"fmt"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ride-hailing/pkg/authx"
	"ride-hailing/pkg/smsx"
	"ride-hailing/services/auth-service/internal/biz"
	"ride-hailing/services/auth-service/internal/conf"
)

type Data struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewData(db *gorm.DB, rdb *redis.Client) *Data {
	return &Data{DB: db, Redis: rdb}
}

func NewDB(c *conf.Data, logger *zap.Logger) (*gorm.DB, error) {
	if c == nil || c.Database == nil {
		return nil, fmt.Errorf("database config is required")
	}
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&accountModel{}, &roleModel{}, &userRoleModel{}, &permissionModel{}, &rolePermissionModel{}, &sessionModel{}); err != nil {
		return nil, err
	}
	if err := SeedBuiltinRBAC(context.Background(), db); err != nil {
		return nil, err
	}
	return db, nil
}

func NewRedis(c *conf.Data) (*redis.Client, error) {
	addr := "127.0.0.1:6379"
	password := ""
	db := 0
	if c != nil && c.Redis != nil {
		if c.Redis.Addr != "" {
			addr = c.Redis.Addr
		}
		password = c.Redis.Password
		db = c.Redis.DB
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return rdb, nil
}

func NewTokenManager(c *conf.Auth) *authx.Manager {
	cfg := authx.JWTConfig{}
	if c != nil && c.Jwt != nil {
		cfg.Enabled = c.Jwt.Enabled
		cfg.Secret = c.Jwt.Secret
		cfg.Issuer = c.Jwt.Issuer
		cfg.ExpireSeconds = c.Jwt.ExpireSeconds
	}
	return authx.NewManager(cfg)
}

func NewSMSSender(c *conf.Auth) biz.SMSSender {
	cfg := smsx.IhuyiConfig{}
	if c != nil && c.Sms != nil && c.Sms.Ihuyi != nil {
		cfg.Account = c.Sms.Ihuyi.Account
		cfg.Password = c.Sms.Ihuyi.Password
		cfg.Endpoint = c.Sms.Ihuyi.Endpoint
	}
	return smsx.NewIhuyiClient(cfg)
}

func NewAuthOptions(c *conf.Auth) biz.AuthOptions {
	opts := biz.AuthOptions{RequireSMSCodeVerify: true}
	if c != nil {
		if c.RequireSmsCodeVerify != nil {
			opts.RequireSMSCodeVerify = *c.RequireSmsCodeVerify
		}
		if c.Sms != nil && c.Sms.CodeExpireSeconds > 0 {
			opts.SMSCodeExpire = time.Duration(c.Sms.CodeExpireSeconds) * time.Second
		}
		if c.Sms != nil && c.Sms.SendCooldownSeconds > 0 {
			opts.SMSCodeSendCooldown = time.Duration(c.Sms.SendCooldownSeconds) * time.Second
		}
		if c.Sms != nil && c.Sms.MaxVerifyAttempts > 0 {
			opts.SMSCodeMaxAttempts = c.Sms.MaxVerifyAttempts
		}
		if c.Sms != nil && c.Sms.VerifyLockSeconds > 0 {
			opts.SMSCodeFailureLock = time.Duration(c.Sms.VerifyLockSeconds) * time.Second
		}
		if c.RefreshTokenExpireSeconds > 0 {
			opts.RefreshTokenExpire = time.Duration(c.RefreshTokenExpireSeconds) * time.Second
		}
	}
	return opts
}

var ProviderSet = wire.NewSet(
	NewDB,
	NewRedis,
	NewData,
	NewTokenManager,
	NewSMSSender,
	NewAuthOptions,
	NewAccountRepo,
	NewRBACRepo,
	NewSMSCodeRepo,
	NewSessionRepo,
	wire.Bind(new(biz.AccountRepo), new(*AccountRepo)),
	wire.Bind(new(biz.PermissionRepo), new(*RBACRepo)),
	wire.Bind(new(biz.SMSCodeRepo), new(*SMSCodeRepo)),
	wire.Bind(new(biz.SessionRepo), new(*SessionRepo)),
)
