# Auth Session and Ihuyi SMS Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add real Ihuyi SMS verification, refresh-token sessions, logout, and JWT-only gateway authentication.

**Architecture:** `pkg/smsx` wraps Ihuyi HTTP form submission. `auth-service` owns SMS code persistence and refresh sessions. `gateway-service` proxies public auth routes and validates JWT for business routes with temporary `X-User-Id` compatibility disabled.

**Tech Stack:** Go, Kratos v2, GORM/MySQL, Nacos config, `github.com/golang-jwt/jwt/v5`, Ihuyi Submit.json, uni-app.

## Global Constraints

- Do not modify `admin-platform`.
- Do not listen on or register business services at `8080` or `8848`.
- Use Nacos-first config with `configs/nacos.yaml`; use `configs/config.yaml` as fallback.
- Use MySQL database `01ride`.
- Do not commit to git in this phase.
- Store refresh tokens only as SHA-256 hashes in backend persistence.
- Disable gateway `auth.jwt.compatible_header_enabled`.

---

### Task 1: Ihuyi SMS Package

**Files:**
- Create: `pkg/smsx/ihuyi.go`
- Create: `pkg/smsx/ihuyi_test.go`

**Interfaces:**
- Produces: `type IhuyiConfig`, `type IhuyiClient`, `func NewIhuyiClient(IhuyiConfig) *IhuyiClient`, `func (c *IhuyiClient) SendVerificationCode(ctx context.Context, mobile string, code string) error`.

- [ ] Write tests that verify form fields and success code `2`.
- [ ] Implement HTTP form submission and response parsing.
- [ ] Run `go test ./smsx`.

### Task 2: Auth Service SMS and Sessions

**Files:**
- Modify: `services/auth-service/api/auth/v1/auth.proto`
- Generate: `services/auth-service/api/auth/v1/auth.pb.go`
- Generate: `services/auth-service/api/auth/v1/auth_grpc.pb.go`
- Generate: `services/auth-service/api/auth/v1/auth_http.pb.go`
- Modify: `services/auth-service/internal/conf/conf.go`
- Modify: `services/auth-service/configs/config.yaml`
- Modify: `services/auth-service/internal/biz/*.go`
- Modify: `services/auth-service/internal/data/*.go`
- Modify: `services/auth-service/internal/service/auth.go`

**Interfaces:**
- Produces: SMS code send/verify, refresh token creation, refresh, logout, `auth_sms_code`, `auth_session`.

- [ ] Write failing biz tests for code verification, refresh, logout, and rejected reused code.
- [ ] Write failing data tests for SMS code and session persistence.
- [ ] Implement minimal biz/data/service changes.
- [ ] Run `go test ./...` in `services/auth-service`.

### Task 3: Gateway Auth Routes and JWT-Only Mode

**Files:**
- Modify: `services/gateway-service/internal/data/auth_client.go`
- Modify: `services/gateway-service/internal/biz/auth.go`
- Modify: `services/gateway-service/internal/service/auth.go`
- Modify: `services/gateway-service/internal/server/http.go`
- Modify: `services/gateway-service/configs/config.yaml`

**Interfaces:**
- Produces: `/carpool/auth/sms/send`, `/carpool/auth/refresh`, `/carpool/auth/logout`, `compatible_header_enabled: false`.

- [ ] Write failing client and route tests.
- [ ] Implement proxy methods and public auth route allow-list.
- [ ] Run `go test ./...` in `services/gateway-service`.

### Task 4: Uni-App Token Refresh

**Files:**
- Modify: `passenger-platform/uni-app/src/api/auth.js`
- Modify: `passenger-platform/uni-app/src/store/user.js`
- Modify: `passenger-platform/uni-app/src/utils/request.js`
- Modify: `driver-platform/uni-app/src/api/auth.js`
- Modify: `driver-platform/uni-app/src/store/user.js`
- Modify: `driver-platform/uni-app/src/utils/request.js`

**Interfaces:**
- Produces: send SMS API wrappers, refresh/logout API wrappers, token retry on `401`.

- [ ] Add token storage fields.
- [ ] Attach Authorization header only when access token exists.
- [ ] Refresh once on 401 and retry original request.
- [ ] Run both H5 builds.

### Task 5: Verification

- [ ] Run `go test ./smsx ./authx ./configx ./registry` in `pkg`.
- [ ] Run `go test ./...` in `services/auth-service`.
- [ ] Run `go test ./...` in `services/gateway-service`.
- [ ] Run existing service tests for passenger, driver, trip, order, and review.
- [ ] Scan config for reserved business ports.
- [ ] Run passenger and driver uni-app H5 builds.
