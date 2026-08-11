# WO-08 Performance Foundation Design

## Scope

WO-08 focuses on Go performance observability, load-test assets, and acceptance reporting. This round does not implement the full `location-svc` or `message-svc` real-time stack; WebSocket, driver location reporting, route tracking, offline replay, and dispatch push remain WO-11 runtime implementation scope.

This round delivers a reusable performance foundation:

- Admin backend endpoints for performance summary, load-test reports, runtime metric snapshots, and export task metadata.
- `AutoMigrate` model registration for WO-08 performance report tables.
- Admin-web WO-08 page and route under `admin-platform/web`.
- Passenger and driver uni-app performance pages aligned to report-style acceptance data rather than fixed mock labels.
- Load-test script and report directories with concrete JSON examples for HTTP, WebSocket, and driver location scenarios.
- Documentation updates that mark WO-08 as foundation-complete while keeping real WebSocket and location services assigned to WO-11.

## Architecture

The admin-platform `carpool` module remains the delivery home for management-side WO-08, matching WO-06 and WO-07. A new performance service owns persisted report records and runtime snapshot aggregation. It exposes Gin REST endpoints through the existing `carpool` API/router pattern.

Runtime metrics are collected from the current Go process using standard library APIs: `runtime`, `runtime/metrics`, `runtime/debug`, and selected HTTP handler metadata. pprof is not exposed publicly in this task; instead the task records pprof endpoint guidance and report links so production access can stay behind network and permission controls.

Frontend pages read the same report schema. Admin users see report status, scenario targets, runtime snapshots, and artifact links. Passenger and driver pages show WO-08 acceptance cards for WebSocket/map/location scenarios, ready to switch from static examples to backend data when the mobile gateway exposes the report API.

## Data Model

`PerformanceReport` records one load-test or profiling result:

- `report_no`: stable business identifier.
- `workorder_no`: defaults to `WO-08`.
- `scenario`: `admin_http`, `passenger_ws_map`, `driver_location`, `driver_ws_dispatch`, or `mixed`.
- `target_service`: service or app under test.
- `tool`: `jmeter`, `k6`, `wrk`, `pprof`, `trace`, or `manual`.
- `qps`, `p50_ms`, `p90_ms`, `p99_ms`, `error_rate`, `goroutines_before`, `goroutines_after`, `heap_before_mb`, `heap_after_mb`.
- `verdict`: `PASS`, `WARN`, or `FAIL`.
- `artifact_path`: repository path to JSON, JMX, k6 script, pprof, or report output.
- `notes`: short review conclusion.

`PerformanceScenario` records acceptance targets:

- `scenario`: unique code.
- `name`: display name.
- `target_qps`, `target_p99_ms`, `max_error_rate`, `max_goroutine_delta_percent`.
- `scope`: `admin`, `passenger`, `driver`, or `backend`.
- `enabled`: whether the scenario participates in dashboard calculations.

Both models are registered in admin-platform `AutoMigrate` because the current project decision is to keep automatic table migration.

## API Design

Admin endpoints use the existing response wrapper and route group style:

- `GET /carpool/performance/summary`: returns scenario counts, pass rate, latest runtime snapshot, and latest report list.
- `GET /carpool/performance/report/list`: paginated report list with filters for scenario, target service, verdict, and tool.
- `POST /carpool/performance/report`: creates or imports a report record.
- `GET /carpool/performance/scenario/list`: lists acceptance targets.
- `GET /carpool/performance/runtime`: returns runtime snapshot for goroutines, heap, GC, CPU count, and Go version.
- `POST /carpool/performance/export`: returns an export task id, matching WO-06/WO-07 export style.

Mobile apps do not receive new backend endpoints in this round. Their pages are adjusted to the shared report fields and can later call gateway endpoints after WO-11 real-time services exist.

## Data Flow

1. Engineers run scripts under `docs/performance/scripts`.
2. Scripts write report JSON under `docs/performance/reports`.
3. Admin operators import report JSON or create a report entry through the WO-08 page.
4. Admin dashboard calculates pass rate and highlights failing P0 targets.
5. Runtime snapshot endpoint gives current process diagnostics for quick sanity checks.
6. Technical review and difference documents reference the report artifact paths.

## Error Handling

Report creation validates scenario code, target service, tool, verdict, latency numbers, error rate range, and artifact path. Invalid data returns a clear business error and does not create partial records.

Runtime snapshot failure is fail-soft: the endpoint returns available fields and an explanatory warning string if an optional metric cannot be read. It must not panic or expose sensitive process internals beyond aggregate counters.

Export is asynchronous by convention: the API returns a task id even when the export file is not generated immediately. The task id is deterministic enough for tests and traceable in logs.

## Testing

Backend tests cover:

- Report lifecycle: create report, list by scenario/verdict, summary pass-rate calculation.
- Scenario seed behavior: WO-08 default scenarios are present and idempotent.
- Runtime snapshot: required fields are non-empty and numeric fields are sane.
- Validation: invalid verdict, negative latency, and out-of-range error rate are rejected.
- AutoMigrate: performance models are included in initialization tests.

Frontend verification covers:

- Admin-web production build includes WO-08 route and page.
- Passenger and driver uni-app H5 builds still pass.
- Source search confirms old fixed-only mock copy is replaced by report-style acceptance copy.

## Non-Goals

- Building `location-svc`, `message-svc`, Redis GEO, Redis Stream, or WebSocket hub runtime.
- Running real 50k-connection load tests in the local workspace.
- Introducing SQL migration files.
- Replacing existing GVA framework internals or changing global router design.
