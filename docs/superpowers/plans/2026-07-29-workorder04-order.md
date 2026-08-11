# Workorder04 Order Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build real workorder04 order management APIs and pages for order query, detail, refund application, manual review, batch refund, and audit history.

**Architecture:** Add isolated order management modules using the existing GVA admin pattern and lightweight passenger/driver gateway modules. Keep verified workorder01-03 flows stable and link cross-module data by `orderNo`.

**Tech Stack:** Go, Gin, GORM, SQLite-compatible GORM tags, Vue3, Vant, GVA, Element Plus.

## Global Constraints

- Do not break verified workorder01, workorder02, or workorder03 APIs.
- Use admin GVA response wrapper for management APIs.
- Use passenger/driver gateway response `{ code, data, msg }`.
- Use TDD for service behavior before production code.
- Use SQLite-compatible tags: no `ON UPDATE CURRENT_TIMESTAMP`.
- Workorder04 code comments should mention workorder04 where business rules are not self-evident.
- Batch refund maximum is 100 order numbers.

---

### Task 1: Admin Order Domain

**Files:**
- Create: `admin-platform/server/model/carpool/order.go`
- Create: `admin-platform/server/model/carpool/request/order.go`
- Create: `admin-platform/server/service/carpool/order_test.go`
- Create: `admin-platform/server/service/carpool/order.go`
- Modify: `admin-platform/server/service/carpool/enter.go`
- Modify: `admin-platform/server/initialize/gorm.go`
- Modify: `admin-platform/server/initialize/ensure_tables.go`

**Interfaces:**
- Produces: `OrderService.ListOrders`, `OrderService.GetOrderDetail`, `OrderService.ListRefunds`, `OrderService.ApplyRefund`, `OrderService.ReviewRefund`, `OrderService.BatchRefund`, `OrderService.ExportOrders`.

- [ ] Write failing admin service tests for idempotent refund, fee calculation, completed-order manual review, status history, and batch refund.
- [ ] Run `go test -c ./service/carpool -o C:\Users\李小龙\Desktop\RideHailing\.tmp\admin-carpool-order.test.exe` and confirm missing symbols fail.
- [ ] Implement models, request structs, service methods, seed data, status history writes, and table migration registration.
- [ ] Run admin carpool test compile and `go build ./...`.

### Task 2: Admin Order API And Router

**Files:**
- Create: `admin-platform/server/api/v1/carpool/order.go`
- Create: `admin-platform/server/router/carpool/order.go`
- Modify: `admin-platform/server/api/v1/carpool/enter.go`
- Modify: `admin-platform/server/router/carpool/enter.go`
- Modify: `admin-platform/server/initialize/router_biz.go`

**Interfaces:**
- HTTP:
  - `GET /carpool/order/list`
  - `GET /carpool/order/:orderNo`
  - `GET /carpool/order/:orderNo/history`
  - `GET /carpool/order/refund/list`
  - `POST /carpool/order/refund/apply`
  - `POST /carpool/order/refund/review`
  - `POST /carpool/order/refund/batch`
  - `POST /carpool/order/export`

- [ ] Add API handlers with binding validation and GVA responses.
- [ ] Register routes with operation record middleware for write/export actions.
- [ ] Run `go build ./...`.

### Task 3: Passenger Refund Gateway

**Files:**
- Create: `passenger-platform/service/internal/workorder04/model/model.go`
- Create: `passenger-platform/service/internal/workorder04/service/service_test.go`
- Create: `passenger-platform/service/internal/workorder04/service/service.go`
- Create: `passenger-platform/service/internal/gateway/handler/workorder04_handler.go`
- Create: `passenger-platform/service/internal/gateway/router/workorder04_routes.go`
- Modify: `passenger-platform/service/internal/gateway/router/v1.go`
- Modify: `passenger-platform/service/cmd/gateway/main.go`

**Interfaces:**
- HTTP:
  - `POST /carpool/orders/:id/refund`
  - `GET /carpool/orders/:id/refund`

- [ ] Write failing passenger service tests for idempotent refund apply and refund status.
- [ ] Implement SQLite-compatible model, service, seed, handler, and routes.
- [ ] Run passenger workorder04 tests and `go build ./...`.

### Task 4: Driver Order Read Gateway

**Files:**
- Create: `driver-platform/service/internal/workorder04/model/model.go`
- Create: `driver-platform/service/internal/workorder04/service/service_test.go`
- Create: `driver-platform/service/internal/workorder04/service/service.go`
- Create: `driver-platform/service/internal/gateway/handler/workorder04_handler.go`
- Create: `driver-platform/service/internal/gateway/router/workorder04_routes.go`
- Modify: `driver-platform/service/internal/gateway/router/v1.go`
- Modify: `driver-platform/service/cmd/gateway/main.go`

**Interfaces:**
- HTTP:
  - `GET /carpool/orders/mine`
  - `GET /carpool/orders/:id`

- [ ] Write failing driver service tests for order list/detail visibility.
- [ ] Implement model, service, seed, handler, and routes.
- [ ] Run driver workorder04 test compile and `go build ./...`.

### Task 5: Admin Frontend Page

**Files:**
- Create: `admin-platform/web/src/api/rideHailing/workorder04.js`
- Create: `admin-platform/web/src/view/admin/workorder04/order.vue`
- Modify: `admin-platform/web/src/router/index.js`
- Modify: `admin-platform/web/src/pathInfo.json`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/index.vue`

**Interfaces:**
- Consumes all admin order APIs from Task 2.

- [ ] Add API wrapper functions for list/detail/history/refund/review/batch/export.
- [ ] Build order management page with filters, table selection, detail drawer, refund tab, refund dialog, review dialog, batch refund, pagination, and export.
- [ ] Register route and workorder entry.
- [ ] Run admin frontend `npm.cmd run build`.

### Task 6: Full Verification

**Files:**
- No production file edits unless verification exposes defects.

- [ ] Run passenger 02/03/04 service tests and `go build ./...`.
- [ ] Run driver 02 service test, 03/04 test compile, and `go build ./...`.
- [ ] Run admin carpool test compile and `go build ./...`.
- [ ] Run passenger, driver, and admin frontend builds.
- [ ] Summarize implemented files, API counts, function flows, and verification evidence.
