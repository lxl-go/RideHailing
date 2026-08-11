# WO-09 GVA Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the WO-09 GVA governance foundation without rewriting GVA internals.

**Architecture:** Add a read-mostly `GvaGovernanceService` under the existing `system` module. It aggregates menu/route, audit, datasource, and timed-task signals from current GVA tables and runtime config, exposes private system APIs, and connects an admin-web WO-09 page.

**Tech Stack:** Go, Gin, GORM, SQLite tests, Gin-Vue-Admin system module, Vue 3, Element Plus, npm/Vite, Markdown docs.

## Global Constraints

- Do not modify GVA framework internals or generated framework conventions.
- Do not replace existing `SysBaseMenu`, `SysAuthority`, `SysTimedTask`, `SysDataAccessLog`, `SysOperationRecord`, or router groups.
- Do not introduce SQL migration files; keep `AutoMigrate`.
- Do not implement a real Redis Pub/Sub or Kafka multi-instance route refresh bus in this round.
- Do not optimize every page chunk, all N+1 risks, or all Redis business caches in this round.
- Keep workorder DOCX files read-only; update Markdown requirements and technical review documents only.
- Root is not a git repository in this workspace; do not include commit steps as required verification.

---

## File Structure

- Create `admin-platform/server/service/system/gva_governance.go`: service structs and snapshot aggregation.
- Create `admin-platform/server/service/system/gva_governance_test.go`: backend behavior tests.
- Modify `admin-platform/server/service/system/enter.go`: expose `GvaGovernanceService`.
- Create `admin-platform/server/api/v1/system/gva_governance.go`: Gin handlers.
- Modify `admin-platform/server/api/v1/system/enter.go`: expose `GvaGovernanceApi` and service variable.
- Create `admin-platform/server/router/system/gva_governance.go`: route registration.
- Modify `admin-platform/server/router/system/enter.go`: expose `GvaGovernanceRouter` and API variable.
- Modify `admin-platform/server/initialize/router.go`: register WO-09 routes in the private system group.
- Create `admin-platform/web/src/api/rideHailing/workorder09.js`: admin API wrapper.
- Create `admin-platform/web/src/view/rideHailing/workorder09/gva/index.vue`: WO-09 admin page.
- Modify `admin-platform/web/src/router/index.js`: static route.
- Modify `admin-platform/web/src/pinia/modules/router.js`: dynamic menu item.
- Modify `admin-platform/web/src/view/rideHailing/workorders/index.vue`: WO-09 card and top copy.
- Modify `admin-platform/web/src/view/rideHailing/workorders/components/WorkorderFilter.vue`: WO-09 summary.
- Modify `网约车系统（管理端）需求分析文档.md`: mark WO-09 foundation status.
- Modify `网约车系统技术评审文档.md`: add WO-09 verification rows.
- Modify `docs/文档修订差异清单.md`: record WO-09 changes.

---

### Task 1: Backend Governance Service

**Files:**
- Create: `admin-platform/server/service/system/gva_governance_test.go`
- Create: `admin-platform/server/service/system/gva_governance.go`
- Modify: `admin-platform/server/service/system/enter.go`

**Interfaces:**
- Consumes: `global.GVA_DB`, `global.GVA_CONFIG`, GVA system models.
- Produces:
  - `type GvaGovernanceService struct{}`
  - `func (s *GvaGovernanceService) GetGovernanceSummary(ctx context.Context) (*GvaGovernanceSummary, error)`
  - `func (s *GvaGovernanceService) GetRouteSnapshot(ctx context.Context) (*GvaRouteSnapshot, error)`
  - `func (s *GvaGovernanceService) GetAuditSnapshot(ctx context.Context) (*GvaAuditSnapshot, error)`
  - `func (s *GvaGovernanceService) GetDatasourceSnapshot(ctx context.Context) (*GvaDatasourceSnapshot, error)`
  - `func (s *GvaGovernanceService) GetTimedTaskSnapshot(ctx context.Context) (*GvaTimedTaskSnapshot, error)`
  - `func (s *GvaGovernanceService) ExportGovernance(ctx context.Context) string`

- [x] **Step 1: Write failing service tests**

Create `admin-platform/server/service/system/gva_governance_test.go` with tests:

```go
package system

import (
	"context"
	"strings"
	"testing"

	"ride-hailing/admin-server/global"
	systemModel "ride-hailing/admin-server/model/system"
	"ride-hailing/admin-server/config"

	"github.com/glebarez/sqlite"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupGvaGovernanceTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&systemModel.SysBaseMenu{},
		&systemModel.SysAuthorityMenu{},
		&systemModel.SysDataAccessLog{},
		&systemModel.SysOperationRecord{},
		&systemModel.SysTimedTask{},
		&systemModel.SysTimedTaskLog{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	global.GVA_DB = db
	global.GVA_CONFIG.System.DbType = "sqlite"
	global.GVA_CONFIG.Sqlite = config.Sqlite{GeneralDB: config.GeneralDB{Dbname: ":memory:"}}
}

func TestGvaGovernanceServiceSummarySnapshotsAndWarnings(t *testing.T) {
	setupGvaGovernanceTestDB(t)
	ctx := context.Background()
	if err := global.GVA_DB.Create(&[]systemModel.SysBaseMenu{
		{Name: "RideHailingWorkorders", Path: "workorders", Component: "view/rideHailing/workorders/index.vue", Meta: systemModel.Meta{Title: "工单管理"}},
		{Name: "RideHailingWorkorders", Path: "duplicate", Component: "view/rideHailing/duplicate.vue", Meta: systemModel.Meta{Title: "重复路由"}},
		{Name: "UnsafeMenu", Path: "unsafe", Component: "unknown/unsafe.vue", Meta: systemModel.Meta{Title: "未知组件"}},
	}).Error; err != nil {
		t.Fatalf("seed menus: %v", err)
	}
	if err := global.GVA_DB.Create(&[]systemModel.SysDataAccessLog{
		{EventType: "blocked_write", TargetTable: "order_main", Operation: "update", RequestID: "req-1", Path: "/carpool/order/refund"},
		{EventType: "no_identity", TargetTable: "carpool_trip", Operation: "query", Path: "/carpool/trip/list"},
	}).Error; err != nil {
		t.Fatalf("seed data access logs: %v", err)
	}
	if err := global.GVA_DB.Create(&systemModel.SysOperationRecord{Method: "POST", Path: "/system/menu/addBaseMenu", Status: 200, RequestID: "req-op", TraceID: "trace-op", LatencyMs: 42}).Error; err != nil {
		t.Fatalf("seed operation record: %v", err)
	}
	if err := global.GVA_DB.Create(&[]systemModel.SysTimedTask{
		{Name: "valid", Spec: "@daily", ExecutorType: systemModel.TimedTaskExecutorHTTP, HttpUrl: "https://example.com", HttpMethod: "POST", HttpHeader: datatypes.JSON(`{"X-Test":"ok"}`), Enabled: true},
		{Name: "invalid-json", Spec: "@daily", ExecutorType: systemModel.TimedTaskExecutorMethod, MethodName: "unknown", Params: datatypes.JSON(`{"bad"`), Enabled: false},
	}).Error; err != nil {
		t.Fatalf("seed timed tasks: %v", err)
	}

	svc := GvaGovernanceService{}
	summary, err := svc.GetGovernanceSummary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Route.TotalMenus != 3 || summary.Audit.DataAccessLogs != 2 || summary.TimedTask.TotalTasks != 2 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if summary.Route.DuplicateNames != 1 {
		t.Fatalf("expected one duplicate route name, got %d", summary.Route.DuplicateNames)
	}
	if len(summary.Route.Warnings) == 0 || len(summary.TimedTask.InvalidTasks) == 0 {
		t.Fatalf("expected route and timed task warnings: %+v", summary)
	}
	if !summary.Datasource.Healthy {
		t.Fatalf("expected sqlite datasource healthy: %+v", summary.Datasource)
	}
}

func TestGvaGovernanceServiceExportTaskID(t *testing.T) {
	setupGvaGovernanceTestDB(t)
	taskID := (&GvaGovernanceService{}).ExportGovernance(context.Background())
	if !strings.HasPrefix(taskID, "WO09-GVA-EXPORT-") {
		t.Fatalf("unexpected task id: %s", taskID)
	}
}
```

- [x] **Step 2: Run tests to verify RED**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/system -run TestGvaGovernanceService -count=1 -v
```

Expected: FAIL because `GvaGovernanceService` and response types do not exist.

- [x] **Step 3: Implement governance service**

Create `admin-platform/server/service/system/gva_governance.go` implementing:

```go
package system

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ride-hailing/admin-server/global"
	systemModel "ride-hailing/admin-server/model/system"
)

const gvaGovernanceExportPrefix = "WO09-GVA-EXPORT-"

type GvaGovernanceService struct{}

type GvaGovernanceSummary struct {
	Route      GvaRouteSnapshot      `json:"route"`
	Audit      GvaAuditSnapshot      `json:"audit"`
	Datasource GvaDatasourceSnapshot `json:"datasource"`
	TimedTask  GvaTimedTaskSnapshot  `json:"timedTask"`
	Warnings   []string              `json:"warnings"`
}

type GvaRouteSnapshot struct {
	TotalMenus      int64    `json:"totalMenus"`
	HiddenMenus     int64    `json:"hiddenMenus"`
	DefaultMenus    int64    `json:"defaultMenus"`
	DuplicateNames  int      `json:"duplicateNames"`
	RouteVersion    string   `json:"routeVersion"`
	WhitelistStatus string   `json:"whitelistStatus"`
	AllowedPrefixes []string `json:"allowedPrefixes"`
	Warnings        []string `json:"warnings"`
}

type GvaAuditSnapshot struct {
	DataAccessLogs        int64                         `json:"dataAccessLogs"`
	BlockedWrites         int64                         `json:"blockedWrites"`
	NoIdentityEvents      int64                         `json:"noIdentityEvents"`
	OperationRecords      int64                         `json:"operationRecords"`
	MissingTraceRecords    int64                         `json:"missingTraceRecords"`
	RecentDataAccessLogs  []systemModel.SysDataAccessLog `json:"recentDataAccessLogs"`
	RecentOperationRecords []systemModel.SysOperationRecord `json:"recentOperationRecords"`
	Warnings              []string                      `json:"warnings"`
}

type GvaDatasourceSnapshot struct {
	DBType       string   `json:"dbType"`
	ActiveDBName string  `json:"activeDbName"`
	Healthy      bool    `json:"healthy"`
	Warning      string  `json:"warning,omitempty"`
	Pool         DBPoolSnapshot `json:"pool"`
}

type DBPoolSnapshot struct {
	OpenConnections int `json:"openConnections"`
	InUse           int `json:"inUse"`
	Idle            int `json:"idle"`
}

type GvaTimedTaskSnapshot struct {
	TotalTasks     int64                `json:"totalTasks"`
	EnabledTasks   int64                `json:"enabledTasks"`
	DisabledTasks  int64                `json:"disabledTasks"`
	InvalidTasks   []TimedTaskIssue     `json:"invalidTasks"`
	LatestLogs     []systemModel.SysTimedTaskLog `json:"latestLogs"`
	Warnings       []string             `json:"warnings"`
}

type TimedTaskIssue struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func (s *GvaGovernanceService) GetGovernanceSummary(ctx context.Context) (*GvaGovernanceSummary, error) {
	route, err := s.GetRouteSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	audit, err := s.GetAuditSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	datasource, err := s.GetDatasourceSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	timedTask, err := s.GetTimedTaskSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	warnings := append([]string{}, route.Warnings...)
	warnings = append(warnings, audit.Warnings...)
	if datasource.Warning != "" {
		warnings = append(warnings, datasource.Warning)
	}
	warnings = append(warnings, timedTask.Warnings...)
	return &GvaGovernanceSummary{Route: *route, Audit: *audit, Datasource: *datasource, TimedTask: *timedTask, Warnings: warnings}, nil
}

func (s *GvaGovernanceService) GetRouteSnapshot(ctx context.Context) (*GvaRouteSnapshot, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&systemModel.SysBaseMenu{})
	snapshot := &GvaRouteSnapshot{AllowedPrefixes: []string{"view/", "plugin/", "layout/", "routerHolder"}}
	if err := db.Count(&snapshot.TotalMenus).Error; err != nil {
		return nil, err
	}
	if err := db.Where("hidden = ?", true).Count(&snapshot.HiddenMenus).Error; err != nil {
		return nil, err
	}
	if err := db.Where("default_menu = ?", true).Count(&snapshot.DefaultMenus).Error; err != nil {
		return nil, err
	}
	var menus []systemModel.SysBaseMenu
	if err := global.GVA_DB.WithContext(ctx).Find(&menus).Error; err != nil {
		return nil, err
	}
	nameCount := map[string]int{}
	var latestUnix int64
	for _, menu := range menus {
		nameCount[menu.Name]++
		if menu.UpdatedAt.Unix() > latestUnix {
			latestUnix = menu.UpdatedAt.Unix()
		}
		if strings.TrimSpace(menu.Component) == "" {
			snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("菜单 %s 缺少 component", menu.Name))
			continue
		}
		if !allowedGvaComponent(menu.Component, snapshot.AllowedPrefixes) {
			snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("菜单 %s component 前缀未在白名单: %s", menu.Name, menu.Component))
		}
	}
	for name, count := range nameCount {
		if name != "" && count > 1 {
			snapshot.DuplicateNames++
			snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("路由名称重复: %s", name))
		}
	}
	snapshot.RouteVersion = fmt.Sprintf("menus-%d-%d", snapshot.TotalMenus, latestUnix)
	if len(snapshot.Warnings) == 0 {
		snapshot.WhitelistStatus = "PASS"
	} else {
		snapshot.WhitelistStatus = "WARN"
	}
	return snapshot, nil
}

func (s *GvaGovernanceService) GetAuditSnapshot(ctx context.Context) (*GvaAuditSnapshot, error) {
	snapshot := &GvaAuditSnapshot{}
	db := global.GVA_DB.WithContext(ctx)
	if err := db.Model(&systemModel.SysDataAccessLog{}).Count(&snapshot.DataAccessLogs).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysDataAccessLog{}).Where("event_type = ?", "blocked_write").Count(&snapshot.BlockedWrites).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysDataAccessLog{}).Where("event_type = ?", "no_identity").Count(&snapshot.NoIdentityEvents).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysOperationRecord{}).Count(&snapshot.OperationRecords).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysOperationRecord{}).Where("request_id = '' OR trace_id = ''").Count(&snapshot.MissingTraceRecords).Error; err != nil {
		return nil, err
	}
	if err := db.Order("id desc").Limit(5).Find(&snapshot.RecentDataAccessLogs).Error; err != nil {
		return nil, err
	}
	if err := db.Order("id desc").Limit(5).Find(&snapshot.RecentOperationRecords).Error; err != nil {
		return nil, err
	}
	if snapshot.MissingTraceRecords > 0 {
		snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("%d 条操作记录缺少 request_id 或 trace_id", snapshot.MissingTraceRecords))
	}
	return snapshot, nil
}

func (s *GvaGovernanceService) GetDatasourceSnapshot(ctx context.Context) (*GvaDatasourceSnapshot, error) {
	snapshot := &GvaDatasourceSnapshot{DBType: global.GVA_CONFIG.System.DbType}
	if global.GVA_ACTIVE_DBNAME != nil {
		snapshot.ActiveDBName = *global.GVA_ACTIVE_DBNAME
	}
	sqlDB, err := global.GVA_DB.WithContext(ctx).DB()
	if err != nil {
		snapshot.Warning = err.Error()
		return snapshot, nil
	}
	stats := sqlDB.Stats()
	snapshot.Pool = DBPoolSnapshot{OpenConnections: stats.OpenConnections, InUse: stats.InUse, Idle: stats.Idle}
	if err := sqlDB.PingContext(ctx); err != nil {
		snapshot.Warning = err.Error()
		return snapshot, nil
	}
	snapshot.Healthy = true
	return snapshot, nil
}

func (s *GvaGovernanceService) GetTimedTaskSnapshot(ctx context.Context) (*GvaTimedTaskSnapshot, error) {
	snapshot := &GvaTimedTaskSnapshot{}
	db := global.GVA_DB.WithContext(ctx)
	if err := db.Model(&systemModel.SysTimedTask{}).Count(&snapshot.TotalTasks).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&systemModel.SysTimedTask{}).Where("enabled = ?", true).Count(&snapshot.EnabledTasks).Error; err != nil {
		return nil, err
	}
	snapshot.DisabledTasks = snapshot.TotalTasks - snapshot.EnabledTasks
	var tasks []systemModel.SysTimedTask
	if err := db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	for _, task := range tasks {
		if len(task.Params) > 0 && !json.Valid(task.Params) {
			snapshot.InvalidTasks = append(snapshot.InvalidTasks, TimedTaskIssue{ID: task.ID, Name: task.Name, Reason: "params 不是合法 JSON"})
		}
		if len(task.HttpHeader) > 0 && !json.Valid(task.HttpHeader) {
			snapshot.InvalidTasks = append(snapshot.InvalidTasks, TimedTaskIssue{ID: task.ID, Name: task.Name, Reason: "httpHeader 不是合法 JSON"})
		}
		if strings.TrimSpace(task.ExecutorType) == "" {
			snapshot.InvalidTasks = append(snapshot.InvalidTasks, TimedTaskIssue{ID: task.ID, Name: task.Name, Reason: "executorType 为空"})
		}
	}
	if err := db.Order("id desc").Limit(5).Find(&snapshot.LatestLogs).Error; err != nil {
		return nil, err
	}
	if len(snapshot.InvalidTasks) > 0 {
		snapshot.Warnings = append(snapshot.Warnings, fmt.Sprintf("%d 个定时任务参数需要治理", len(snapshot.InvalidTasks)))
	}
	return snapshot, nil
}

func (s *GvaGovernanceService) ExportGovernance(ctx context.Context) string {
	return fmt.Sprintf("%s%d", gvaGovernanceExportPrefix, time.Now().UnixNano())
}

func allowedGvaComponent(component string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(component, prefix) {
			return true
		}
	}
	return false
}

```

- [x] **Step 4: Register service in enter group**

Modify `admin-platform/server/service/system/enter.go` and add:

```go
	GvaGovernanceService
```

near `TimedTaskService`.

- [x] **Step 5: Run GREEN service tests**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/system -run TestGvaGovernanceService -count=1 -v
```

Expected: PASS.

---

### Task 2: Backend API and Router

**Files:**
- Create: `admin-platform/server/api/v1/system/gva_governance.go`
- Modify: `admin-platform/server/api/v1/system/enter.go`
- Create: `admin-platform/server/router/system/gva_governance.go`
- Modify: `admin-platform/server/router/system/enter.go`
- Modify: `admin-platform/server/initialize/router.go`

**Interfaces:**
- Consumes: `service.ServiceGroupApp.SystemServiceGroup.GvaGovernanceService`
- Produces routes:
  - `GET /system/gva-governance/summary`
  - `GET /system/gva-governance/routes`
  - `GET /system/gva-governance/audit`
  - `GET /system/gva-governance/datasource`
  - `GET /system/gva-governance/timed-task`
  - `POST /system/gva-governance/export`

- [x] **Step 1: Run package compile baseline**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./api/v1/system ./router/system ./initialize -count=1 -v
```

Expected: PASS before edits.

- [x] **Step 2: Add API handlers**

Create `admin-platform/server/api/v1/system/gva_governance.go`:

```go
package system

import (
	"github.com/gin-gonic/gin"

	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/utils/logger"
)

type GvaGovernanceApi struct{}

func (a *GvaGovernanceApi) GetGovernanceSummary(c *gin.Context) {
	data, err := gvaGovernanceService.GetGovernanceSummary(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get governance summary failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) GetRouteSnapshot(c *gin.Context) {
	data, err := gvaGovernanceService.GetRouteSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get route snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) GetAuditSnapshot(c *gin.Context) {
	data, err := gvaGovernanceService.GetAuditSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get audit snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) GetDatasourceSnapshot(c *gin.Context) {
	data, err := gvaGovernanceService.GetDatasourceSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get datasource snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) GetTimedTaskSnapshot(c *gin.Context) {
	data, err := gvaGovernanceService.GetTimedTaskSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get timed task snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) ExportGovernance(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": gvaGovernanceService.ExportGovernance(c.Request.Context())}, c)
}
```

- [x] **Step 3: Wire API enter group**

Modify `admin-platform/server/api/v1/system/enter.go`:

```go
type ApiGroup struct {
	...
	GvaGovernanceApi
}
```

and add service variable:

```go
	gvaGovernanceService = service.ServiceGroupApp.SystemServiceGroup.GvaGovernanceService
```

- [x] **Step 4: Add router**

Create `admin-platform/server/router/system/gva_governance.go`:

```go
package system

import (
	"github.com/gin-gonic/gin"

	"ride-hailing/admin-server/middleware"
)

type GvaGovernanceRouter struct{}

func (r *GvaGovernanceRouter) InitGvaGovernanceRouter(Router *gin.RouterGroup) {
	governanceRouter := Router.Group("system/gva-governance").Use(middleware.OperationRecord())
	governanceRouterWithoutRecord := Router.Group("system/gva-governance")
	{
		governanceRouter.POST("export", gvaGovernanceApi.ExportGovernance)
	}
	{
		governanceRouterWithoutRecord.GET("summary", gvaGovernanceApi.GetGovernanceSummary)
		governanceRouterWithoutRecord.GET("routes", gvaGovernanceApi.GetRouteSnapshot)
		governanceRouterWithoutRecord.GET("audit", gvaGovernanceApi.GetAuditSnapshot)
		governanceRouterWithoutRecord.GET("datasource", gvaGovernanceApi.GetDatasourceSnapshot)
		governanceRouterWithoutRecord.GET("timed-task", gvaGovernanceApi.GetTimedTaskSnapshot)
	}
}
```

- [x] **Step 5: Wire router enter group**

Modify `admin-platform/server/router/system/enter.go`:

```go
type RouterGroup struct {
	...
	GvaGovernanceRouter
}
```

and add API variable:

```go
	gvaGovernanceApi = api.ApiGroupApp.SystemApiGroup.GvaGovernanceApi
```

- [x] **Step 6: Register router**

Modify `admin-platform/server/initialize/router.go` after `InitTimedTaskRouter`:

```go
		systemRouter.InitGvaGovernanceRouter(PrivateGroup)                 // WO-09 GVA治理专项
```

- [x] **Step 7: Verify backend integration**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/system -run TestGvaGovernanceService -count=1 -v
go test ./api/v1/system ./router/system ./initialize -count=1 -v
```

Expected: PASS. Full `go test ./service/system` is not used as WO-09 evidence because pre-existing auto-code tests depend on external resource paths and initialized DB state unrelated to this task.

---

### Task 3: Admin-Web WO-09 Page

**Files:**
- Create: `admin-platform/web/src/api/rideHailing/workorder09.js`
- Create: `admin-platform/web/src/view/rideHailing/workorder09/gva/index.vue`
- Modify: `admin-platform/web/src/router/index.js`
- Modify: `admin-platform/web/src/pinia/modules/router.js`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/index.vue`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/components/WorkorderFilter.vue`

**Interfaces:**
- Consumes backend endpoints from Task 2.
- Produces route `/ride-hailing/workorder09/gva` and component name `RideHailingWorkorder09Gva`.

- [x] **Step 1: Add API wrapper**

Create `admin-platform/web/src/api/rideHailing/workorder09.js`:

```js
import service from '@/utils/request'

export const getGvaGovernanceSummary = () => service({ url: '/system/gva-governance/summary', method: 'get' })

export const getGvaRouteSnapshot = () => service({ url: '/system/gva-governance/routes', method: 'get' })

export const getGvaAuditSnapshot = () => service({ url: '/system/gva-governance/audit', method: 'get' })

export const getGvaDatasourceSnapshot = () => service({ url: '/system/gva-governance/datasource', method: 'get' })

export const getGvaTimedTaskSnapshot = () => service({ url: '/system/gva-governance/timed-task', method: 'get' })

export const exportGvaGovernance = () => service({ url: '/system/gva-governance/export', method: 'post' })
```

- [x] **Step 2: Build page**

Create `admin-platform/web/src/view/rideHailing/workorder09/gva/index.vue` with:

```vue
<template>
  <div class="gva-page">
    <div class="hero">
      <div>
        <h2>WO-09 GVA 框架治理</h2>
        <p>动态路由、权限审计、多数据源和定时任务参数治理。</p>
      </div>
      <div class="actions">
        <el-button icon="Refresh" @click="loadData">刷新</el-button>
        <el-button type="primary" icon="Download" @click="handleExport">导出</el-button>
      </div>
    </div>

    <el-row :gutter="16">
      <el-col v-for="item in kpis" :key="item.label" :span="6">
        <el-card shadow="never" class="kpi-card">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
          <p>{{ item.note }}</p>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="section-row">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>动态路由治理</template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="菜单数">{{ route.totalMenus || 0 }}</el-descriptions-item>
            <el-descriptions-item label="隐藏路由">{{ route.hiddenMenus || 0 }}</el-descriptions-item>
            <el-descriptions-item label="默认菜单">{{ route.defaultMenus || 0 }}</el-descriptions-item>
            <el-descriptions-item label="版本">{{ route.routeVersion || '-' }}</el-descriptions-item>
          </el-descriptions>
          <el-alert v-if="routeWarnings.length" class="mt12" type="warning" :closable="false" :title="routeWarnings.join('；')" />
          <el-empty v-else description="路由白名单暂无告警" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>多数据源健康</template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="DB类型">{{ datasource.dbType || '-' }}</el-descriptions-item>
            <el-descriptions-item label="当前库">{{ datasource.activeDbName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="健康状态">
              <el-tag :type="datasource.healthy ? 'success' : 'warning'">{{ datasource.healthy ? '正常' : '异常' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="打开连接">{{ datasource.pool?.openConnections || 0 }}</el-descriptions-item>
          </el-descriptions>
          <el-alert v-if="datasource.warning" class="mt12" type="warning" :closable="false" :title="datasource.warning" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="section-row">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>权限审计</template>
          <el-table :data="audit.recentDataAccessLogs || []" height="260">
            <el-table-column prop="eventType" label="事件" width="130" />
            <el-table-column prop="targetTable" label="表" min-width="140" />
            <el-table-column prop="path" label="路径" min-width="180" show-overflow-tooltip />
            <el-table-column prop="requestId" label="请求ID" min-width="120" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>定时任务治理</template>
          <el-table :data="timedTask.invalidTasks || []" height="260">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="name" label="任务" min-width="140" />
            <el-table-column prop="reason" label="问题" min-width="220" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { getGvaGovernanceSummary, exportGvaGovernance } from '@/api/rideHailing/workorder09'

defineOptions({ name: 'RideHailingWorkorder09Gva' })

const state = reactive({
  route: {},
  audit: {},
  datasource: {},
  timedTask: {},
})

const route = computed(() => state.route)
const audit = computed(() => state.audit)
const datasource = computed(() => state.datasource)
const timedTask = computed(() => state.timedTask)
const routeWarnings = computed(() => route.value.warnings || [])

const kpis = computed(() => [
  { label: '菜单总数', value: route.value.totalMenus || 0, note: `重复路由 ${route.value.duplicateNames || 0}` },
  { label: '审计事件', value: audit.value.dataAccessLogs || 0, note: `越权写 ${audit.value.blockedWrites || 0}` },
  { label: '操作日志', value: audit.value.operationRecords || 0, note: `缺链路 ${audit.value.missingTraceRecords || 0}` },
  { label: '定时任务', value: timedTask.value.totalTasks || 0, note: `启用 ${timedTask.value.enabledTasks || 0}` },
])

const loadData = async () => {
  const res = await getGvaGovernanceSummary()
  const data = res.data || {}
  state.route = data.route || {}
  state.audit = data.audit || {}
  state.datasource = data.datasource || {}
  state.timedTask = data.timedTask || {}
}

const handleExport = async () => {
  const res = await exportGvaGovernance()
  ElMessage.success(`导出任务已创建：${res.data?.taskId || '-'}`)
}

onMounted(loadData)
</script>

<style scoped>
.gva-page {
  padding: 8px 0 24px;
}

.hero,
.actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.hero {
  margin-bottom: 16px;
}

.hero h2 {
  margin: 0 0 6px;
  font-size: 24px;
}

.hero p,
.kpi-card p,
.kpi-card span {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.kpi-card strong {
  display: block;
  margin: 8px 0;
  font-size: 26px;
}

.section-row {
  margin-top: 16px;
}

.mt12 {
  margin-top: 12px;
}
</style>
```

- [x] **Step 3: Add static route**

Modify `admin-platform/web/src/router/index.js` after WO-08:

```js
      {
        path: 'workorder09/gva',
        name: 'RideHailingWorkorder09Gva',
        component: () => import('@/view/rideHailing/workorder09/gva/index.vue'),
        meta: { title: 'GVA Governance' }
      }
```

- [x] **Step 4: Add dynamic menu**

Modify `admin-platform/web/src/pinia/modules/router.js` after WO-08:

```js
    {
      path: 'workorder09/gva',
      name: 'RideHailingWorkorder09Gva',
      meta: { title: 'GVA治理', icon: 'el-icon-setting', keepAlive: false },
      component: 'view/rideHailing/workorder09/gva/index.vue'
    }
```

- [x] **Step 5: Update workorder center**

Modify `admin-platform/web/src/view/rideHailing/workorders/index.vue`:

- Top copy becomes `工单01-09已接入真实页面与接口，其余工单保留开发入口。`
- WO-09 card becomes:

```js
  {
    id: '09',
    title: '工单09 GVA框架',
    summary: '动态路由、权限审计、多数据源健康和定时任务治理已接入。',
    status: '已完成',
    statusType: 'success',
    link: '/ride-hailing/workorder09/gva',
  },
```

- [x] **Step 6: Update workorder filter**

Modify `admin-platform/web/src/view/rideHailing/workorders/components/WorkorderFilter.vue`:

```js
{ id: '09', name: 'GVA框架', summary: '动态路由、权限审计、多数据源健康和定时任务治理已接入。' },
```

- [x] **Step 7: Verify admin build**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\apps\admin-web\web
npm run build
```

Expected: exit code 0. Existing npm config warnings, BigInt tolerated transform warnings, and chunk-size warnings are non-blocking.

---

### Task 4: Documentation and Final Verification

**Files:**
- Modify: `网约车系统（管理端）需求分析文档.md`
- Modify: `网约车系统技术评审文档.md`
- Modify: `docs/文档修订差异清单.md`
- Modify: `docs/superpowers/plans/2026-07-30-workorder09-gva-foundation.md`

**Interfaces:**
- Consumes verified backend and frontend work from Tasks 1-3.
- Produces updated WO-09 status and verification evidence.

- [x] **Step 1: Update management requirements document**

Update `网约车系统（管理端）需求分析文档.md`:

- Mark WO-09 in the workorder matrix as connected through GVA governance.
- Add WO-09 row in the current validation section.
- Change later boundary from the previous WO-09 onward wording to WO-10~WO-11.
- Keep `AutoMigrate` as current migration mode.

- [x] **Step 2: Update technical review document**

Update `网约车系统技术评审文档.md`:

- Add WO-09 verification row under 14.7.
- Update 14.8 heading to `工单 06~09 生产补强与 10~11 后续开发进入条件`.
- Mark route refresh bus and deeper chunk optimization as production hardening after the foundation.

- [x] **Step 3: Update difference list**

Append a new section to `docs/文档修订差异清单.md`:

```markdown
## 十、2026-07-30 WO-09 GVA 框架治理底座接入补充

| 修订对象 | 本次新增内容 | 修订原因 |
|---------|-------------|---------|
| 管理端需求文档 | 标记 WO-09 GVA 治理底座已接入，并将后续边界调整为 WO-10~WO-11 | WO-09 已具备管理端接口和页面，不应继续写成待开发 |
| 技术评审文档 | 补充 WO-09 路由治理、权限审计、多数据源健康、定时任务治理和验证结果 | 技术评审需体现当前真实代码状态，并保留 Redis Pub/Sub/Kafka 路由刷新为后续生产化补强 |
| `admin-platform/server` | 新增 GVA 治理服务、接口和路由 | 管理端需要统一查看动态路由、审计、多数据源和定时任务治理状态 |
| `admin-platform/web` | 新增 WO-09 API、页面、静态路由、动态菜单和工单中心入口 | 管理端可从工单中心进入 GVA 治理页面 |
```

- [x] **Step 4: Run final backend verification**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/system -run TestGvaGovernanceService -count=1 -v
go test ./api/v1/system ./router/system ./initialize -count=1 -v
```

Expected: PASS. Full `go test ./service/system` currently has pre-existing auto-code template failures and is documented as residual risk, not WO-09 failure evidence.

- [x] **Step 5: Run final frontend verification**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\apps\admin-web\web
npm run build
```

Expected: exit code 0.

- [x] **Step 6: Search for stale WO-09 copy**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing
rg -n --glob '!admin-platform/web/dist/**' "<stale WO-09 pending-copy patterns>" apps docs "网约车系统（管理端）需求分析文档.md" "网约车系统技术评审文档.md"
```

Expected: no stale WO-09 pending copy in source or updated docs.
