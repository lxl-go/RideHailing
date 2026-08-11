# Auth Service and JWT Gateway Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build auth-service token issuing and gateway JWT protection for passenger/driver app APIs.

**Architecture:** Shared JWT behavior lives in `pkg/authx`. `auth-service` signs tokens and owns `auth_account` persistence. `gateway-service` proxies login and validates JWT before business route handlers, keeping `X-User-Id` compatibility during migration.

**Tech Stack:** Go, Kratos v2, GORM/MySQL, Nacos config/registry, `github.com/golang-jwt/jwt/v5`, uni-app request wrappers.

## Global Constraints

- Do not modify `admin-platform`.
- Do not listen on or register business services at `8080` or `8848`.
- Use Nacos-first config with `configs/nacos.yaml`; use `configs/config.yaml` as fallback.
- Use MySQL database `01ride`.
- Do not commit to git in this phase.
- Keep `X-User-Id` compatibility until JWT rollout is complete.

---

### Task 1: Shared JWT Package

**Files:**
- Create: `pkg/authx/jwt.go`
- Create: `pkg/authx/jwt_test.go`
- Modify: `pkg/go.mod`

**Interfaces:**
- Produces: `type JWTConfig`, `type Claims`, `type Manager`, `func NewManager(JWTConfig) *Manager`, `func (m *Manager) Generate(Claims) (TokenPair, error)`, `func (m *Manager) ParseBearer(string) (Claims, error)`.

- [ ] Write failing tests for token generation and bearer parsing.
- [ ] Add minimal implementation using HS256.
- [ ] Run `go test ./authx`.

### Task 2: Auth Service

**Files:**
- Create: `services/auth-service/api/auth/v1/auth.proto`
- Generate: `auth.pb.go`, `auth_grpc.pb.go`, `auth_http.pb.go`
- Create: `services/auth-service/cmd/auth/*`
- Create: `services/auth-service/configs/config.yaml`
- Create: `services/auth-service/configs/nacos.yaml`
- Create: `services/auth-service/internal/conf/conf.go`
- Create: `services/auth-service/internal/biz/*`
- Create: `services/auth-service/internal/data/*`
- Create: `services/auth-service/internal/service/*`
- Create: `services/auth-service/internal/server/*`
- Modify: `go.work`

**Interfaces:**
- Produces: `POST /v1/auth/login`, `POST /v1/auth/verify`, service name `auth-service`, ports HTTP `9010`, gRPC `9110`.

- [ ] Write failing biz tests for login create/reuse and role validation.
- [ ] Write failing data tests for account persistence.
- [ ] Implement auth-service.
- [ ] Run `go test ./...` inside `services/auth-service`.

### Task 3: Gateway Auth Integration

**Files:**
- Modify: `services/gateway-service/go.mod`
- Modify: `services/gateway-service/configs/config.yaml`
- Modify: `services/gateway-service/configs/nacos.yaml`
- Modify: `services/gateway-service/internal/conf/conf.go`
- Create: `services/gateway-service/internal/data/auth_client.go`
- Create: `services/gateway-service/internal/biz/auth.go`
- Create: `services/gateway-service/internal/service/auth.go`
- Create: `services/gateway-service/internal/server/auth_filter.go`
- Modify: `services/gateway-service/internal/server/http.go`
- Modify: `services/gateway-service/cmd/gateway/init.go`

**Interfaces:**
- Produces: `POST /carpool/auth/login`, gateway JWT filter, authenticated user id extraction from context/header.

- [ ] Write failing auth client tests.
- [ ] Write failing gateway filter tests.
- [ ] Implement gateway auth proxy and filter.
- [ ] Run `go test ./...` inside `services/gateway-service`.

### Task 4: Frontend Token Adoption

**Files:**
- Modify passenger/driver uni-app request utilities where current API headers are set.
- Add login API wrapper only where it is absent.

**Interfaces:**
- Produces: `Authorization: Bearer <token>` on app API requests while keeping old `X-User-Id` fallback.

- [ ] Locate existing request helpers.
- [ ] Store access token from login response.
- [ ] Attach authorization header in requests.
- [ ] Run available uni-app builds.

### Task 5: Verification

**Files:**
- No production file changes.

- [ ] Run `go test ./configx ./registry ./authx` in `pkg`.
- [ ] Run `go test ./...` in `services/auth-service`.
- [ ] Run `go test ./...` in `services/gateway-service`.
- [ ] Run existing service tests for passenger, driver, trip, order, and review.
- [ ] Scan ports for `8080` and `8848`.
- [ ] Run front-end builds touched by this phase.
