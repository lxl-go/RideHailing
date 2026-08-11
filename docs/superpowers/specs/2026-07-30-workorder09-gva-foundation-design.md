# WO-09 GVA Foundation Design

## Scope

WO-09 focuses on the first production-grade GVA framework governance foundation. This round does not rewrite Gin-Vue-Admin internals and does not replace the existing system menu, authority, timed task, router, or AutoMigrate mechanisms.

This round delivers:

- Dynamic menu and route governance metadata: route version, route whitelist snapshot, and refresh/audit status.
- Permission audit enhancement: aggregate data access audit and operation record signals into a WO-09 review surface.
- Multi-datasource governance snapshot: expose active datasource, configured pool targets, and health status without changing the current datasource selection model.
- Timed task / XxlJob-style parameter governance: validate JSON params, executor type, method/http settings, enabled state, latest run status, and rollback guidance using the existing `sys_timed_tasks` module.
- Admin-web WO-09 page and route under `admin-platform/web`.
- Documentation updates that mark WO-09 foundation complete and keep deeper performance tuning, CDN/chunk optimization, and full multi-instance Redis Pub/Sub refresh as later production hardening.

## Non-Goals

- Do not modify GVA framework internals or generated framework conventions.
- Do not replace existing `SysBaseMenu`, `SysAuthority`, `SysTimedTask`, `SysDataAccessLog`, `SysOperationRecord`, or router groups.
- Do not introduce SQL migration files; keep `AutoMigrate`.
- Do not implement a real Redis Pub/Sub or Kafka multi-instance route refresh bus in this round.
- Do not optimize every page chunk, all N+1 risks, or all Redis business caches in this round.

## Architecture

Add a focused WO-09 governance module inside `admin-platform/server/service/system`, exposed through `api/v1/system` and `router/system`. It reads existing GVA system tables and runtime config, then returns normalized review snapshots for the admin page.

The module is read-mostly. It may write a lightweight governance review record if an existing table pattern supports it cleanly, but the core acceptance is available without extra runtime dependency. Existing `AutoMigrate` remains the table synchronization mechanism for any new model added.

Admin-web receives a new route `/ride-hailing/workorder09/gva` and a page that surfaces:

- Route governance cards: menu count, hidden route count, default menu count, route version marker, whitelist status.
- Permission audit cards: data access log count, blocked write count, recent operation records.
- Datasource cards: db type, active db name, pool target values, health state.
- Timed task cards: total tasks, enabled tasks, latest run status, JSON parameter validation status.
- Acceptance checklist: dynamic route whitelist, permission audit fields, datasource health, timed task parameter governance, frontend build result.

## Backend Components

### `GvaGovernanceService`

New service methods:

- `GetGovernanceSummary(ctx context.Context) (*GvaGovernanceSummary, error)`
- `GetRouteSnapshot(ctx context.Context) (*GvaRouteSnapshot, error)`
- `GetAuditSnapshot(ctx context.Context) (*GvaAuditSnapshot, error)`
- `GetDatasourceSnapshot(ctx context.Context) (*GvaDatasourceSnapshot, error)`
- `GetTimedTaskSnapshot(ctx context.Context) (*GvaTimedTaskSnapshot, error)`
- `ExportGovernance(ctx context.Context) string`

The service reads:

- `system.SysBaseMenu`
- `system.SysAuthorityMenu`
- `system.SysDataAccessLog`
- `system.SysOperationRecord`
- `system.SysTimedTask`
- `system.SysTimedTaskLog`
- `global.GVA_CONFIG`
- `global.GVA_DB`

### API

New endpoints under the existing private GVA system router group:

- `GET /system/gva-governance/summary`
- `GET /system/gva-governance/routes`
- `GET /system/gva-governance/audit`
- `GET /system/gva-governance/datasource`
- `GET /system/gva-governance/timed-task`
- `POST /system/gva-governance/export`

All endpoints use the existing response wrapper and GVA middleware. Write-like export uses operation record middleware.

## Frontend Components

### Admin API Wrapper

Create `admin-platform/web/src/api/rideHailing/workorder09.js` with functions matching the six backend endpoints.

### Admin Page

Create `admin-platform/web/src/view/rideHailing/workorder09/gva/index.vue`.

The page follows the WO-06~WO-08 operational style:

- KPI cards at top.
- Snapshot panels for route, audit, datasource, timed task.
- Tables for whitelist entries and recent audit/task signals.
- Export and refresh actions.
- No mock-only business claims: when backend returns empty audit/task data, show empty state as “暂无记录”.

### Route and Menu

Wire:

- Static route in `admin-platform/web/src/router/index.js`
- Dynamic menu in `admin-platform/web/src/pinia/modules/router.js`
- Workorder card in `admin-platform/web/src/view/rideHailing/workorders/index.vue`
- Filter summary in `admin-platform/web/src/view/rideHailing/workorders/components/WorkorderFilter.vue`

WO-09 card becomes “已完成” only after backend tests and admin-web build pass.

## Data Flow

1. Admin opens the WO-09 GVA page.
2. Page requests the governance summary and individual snapshots.
3. Backend reads existing GVA system tables and runtime config.
4. Backend returns normalized metrics and acceptance flags.
5. Admin page renders framework governance status and links the result to WO-09 documentation.
6. Export returns a traceable task id; actual async file generation may be added later with the existing export task pattern.

## Validation Rules

Route governance:

- Route version can be derived from latest menu update time and menu count.
- Whitelist snapshot must list static ride-hailing routes and GVA dynamic menu component paths that are allowed.
- Empty component paths, duplicate route names, or unknown component prefixes are flagged as warnings.

Permission audit:

- `blocked_write`, `no_identity`, and recent operation records are counted separately.
- Audit snapshot includes request id, method, path, user id, authority id, and detail where available.
- Missing trace/request id is marked as a governance warning, not a request failure.

Datasource:

- Snapshot includes current db type, active db name, configured pool targets when available, and ping result.
- If DB ping fails, endpoint returns the snapshot with `healthy=false` and a warning.

Timed task:

- JSON params and HTTP headers are validated when present.
- Invalid cron/JSON/task parameter records appear in `invalidTasks`.
- Disabled tasks are not treated as failures; they are counted separately.

## Error Handling

Read endpoints are fail-soft where possible:

- If one sub-snapshot fails, summary returns other sections and includes warnings.
- DB query errors still return a standard `FailWithMessage` for the affected endpoint.
- Export never claims file generation; it only returns a task id with `WO09-GVA-EXPORT-` prefix.

Sensitive values are not returned:

- Datasource passwords, Redis password, auth secrets, and HTTP task headers such as `Authorization` are masked.

## Testing

Backend tests:

- Summary aggregates route, audit, datasource, and timed task counts from an in-memory sqlite DB.
- Route snapshot detects duplicate route names and invalid component prefixes.
- Timed task snapshot flags invalid JSON params and counts enabled/disabled tasks correctly.
- Datasource snapshot reports current sqlite DB as healthy.
- Export task id uses `WO09-GVA-EXPORT-` prefix.

Frontend verification:

- Admin-web build succeeds after route/menu/page wiring.
- Source search confirms WO-09 no longer appears as “待开发” in admin-workorder entry.

Documentation:

- Management requirements document marks WO-09 foundation as connected.
- Technical review document adds WO-09 verification rows and keeps deep route refresh bus/performance tuning as later hardening.
- Difference list records WO-09 backend, frontend, docs, and verification changes.

## Acceptance Criteria

- `go test ./service/system -run TestGvaGovernanceService -count=1 -v` exits 0.
- `go test ./api/v1/system ./router/system ./initialize -count=1 -v` exits 0.
- `npm run build` in `admin-platform/web` exits 0.
- WO-09 route is reachable through `/ride-hailing/workorder09/gva`.
- WO-09 workorder card shows a real link after verification.
- Docs clearly state that `AutoMigrate` remains the current table migration mode.
- Docs clearly state that full Redis Pub/Sub/Kafka route refresh and deeper frontend chunk optimization are later production hardening, not this foundation scope.
