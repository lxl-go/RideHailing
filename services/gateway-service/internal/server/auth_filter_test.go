package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"ride-hailing/pkg/authx"
	"ride-hailing/services/gateway-service/internal/conf"
)

func TestAuthFilterInjectsUserIDFromJWT(t *testing.T) {
	manager := authx.NewManager(authx.JWTConfig{Secret: "test-secret", Issuer: "ride-hailing-test"})
	token, err := manager.Generate(authx.Claims{UserID: "1001", Role: "passenger"})
	require.NoError(t, err)

	filter := NewAuthFilter(&conf.Auth{Jwt: &conf.Auth_JWT{
		Enabled:                 true,
		Secret:                  "test-secret",
		Issuer:                  "ride-hailing-test",
		CompatibleHeaderEnabled: false,
	}})
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "1001", r.Header.Get("X-User-Id"))
		user, ok := CurrentUserFromRequest(r)
		require.True(t, ok)
		require.Equal(t, int64(1001), user.UserID)
		require.Equal(t, "passenger", user.Role)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/carpool/passengers/me", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestAuthFilterAcceptsWebSocketAccessTokenQuery(t *testing.T) {
	manager := authx.NewManager(authx.JWTConfig{Secret: "test-secret", Issuer: "ride-hailing-test"})
	token, err := manager.Generate(authx.Claims{UserID: "7001", Role: "driver"})
	require.NoError(t, err)

	filter := NewAuthFilter(&conf.Auth{Jwt: &conf.Auth_JWT{
		Enabled:                 true,
		Secret:                  "test-secret",
		Issuer:                  "ride-hailing-test",
		CompatibleHeaderEnabled: false,
	}})
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer "+token.AccessToken, r.Header.Get("Authorization"))
		user, ok := CurrentUserFromRequest(r)
		require.True(t, ok)
		require.Equal(t, int64(7001), user.UserID)
		require.Equal(t, "driver", user.Role)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/driver/orders/2087079500750913536/chat/ws?role=driver&user_id=7001&access_token="+token.AccessToken, nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestAuthFilterRejectsMissingTokenWhenCompatibilityDisabled(t *testing.T) {
	filter := NewAuthFilter(&conf.Auth{Jwt: &conf.Auth_JWT{
		Enabled:                 true,
		Secret:                  "test-secret",
		CompatibleHeaderEnabled: false,
	}})
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/carpool/passengers/me", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAuthFilterAllowsLoginPathWithoutToken(t *testing.T) {
	filter := NewAuthFilter(&conf.Auth{Jwt: &conf.Auth_JWT{Enabled: true, Secret: "test-secret"}})
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/carpool/auth/login", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestAuthFilterAllowsPublicAuthPathsWithoutToken(t *testing.T) {
	filter := NewAuthFilter(&conf.Auth{Jwt: &conf.Auth_JWT{Enabled: true, Secret: "test-secret"}})
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, path := range []string{"/carpool/auth/sms/send", "/carpool/auth/refresh", "/carpool/auth/logout"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		require.Equal(t, http.StatusNoContent, res.Code, path)
	}
}

func TestAuthFilterAllowsPaymentCallbackPathsWithoutToken(t *testing.T) {
	filter := NewAuthFilter(&conf.Auth{Jwt: &conf.Auth_JWT{Enabled: true, Secret: "test-secret"}})
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/pay/success"},
		{method: http.MethodPost, path: "/api/v1/pay/notify"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		require.Equal(t, http.StatusNoContent, res.Code, tc.path)
	}
}

func TestAuthFilterAllowsTemporaryUserHeaderWhenCompatible(t *testing.T) {
	filter := NewAuthFilter(&conf.Auth{Jwt: &conf.Auth_JWT{
		Enabled:                 true,
		Secret:                  "test-secret",
		CompatibleHeaderEnabled: true,
	}})
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := CurrentUserFromRequest(r)
		require.True(t, ok)
		require.Equal(t, int64(3001), user.UserID)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/carpool/drivers/me", nil)
	req.Header.Set("X-User-Id", "3001")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestAuthFilterRejectsInvalidTokenEvenWhenTemporaryHeaderExists(t *testing.T) {
	filter := NewAuthFilter(&conf.Auth{Jwt: &conf.Auth_JWT{
		Enabled:                 true,
		Secret:                  "test-secret",
		CompatibleHeaderEnabled: true,
	}})
	handler := filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/carpool/passengers/me", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	req.Header.Set("X-User-Id", "3001")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
}

type fakePermissionChecker struct {
	allowed        bool
	err            error
	calls          int
	userID         int64
	permissionCode string
}

func (f *fakePermissionChecker) CheckPermission(_ context.Context, userID int64, permissionCode string) (bool, error) {
	f.calls++
	f.userID = userID
	f.permissionCode = permissionCode
	if f.err != nil {
		return false, f.err
	}
	return f.allowed, nil
}

func TestRequirePermissionReturnsUnauthorizedWithoutCurrentUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/carpool/orders", nil)
	res := httptest.NewRecorder()

	require.False(t, requirePermission(res, req, &fakePermissionChecker{allowed: true}, "order:create"))
	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestRequirePermissionReturnsForbiddenWhenDenied(t *testing.T) {
	checker := &fakePermissionChecker{allowed: false}
	req := withCurrentUser(httptest.NewRequest(http.MethodPost, "/carpool/orders", nil), CurrentUser{UserID: 1001, Role: "passenger"})
	res := httptest.NewRecorder()

	require.False(t, requirePermission(res, req, checker, "order:create"))
	require.Equal(t, http.StatusForbidden, res.Code)
	require.Equal(t, int64(1001), checker.userID)
	require.Equal(t, "order:create", checker.permissionCode)
}

func TestRequirePermissionAllowsWhenGranted(t *testing.T) {
	checker := &fakePermissionChecker{allowed: true}
	req := withCurrentUser(httptest.NewRequest(http.MethodPost, "/carpool/orders", nil), CurrentUser{UserID: 1001, Role: "passenger"})
	res := httptest.NewRecorder()

	require.True(t, requirePermission(res, req, checker, "order:create"))
	require.Equal(t, http.StatusOK, res.Code)
}

func TestCachedPermissionCheckerReusesAllowedDecisionWithinTTL(t *testing.T) {
	now := time.Unix(100, 0)
	upstream := &fakePermissionChecker{allowed: true}
	checker := NewCachedPermissionChecker(upstream, PermissionPolicy{
		CacheTTL: 30 * time.Second,
		Now:      func() time.Time { return now },
	})

	allowed, err := checker.CheckPermission(context.Background(), 1001, "order:create")
	require.NoError(t, err)
	require.True(t, allowed)

	now = now.Add(10 * time.Second)
	allowed, err = checker.CheckPermission(context.Background(), 1001, " order:create ")
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 1, upstream.calls)
}

func TestCachedPermissionCheckerRechecksAfterTTLExpires(t *testing.T) {
	now := time.Unix(100, 0)
	upstream := &fakePermissionChecker{allowed: true}
	checker := NewCachedPermissionChecker(upstream, PermissionPolicy{
		CacheTTL: 30 * time.Second,
		Now:      func() time.Time { return now },
	})

	allowed, err := checker.CheckPermission(context.Background(), 1001, "order:create")
	require.NoError(t, err)
	require.True(t, allowed)

	now = now.Add(31 * time.Second)
	allowed, err = checker.CheckPermission(context.Background(), 1001, "order:create")
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 2, upstream.calls)
}

func TestCachedPermissionCheckerOpensCircuitAfterConsecutiveFailures(t *testing.T) {
	now := time.Unix(100, 0)
	upstreamErr := errors.New("auth unavailable")
	upstream := &fakePermissionChecker{err: upstreamErr}
	checker := NewCachedPermissionChecker(upstream, PermissionPolicy{
		FailureThreshold: 2,
		CircuitOpenTTL:   time.Minute,
		Now:              func() time.Time { return now },
	})

	allowed, err := checker.CheckPermission(context.Background(), 1001, "order:create")
	require.ErrorIs(t, err, upstreamErr)
	require.False(t, allowed)

	allowed, err = checker.CheckPermission(context.Background(), 1001, "order:create")
	require.ErrorIs(t, err, upstreamErr)
	require.False(t, allowed)

	upstream.err = nil
	upstream.allowed = true
	allowed, err = checker.CheckPermission(context.Background(), 1001, "order:create")
	require.ErrorIs(t, err, ErrPermissionCircuitOpen)
	require.False(t, allowed)
	require.Equal(t, 2, upstream.calls)

	now = now.Add(time.Minute + time.Second)
	allowed, err = checker.CheckPermission(context.Background(), 1001, "order:create")
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 3, upstream.calls)
}
