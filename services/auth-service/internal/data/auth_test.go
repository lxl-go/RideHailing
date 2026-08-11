package data

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ride-hailing/services/auth-service/internal/biz"
)

func newAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&accountModel{}, &roleModel{}, &userRoleModel{}, &permissionModel{}, &rolePermissionModel{}))
	return db
}

func TestAccountRepoCreateGetAndUpdateLastLogin(t *testing.T) {
	db := newAuthTestDB(t)

	repo := NewAccountRepo(db, zap.NewNop())
	account := &biz.Account{
		Principal: "13800138000",
		Role:      biz.RolePassenger,
		Status:    biz.AccountStatusEnabled,
	}
	require.NoError(t, repo.Create(context.Background(), account))
	require.NotZero(t, account.ID)
	require.False(t, account.LastLoginAt.IsZero())

	found, err := repo.GetByPrincipalAndRole(context.Background(), "13800138000", biz.RolePassenger)
	require.NoError(t, err)
	require.Equal(t, account.ID, found.ID)
	require.Equal(t, biz.RolePassenger, found.Role)
	require.False(t, found.LastLoginAt.IsZero())

	require.NoError(t, repo.UpdateLastLogin(context.Background(), account.ID))
	updated, err := repo.GetByPrincipalAndRole(context.Background(), "13800138000", biz.RolePassenger)
	require.NoError(t, err)
	require.False(t, updated.LastLoginAt.IsZero())

	require.NoError(t, repo.EnsureUserRole(context.Background(), account.ID, biz.RolePassenger))
	hasRole, err := repo.UserHasRole(context.Background(), account.ID, biz.RolePassenger)
	require.NoError(t, err)
	require.True(t, hasRole)
}

func TestRBACModelsUseFiveStandardTables(t *testing.T) {
	require.Equal(t, "auth_user", accountModel{}.TableName())
	require.Equal(t, "auth_role", roleModel{}.TableName())
	require.Equal(t, "auth_user_role", userRoleModel{}.TableName())
	require.Equal(t, "auth_permission", permissionModel{}.TableName())
	require.Equal(t, "auth_role_permission", rolePermissionModel{}.TableName())
}

func TestSeedBuiltinRBACIsIdempotent(t *testing.T) {
	db := newAuthTestDB(t)
	require.NoError(t, SeedBuiltinRBAC(context.Background(), db))
	require.NoError(t, SeedBuiltinRBAC(context.Background(), db))

	var roleCount int64
	require.NoError(t, db.Model(&roleModel{}).Where("code IN ?", []string{"passenger", "driver"}).Count(&roleCount).Error)
	require.Equal(t, int64(2), roleCount)

	var permissionCount int64
	require.NoError(t, db.Model(&permissionModel{}).Where("code = ?", "order:create").Count(&permissionCount).Error)
	require.Equal(t, int64(1), permissionCount)
}

func TestCheckUserPermissionUsesRolePermissionBindings(t *testing.T) {
	db := newAuthTestDB(t)
	require.NoError(t, SeedBuiltinRBAC(context.Background(), db))

	accountRepo := NewAccountRepo(db, zap.NewNop())
	account := &biz.Account{
		Principal: "13800138000",
		Role:      biz.RolePassenger,
		Status:    biz.AccountStatusEnabled,
	}
	require.NoError(t, accountRepo.Create(context.Background(), account))
	require.NoError(t, accountRepo.EnsureUserRole(context.Background(), account.ID, biz.RolePassenger))

	rbacRepo := NewRBACRepo(db)
	allowed, err := rbacRepo.CheckUserPermission(context.Background(), account.ID, "order:create")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = rbacRepo.CheckUserPermission(context.Background(), account.ID, "trip:publish")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestSharedTripReadPermissionsAreGrantedToPassengerAndDriver(t *testing.T) {
	db := newAuthTestDB(t)
	require.NoError(t, SeedBuiltinRBAC(context.Background(), db))

	accountRepo := NewAccountRepo(db, zap.NewNop())
	passenger := &biz.Account{Principal: "13800138001", Role: biz.RolePassenger, Status: biz.AccountStatusEnabled}
	driver := &biz.Account{Principal: "13800138002", Role: biz.RoleDriver, Status: biz.AccountStatusEnabled}
	require.NoError(t, accountRepo.Create(context.Background(), passenger))
	require.NoError(t, accountRepo.Create(context.Background(), driver))
	require.NoError(t, accountRepo.EnsureUserRole(context.Background(), passenger.ID, biz.RolePassenger))
	require.NoError(t, accountRepo.EnsureUserRole(context.Background(), driver.ID, biz.RoleDriver))

	rbacRepo := NewRBACRepo(db)
	for _, userID := range []int64{passenger.ID, driver.ID} {
		allowed, err := rbacRepo.CheckUserPermission(context.Background(), userID, "trip:search")
		require.NoError(t, err)
		require.True(t, allowed)

		allowed, err = rbacRepo.CheckUserPermission(context.Background(), userID, "trip:view_detail")
		require.NoError(t, err)
		require.True(t, allowed)
	}
}

func TestSMSCodeRepoUsesRedisForLimitCodeAndFailureLock(t *testing.T) {
	rdb := newRedisTestClient(t)
	ctx := context.Background()
	repo := NewSMSCodeRepo(rdb, zap.NewNop())
	mobile := "13800138099"
	role := biz.RolePassenger
	require.NoError(t, rdb.Del(ctx, smsLimitKey(mobile, role), smsCodeKey(mobile, role), smsFailKey(mobile, role), smsLockKey(mobile, role)).Err())
	t.Cleanup(func() {
		_ = rdb.Del(ctx, smsLimitKey(mobile, role), smsCodeKey(mobile, role), smsFailKey(mobile, role), smsLockKey(mobile, role)).Err()
	})

	ok, err := repo.ReserveSend(ctx, mobile, role, time.Now(), time.Minute)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = repo.ReserveSend(ctx, mobile, role, time.Now(), time.Minute)
	require.NoError(t, err)
	require.False(t, ok)

	item := &biz.SMSCode{
		Mobile:    mobile,
		Role:      role,
		Code:      "123456",
		Status:    biz.SMSCodeStatusActive,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	require.NoError(t, repo.Store(ctx, item, 5*time.Minute))
	storedCode, err := rdb.Get(ctx, smsCodeKey(mobile, role)).Result()
	require.NoError(t, err)
	require.Equal(t, "123456", storedCode)
	found, err := repo.GetActive(ctx, mobile, role, time.Now())
	require.NoError(t, err)
	require.Equal(t, "123456", found.Code)

	locked, err := repo.RecordFailure(ctx, mobile, role, time.Now(), 3, 5*time.Minute)
	require.NoError(t, err)
	require.False(t, locked)
	locked, err = repo.RecordFailure(ctx, mobile, role, time.Now(), 3, 5*time.Minute)
	require.NoError(t, err)
	require.False(t, locked)
	locked, err = repo.RecordFailure(ctx, mobile, role, time.Now(), 3, 5*time.Minute)
	require.NoError(t, err)
	require.True(t, locked)

	locked, err = repo.IsLocked(ctx, mobile, role, time.Now())
	require.NoError(t, err)
	require.True(t, locked)
	require.NoError(t, repo.DeleteActive(ctx, mobile, role))
	exists, err := rdb.Exists(ctx, smsCodeKey(mobile, role)).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), exists)
	exists, err = rdb.Exists(ctx, smsLimitKey(mobile, role)).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), exists)
	_, err = repo.GetActive(ctx, mobile, role, time.Now())
	require.ErrorIs(t, err, biz.ErrSMSCodeNotFound)
	require.NoError(t, repo.ClearFailures(ctx, mobile, role))
	locked, err = repo.IsLocked(ctx, mobile, role, time.Now())
	require.NoError(t, err)
	require.False(t, locked)
}

func TestSMSCodeRepoRejectsNonSixDigitCode(t *testing.T) {
	rdb := newRedisTestClient(t)
	ctx := context.Background()
	repo := NewSMSCodeRepo(rdb, zap.NewNop())
	mobile := "13800138100"
	role := biz.RoleDriver
	key := smsCodeKey(mobile, role)
	require.NoError(t, rdb.Del(ctx, key).Err())
	t.Cleanup(func() {
		_ = rdb.Del(ctx, key).Err()
	})

	err := repo.Store(ctx, &biz.SMSCode{
		Mobile:    mobile,
		Role:      role,
		Code:      "c64d45097e73b29e515193c62229ee6944728dcaa9fa910152fb8ee630e871a3",
		Status:    biz.SMSCodeStatusActive,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}, 5*time.Minute)

	require.ErrorIs(t, err, biz.ErrInvalidSMSCode)
	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), exists)
}

func TestSMSCodeRepoDeletesStoredNonSixDigitCodeOnRead(t *testing.T) {
	rdb := newRedisTestClient(t)
	ctx := context.Background()
	repo := NewSMSCodeRepo(rdb, zap.NewNop())
	mobile := "13800138101"
	role := biz.RoleDriver
	key := smsCodeKey(mobile, role)
	require.NoError(t, rdb.Set(ctx, key, "7fab8cdeb9f69308f06aca0fdb78546d8f6467ed51381fcd5f89ad738a8eab1f", 5*time.Minute).Err())
	t.Cleanup(func() {
		_ = rdb.Del(ctx, key).Err()
	})

	_, err := repo.GetActive(ctx, mobile, role, time.Now())

	require.ErrorIs(t, err, biz.ErrSMSCodeNotFound)
	exists, err := rdb.Exists(ctx, key).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), exists)
}

func newRedisTestClient(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func TestSessionRepoCreateGetAndRevoke(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sessionModel{}))

	repo := NewSessionRepo(db, zap.NewNop())
	session := &biz.AuthSession{
		AccountID:         1001,
		Role:              biz.RolePassenger,
		RefreshTokenHash:  "refresh-hash",
		Status:            biz.SessionStatusActive,
		RefreshExpireTime: time.Now().Add(7 * 24 * time.Hour),
	}
	require.NoError(t, repo.Create(context.Background(), session))
	require.NotZero(t, session.ID)

	found, err := repo.GetByRefreshTokenHash(context.Background(), "refresh-hash")
	require.NoError(t, err)
	require.Equal(t, int64(1001), found.AccountID)

	require.NoError(t, repo.Revoke(context.Background(), session.ID))
	revoked, err := repo.GetByRefreshTokenHash(context.Background(), "refresh-hash")
	require.NoError(t, err)
	require.Equal(t, biz.SessionStatusRevoked, revoked.Status)
	require.NotNil(t, revoked.RevokedAt)
}
