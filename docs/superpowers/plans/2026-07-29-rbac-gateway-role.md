# RBAC and Gateway Role Isolation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add RBAC five-table persistence, gateway current-user context, role isolation, and real uni-app login entry pages.

**Architecture:** `auth-service` owns RBAC tables and user-role assignment at login. `gateway-service` parses JWT once in the auth filter, stores `CurrentUser` in context, and enforces passenger/driver roles at route boundaries. Uni-app clients use the existing auth API wrappers through a visible login page.

**Tech Stack:** Go, Kratos v2, GORM/MySQL, Nacos config, JWT, uni-app Vue 3.

## Global Constraints

- Do not modify `admin-platform`.
- Do not listen on or register business services at `8080` or `8848`.
- Use Nacos-first config with `configs/nacos.yaml`; use `configs/config.yaml` as fallback.
- Use MySQL database `01ride`.
- Do not commit to git in this phase.
- RBAC persistence must include exactly these five logical tables: user, role, user-role relation, permission, role-permission relation.

---

### Task 1: Auth Service RBAC Persistence

**Files:**
- Modify: `services/auth-service/internal/data/auth.go`
- Create: `services/auth-service/internal/data/rbac.go`
- Modify: `services/auth-service/internal/data/data.go`
- Modify: `services/auth-service/internal/data/auth_test.go`
- Modify: `services/auth-service/internal/biz/repo.go`
- Modify: `services/auth-service/internal/biz/auth.go`
- Modify: `services/auth-service/internal/biz/auth_test.go`

**Interfaces:**
- Produces: `auth_user`, `auth_role`, `auth_user_role`, `auth_permission`, `auth_role_permission`; `EnsureUserRole(ctx, userID, role)`.

### Task 2: Gateway Current User

**Files:**
- Modify: `services/gateway-service/internal/server/auth_filter.go`
- Modify: `services/gateway-service/internal/server/auth_filter_test.go`
- Modify: `services/gateway-service/internal/server/http.go`

**Interfaces:**
- Produces: `type CurrentUser`, `CurrentUserFromRequest(*http.Request)`, `requireRole`.

### Task 3: Role Isolation

**Files:**
- Modify: `services/gateway-service/internal/server/http.go`
- Modify: `services/gateway-service/internal/server/auth_filter_test.go`

**Interfaces:**
- Passenger-only routes reject `driver`; driver-only routes reject `passenger`.

### Task 4: Uni-App Login Pages

**Files:**
- Create: `passenger-platform/uni-app/src/pages/login/login.vue`
- Modify: `passenger-platform/uni-app/src/pages.json`
- Create: `driver-platform/uni-app/src/pages/login/login.vue`
- Modify: `driver-platform/uni-app/src/pages.json`

**Interfaces:**
- Produces: mobile input, send code, login, session storage, redirect home.

### Task 5: Verification

- Run `go test ./... -count=1` in `services/auth-service`.
- Run `go test ./... -count=1` in `services/gateway-service`.
- Run existing business service tests.
- Run passenger and driver H5 builds.
- Scan for reserved business ports.
