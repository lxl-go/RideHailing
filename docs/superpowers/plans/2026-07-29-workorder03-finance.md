# Workorder03 Finance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build real workorder03 finance APIs for passenger refund progress, driver income ledger, and admin finance reconciliation.

**Architecture:** Add isolated `finance` modules to passenger gateway, driver gateway, and admin GVA backend. Use GORM persistence with seed data and keep the module linked by order numbers rather than changing already verified order flows.

**Tech Stack:** Go, Gin, GORM, SQLite-compatible schema, Vue3, Vant, GVA, Element Plus.

## Global Constraints

- Do not break verified workorder01/workorder02 APIs.
- Use unified gateway response `{ code, data, msg }`.
- Use GVA response wrapper for admin backend.
- Use tests before implementation for finance service behavior.
- Keep frontend API files aligned with existing request wrappers.

---

### Task 1: Passenger Refund Progress

**Files:**
- Create: `passenger-platform/service/internal/finance/model/model.go`
- Create: `passenger-platform/service/internal/finance/service/service.go`
- Create: `passenger-platform/service/internal/finance/service/service_test.go`
- Create: `passenger-platform/service/internal/gateway/handler/finance_handler.go`
- Create: `passenger-platform/service/internal/gateway/router/finance_routes.go`
- Modify: `passenger-platform/service/internal/gateway/router/v1.go`
- Modify: `passenger-platform/service/cmd/gateway/main.go`
- Modify: `passenger-platform/web/src/views/passenger/workorder03/refundProgress.vue`
- Create: `passenger-platform/web/src/api/workorder03.ts`

**Interfaces:**
- Produces: `finance.NewFinanceService(db)`, `finance.AutoMigrate(db)`, `finance.SeedDefaults(db)`, `GetRefundProgress(passengerID uint64, orderNo string)`.
- HTTP: `GET /carpool/finance/refunds/progress?orderNo=...`.

- [ ] Write failing service test for refund progress timeline and coupon return.
- [ ] Run passenger finance service test and verify missing symbols fail.
- [ ] Implement passenger finance model/service/migration/seed.
- [ ] Register passenger gateway route and handler.
- [ ] Switch passenger refund page to real API.
- [ ] Run passenger service tests, `go build ./...`, and frontend build.

### Task 2: Driver Income Ledger

**Files:**
- Create: `driver-platform/service/internal/finance/model/model.go`
- Create: `driver-platform/service/internal/finance/service/service.go`
- Create: `driver-platform/service/internal/finance/service/service_test.go`
- Create: `driver-platform/service/internal/gateway/handler/finance_handler.go`
- Create: `driver-platform/service/internal/gateway/router/finance_routes.go`
- Modify: `driver-platform/service/internal/gateway/router/v1.go`
- Modify: `driver-platform/service/cmd/gateway/main.go`
- Modify: `driver-platform/web/src/views/driver/workorder03/incomeLedger.vue`
- Create: `driver-platform/web/src/api/workorder03.ts`

**Interfaces:**
- Produces: `finance.NewFinanceService(db)`, `finance.AutoMigrate(db)`, `finance.SeedDefaults(db)`, `ListIncomeLedger(driverID uint64)`.
- HTTP: `GET /carpool/finance/income/ledger`.

- [ ] Write failing service test for summary totals and per-order ledger rows.
- [ ] Run driver finance service test and verify missing symbols fail.
- [ ] Implement driver finance model/service/migration/seed.
- [ ] Register driver gateway route and handler.
- [ ] Switch driver income page to real API.
- [ ] Run driver service tests, `go build ./...`, and frontend build.

### Task 3: Admin Finance Reconciliation

**Files:**
- Create: `admin-platform/server/model/carpool/finance.go`
- Create: `admin-platform/server/model/carpool/request/finance.go`
- Create: `admin-platform/server/service/carpool/finance.go`
- Create: `admin-platform/server/service/carpool/finance_test.go`
- Create: `admin-platform/server/api/v1/carpool/finance.go`
- Create: `admin-platform/server/router/carpool/finance.go`
- Modify: `admin-platform/server/model/carpool/enter.go` equivalents: service/api/router groups.
- Modify: `admin-platform/server/initialize/gorm.go`
- Modify: `admin-platform/server/initialize/router_biz.go`
- Create: `admin-platform/web/src/api/rideHailing/workorder03.js`
- Create: `admin-platform/web/src/view/admin/workorder03/finance.vue`

**Interfaces:**
- Produces: admin `FinanceService` with list, summary, abnormal list, export task.
- HTTP: `/carpool/finance/transaction/list`, `/carpool/finance/refund/list`, `/carpool/finance/summary`, `/carpool/finance/abnormal/list`, `/carpool/finance/export`.

- [ ] Write failing admin finance service test for summary and abnormal transaction filtering.
- [ ] Run admin carpool test compile to verify missing symbols fail.
- [ ] Implement admin finance model/service/API/router/migration.
- [ ] Add admin frontend page/API following GVA style.
- [ ] Run admin `go build ./...`, test compile, and frontend build.
