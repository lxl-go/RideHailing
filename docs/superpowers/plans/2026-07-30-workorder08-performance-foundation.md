# WO-08 Performance Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build WO-08 performance observability and load-test acceptance foundations without implementing the WO-11 real-time WebSocket/location runtime.

**Architecture:** Add a focused performance module inside admin-platform `carpool`, following WO-06/WO-07 model/service/api/router patterns. Persist performance scenarios and reports through GORM `AutoMigrate`, expose admin REST endpoints, and connect admin-web plus existing uni-app performance pages to report-style acceptance data. Store executable load-test examples and report JSON under a `docs/performance/` directory.

**Tech Stack:** Go, Gin, GORM, SQLite tests, runtime/runtime metrics standard library, Vue 3, Element Plus, uni-app Vue, npm/Vite builds, Markdown docs.

## Global Constraints

- Keep `AutoMigrate`; do not introduce versioned SQL migration files.
- Do not modify workorder DOCX files; read them only as requirements reference.
- Keep WO-08 scoped to performance foundation, scripts, reports, and page entry; full `location-svc`, `message-svc`, Redis GEO, Redis Stream, and WebSocket hub runtime stay in WO-11.
- Add tests before implementation for backend business behavior.
- Use existing `carpool` module conventions and route grouping.
- Admin frontend lives in `admin-platform/web`; passenger and driver uni-apps live in `apps/passenger-uni-app/uni-app` and `apps/driver-uni-app/uni-app`.

---

## File Structure

- Create `admin-platform/server/model/carpool/performance.go`: GORM models for performance reports and scenarios.
- Create `admin-platform/server/model/carpool/request/performance.go`: list and create request structs.
- Create `admin-platform/server/service/carpool/performance.go`: service methods for report lifecycle, summary, scenario seed, runtime snapshot, and export id.
- Create `admin-platform/server/service/carpool/performance_test.go`: test report lifecycle, validation, runtime snapshot, and seed idempotency.
- Create `admin-platform/server/api/v1/carpool/performance.go`: Gin API handlers.
- Create `admin-platform/server/router/carpool/performance.go`: route registration.
- Modify `admin-platform/server/service/carpool/enter.go`: expose `PerformanceService`.
- Modify `admin-platform/server/api/v1/carpool/enter.go`: expose `PerformanceApi`.
- Modify `admin-platform/server/router/carpool/enter.go`: expose `PerformanceRouter`.
- Modify `admin-platform/server/initialize/router_biz.go`: register performance router.
- Modify `admin-platform/server/initialize/gorm.go`: add performance models to `AutoMigrate`.
- Modify `admin-platform/server/initialize/ensure_tables.go`: add performance table ensure checks if the local pattern already lists carpool tables.
- Create `admin-platform/web/src/api/rideHailing/workorder08.js`: admin API wrapper.
- Create `admin-platform/web/src/view/rideHailing/workorder08/performance/index.vue`: WO-08 admin page.
- Modify `admin-platform/web/src/router/index.js`: add static route.
- Modify `admin-platform/web/src/pinia/modules/router.js`: add dynamic menu item.
- Modify `admin-platform/web/src/view/rideHailing/workorders/index.vue`: update WO-08 card and link.
- Modify `admin-platform/web/src/view/rideHailing/workorders/components/WorkorderFilter.vue`: update WO-08 summary.
- Modify `apps/passenger-uni-app/uni-app/src/pages/performance/performance.vue`: align page to acceptance report fields.
- Modify `apps/driver-uni-app/uni-app/src/pages/performance/performance.vue`: align page to acceptance report fields.
- Create `docs/performance/scripts/http-admin-smoke.k6.js`: HTTP smoke load-test example.
- Create `docs/performance/scripts/passenger-ws-map.k6.js`: WebSocket/map scenario example using k6 syntax.
- Create `docs/performance/scripts/driver-location-report.k6.js`: driver location HTTP scenario example.
- Create `docs/performance/reports/wo08-admin-http.sample.json`: sample PASS report.
- Create `docs/performance/reports/wo08-passenger-ws-map.sample.json`: sample WARN report.
- Create `docs/performance/reports/wo08-driver-location.sample.json`: sample PASS report.
- Modify `网约车系统（管理端）需求分析文档.md`: mark WO-08 foundation status.
- Modify `网约车系统技术评审文档.md`: add WO-08 verification row and artifact paths.
- Modify `docs/文档修订差异清单.md`: add WO-08 change record.

---

### Task 1: Backend Performance Service

**Files:**
- Create: `admin-platform/server/model/carpool/performance.go`
- Create: `admin-platform/server/model/carpool/request/performance.go`
- Create: `admin-platform/server/service/carpool/performance.go`
- Create: `admin-platform/server/service/carpool/performance_test.go`
- Modify: `admin-platform/server/service/carpool/enter.go`

**Interfaces:**
- Consumes: `*gorm.DB`, existing `carpool` service package conventions.
- Produces: `type PerformanceService struct{ db *gorm.DB }`, `NewPerformanceService(db *gorm.DB) *PerformanceService`, `CreateReport`, `ListReports`, `GetSummary`, `ListScenarios`, `GetRuntimeSnapshot`, `ExportTaskID`, `SeedPerformanceDefaults`.

- [x] **Step 1: Write failing lifecycle and validation tests**

Create `admin-platform/server/service/carpool/performance_test.go` with tests that use sqlite memory DB, call `AutoMigrate(&carpoolModel.PerformanceReport{}, &carpoolModel.PerformanceScenario{})`, seed scenarios, create one PASS and one WARN report, verify filtered list and summary, and verify invalid verdict/latency/error-rate rejection.

- [x] **Step 2: Run tests to verify RED**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool -run TestPerformanceService -count=1 -v
```

Expected: FAIL because performance types and service do not exist.

- [x] **Step 3: Add GORM models**

Implement `PerformanceReport` and `PerformanceScenario` with table names `performance_report` and `performance_scenario`. Use indexed fields for `report_no`, `workorder_no`, `scenario`, `target_service`, and `verdict`.

- [x] **Step 4: Add request structs**

Implement `PerformanceReportSearch`, `SavePerformanceReportRequest`, and `PerformanceScenarioSearch`. Use JSON/form tags matching existing request style.

- [x] **Step 5: Implement minimal service**

Implement validation, scenario seed defaults, report creation, filtered pagination, summary pass-rate calculation, runtime snapshot using `runtime` and `runtime/metrics`, and export task id generation using a stable `WO08-EXPORT-` prefix.

- [x] **Step 6: Register service entry**

Add `PerformanceService *PerformanceService` to `ServiceGroup` and initialize it in `NewServiceGroup`.

- [x] **Step 7: Run GREEN tests**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool -run TestPerformanceService -count=1 -v
```

Expected: PASS for all `TestPerformanceService...` tests.

---

### Task 2: Backend API, Router, and AutoMigrate

**Files:**
- Create: `admin-platform/server/api/v1/carpool/performance.go`
- Create: `admin-platform/server/router/carpool/performance.go`
- Modify: `admin-platform/server/api/v1/carpool/enter.go`
- Modify: `admin-platform/server/router/carpool/enter.go`
- Modify: `admin-platform/server/initialize/router_biz.go`
- Modify: `admin-platform/server/initialize/gorm.go`
- Modify: `admin-platform/server/initialize/ensure_tables.go`

**Interfaces:**
- Consumes: Task 1 service methods.
- Produces routes:
  - `GET carpool/performance/summary`
  - `GET carpool/performance/report/list`
  - `POST carpool/performance/report`
  - `GET carpool/performance/scenario/list`
  - `GET carpool/performance/runtime`
  - `POST carpool/performance/export`

- [x] **Step 1: Write failing initialization/API compilation test**

Run existing package tests before implementation:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./api/v1/carpool ./router/carpool ./initialize -count=1 -v
```

Expected: PASS before route changes. After adding route references but before files are complete, package compilation would fail, proving the integration points are exercised.

- [x] **Step 2: Add API handlers**

Implement handlers that bind query/body structs, call the service, and use existing `response.OkWithData`, `response.OkWithMessage`, and `response.FailWithMessage` patterns.

- [x] **Step 3: Add router registration**

Create `PerformanceRouter.InitPerformanceRouter(Router *gin.RouterGroup)` and register the six endpoints under `carpool/performance`.

- [x] **Step 4: Wire enter groups**

Expose `PerformanceApi` and `PerformanceRouter` in existing group structs.

- [x] **Step 5: Wire business router**

Call performance router registration in `initialize/router_biz.go` alongside analytics and marketing routers.

- [x] **Step 6: Wire AutoMigrate**

Add `&carpool.PerformanceReport{}` and `&carpool.PerformanceScenario{}` to `initialize/gorm.go`. Add ensure-table entries only if the file already lists carpool model table names.

- [x] **Step 7: Verify integrated backend packages**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool ./api/v1/carpool ./router/carpool ./initialize -count=1 -v
```

Expected: PASS. Existing GORM `record not found` logs from order tests are acceptable if package exit code is 0.

---

### Task 3: Admin-Web WO-08 Page and Route

**Files:**
- Create: `admin-platform/web/src/api/rideHailing/workorder08.js`
- Create: `admin-platform/web/src/view/rideHailing/workorder08/performance/index.vue`
- Modify: `admin-platform/web/src/router/index.js`
- Modify: `admin-platform/web/src/pinia/modules/router.js`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/index.vue`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/components/WorkorderFilter.vue`

**Interfaces:**
- Consumes backend routes from Task 2.
- Produces route `/ride-hailing/workorder08/performance` and component name `RideHailingWorkorder08Performance`.

- [x] **Step 1: Add API wrapper**

Create functions: `getPerformanceSummary`, `listPerformanceReports`, `createPerformanceReport`, `listPerformanceScenarios`, `getRuntimeSnapshot`, and `exportPerformanceReports`.

- [x] **Step 2: Build page**

Create an Element Plus page with KPI cards, runtime snapshot panel, scenario target table, report list table, and create-report dialog. Keep layout consistent with WO-06 and WO-07 management pages.

- [x] **Step 3: Add route and menu**

Add static route after WO-07 and dynamic menu item with title `性能压测`, icon `el-icon-odometer`, and component `view/rideHailing/workorder08/performance/index.vue`.

- [x] **Step 4: Update workorder entry text**

Set WO-08 card to success and link `/ride-hailing/workorder08/performance`. Update overview copy from `工单01-07` to `工单01-08`.

- [x] **Step 5: Verify admin build**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\apps\admin-web\web
npm run build
```

Expected: exit code 0. Known npm config warnings, BigInt transform warnings, and chunk-size warnings are non-blocking.

---

### Task 4: Mobile Performance Pages and Load-Test Artifacts

**Files:**
- Modify: `apps/passenger-uni-app/uni-app/src/pages/performance/performance.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/performance/performance.vue`
- Create: `docs/performance/scripts/http-admin-smoke.k6.js`
- Create: `docs/performance/scripts/passenger-ws-map.k6.js`
- Create: `docs/performance/scripts/driver-location-report.k6.js`
- Create: `docs/performance/reports/wo08-admin-http.sample.json`
- Create: `docs/performance/reports/wo08-passenger-ws-map.sample.json`
- Create: `docs/performance/reports/wo08-driver-location.sample.json`

**Interfaces:**
- Consumes WO-08 report schema from the design.
- Produces repo-local sample artifacts referenced by docs and admin report records.

- [x] **Step 1: Update passenger performance page**

Show report-style rows for `passenger_ws_map`: target, current sample result, verdict, reconnect success, map first-frame, marker update, and artifact path.

- [x] **Step 2: Update driver performance page**

Show report-style rows for `driver_location` and `driver_ws_dispatch`: target QPS/message rate, P99, success rate, goroutine delta, and artifact path.

- [x] **Step 3: Add k6 script examples**

Create one HTTP script for admin summary/list requests, one WebSocket script for passenger map tracking, and one location-report script. Scripts must read target host from environment variables and include default local URLs.

- [x] **Step 4: Add sample report JSON**

Add sample reports matching the backend fields. Use `PASS` for admin HTTP and driver location, `WARN` for passenger WebSocket/map to show dashboard warning behavior.

- [x] **Step 5: Verify uni-app builds**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\apps\passenger-uni-app\uni-app
npm run build:h5
cd C:\Users\李小龙\Desktop\RideHailing\apps\driver-uni-app\uni-app
npm run build:h5
```

Expected: both commands exit code 0. Existing package-manager warnings are non-blocking.

---

### Task 5: Documentation and Final Verification

**Files:**
- Modify: `网约车系统（管理端）需求分析文档.md`
- Modify: `网约车系统技术评审文档.md`
- Modify: `docs/文档修订差异清单.md`

**Interfaces:**
- Consumes verified code and artifact paths from Tasks 1-4.
- Produces updated delivery status for WO-08 foundation.

- [x] **Step 1: Update management requirements document**

Add WO-08 to the current verification matrix as performance foundation connected. Keep real WebSocket/location implementation listed under WO-11.

- [x] **Step 2: Update technical review document**

Add WO-08 backend, admin-web, mobile page, and artifact verification rows. Reference `docs/performance/scripts` and `docs/performance/reports`.

- [x] **Step 3: Update difference list**

Add a new WO-08 section with backend, frontend, mobile, performance artifacts, and documentation changes.

- [x] **Step 4: Run final backend verification**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool ./api/v1/carpool ./router/carpool ./initialize -count=1 -v
```

Expected: PASS.

- [x] **Step 5: Run final frontend verification**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\apps\admin-web\web
npm run build
cd C:\Users\李小龙\Desktop\RideHailing\apps\passenger-uni-app\uni-app
npm run build:h5
cd C:\Users\李小龙\Desktop\RideHailing\apps\driver-uni-app\uni-app
npm run build:h5
```

Expected: all commands exit code 0.

- [x] **Step 6: Search for stale WO-08 copy**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing
rg -n "<stale WO-08 pending-copy patterns>" apps docs "网约车系统（管理端）需求分析文档.md" "网约车系统技术评审文档.md"
```

Expected: no stale “待开发” copy for WO-08 in admin-web source or updated documents; WO-11 real-time scope references may remain.
