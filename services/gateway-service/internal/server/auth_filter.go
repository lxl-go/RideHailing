package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"ride-hailing/pkg/authx"
	"ride-hailing/services/gateway-service/internal/conf"
)

type userIDContextKey struct{}
type currentUserContextKey struct{}

type CurrentUser struct {
	UserID int64
	Role   string
	JTI    string
}

type permissionChecker interface {
	CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
}

var ErrPermissionCircuitOpen = errors.New("permission checker circuit open")

type PermissionPolicy struct {
	CacheTTL         time.Duration
	FailureThreshold int
	CircuitOpenTTL   time.Duration
	Now              func() time.Time
}

type permissionCacheEntry struct {
	allowed   bool
	expiresAt time.Time
}

type CachedPermissionChecker struct {
	upstream permissionChecker
	policy   PermissionPolicy

	mu           sync.Mutex
	cache        map[string]permissionCacheEntry
	failures     int
	circuitUntil time.Time
}

func NewCachedPermissionChecker(upstream permissionChecker, policy PermissionPolicy) *CachedPermissionChecker {
	if policy.Now == nil {
		policy.Now = time.Now
	}
	if policy.FailureThreshold < 0 {
		policy.FailureThreshold = 0
	}
	return &CachedPermissionChecker{
		upstream: upstream,
		policy:   policy,
		cache:    map[string]permissionCacheEntry{},
	}
}

func (c *CachedPermissionChecker) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	if c == nil || c.upstream == nil {
		return false, nil
	}
	permissionCode = strings.TrimSpace(permissionCode)
	if userID <= 0 || permissionCode == "" {
		return false, nil
	}

	now := c.policy.Now()
	key := permissionCacheKey(userID, permissionCode)

	c.mu.Lock()
	if now.Before(c.circuitUntil) {
		c.mu.Unlock()
		return false, ErrPermissionCircuitOpen
	}
	if entry, ok := c.cache[key]; ok && now.Before(entry.expiresAt) {
		c.mu.Unlock()
		return entry.allowed, nil
	}
	c.mu.Unlock()

	allowed, err := c.upstream.CheckPermission(ctx, userID, permissionCode)
	if err != nil {
		c.recordFailure(now)
		return false, err
	}

	c.mu.Lock()
	c.failures = 0
	if c.policy.CacheTTL > 0 {
		c.cache[key] = permissionCacheEntry{allowed: allowed, expiresAt: now.Add(c.policy.CacheTTL)}
	}
	c.mu.Unlock()
	return allowed, nil
}

func (c *CachedPermissionChecker) recordFailure(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	if c.policy.FailureThreshold > 0 && c.failures >= c.policy.FailureThreshold && c.policy.CircuitOpenTTL > 0 {
		c.circuitUntil = now.Add(c.policy.CircuitOpenTTL)
	}
}

func permissionCacheKey(userID int64, permissionCode string) string {
	return strconv.FormatInt(userID, 10) + ":" + strings.TrimSpace(permissionCode)
}

func permissionPolicyFromConfig(c *conf.Auth) PermissionPolicy {
	policy := PermissionPolicy{
		CacheTTL:         30 * time.Second,
		FailureThreshold: 3,
		CircuitOpenTTL:   10 * time.Second,
	}
	if c == nil || c.Permission == nil {
		return policy
	}
	if d, err := time.ParseDuration(c.Permission.CacheTTL); err == nil && d >= 0 {
		policy.CacheTTL = d
	}
	if c.Permission.FailureThreshold >= 0 {
		policy.FailureThreshold = c.Permission.FailureThreshold
	}
	if d, err := time.ParseDuration(c.Permission.CircuitOpenTTL); err == nil && d >= 0 {
		policy.CircuitOpenTTL = d
	}
	return policy
}

func NewAuthFilter(c *conf.Auth) khttp.FilterFunc {
	cfg := authx.JWTConfig{}
	compatibleHeader := true
	enabled := false
	if c != nil && c.Jwt != nil {
		enabled = c.Jwt.Enabled
		cfg.Enabled = c.Jwt.Enabled
		cfg.Secret = c.Jwt.Secret
		cfg.Issuer = c.Jwt.Issuer
		cfg.ExpireSeconds = c.Jwt.ExpireSeconds
		compatibleHeader = c.Jwt.CompatibleHeaderEnabled
	}
	manager := authx.NewManager(cfg)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled || isPublicPath(r) {
				next.ServeHTTP(w, r)
				return
			}
			applyWebSocketQueryToken(r)

			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			claims, err := manager.ParseBearer(authorization)
			if err == nil {
				userID, parseErr := strconv.ParseInt(claims.UserID, 10, 64)
				if parseErr == nil && userID > 0 {
					next.ServeHTTP(w, withCurrentUser(r, CurrentUser{UserID: userID, Role: claims.Role, JTI: claims.JTI}))
					return
				}
			}

			if compatibleHeader && authorization == "" {
				userID, parseErr := strconv.ParseInt(strings.TrimSpace(r.Header.Get("X-User-Id")), 10, 64)
				if parseErr == nil && userID > 0 {
					next.ServeHTTP(w, withCurrentUser(r, CurrentUser{UserID: userID}))
					return
				}
			}

			writeUnauthorized(w)
		})
	}
}

func applyWebSocketQueryToken(r *http.Request) {
	if r == nil || strings.TrimSpace(r.Header.Get("Authorization")) != "" || !isWebSocketRequest(r) {
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if token == "" {
		return
	}
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}
	r.Header.Set("Authorization", token)
}

func isWebSocketRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket")
}

func UserIDFromRequest(r *http.Request) int64 {
	if r == nil {
		return 0
	}
	if userID, ok := r.Context().Value(userIDContextKey{}).(int64); ok && userID > 0 {
		return userID
	}
	userID, _ := strconv.ParseInt(strings.TrimSpace(r.Header.Get("X-User-Id")), 10, 64)
	return userID
}

func CurrentUserFromRequest(r *http.Request) (CurrentUser, bool) {
	if r == nil {
		return CurrentUser{}, false
	}
	user, ok := r.Context().Value(currentUserContextKey{}).(CurrentUser)
	return user, ok && user.UserID > 0
}

func withCurrentUser(r *http.Request, user CurrentUser) *http.Request {
	ctx := context.WithValue(r.Context(), userIDContextKey{}, user.UserID)
	ctx = context.WithValue(ctx, currentUserContextKey{}, user)
	r = r.Clone(ctx)
	r.Header.Set("X-User-Id", strconv.FormatInt(user.UserID, 10))
	return r
}

func currentUserID(r *http.Request) int64 {
	user, ok := CurrentUserFromRequest(r)
	if !ok {
		return 0
	}
	return user.UserID
}

func requirePermission(w http.ResponseWriter, r *http.Request, checker permissionChecker, permissionCode string) bool {
	user, ok := CurrentUserFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return false
	}
	permissionCode = strings.TrimSpace(permissionCode)
	if checker == nil || permissionCode == "" {
		writeForbidden(w)
		return false
	}
	allowed, err := checker.CheckPermission(r.Context(), user.UserID, permissionCode)
	if err != nil || !allowed {
		writeForbidden(w)
		return false
	}
	return true
}

func isPublicPath(r *http.Request) bool {
	if r.Method == http.MethodOptions {
		return true
	}
	switch r.URL.Path {
	case "/carpool/auth/sms/send", "/carpool/auth/login", "/carpool/auth/refresh", "/carpool/auth/logout", "/api/v1/maps/static", "/pay/success", "/api/v1/pay/notify":
		return true
	default:
		return false
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": http.StatusUnauthorized,
		"data": nil,
		"msg":  "unauthorized",
	})
}

func writeForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": http.StatusForbidden,
		"data": nil,
		"msg":  "forbidden",
	})
}
