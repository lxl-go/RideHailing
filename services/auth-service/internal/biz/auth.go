package biz

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"ride-hailing/pkg/authx"
)

const (
	AccountStatusEnabled  = 1
	AccountStatusDisabled = 2
)

type Role string

const (
	RolePassenger Role = "passenger"
	RoleDriver    Role = "driver"
)

type Account struct {
	ID          int64
	Principal   string
	Role        Role
	Status      int
	LastLoginAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	SMSCodeStatusActive = 1
	SMSCodeStatusUsed   = 2
)

type SMSCode struct {
	ID        int64
	Mobile    string
	Role      Role
	Code      string
	Status    int
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	SessionStatusActive  = 1
	SessionStatusRevoked = 2
)

type AuthSession struct {
	ID                int64
	AccountID         int64
	Role              Role
	RefreshTokenHash  string
	Status            int
	RefreshExpireTime time.Time
	RevokedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type AuthOptions struct {
	SMSCodeExpire        time.Duration
	SMSCodeSendCooldown  time.Duration
	SMSCodeMaxAttempts   int
	SMSCodeFailureLock   time.Duration
	RefreshTokenExpire   time.Duration
	RequireSMSCodeVerify bool
}

type SendLoginCodeCommand struct {
	Mobile string
	Role   Role
}

type LoginCommand struct {
	Principal string
	Role      Role
	Code      string
}

type Session struct {
	UserID       int64
	Role         Role
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

type TokenClaims struct {
	UserID int64
	Role   Role
	JTI    string
}

type AuthUsecase struct {
	log         *zap.Logger
	repo        AccountRepo
	smsCodes    SMSCodeRepo
	sessions    SessionRepo
	permissions PermissionRepo
	smsSender   SMSSender
	tokens      *authx.Manager
	opts        AuthOptions
	codeNow     func() time.Time
	codeFactory func() (string, error)
}

func NewAuthUsecase(log *zap.Logger, repo AccountRepo, smsCodes SMSCodeRepo, sessions SessionRepo, smsSender SMSSender, permissions PermissionRepo, tokens *authx.Manager, opts AuthOptions) *AuthUsecase {
	if opts.SMSCodeExpire <= 0 {
		opts.SMSCodeExpire = 5 * time.Minute
	}
	if opts.SMSCodeSendCooldown <= 0 {
		opts.SMSCodeSendCooldown = time.Minute
	}
	if opts.SMSCodeMaxAttempts <= 0 {
		opts.SMSCodeMaxAttempts = 3
	}
	if opts.SMSCodeFailureLock <= 0 {
		opts.SMSCodeFailureLock = 5 * time.Minute
	}
	if opts.RefreshTokenExpire <= 0 {
		opts.RefreshTokenExpire = 7 * 24 * time.Hour
	}
	return &AuthUsecase{
		log:         log,
		repo:        repo,
		smsCodes:    smsCodes,
		sessions:    sessions,
		permissions: permissions,
		smsSender:   smsSender,
		tokens:      tokens,
		opts:        opts,
		codeNow:     time.Now,
		codeFactory: generateSMSCode,
	}
}

func (uc *AuthUsecase) SendLoginCode(ctx context.Context, cmd SendLoginCodeCommand) error {
	mobile := strings.TrimSpace(cmd.Mobile)
	if mobile == "" {
		return ErrInvalidPrincipal
	}
	role := normalizeRole(cmd.Role)
	if !validRole(role) {
		return ErrInvalidRole
	}
	now := uc.codeNow()
	ok, err := uc.smsCodes.ReserveSend(ctx, mobile, role, now, uc.opts.SMSCodeSendCooldown)
	if err != nil {
		return err
	}
	if !ok {
		return ErrSMSCodeTooFrequent
	}
	code, err := uc.codeFactory()
	if err != nil {
		return err
	}
	item := &SMSCode{
		Mobile:    mobile,
		Role:      role,
		Code:      code,
		Status:    SMSCodeStatusActive,
		ExpiresAt: now.Add(uc.opts.SMSCodeExpire),
	}
	if err := uc.smsCodes.Store(ctx, item, uc.opts.SMSCodeExpire); err != nil {
		return err
	}
	if err := uc.smsSender.SendVerificationCode(ctx, mobile, code); err != nil {
		_ = uc.smsCodes.DeleteActive(ctx, mobile, role)
		return fmt.Errorf("%w: %v", ErrSMSSendFailed, err)
	}
	return nil
}

func (uc *AuthUsecase) Login(ctx context.Context, cmd LoginCommand) (*Session, error) {
	principal := strings.TrimSpace(cmd.Principal)
	if principal == "" {
		return nil, ErrInvalidPrincipal
	}
	role := normalizeRole(cmd.Role)
	if !validRole(role) {
		return nil, ErrInvalidRole
	}
	if uc.opts.RequireSMSCodeVerify {
		if err := uc.verifyLoginCode(ctx, principal, role, cmd.Code); err != nil {
			return nil, err
		}
	}

	account, err := uc.repo.GetByPrincipalAndRole(ctx, principal, role)
	if err != nil {
		if !errors.Is(err, ErrAccountNotFound) {
			return nil, err
		}
		account = &Account{
			Principal: principal,
			Role:      role,
			Status:    AccountStatusEnabled,
		}
		if err := uc.repo.Create(ctx, account); err != nil {
			uc.log.Error("create auth account failed", zap.Error(err))
			return nil, err
		}
	}
	if account.Status == AccountStatusDisabled {
		return nil, ErrAccountDisabled
	}
	if err := uc.repo.UpdateLastLogin(ctx, account.ID); err != nil {
		uc.log.Warn("update last login failed", zap.Int64("account_id", account.ID), zap.Error(err))
	}
	if err := uc.repo.EnsureUserRole(ctx, account.ID, account.Role); err != nil {
		return nil, err
	}

	accessToken, err := uc.tokens.Generate(authx.Claims{
		UserID: strconv.FormatInt(account.ID, 10),
		Role:   string(account.Role),
	})
	if err != nil {
		return nil, err
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	session := &AuthSession{
		AccountID:         account.ID,
		Role:              account.Role,
		RefreshTokenHash:  hashRefreshToken(refreshToken),
		Status:            SessionStatusActive,
		RefreshExpireTime: uc.codeNow().Add(uc.opts.RefreshTokenExpire),
	}
	if err := uc.sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	if uc.opts.RequireSMSCodeVerify {
		if err := uc.smsCodes.DeleteActive(ctx, principal, role); err != nil {
			return nil, err
		}
		if err := uc.smsCodes.ClearFailures(ctx, principal, role); err != nil {
			return nil, err
		}
	}
	return &Session{
		UserID:       account.ID,
		Role:         account.Role,
		AccessToken:  accessToken.AccessToken,
		RefreshToken: refreshToken,
		TokenType:    accessToken.TokenType,
		ExpiresIn:    accessToken.ExpiresIn,
	}, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	session, err := uc.activeSession(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	account, err := uc.repo.GetByID(ctx, session.AccountID)
	if err != nil {
		return nil, err
	}
	if account.Status == AccountStatusDisabled {
		return nil, ErrAccountDisabled
	}
	if err := uc.sessions.Revoke(ctx, session.ID); err != nil {
		return nil, err
	}
	accessToken, err := uc.tokens.Generate(authx.Claims{
		UserID: strconv.FormatInt(account.ID, 10),
		Role:   string(account.Role),
	})
	if err != nil {
		return nil, err
	}
	nextRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	nextSession := &AuthSession{
		AccountID:         account.ID,
		Role:              account.Role,
		RefreshTokenHash:  hashRefreshToken(nextRefreshToken),
		Status:            SessionStatusActive,
		RefreshExpireTime: uc.codeNow().Add(uc.opts.RefreshTokenExpire),
	}
	if err := uc.sessions.Create(ctx, nextSession); err != nil {
		return nil, err
	}
	return &Session{
		UserID:       account.ID,
		Role:         account.Role,
		AccessToken:  accessToken.AccessToken,
		RefreshToken: nextRefreshToken,
		TokenType:    accessToken.TokenType,
		ExpiresIn:    accessToken.ExpiresIn,
	}, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	session, err := uc.activeSession(ctx, refreshToken)
	if err != nil {
		return err
	}
	return uc.sessions.Revoke(ctx, session.ID)
}

func (uc *AuthUsecase) VerifyToken(_ context.Context, authorization string) (*TokenClaims, error) {
	claims, err := uc.tokens.ParseBearer(authorization)
	if err != nil {
		return nil, ErrInvalidToken
	}
	userID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil || userID <= 0 {
		return nil, ErrInvalidToken
	}
	role := normalizeRole(Role(claims.Role))
	if !validRole(role) {
		return nil, ErrInvalidToken
	}
	return &TokenClaims{UserID: userID, Role: role, JTI: claims.JTI}, nil
}

func (uc *AuthUsecase) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	permissionCode = strings.TrimSpace(permissionCode)
	if userID <= 0 || permissionCode == "" || uc.permissions == nil {
		return false, nil
	}
	return uc.permissions.CheckUserPermission(ctx, userID, permissionCode)
}

func (uc *AuthUsecase) verifyLoginCode(ctx context.Context, mobile string, role Role, code string) error {
	now := uc.codeNow()
	locked, err := uc.smsCodes.IsLocked(ctx, mobile, role, now)
	if err != nil {
		return err
	}
	if locked {
		return ErrSMSLoginLocked
	}
	item, err := uc.smsCodes.GetActive(ctx, mobile, role, now)
	if err != nil {
		if errors.Is(err, ErrSMSCodeNotFound) {
			return ErrInvalidSMSCode
		}
		return err
	}
	if item.Code != strings.TrimSpace(code) {
		locked, err := uc.smsCodes.RecordFailure(ctx, mobile, role, now, uc.opts.SMSCodeMaxAttempts, uc.opts.SMSCodeFailureLock)
		if err != nil {
			return err
		}
		if locked {
			return ErrSMSLoginLocked
		}
		return ErrInvalidSMSCode
	}
	return nil
}

func (uc *AuthUsecase) activeSession(ctx context.Context, refreshToken string) (*AuthSession, error) {
	session, err := uc.sessions.GetByRefreshTokenHash(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if session.Status == SessionStatusRevoked {
		return nil, ErrSessionRevoked
	}
	if !session.RefreshExpireTime.After(uc.codeNow()) {
		return nil, ErrSessionExpired
	}
	return session, nil
}

func normalizeRole(role Role) Role {
	return Role(strings.ToLower(strings.TrimSpace(string(role))))
}

func validRole(role Role) bool {
	return role == RolePassenger || role == RoleDriver
}

func generateSMSCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(n.Int64()+100000, 10), nil
}

func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return fmt.Sprintf("%x", sum[:])
}
