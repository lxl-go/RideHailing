# Dynamic RBAC Permissions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace gateway role-only authorization with RBAC permission-code checks backed by `auth-service`.

**Architecture:** `auth-service` owns RBAC persistence, seeds built-in role-permission bindings, and exposes `CheckPermission`. `gateway-service` keeps JWT authentication at the edge, maps each protected `/carpool/**` operation to a permission code, calls `auth-service`, and fails closed on denied permissions.

**Tech Stack:** Go, Kratos HTTP/gRPC transport, protobuf, GORM, MySQL-compatible schema, existing fallback/Nacos config, existing Go test setup.

## Global Constraints

- Keep `admin-platform` unchanged.
- Keep passenger and driver uni-app API contracts unchanged.
- Keep JWT as the authentication carrier.
- Keep `gateway-service` as the external traffic entry.
- Keep `auth-service` as the owner of user, role, and permission data.
- Add built-in permission seed data for passenger and driver roles.
- Add an `auth-service` permission check API for gateway use.
- Replace gateway route `requireRole` checks with permission-code checks.
- Do not build RBAC management screens or admin APIs in this phase.
- Do not introduce tenant, department, data-scope, or field-level authorization in this phase.
- Do not modify `admin-platform`.
- Do not commit changes to git.

---

## File Structure

- Modify `services/auth-service/api/auth/v1/auth.proto`: add `CheckPermission` RPC and request/reply messages.
- Regenerate `services/auth-service/api/auth/v1/auth.pb.go`, `auth_grpc.pb.go`, and `auth_http.pb.go` from proto.
- Modify `services/auth-service/internal/biz/repo.go`: add RBAC permission repository interface.
- Modify `services/auth-service/internal/biz/auth.go`: add `CheckPermission` usecase method.
- Modify `services/auth-service/internal/data/rbac.go`: add permission constants, seed definitions, idempotent seed function, and permission lookup query.
- Modify `services/auth-service/internal/data/data.go`: call RBAC seed during data initialization and wire repo into `AuthUsecase`.
- Modify `services/auth-service/internal/service/auth.go`: expose `CheckPermission` at transport service layer.
- Modify `services/auth-service/internal/data/auth_test.go`: add RBAC seed and permission query integration tests.
- Modify `services/auth-service/internal/biz/auth_test.go`: add usecase permission check tests with a fake repo.
- Modify `services/gateway-service/internal/data/auth_client.go`: add `CheckPermission` to HTTP and gRPC auth clients.
- Modify `services/gateway-service/internal/biz/auth.go`: pass through `CheckPermission`.
- Modify `services/gateway-service/internal/service/auth.go`: expose `CheckPermission` to gateway server handlers.
- Modify `services/gateway-service/internal/server/auth_filter.go`: replace `requireRole` helper with `requirePermission`.
- Modify `services/gateway-service/internal/server/http.go`: replace route role checks with permission codes.
- Modify `services/gateway-service/internal/server/auth_filter_test.go`: add permission enforcement tests.
- Modify `services/gateway-service/internal/data/auth_client_test.go`: add HTTP client path test for `/v1/auth/permission/check`.

---

### Task 1: Auth-Service Permission Contract

**Files:**
- Modify: `services/auth-service/api/auth/v1/auth.proto`
- Generated: `services/auth-service/api/auth/v1/auth.pb.go`
- Generated: `services/auth-service/api/auth/v1/auth_grpc.pb.go`
- Generated: `services/auth-service/api/auth/v1/auth_http.pb.go`
- Modify: `services/auth-service/internal/service/auth.go`

**Interfaces:**
- Produces: `CheckPermission(ctx context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionReply, error)`
- Produces: `CheckPermissionRequest{UserId int64, PermissionCode string}`
- Produces: `CheckPermissionReply{Allowed bool}`

- [ ] **Step 1: Write the failing service-layer test**

Add this test to `services/auth-service/internal/service/auth_test.go` if the file does not exist, otherwise append it:

```go
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	authv1 "ride-hailing/services/auth-service/api/auth/v1"
	"ride-hailing/services/auth-service/internal/biz"
)

type permissionUsecase interface {
	CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
}

func TestCheckPermissionReturnsAllowedReply(t *testing.T) {
	uc := &fakeAuthUsecaseForPermission{allowed: true}
	svc := &AuthService{uc: uc}

	reply, err := svc.CheckPermission(context.Background(), &authv1.CheckPermissionRequest{
		UserId:         1001,
		PermissionCode: "order:create",
	})

	require.NoError(t, err)
	require.True(t, reply.Allowed)
	require.Equal(t, int64(1001), uc.userID)
	require.Equal(t, "order:create", uc.permissionCode)
}

type fakeAuthUsecaseForPermission struct {
	allowed        bool
	userID         int64
	permissionCode string
}

func (f *fakeAuthUsecaseForPermission) SendLoginCode(context.Context, string, string) error {
	return nil
}

func (f *fakeAuthUsecaseForPermission) Login(context.Context, string, string, string) (*biz.Session, error) {
	return nil, nil
}

func (f *fakeAuthUsecaseForPermission) VerifyToken(context.Context, string) (*biz.TokenClaims, error) {
	return nil, nil
}

func (f *fakeAuthUsecaseForPermission) RefreshToken(context.Context, string) (*biz.Session, error) {
	return nil, nil
}

func (f *fakeAuthUsecaseForPermission) Logout(context.Context, string) error {
	return nil
}

func (f *fakeAuthUsecaseForPermission) CheckPermission(_ context.Context, userID int64, permissionCode string) (bool, error) {
	f.userID = userID
	f.permissionCode = permissionCode
	return f.allowed, nil
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
cd services/auth-service
go test ./internal/service -run TestCheckPermissionReturnsAllowedReply -count=1
```

Expected: FAIL because `authv1.CheckPermissionRequest`, `AuthService.CheckPermission`, or the service usecase shape is not available yet.

- [ ] **Step 3: Update proto contract**

In `services/auth-service/api/auth/v1/auth.proto`, add this RPC before the closing `}` of `service AuthService`:

```proto
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionReply) {
    option (google.api.http) = {
      post: "/v1/auth/permission/check"
      body: "*"
    };
  }
```

Add these messages after `LogoutReply`:

```proto
message CheckPermissionRequest {
  int64 user_id = 1;
  string permission_code = 2;
}

message CheckPermissionReply {
  bool allowed = 1;
}
```

- [ ] **Step 4: Regenerate proto code**

Run from `services/auth-service`:

```powershell
make api
```

If `make` is unavailable on Windows, run the existing equivalent project command used by other services:

```powershell
protoc --proto_path=./api --proto_path=../../third_party --go_out=paths=source_relative:./api --go-http_out=paths=source_relative:./api --go-grpc_out=paths=source_relative:./api ./api/auth/v1/auth.proto
```

Expected: generated files include `CheckPermission`.

- [ ] **Step 5: Add service method**

Update `services/auth-service/internal/service/auth.go`:

```go
func (s *AuthService) CheckPermission(ctx context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionReply, error) {
	allowed, err := s.uc.CheckPermission(ctx, req.UserId, req.PermissionCode)
	if err != nil {
		return nil, err
	}
	return &authv1.CheckPermissionReply{Allowed: allowed}, nil
}
```

If `AuthService.uc` is currently concrete `*biz.AuthUsecase`, keep that type and add the method to `biz.AuthUsecase` in Task 2 before rerunning the package tests.

- [ ] **Step 6: Run test to verify it passes after Task 2 interfaces exist**

Run:

```powershell
cd services/auth-service
go test ./internal/service -run TestCheckPermissionReturnsAllowedReply -count=1
```

Expected: PASS after Task 2 completes the usecase method.

---

### Task 2: Auth-Service RBAC Seed and Permission Query

**Files:**
- Modify: `services/auth-service/internal/biz/repo.go`
- Modify: `services/auth-service/internal/biz/auth.go`
- Modify: `services/auth-service/internal/biz/auth_test.go`
- Modify: `services/auth-service/internal/data/rbac.go`
- Modify: `services/auth-service/internal/data/data.go`
- Modify: `services/auth-service/internal/data/auth_test.go`

**Interfaces:**
- Consumes: `authv1.CheckPermissionRequest` and `authv1.CheckPermissionReply` from Task 1.
- Produces: `type PermissionRepo interface { CheckUserPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) }`
- Produces: `func SeedBuiltinRBAC(ctx context.Context, db *gorm.DB) error`
- Produces: `func (uc *AuthUsecase) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)`

- [ ] **Step 1: Write failing data tests**

Append these tests to `services/auth-service/internal/data/auth_test.go`:

```go
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

	accountRepo := NewAccountRepo(db)
	account, err := accountRepo.FindOrCreate(context.Background(), "13800138000", "passenger")
	require.NoError(t, err)
	require.NoError(t, accountRepo.EnsureUserRole(context.Background(), account.ID, "passenger"))

	rbacRepo := NewRBACRepo(db)
	allowed, err := rbacRepo.CheckUserPermission(context.Background(), account.ID, "order:create")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = rbacRepo.CheckUserPermission(context.Background(), account.ID, "trip:publish")
	require.NoError(t, err)
	require.False(t, allowed)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```powershell
cd services/auth-service
go test ./internal/data -run "TestSeedBuiltinRBACIsIdempotent|TestCheckUserPermissionUsesRolePermissionBindings" -count=1
```

Expected: FAIL because `SeedBuiltinRBAC`, `NewRBACRepo`, and `CheckUserPermission` do not exist.

- [ ] **Step 3: Add permission repository interface**

Add to `services/auth-service/internal/biz/repo.go`:

```go
type PermissionRepo interface {
	CheckUserPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
}
```

- [ ] **Step 4: Add usecase dependency and method**

Update `AuthUsecase` in `services/auth-service/internal/biz/auth.go` to include:

```go
permissions PermissionRepo
```

Update `NewAuthUsecase` signature to accept `permissions PermissionRepo` and assign it:

```go
func NewAuthUsecase(accounts AccountRepo, codes SMSCodeRepo, sessions SessionRepo, permissions PermissionRepo, opts AuthOptions) *AuthUsecase {
	// keep existing option defaults
	return &AuthUsecase{
		accounts:    accounts,
		codes:       codes,
		sessions:    sessions,
		permissions: permissions,
		opts:        opts,
		now:         time.Now,
	}
}
```

Add:

```go
func (uc *AuthUsecase) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	if userID <= 0 || strings.TrimSpace(permissionCode) == "" {
		return false, nil
	}
	if uc.permissions == nil {
		return false, nil
	}
	return uc.permissions.CheckUserPermission(ctx, userID, strings.TrimSpace(permissionCode))
}
```

- [ ] **Step 5: Update auth usecase tests**

In `services/auth-service/internal/biz/auth_test.go`, add a fake permission repo:

```go
type fakePermissionRepo struct {
	allowed map[string]bool
}

func (r *fakePermissionRepo) CheckUserPermission(_ context.Context, _ int64, permissionCode string) (bool, error) {
	return r.allowed[permissionCode], nil
}
```

Update helper construction so `NewAuthUsecase` receives `&fakePermissionRepo{allowed: map[string]bool{}}`.

Add:

```go
func TestCheckPermissionDeniesEmptyInput(t *testing.T) {
	uc, _ := newTestAuthUsecase(t)

	allowed, err := uc.CheckPermission(context.Background(), 0, "order:create")
	require.NoError(t, err)
	require.False(t, allowed)

	allowed, err = uc.CheckPermission(context.Background(), 1001, "")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestCheckPermissionDelegatesToPermissionRepo(t *testing.T) {
	uc, deps := newTestAuthUsecase(t)
	deps.permissions.allowed["order:create"] = true

	allowed, err := uc.CheckPermission(context.Background(), 1001, " order:create ")

	require.NoError(t, err)
	require.True(t, allowed)
}
```

If `newTestAuthUsecase` returns a different dependency struct, extend it with `permissions *fakePermissionRepo`.

- [ ] **Step 6: Implement RBAC seed and repo**

In `services/auth-service/internal/data/rbac.go`, add:

```go
import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	RolePassenger = "passenger"
	RoleDriver    = "driver"
)

type builtinPermission struct {
	Code     string
	Name     string
	Resource string
	Action   string
	Roles    []string
}

var builtinPermissions = []builtinPermission{
	{Code: "trip:search", Name: "Search trips", Resource: "trip", Action: "search", Roles: []string{RolePassenger, RoleDriver}},
	{Code: "trip:view_detail", Name: "View trip detail", Resource: "trip", Action: "view_detail", Roles: []string{RolePassenger, RoleDriver}},
	{Code: "trip:publish", Name: "Publish trip", Resource: "trip", Action: "publish", Roles: []string{RoleDriver}},
	{Code: "trip:list_driver_self", Name: "List own driver trips", Resource: "trip", Action: "list_driver_self", Roles: []string{RoleDriver}},
	{Code: "trip:update_status_self", Name: "Update own trip status", Resource: "trip", Action: "update_status_self", Roles: []string{RoleDriver}},
	{Code: "order:create", Name: "Create order", Resource: "order", Action: "create", Roles: []string{RolePassenger}},
	{Code: "order:cancel_self", Name: "Cancel own order", Resource: "order", Action: "cancel_self", Roles: []string{RolePassenger}},
	{Code: "order:list_passenger_self", Name: "List own passenger orders", Resource: "order", Action: "list_passenger_self", Roles: []string{RolePassenger}},
	{Code: "order:view_passenger_self", Name: "View own passenger order", Resource: "order", Action: "view_passenger_self", Roles: []string{RolePassenger}},
	{Code: "order:list_driver_pending", Name: "List pending driver orders", Resource: "order", Action: "list_driver_pending", Roles: []string{RoleDriver}},
	{Code: "order:accept_driver_self", Name: "Accept own driver order", Resource: "order", Action: "accept_driver_self", Roles: []string{RoleDriver}},
	{Code: "order:reject_driver_self", Name: "Reject own driver order", Resource: "order", Action: "reject_driver_self", Roles: []string{RoleDriver}},
	{Code: "review:submit", Name: "Submit review", Resource: "review", Action: "submit", Roles: []string{RolePassenger, RoleDriver}},
	{Code: "passenger:profile:view_self", Name: "View passenger profile", Resource: "passenger_profile", Action: "view_self", Roles: []string{RolePassenger}},
	{Code: "passenger:profile:update_self", Name: "Update passenger profile", Resource: "passenger_profile", Action: "update_self", Roles: []string{RolePassenger}},
	{Code: "driver:profile:view_self", Name: "View driver profile", Resource: "driver_profile", Action: "view_self", Roles: []string{RoleDriver}},
	{Code: "driver:profile:update_self", Name: "Update driver profile", Resource: "driver_profile", Action: "update_self", Roles: []string{RoleDriver}},
	{Code: "driver:certification:submit_self", Name: "Submit driver certification", Resource: "driver_certification", Action: "submit_self", Roles: []string{RoleDriver}},
	{Code: "driver:certification:view_self", Name: "View driver certification", Resource: "driver_certification", Action: "view_self", Roles: []string{RoleDriver}},
	{Code: "driver:vehicle:manage_self", Name: "Manage driver vehicle", Resource: "driver_vehicle", Action: "manage_self", Roles: []string{RoleDriver}},
	{Code: "driver:vehicle:list_self", Name: "List driver vehicles", Resource: "driver_vehicle", Action: "list_self", Roles: []string{RoleDriver}},
}

func SeedBuiltinRBAC(ctx context.Context, db *gorm.DB) error {
	roles := []roleModel{
		{Code: RolePassenger, Name: "Passenger", Status: 1},
		{Code: RoleDriver, Name: "Driver", Status: 1},
	}
	for _, role := range roles {
		if err := db.WithContext(ctx).Where("code = ?", role.Code).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}

	roleByCode := map[string]roleModel{}
	var existingRoles []roleModel
	if err := db.WithContext(ctx).Where("code IN ?", []string{RolePassenger, RoleDriver}).Find(&existingRoles).Error; err != nil {
		return err
	}
	for _, role := range existingRoles {
		roleByCode[role.Code] = role
	}

	for _, item := range builtinPermissions {
		permission := permissionModel{
			Code:     item.Code,
			Name:     item.Name,
			Resource: item.Resource,
			Action:   item.Action,
			Status:   1,
		}
		if err := db.WithContext(ctx).Where("code = ?", permission.Code).FirstOrCreate(&permission).Error; err != nil {
			return err
		}
		for _, roleCode := range item.Roles {
			role := roleByCode[roleCode]
			binding := rolePermissionModel{RoleID: role.ID, PermissionID: permission.ID}
			if err := db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).FirstOrCreate(&binding).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type RBACRepo struct {
	db *gorm.DB
}

func NewRBACRepo(db *gorm.DB) *RBACRepo {
	return &RBACRepo{db: db}
}

func (r *RBACRepo) CheckUserPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("auth_user AS u").
		Joins("JOIN auth_user_role AS ur ON ur.user_id = u.id").
		Joins("JOIN auth_role AS r ON r.id = ur.role_id").
		Joins("JOIN auth_role_permission AS rp ON rp.role_id = r.id").
		Joins("JOIN auth_permission AS p ON p.id = rp.permission_id").
		Where("u.id = ? AND u.status = ? AND r.status = ? AND p.status = ? AND p.code = ?", userID, 1, 1, 1, permissionCode).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

var _ = time.Time{}
```

Remove `var _ = time.Time{}` if the file already uses `time` through model definitions in the same file.

- [ ] **Step 7: Wire seed and repo into data initialization**

In `services/auth-service/internal/data/data.go`, after `AutoMigrate`, call:

```go
if err := SeedBuiltinRBAC(context.Background(), db); err != nil {
	panic(err)
}
```

Update `NewAuthUsecase` construction to pass:

```go
NewRBACRepo(db)
```

- [ ] **Step 8: Run auth-service tests**

Run:

```powershell
cd services/auth-service
go test ./... -count=1
```

Expected: PASS.

---

### Task 3: Gateway Auth Client Permission Call

**Files:**
- Modify: `services/gateway-service/internal/data/auth_client.go`
- Modify: `services/gateway-service/internal/data/auth_client_test.go`
- Modify: `services/gateway-service/internal/biz/auth.go`
- Modify: `services/gateway-service/internal/service/auth.go`

**Interfaces:**
- Consumes: `authv1.CheckPermissionRequest` from Task 1.
- Produces: `CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)` on gateway auth client, biz, and service layers.

- [ ] **Step 1: Write failing HTTP client test**

Append to `services/gateway-service/internal/data/auth_client_test.go`:

```go
func TestAuthHTTPClientCheckPermissionUsesPermissionCheckPath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, http.MethodPost, r.Method)
		var req authv1.CheckPermissionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, int64(1001), req.UserId)
		require.Equal(t, "order:create", req.PermissionCode)
		_ = json.NewEncoder(w).Encode(&authv1.CheckPermissionReply{Allowed: true})
	}))
	defer server.Close()

	client := NewAuthHTTPClient(server.URL)
	allowed, err := client.CheckPermission(t.Context(), 1001, "order:create")

	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, "/v1/auth/permission/check", gotPath)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
cd services/gateway-service
go test ./internal/data -run TestAuthHTTPClientCheckPermissionUsesPermissionCheckPath -count=1
```

Expected: FAIL because `CheckPermission` does not exist on `AuthHTTPClient`.

- [ ] **Step 3: Add client methods**

Update `AuthClient` interface in `services/gateway-service/internal/data/auth_client.go`:

```go
CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
```

Add to `AuthGRPCClient`:

```go
func (c *AuthGRPCClient) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	reply, err := c.client.CheckPermission(ctx, &authv1.CheckPermissionRequest{UserId: userID, PermissionCode: permissionCode})
	if err != nil {
		return false, err
	}
	return reply.Allowed, nil
}
```

Add to `AuthHTTPClient`:

```go
func (c *AuthHTTPClient) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	var reply authv1.CheckPermissionReply
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/permission/check", &authv1.CheckPermissionRequest{
		UserId:         userID,
		PermissionCode: permissionCode,
	}, &reply)
	if err != nil {
		return false, err
	}
	return reply.Allowed, nil
}
```

- [ ] **Step 4: Add biz and service pass-through**

In `services/gateway-service/internal/biz/auth.go`:

```go
func (uc *AuthUsecase) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	return uc.client.CheckPermission(ctx, userID, permissionCode)
}
```

In `services/gateway-service/internal/service/auth.go`:

```go
func (s *AuthService) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	return s.uc.CheckPermission(ctx, userID, permissionCode)
}
```

- [ ] **Step 5: Run gateway data tests**

Run:

```powershell
cd services/gateway-service
go test ./internal/data -run TestAuthHTTPClientCheckPermissionUsesPermissionCheckPath -count=1
```

Expected: PASS.

---

### Task 4: Gateway Permission Enforcement

**Files:**
- Modify: `services/gateway-service/internal/server/auth_filter.go`
- Modify: `services/gateway-service/internal/server/auth_filter_test.go`
- Modify: `services/gateway-service/internal/server/http.go`

**Interfaces:**
- Consumes: `CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)` from Task 3.
- Produces: `requirePermission(w http.ResponseWriter, r *http.Request, authSvc permissionChecker, permissionCode string) bool`

- [ ] **Step 1: Write failing permission helper tests**

Append to `services/gateway-service/internal/server/auth_filter_test.go`:

```go
type fakePermissionChecker struct {
	allowed        bool
	userID         int64
	permissionCode string
}

func (f *fakePermissionChecker) CheckPermission(_ context.Context, userID int64, permissionCode string) (bool, error) {
	f.userID = userID
	f.permissionCode = permissionCode
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```powershell
cd services/gateway-service
go test ./internal/server -run "TestRequirePermission" -count=1
```

Expected: FAIL because `requirePermission` does not exist.

- [ ] **Step 3: Implement permission helper**

In `services/gateway-service/internal/server/auth_filter.go`, add:

```go
type permissionChecker interface {
	CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
}

func requirePermission(w http.ResponseWriter, r *http.Request, checker permissionChecker, permissionCode string) bool {
	user, ok := CurrentUserFromRequest(r)
	if !ok {
		writeUnauthorized(w)
		return false
	}
	if checker == nil || strings.TrimSpace(permissionCode) == "" {
		writeForbidden(w)
		return false
	}
	allowed, err := checker.CheckPermission(r.Context(), user.UserID, strings.TrimSpace(permissionCode))
	if err != nil || !allowed {
		writeForbidden(w)
		return false
	}
	return true
}
```

Remove `requireRole` only after all route call sites are migrated.

- [ ] **Step 4: Replace route checks in `http.go`**

Use this mapping:

```go
GET /carpool/trips -> trip:search
GET /carpool/trips/{id} -> trip:view_detail
POST /carpool/trips -> trip:publish
GET /carpool/trips/mine -> trip:list_driver_self
PUT /carpool/trips/{id}/status -> trip:update_status_self
POST /carpool/orders -> order:create
POST /carpool/orders/{id}/cancel -> order:cancel_self
GET /carpool/orders -> order:list_passenger_self
GET /carpool/orders/pending -> order:list_driver_pending
GET /carpool/orders/{id} -> order:view_passenger_self
POST /carpool/orders/{id}/accept -> order:accept_driver_self
POST /carpool/orders/{id}/reject -> order:reject_driver_self
POST /carpool/reviews -> review:submit
GET /carpool/passengers/me -> passenger:profile:view_self
PUT /carpool/passengers/me -> passenger:profile:update_self
GET /carpool/drivers/me -> driver:profile:view_self
PUT /carpool/drivers/me -> driver:profile:update_self
POST /carpool/drivers/certification -> driver:certification:submit_self
GET /carpool/drivers/certification -> driver:certification:view_self
POST /carpool/drivers/vehicles -> driver:vehicle:manage_self
GET /carpool/drivers/vehicles -> driver:vehicle:list_self
```

Example replacement:

```go
if !requirePermission(ctx.Response(), ctx.Request(), authSvc, "order:create") {
	return nil
}
```

Update route registration functions to accept `authSvc *service.AuthService` where needed:

```go
registerTripRoutes(srv, authSvc, tripSvc)
registerOrderRoutes(srv, authSvc, orderSvc)
registerReviewRoutes(srv, authSvc, reviewSvc)
registerPassengerRoutes(srv, authSvc, passengerSvc)
registerDriverRoutes(srv, authSvc, driverSvc)
```

- [ ] **Step 5: Remove role helper**

After all call sites are migrated, remove `requireRole` from `auth_filter.go`.

- [ ] **Step 6: Run gateway server tests**

Run:

```powershell
cd services/gateway-service
go test ./internal/server -count=1
```

Expected: PASS.

---

### Task 5: Final Verification and Regression Scan

**Files:**
- Read-only scan across `services/auth-service` and `services/gateway-service`.

**Interfaces:**
- Consumes: all tasks above.
- Produces: verified dynamic RBAC permission enforcement.

- [ ] **Step 1: Run auth-service full tests**

Run:

```powershell
cd services/auth-service
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 2: Run gateway-service full tests**

Run:

```powershell
cd services/gateway-service
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 3: Scan for old gateway role checks**

Run from repository root:

```powershell
rg -n "requireRole\\(" services\\gateway-service\\internal
```

Expected: no matches.

- [ ] **Step 4: Scan for protected review route permission**

Run from repository root:

```powershell
rg -n "review:submit|requirePermission" services\\gateway-service\\internal\\server
```

Expected: `review:submit` appears in review route and `requirePermission` is used by protected handlers.

- [ ] **Step 5: Scan generated auth contract**

Run from repository root:

```powershell
rg -n "CheckPermission|/v1/auth/permission/check" services\\auth-service\\api\\auth\\v1 services\\gateway-service\\internal
```

Expected: generated proto files, auth-service service layer, and gateway client/server code all contain `CheckPermission`.
