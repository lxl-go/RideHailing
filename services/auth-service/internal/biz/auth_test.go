package biz

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ride-hailing/pkg/authx"
)

type fakeAccountRepo struct {
	nextID    int64
	records   map[string]*Account
	userRoles map[int64]map[Role]bool
}

func newFakeAccountRepo() *fakeAccountRepo {
	return &fakeAccountRepo{nextID: 1000, records: map[string]*Account{}, userRoles: map[int64]map[Role]bool{}}
}

func (r *fakeAccountRepo) GetByPrincipalAndRole(_ context.Context, principal string, role Role) (*Account, error) {
	item, ok := r.records[string(role)+":"+principal]
	if !ok {
		return nil, ErrAccountNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakeAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	for _, item := range r.records {
		if item.ID == id {
			copy := *item
			return &copy, nil
		}
	}
	return nil, ErrAccountNotFound
}

func (r *fakeAccountRepo) Create(_ context.Context, account *Account) error {
	r.nextID++
	copy := *account
	copy.ID = r.nextID
	r.records[string(account.Role)+":"+account.Principal] = &copy
	account.ID = copy.ID
	return nil
}

func (r *fakeAccountRepo) UpdateLastLogin(_ context.Context, id int64) error {
	for _, item := range r.records {
		if item.ID == id {
			return nil
		}
	}
	return ErrAccountNotFound
}

func (r *fakeAccountRepo) EnsureUserRole(_ context.Context, userID int64, role Role) error {
	if r.userRoles[userID] == nil {
		r.userRoles[userID] = map[Role]bool{}
	}
	r.userRoles[userID][role] = true
	return nil
}

func (r *fakeAccountRepo) UserHasRole(_ context.Context, userID int64, role Role) (bool, error) {
	return r.userRoles[userID][role], nil
}

type fakeSMSCodeRepo struct {
	nextID     int64
	items      map[int64]*SMSCode
	sendLimits map[string]time.Time
	failures   map[string]int
	locks      map[string]time.Time
}

func newFakeSMSCodeRepo() *fakeSMSCodeRepo {
	return &fakeSMSCodeRepo{
		items:      map[int64]*SMSCode{},
		sendLimits: map[string]time.Time{},
		failures:   map[string]int{},
		locks:      map[string]time.Time{},
	}
}

func (r *fakeSMSCodeRepo) Create(_ context.Context, item *SMSCode) error {
	return r.Store(context.Background(), item, time.Until(item.ExpiresAt))
}

func (r *fakeSMSCodeRepo) Store(_ context.Context, item *SMSCode, _ time.Duration) error {
	r.nextID++
	copy := *item
	copy.ID = r.nextID
	r.items[copy.ID] = &copy
	item.ID = copy.ID
	return nil
}

func (r *fakeSMSCodeRepo) GetLatestActive(_ context.Context, mobile string, role Role, now time.Time) (*SMSCode, error) {
	return r.GetActive(context.Background(), mobile, role, now)
}

func (r *fakeSMSCodeRepo) GetActive(_ context.Context, mobile string, role Role, now time.Time) (*SMSCode, error) {
	var latest *SMSCode
	for _, item := range r.items {
		if item.Mobile == mobile && item.Role == role && item.Status == SMSCodeStatusActive && item.ExpiresAt.After(now) {
			copy := *item
			if latest == nil || copy.ID > latest.ID {
				latest = &copy
			}
		}
	}
	if latest == nil {
		return nil, ErrSMSCodeNotFound
	}
	return latest, nil
}

func (r *fakeSMSCodeRepo) MarkUsed(_ context.Context, id int64) error {
	item, ok := r.items[id]
	if !ok {
		return ErrSMSCodeNotFound
	}
	item.Status = SMSCodeStatusUsed
	return nil
}

func (r *fakeSMSCodeRepo) DeleteActive(_ context.Context, mobile string, role Role) error {
	delete(r.sendLimits, smsCodeTestKey(mobile, role))
	item, err := r.GetActive(context.Background(), mobile, role, time.Unix(0, 0))
	if err != nil {
		return err
	}
	return r.MarkUsed(context.Background(), item.ID)
}

func (r *fakeSMSCodeRepo) ReserveSend(_ context.Context, mobile string, role Role, now time.Time, ttl time.Duration) (bool, error) {
	key := smsCodeTestKey(mobile, role)
	if expiresAt, ok := r.sendLimits[key]; ok && expiresAt.After(now) {
		return false, nil
	}
	r.sendLimits[key] = now.Add(ttl)
	return true, nil
}

func (r *fakeSMSCodeRepo) RecordFailure(_ context.Context, mobile string, role Role, now time.Time, maxAttempts int, lockTTL time.Duration) (bool, error) {
	key := smsCodeTestKey(mobile, role)
	if lockedUntil, ok := r.locks[key]; ok && lockedUntil.After(now) {
		return true, nil
	}
	r.failures[key]++
	if r.failures[key] >= maxAttempts {
		r.locks[key] = now.Add(lockTTL)
		return true, nil
	}
	return false, nil
}

func (r *fakeSMSCodeRepo) IsLocked(_ context.Context, mobile string, role Role, now time.Time) (bool, error) {
	lockedUntil, ok := r.locks[smsCodeTestKey(mobile, role)]
	return ok && lockedUntil.After(now), nil
}

func (r *fakeSMSCodeRepo) ClearFailures(_ context.Context, mobile string, role Role) error {
	delete(r.failures, smsCodeTestKey(mobile, role))
	delete(r.locks, smsCodeTestKey(mobile, role))
	return nil
}

func smsCodeTestKey(mobile string, role Role) string {
	return string(role) + ":" + mobile
}

type fakeSessionRepo struct {
	nextID    int64
	items     map[string]*AuthSession
	createErr error
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{items: map[string]*AuthSession{}}
}

func (r *fakeSessionRepo) Create(_ context.Context, session *AuthSession) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.nextID++
	copy := *session
	copy.ID = r.nextID
	r.items[copy.RefreshTokenHash] = &copy
	session.ID = copy.ID
	return nil
}

func (r *fakeSessionRepo) GetByRefreshTokenHash(_ context.Context, hash string) (*AuthSession, error) {
	item, ok := r.items[hash]
	if !ok {
		return nil, ErrSessionNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakeSessionRepo) Revoke(_ context.Context, id int64) error {
	for _, item := range r.items {
		if item.ID == id {
			item.Status = SessionStatusRevoked
			now := time.Now()
			item.RevokedAt = &now
			return nil
		}
	}
	return ErrSessionNotFound
}

type fakeSMSSender struct {
	mobile string
	code   string
	err    error
	calls  int
}

func (s *fakeSMSSender) SendVerificationCode(_ context.Context, mobile string, code string) error {
	s.calls++
	s.mobile = mobile
	s.code = code
	return s.err
}

type fakePermissionRepo struct {
	allowed map[string]bool
	gotCode string
}

func (r *fakePermissionRepo) CheckUserPermission(_ context.Context, _ int64, permissionCode string) (bool, error) {
	r.gotCode = permissionCode
	return r.allowed[permissionCode], nil
}

func newTestUsecase(accounts *fakeAccountRepo, codes *fakeSMSCodeRepo, sessions *fakeSessionRepo, sender *fakeSMSSender) *AuthUsecase {
	return NewAuthUsecase(zap.NewNop(), accounts, codes, sessions, sender, &fakePermissionRepo{allowed: map[string]bool{}}, authx.NewManager(authx.JWTConfig{
		Secret:        "test-secret",
		Issuer:        "ride-hailing-test",
		ExpireSeconds: 3600,
	}), AuthOptions{
		SMSCodeExpire:        5 * time.Minute,
		RefreshTokenExpire:   7 * 24 * time.Hour,
		RequireSMSCodeVerify: true,
	})
}

func TestSendLoginCodeStoresCodeAndCallsSender(t *testing.T) {
	codes := newFakeSMSCodeRepo()
	sender := &fakeSMSSender{}
	uc := newTestUsecase(newFakeAccountRepo(), codes, newFakeSessionRepo(), sender)
	uc.codeFactory = func() (string, error) { return "123456", nil }

	err := uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RolePassenger})

	require.NoError(t, err)
	require.Equal(t, "13800138000", sender.mobile)
	require.Equal(t, "123456", sender.code)
	require.Len(t, codes.items, 1)
	for _, item := range codes.items {
		require.Equal(t, "123456", item.Code)
	}
}

func TestGenerateSMSCodeReturnsSixDigitNumber(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := generateSMSCode()
		require.NoError(t, err)
		require.Regexp(t, `^\d{6}$`, code)
		value, err := strconv.Atoi(code)
		require.NoError(t, err)
		require.GreaterOrEqual(t, value, 100000)
		require.LessOrEqual(t, value, 999999)
	}
}

func TestSendLoginCodeRejectsRepeatedRequestWithinOneMinute(t *testing.T) {
	codes := newFakeSMSCodeRepo()
	uc := newTestUsecase(newFakeAccountRepo(), codes, newFakeSessionRepo(), &fakeSMSSender{})
	now := time.Unix(100, 0)
	uc.codeNow = func() time.Time { return now }

	require.NoError(t, uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RolePassenger}))
	err := uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RolePassenger})

	require.ErrorIs(t, err, ErrSMSCodeTooFrequent)
	require.Len(t, codes.items, 1)
}

func TestSendLoginCodeClearsCooldownWhenSenderFails(t *testing.T) {
	codes := newFakeSMSCodeRepo()
	sender := &fakeSMSSender{err: errors.New("sms provider unavailable")}
	uc := newTestUsecase(newFakeAccountRepo(), codes, newFakeSessionRepo(), sender)
	now := time.Unix(100, 0)
	uc.codeNow = func() time.Time { return now }

	err := uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RoleDriver})
	require.ErrorContains(t, err, "sms provider unavailable")

	err = uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RoleDriver})

	require.ErrorContains(t, err, "sms provider unavailable")
	require.Equal(t, 2, sender.calls)
}

func TestLoginVerifiesSMSCodeCreatesAccountAndSession(t *testing.T) {
	accounts := newFakeAccountRepo()
	codes := newFakeSMSCodeRepo()
	sessions := newFakeSessionRepo()
	sender := &fakeSMSSender{}
	uc := newTestUsecase(accounts, codes, sessions, sender)
	require.NoError(t, uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RolePassenger}))

	session, err := uc.Login(context.Background(), LoginCommand{
		Principal: "13800138000",
		Role:      RolePassenger,
		Code:      sender.code,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1001), session.UserID)
	require.Equal(t, RolePassenger, session.Role)
	require.NotEmpty(t, session.AccessToken)
	require.NotEmpty(t, session.RefreshToken)
	require.Equal(t, int64(3600), session.ExpiresIn)
	require.Len(t, accounts.records, 1)
	require.True(t, accounts.userRoles[1001][RolePassenger])
	require.Len(t, sessions.items, 1)
	_, err = codes.GetActive(context.Background(), "13800138000", RolePassenger, time.Now())
	require.ErrorIs(t, err, ErrSMSCodeNotFound)
}

func TestLoginKeepsSMSCodeWhenSessionCreationFails(t *testing.T) {
	codes := newFakeSMSCodeRepo()
	sessions := newFakeSessionRepo()
	sessions.createErr = errors.New("session create failed")
	sender := &fakeSMSSender{}
	uc := newTestUsecase(newFakeAccountRepo(), codes, sessions, sender)
	require.NoError(t, uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RolePassenger}))

	_, err := uc.Login(context.Background(), LoginCommand{
		Principal: "13800138000",
		Role:      RolePassenger,
		Code:      sender.code,
	})

	require.ErrorContains(t, err, "session create failed")
	found, err := codes.GetActive(context.Background(), "13800138000", RolePassenger, time.Now())
	require.NoError(t, err)
	require.Equal(t, sender.code, found.Code)
}

func TestLoginLocksAfterThreeInvalidSMSCodesForFiveMinutes(t *testing.T) {
	codes := newFakeSMSCodeRepo()
	uc := newTestUsecase(newFakeAccountRepo(), codes, newFakeSessionRepo(), &fakeSMSSender{})
	sender := uc.smsSender.(*fakeSMSSender)
	uc.opts.SMSCodeExpire = 10 * time.Minute
	now := time.Unix(100, 0)
	uc.codeNow = func() time.Time { return now }
	require.NoError(t, uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RolePassenger}))

	for i := 0; i < 2; i++ {
		_, err := uc.Login(context.Background(), LoginCommand{Principal: "13800138000", Role: RolePassenger, Code: "000000"})
		require.ErrorIs(t, err, ErrInvalidSMSCode)
	}
	_, err := uc.Login(context.Background(), LoginCommand{Principal: "13800138000", Role: RolePassenger, Code: "000000"})
	require.ErrorIs(t, err, ErrSMSLoginLocked)

	_, err = uc.Login(context.Background(), LoginCommand{Principal: "13800138000", Role: RolePassenger, Code: sender.code})
	require.ErrorIs(t, err, ErrSMSLoginLocked)

	now = now.Add(5*time.Minute + time.Second)
	session, err := uc.Login(context.Background(), LoginCommand{Principal: "13800138000", Role: RolePassenger, Code: sender.code})
	require.NoError(t, err)
	require.Equal(t, RolePassenger, session.Role)
}

func TestLoginRejectsReusedSMSCode(t *testing.T) {
	uc := newTestUsecase(newFakeAccountRepo(), newFakeSMSCodeRepo(), newFakeSessionRepo(), &fakeSMSSender{})
	sender := uc.smsSender.(*fakeSMSSender)
	require.NoError(t, uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13800138000", Role: RolePassenger}))
	require.NoError(t, mustLogin(uc, "13800138000", RolePassenger, sender.code))

	err := mustLogin(uc, "13800138000", RolePassenger, sender.code)

	require.True(t, errors.Is(err, ErrInvalidSMSCode), "err = %v", err)
}

func TestRefreshTokenIssuesNewAccessToken(t *testing.T) {
	uc := newTestUsecase(newFakeAccountRepo(), newFakeSMSCodeRepo(), newFakeSessionRepo(), &fakeSMSSender{})
	sender := uc.smsSender.(*fakeSMSSender)
	require.NoError(t, uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13900139000", Role: RoleDriver}))
	session, err := loginForSession(uc, "13900139000", RoleDriver, sender.code)
	require.NoError(t, err)

	refreshed, err := uc.RefreshToken(context.Background(), session.RefreshToken)

	require.NoError(t, err)
	require.Equal(t, session.UserID, refreshed.UserID)
	require.Equal(t, RoleDriver, refreshed.Role)
	require.NotEmpty(t, refreshed.AccessToken)
	require.NotEmpty(t, refreshed.RefreshToken)
}

func TestLogoutRevokesRefreshToken(t *testing.T) {
	uc := newTestUsecase(newFakeAccountRepo(), newFakeSMSCodeRepo(), newFakeSessionRepo(), &fakeSMSSender{})
	sender := uc.smsSender.(*fakeSMSSender)
	require.NoError(t, uc.SendLoginCode(context.Background(), SendLoginCodeCommand{Mobile: "13900139000", Role: RoleDriver}))
	session, err := loginForSession(uc, "13900139000", RoleDriver, sender.code)
	require.NoError(t, err)

	require.NoError(t, uc.Logout(context.Background(), session.RefreshToken))
	_, err = uc.RefreshToken(context.Background(), session.RefreshToken)

	require.True(t, errors.Is(err, ErrSessionRevoked), "err = %v", err)
}

func TestLoginRejectsInvalidRole(t *testing.T) {
	uc := newTestUsecase(newFakeAccountRepo(), newFakeSMSCodeRepo(), newFakeSessionRepo(), &fakeSMSSender{})

	_, err := uc.Login(context.Background(), LoginCommand{
		Principal: "13800138000",
		Role:      Role("admin"),
		Code:      "123456",
	})

	require.True(t, errors.Is(err, ErrInvalidRole), "err = %v", err)
}

func TestVerifyTokenReturnsClaims(t *testing.T) {
	manager := authx.NewManager(authx.JWTConfig{Secret: "test-secret", Issuer: "ride-hailing-test"})
	uc := NewAuthUsecase(zap.NewNop(), newFakeAccountRepo(), newFakeSMSCodeRepo(), newFakeSessionRepo(), &fakeSMSSender{}, &fakePermissionRepo{allowed: map[string]bool{}}, manager, AuthOptions{})
	token, err := manager.Generate(authx.Claims{UserID: "1001", Role: string(RolePassenger)})
	require.NoError(t, err)

	claims, err := uc.VerifyToken(context.Background(), "Bearer "+token.AccessToken)

	require.NoError(t, err)
	require.Equal(t, int64(1001), claims.UserID)
	require.Equal(t, RolePassenger, claims.Role)
}

func TestCheckPermissionDeniesEmptyInput(t *testing.T) {
	uc := newTestUsecase(newFakeAccountRepo(), newFakeSMSCodeRepo(), newFakeSessionRepo(), &fakeSMSSender{})

	allowed, err := uc.CheckPermission(context.Background(), 0, "order:create")
	require.NoError(t, err)
	require.False(t, allowed)

	allowed, err = uc.CheckPermission(context.Background(), 1001, "")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestCheckPermissionDelegatesToPermissionRepo(t *testing.T) {
	permissions := &fakePermissionRepo{allowed: map[string]bool{"order:create": true}}
	uc := NewAuthUsecase(zap.NewNop(), newFakeAccountRepo(), newFakeSMSCodeRepo(), newFakeSessionRepo(), &fakeSMSSender{}, permissions, authx.NewManager(authx.JWTConfig{
		Secret:        "test-secret",
		Issuer:        "ride-hailing-test",
		ExpireSeconds: 3600,
	}), AuthOptions{})

	allowed, err := uc.CheckPermission(context.Background(), 1001, " order:create ")

	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, "order:create", permissions.gotCode)
}

func mustLogin(uc *AuthUsecase, principal string, role Role, code string) error {
	_, err := uc.Login(context.Background(), LoginCommand{Principal: principal, Role: role, Code: code})
	return err
}

func loginForSession(uc *AuthUsecase, principal string, role Role, code string) (*Session, error) {
	return uc.Login(context.Background(), LoginCommand{Principal: principal, Role: role, Code: code})
}
