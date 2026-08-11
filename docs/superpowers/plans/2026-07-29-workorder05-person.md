# Workorder05 Person Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build real workorder05 admin APIs and page for staff, driver, passenger, role assignment, import validation, duplicate prevention, masking, and batch status operations.

**Architecture:** Add an isolated admin `person` module under `carpool`, following the existing GVA model/request/service/api/router/frontend pattern. Keep GVA system user and verified workorder01-04 behavior stable.

**Tech Stack:** Go, Gin, GORM, SQLite-compatible GORM tags, Vue3, GVA, Element Plus.

## Global Constraints

- Do not break verified workorder01-04 APIs.
- Use admin GVA response wrapper for management APIs.
- Use SQLite-compatible tags: no `ON UPDATE CURRENT_TIMESTAMP`.
- Use service tests before production service code.
- Sensitive phone and ID card values must be masked in API-facing service responses.
- Import commit must be transactional.
- Batch status operation supports at most 100 IDs.

---

### Task 1: Admin Person Domain

**Files:**
- Create: `admin-platform/server/model/carpool/person.go`
- Create: `admin-platform/server/model/carpool/request/person.go`
- Create: `admin-platform/server/service/carpool/person_test.go`
- Create: `admin-platform/server/service/carpool/person.go`
- Modify: `admin-platform/server/service/carpool/enter.go`
- Modify: `admin-platform/server/initialize/gorm.go`
- Modify: `admin-platform/server/initialize/ensure_tables.go`

**Interfaces:**
- Produces: `PersonService.ListPersons`, `PersonService.GetPersonDetail`, `PersonService.CreatePerson`, `PersonService.UpdatePerson`, `PersonService.AssignRoles`, `PersonService.BatchUpdateStatus`, `PersonService.PreviewImport`, `PersonService.CommitImport`, `PersonService.ListImportErrors`, `PersonService.ExportPersons`.

- [ ] Write failing service tests for masking, duplicate validation, multi-role assignment, batch status, import preview, and transactional import failure.
- [ ] Run `go test -c ./service/carpool -o C:\Users\李小龙\Desktop\RideHailing\.tmp\admin-carpool-person.test.exe` and confirm missing symbols fail.
- [ ] Implement models, request structs, service methods, seed data, status operations, import batch/error persistence, and migration registration.
- [ ] Run admin carpool test compile and `go build ./...`.

### Task 2: Admin Person API And Router

**Files:**
- Create: `admin-platform/server/api/v1/carpool/person.go`
- Create: `admin-platform/server/router/carpool/person.go`
- Modify: `admin-platform/server/api/v1/carpool/enter.go`
- Modify: `admin-platform/server/router/carpool/enter.go`
- Modify: `admin-platform/server/initialize/router_biz.go`

**Interfaces:**
- HTTP:
  - `GET /carpool/person/list`
  - `GET /carpool/person/:id`
  - `POST /carpool/person`
  - `PUT /carpool/person/:id`
  - `POST /carpool/person/roles`
  - `POST /carpool/person/batch/status`
  - `POST /carpool/person/import/preview`
  - `POST /carpool/person/import/commit`
  - `GET /carpool/person/import/errors`
  - `POST /carpool/person/export`

- [ ] Add API handlers with binding validation and GVA responses.
- [ ] Register routes with operation record middleware for write/export actions.
- [ ] Run `go build ./...`.

### Task 3: Admin Frontend Page

**Files:**
- Create: `admin-platform/web/src/api/rideHailing/workorder05.js`
- Create: `admin-platform/web/src/view/admin/workorder05/person.vue`
- Modify: `admin-platform/web/src/router/index.js`
- Modify: `admin-platform/web/src/pathInfo.json`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/index.vue`

**Interfaces:**
- Consumes all admin person APIs from Task 2.

- [ ] Add API wrapper functions.
- [ ] Build person management page with tabs, filters, table selection, create/edit dialog, role dialog, detail drawer, import preview/commit, error table, pagination, and export.
- [ ] Register route and workorder entry.
- [ ] Run admin frontend `npm.cmd run build`.

### Task 4: Full Verification

**Files:**
- No production file edits unless verification exposes defects.

- [ ] Run passenger 02/03/04 service tests and `go build ./...`.
- [ ] Run driver 02 service test, 03/04 test compile, and `go build ./...`.
- [ ] Run admin carpool test compile and `go build ./...`.
- [ ] Run passenger, driver, and admin frontend builds.
- [ ] Summarize implemented files, API counts, function flows, and verification evidence.
