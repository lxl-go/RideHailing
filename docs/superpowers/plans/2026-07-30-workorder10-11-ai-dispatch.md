# WO-10/WO-11 AI 智能出行与订单派单 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完整开发 WO-10 AI 智能出行助手与 WO-11 订单中心/派单规则配置，并严格接入用户提供的两套 Coze 私有自定义接口。

**Architecture:** 先在共享 Go 包中封装两套 Coze 私有接口，测试锁死 curl 结构；再在 GVA 管理端 `carpool` 模块落 WO-10/WO-11 管理接口、AutoMigrate 模型与页面；最后改造 passenger/driver uni-app 页面接真实网关接口或明确降级接口。WO-10 和 WO-11 解耦，通过订单 AI 上下文字段联动。

**Tech Stack:** Go, Gin-Vue-Admin, Kratos-style service layering, GORM AutoMigrate, httptest, Vue 3, Element Plus, uni-app, uView, npm/Vite.

## Global Constraints

- 禁止使用扣子官方 `api.coze.cn` OpenAPI；只能调用用户给定的两个私有化域名。
- `CallTravelBot` 只能调用 `https://fff2xdtnzj.coze.site/stream_run`，并只能使用智能体专属 token。
- `CallRainRouteWorkflow` 只能调用 `https://xchnkhx636.coze.site/run`，并只能使用路线工作流专属 token。
- `stream_run` 请求固定 `POST`、固定两个 Header、固定 JSON 层级、固定 `type="query"`、固定 `project_id=7668272524714786851`。
- `stream_run` 仅允许业务层传入 `text` 与 `session_id`。
- `run` 路线工作流请求固定 `POST`、固定两个 Header、固定 7 个 JSON key：`origin`、`destination`、`city`、`weather`、`avoid`、`preference`、`user_role`。
- `Bearer` 与 token 之间必须保留 1 个空格。
- Token、Authorization、完整手机号、完整车牌不得写入日志、测试输出或文档。
- 保留 `AutoMigrate`，不新增 SQL migration。
- 工单 DOCX 只读，不修改。
- 根目录不是 git 仓库，本计划不包含 commit 步骤。

---

## File Structure

### Shared Go Coze Client

- Create: `pkg/cozex/coze.go`
- Create: `pkg/cozex/coze_test.go`
- Modify: `pkg/go.mod` only if tests reveal a missing standard dependency replacement is needed; default不改依赖。

### Admin Backend WO-10

- Create: `admin-platform/server/model/carpool/ai.go`
- Create: `admin-platform/server/model/carpool/request/ai.go`
- Create: `admin-platform/server/service/carpool/ai.go`
- Create: `admin-platform/server/service/carpool/ai_test.go`
- Create: `admin-platform/server/api/v1/carpool/ai.go`
- Create: `admin-platform/server/router/carpool/ai.go`
- Modify: `admin-platform/server/service/carpool/enter.go`
- Modify: `admin-platform/server/api/v1/carpool/enter.go`
- Modify: `admin-platform/server/router/carpool/enter.go`
- Modify: `admin-platform/server/initialize/gorm.go`
- Modify: `admin-platform/server/initialize/router_biz.go`

### Admin Backend WO-11

- Create: `admin-platform/server/model/carpool/dispatch.go`
- Create: `admin-platform/server/model/carpool/request/dispatch.go`
- Create: `admin-platform/server/service/carpool/dispatch.go`
- Create: `admin-platform/server/service/carpool/dispatch_test.go`
- Create: `admin-platform/server/api/v1/carpool/dispatch.go`
- Create: `admin-platform/server/router/carpool/dispatch.go`
- Modify: same enter/gorm/router files as WO-10.
- Modify: `admin-platform/server/model/carpool/order.go` to add AI context fields only if not already present.

### Admin Web

- Create: `admin-platform/web/src/api/rideHailing/workorder10.js`
- Create: `admin-platform/web/src/api/rideHailing/workorder11.js`
- Create: `admin-platform/web/src/view/rideHailing/workorder10/ai/index.vue`
- Create: `admin-platform/web/src/view/rideHailing/workorder11/dispatch/index.vue`
- Modify: `admin-platform/web/src/router/index.js`
- Modify: `admin-platform/web/src/pinia/modules/router.js`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/index.vue`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/components/WorkorderFilter.vue`

### Passenger Uni-App

- Create: `apps/passenger-uni-app/uni-app/src/api/workorder10.js`
- Create: `apps/passenger-uni-app/uni-app/src/api/workorder11.js`
- Modify: `apps/passenger-uni-app/uni-app/src/pages/aiAssistant/aiAssistant.vue`
- Modify: `apps/passenger-uni-app/uni-app/src/pages/floodReport/floodReport.vue`
- Modify: `apps/passenger-uni-app/uni-app/src/pages/tracking/tracking.vue`

### Driver Uni-App

- Create: `apps/driver-uni-app/uni-app/src/api/workorder10.js`
- Create: `apps/driver-uni-app/uni-app/src/api/workorder11.js`
- Modify: `apps/driver-uni-app/uni-app/src/pages/aiAlerts/aiAlerts.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/orders/orders.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/location/location.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/track/track.vue`

### Documentation

- Modify: `网约车系统（管理端）需求分析文档.md`
- Modify: `网约车系统（乘客端）需求分析文档.md`
- Modify: `网约车系统（司机端）需求分析文档.md`
- Modify: `网约车系统技术评审文档.md`
- Modify: `docs/文档修订差异清单.md`
- Modify: this plan file by ticking completed steps during execution.

---

### Task 1: Coze Private Interface Client

**Files:**
- Create: `pkg/cozex/coze_test.go`
- Create: `pkg/cozex/coze.go`

**Interfaces:**
- Produces:
  - `type Config struct`
  - `type Client struct`
  - `func NewClient(cfg Config, httpClient *http.Client) *Client`
  - `func (c *Client) CallTravelBot(ctx context.Context, req TravelBotRequest) (*TravelBotResponse, error)`
  - `func (c *Client) CallRainRouteWorkflow(ctx context.Context, req RainRouteWorkflowRequest) (*RainRouteWorkflowResponse, error)`

- [x] **Step 1: Write failing tests that lock the two curl structures**

Create `pkg/cozex/coze_test.go` with tests that start `httptest.Server`, override URLs through config, and assert:

```go
func TestCallTravelBotBuildsExactStreamRunShape(t *testing.T) {
	// Assert POST /stream_run, Authorization = "Bearer travel-token",
	// Content-Type = "application/json", project_id = 7668272524714786851,
	// outer type = "query", text and session_id are the only business variables.
}

func TestCallRainRouteWorkflowBuildsExactSevenFieldShape(t *testing.T) {
	// Assert POST /run, Authorization = "Bearer workflow-token",
	// JSON keys are exactly origin,destination,city,weather,avoid,preference,user_role,
	// and no project_id is present.
}

func TestCozeClientRejectsMissingDedicatedTokens(t *testing.T) {
	// Assert empty travel token fails before HTTP request for CallTravelBot
	// and empty workflow token fails before HTTP request for CallRainRouteWorkflow.
}
```

- [x] **Step 2: Run RED**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\pkg
go test ./cozex -count=1 -v
```

Expected: FAIL because `pkg/cozex` does not exist.

- [x] **Step 3: Implement minimal client**

Create `pkg/cozex/coze.go` with:

```go
package cozex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultTravelBotURL       = "https://fff2xdtnzj.coze.site/stream_run"
	DefaultRainRouteURL       = "https://xchnkhx636.coze.site/run"
	DefaultTravelBotProjectID = int64(7668272524714786851)
)

type Config struct {
	TravelBotURL          string
	RainRouteWorkflowURL  string
	TravelBotToken        string
	RainRouteWorkflowToken string
	TravelBotProjectID    int64
	Timeout               time.Duration
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

type TravelBotRequest struct {
	Text      string
	SessionID string
}

type TravelBotResponse struct {
	RawBody string `json:"rawBody"`
}

type RainRouteWorkflowRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	City        string `json:"city"`
	Weather     string `json:"weather"`
	Avoid       string `json:"avoid"`
	Preference  string `json:"preference"`
	UserRole    string `json:"user_role"`
}

type RainRouteWorkflowResponse struct {
	RawBody string `json:"rawBody"`
}
```

The implementation must:

- Fill default URLs and project id when config is empty.
- Create `Authorization` with exactly `"Bearer "+token`.
- Marshal travel bot request using fixed nested structs, not map mutation.
- Marshal workflow request from `RainRouteWorkflowRequest` so only 7 keys exist.
- Treat non-2xx as errors with response body preview, masking Authorization.

- [x] **Step 4: Run GREEN**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\pkg
go test ./cozex -count=1 -v
```

Expected: PASS.

---

### Task 2: WO-10 Admin Backend Models and Service

**Files:**
- Create: `admin-platform/server/model/carpool/ai.go`
- Create: `admin-platform/server/model/carpool/request/ai.go`
- Create: `admin-platform/server/service/carpool/ai_test.go`
- Create: `admin-platform/server/service/carpool/ai.go`
- Modify: `admin-platform/server/service/carpool/enter.go`
- Modify: `admin-platform/server/initialize/gorm.go`

**Interfaces:**
- Consumes: `pkg/cozex.Client`, `global.GVA_DB`
- Produces:
  - `type AIService struct{}`
  - `func (s *AIService) Chat(ctx context.Context, req carpoolReq.AIChatRequest) (*AIChatResult, error)`
  - `func (s *AIService) PlanRainRoute(ctx context.Context, req carpoolReq.RainRouteRequest) (*AIRoutePlanResult, error)`
  - `func (s *AIService) ChatWithRainRoute(ctx context.Context, req carpoolReq.ChatWithRouteRequest) (*AIChatResult, error)`
  - `func (s *AIService) ReportFlooding(ctx context.Context, req carpoolReq.FloodReportRequest) (*carpoolModel.AiFloodReport, error)`
  - list and summary methods for admin pages.

- [x] **Step 1: Write failing service tests**

Create tests for:

- Chat persists `AiConversationLog` with `success=true`.
- Chat falls back and persists `fallback=true` when provider returns error.
- PlanRainRoute persists `AiRoutePlanLog`.
- ReportFlooding with confidence `<80` sets audit status to pending manual review.

Use an in-memory sqlite DB and a fake Coze provider interface so tests do not call external Coze.

- [x] **Step 2: Run RED**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool -run TestAIService -count=1 -v
```

Expected: FAIL because AI models and service do not exist.

- [x] **Step 3: Add models**

Create `admin-platform/server/model/carpool/ai.go` with:

- `AiConversationLog` table `ai_conversation_log`
- `AiRoutePlanLog` table `ai_route_plan_log`
- `AiFloodReport` table `ai_flood_report`
- `AiFallbackTemplate` table `ai_fallback_template`

All structs include `CreatedAt` and `UpdatedAt`. Use `gorm` indexes on `session_id`, `user_role`, `success`, `fallback`, `route_plan_no`, `report_no`, `audit_status`, and `trace_id`.

- [x] **Step 4: Add request structs**

Create `admin-platform/server/model/carpool/request/ai.go` with:

```go
type AIChatRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
	Text      string `json:"text" binding:"required"`
	UserID    uint64 `json:"userId"`
	UserRole  string `json:"userRole" binding:"required"`
}

type RainRouteRequest struct {
	SessionID   string `json:"sessionId"`
	Origin      string `json:"origin" binding:"required"`
	Destination string `json:"destination" binding:"required"`
	City        string `json:"city" binding:"required"`
	Weather     string `json:"weather"`
	Avoid       string `json:"avoid"`
	Preference  string `json:"preference"`
	UserRole    string `json:"userRole" binding:"required"`
}

type ChatWithRouteRequest struct {
	Chat AIChatRequest `json:"chat"`
	Route RainRouteRequest `json:"route"`
}

type FloodReportRequest struct {
	ReporterID   uint64  `json:"reporterId"`
	ReporterRole string  `json:"reporterRole" binding:"required"`
	City         string  `json:"city" binding:"required"`
	LocationText string  `json:"locationText" binding:"required"`
	PhotoURL     string  `json:"photoUrl"`
	DepthCM      float64 `json:"depthCm"`
	Confidence   float64 `json:"confidence"`
}
```

- [x] **Step 5: Implement AI service**

Implement validation:

- `UserRole` must be `passenger`, `driver`, or `admin`.
- Chat text cannot be blank.
- Route request must include origin, destination, city.
- Fallback templates are used when Coze provider errors.
- All DB writes use `global.GVA_DB.WithContext(ctx).Transaction`.

- [x] **Step 6: Wire service and AutoMigrate**

Modify:

- `service/carpool/enter.go`: add `AIService`
- `initialize/gorm.go`: add four AI models to `AutoMigrate`

- [x] **Step 7: Run GREEN**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool -run TestAIService -count=1 -v
```

Expected: PASS.

---

### Task 3: WO-10 Admin API and Router

**Files:**
- Create: `admin-platform/server/api/v1/carpool/ai.go`
- Create: `admin-platform/server/router/carpool/ai.go`
- Modify: `admin-platform/server/api/v1/carpool/enter.go`
- Modify: `admin-platform/server/router/carpool/enter.go`
- Modify: `admin-platform/server/initialize/router_biz.go`

**Interfaces:**
- Produces admin endpoints under `/carpool/ai`.

- [x] **Step 1: Run compile baseline**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./api/v1/carpool ./router/carpool ./initialize -count=1 -v
```

Expected: PASS before edits.

- [x] **Step 2: Add API handlers**

Create handlers:

- `GetAISummary`
- `Chat`
- `PlanRainRoute`
- `ChatWithRainRoute`
- `ListConversationLogs`
- `ListRoutePlanLogs`
- `ListFloodReports`
- `AuditFloodReport`
- `ExportAI`

Use existing `response.OkWithData`, `response.FailWithMessage`, `utils.Verify` page validation, and structured logger.

- [x] **Step 3: Add router**

Create `router/carpool/ai.go`:

- write routes with `middleware.OperationRecord()`:
  - `POST chat`
  - `POST rain-route`
  - `POST chat-with-route`
  - `POST flood-report/audit`
  - `POST export`
- read routes without operation record:
  - `GET summary`
  - `GET conversation/list`
  - `GET route-plan/list`
  - `GET flood-report/list`

- [x] **Step 4: Wire enter groups and router initialization**

Add `AIApi`, `AIRouter`, `aiApi`, and call `carpoolRouter.InitAIRouter(privateGroup)`.

- [x] **Step 5: Verify integration compile**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./api/v1/carpool ./router/carpool ./initialize -count=1 -v
```

Expected: PASS.

---

### Task 4: WO-11 Dispatch Models and Service

**Files:**
- Create: `admin-platform/server/model/carpool/dispatch.go`
- Create: `admin-platform/server/model/carpool/request/dispatch.go`
- Create: `admin-platform/server/service/carpool/dispatch_test.go`
- Create: `admin-platform/server/service/carpool/dispatch.go`
- Modify: `admin-platform/server/model/carpool/order.go`
- Modify: `admin-platform/server/service/carpool/enter.go`
- Modify: `admin-platform/server/initialize/gorm.go`

**Interfaces:**
- Produces:
  - `type DispatchService struct{}`
  - `func (s *DispatchService) ListOrders(ctx context.Context, search DispatchOrderSearch) ([]OrderMain, int64, error)`
  - `func (s *DispatchService) GetOrderDetail(ctx context.Context, id uint64) (*DispatchOrderDetail, error)`
  - `func (s *DispatchService) CancelOrder(ctx context.Context, req CancelOrderRequest) error`
  - `func (s *DispatchService) ReassignOrder(ctx context.Context, req ReassignOrderRequest) (*DispatchDecision, error)`
  - `func (s *DispatchService) ScoreDrivers(ctx context.Context, req DispatchScoreRequest) (*DispatchDecision, error)`
  - config, audit, and track replay methods.

- [x] **Step 1: Write failing dispatch tests**

Create tests for:

- Earlier scheduled order sorts ahead for dispatch priority.
- Driver with overlapping service window is excluded.
- Day/night weight selection changes score.
- Candidate pool filters by city, fleet, and hot zone.
- Highest score driver is selected with score detail.
- Duplicate `idempotency_key` does not create duplicate dispatch audit.

- [x] **Step 2: Run RED**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool -run TestDispatchService -count=1 -v
```

Expected: FAIL because dispatch models and service do not exist.

- [x] **Step 3: Add dispatch models**

Create models:

- `OrderDispatchAudit`
- `DispatchConfig`
- `DispatchConfigVersion`
- `DriverLocationPoint`
- `RealtimeMessage`

Modify `OrderMain` only for AI context fields:

- `AIContextID`
- `AIRiskLevel`
- `AIRouteSummary`
- `RecommendedVehicleType`

- [x] **Step 4: Add dispatch request structs**

Create search and action requests:

- `DispatchOrderSearch`
- `CancelOrderRequest`
- `ReassignOrderRequest`
- `DispatchConfigRequest`
- `DispatchScoreRequest`
- `DriverCandidate`
- `TrackReplaySearch`

- [x] **Step 5: Implement dispatch service**

Implementation requirements:

- All write methods use DB transactions.
- Cancel/reassign use idempotency key.
- Driver conflict uses time-window overlap.
- Day is 07:00 inclusive to 22:00 exclusive.
- Score formula records individual parts in `score_detail`.
- Missing Redis is not fatal for admin verification; MySQL-backed path must work.

- [x] **Step 6: Wire service and AutoMigrate**

Add `DispatchService` to enter group and dispatch models to `AutoMigrate`.

- [x] **Step 7: Run GREEN**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool -run TestDispatchService -count=1 -v
```

Expected: PASS.

---

### Task 5: WO-11 Dispatch API and Router

**Files:**
- Create: `admin-platform/server/api/v1/carpool/dispatch.go`
- Create: `admin-platform/server/router/carpool/dispatch.go`
- Modify: `admin-platform/server/api/v1/carpool/enter.go`
- Modify: `admin-platform/server/router/carpool/enter.go`
- Modify: `admin-platform/server/initialize/router_biz.go`

**Interfaces:**
- Produces admin endpoints under `/carpool/dispatch`.

- [x] **Step 1: Add API handlers**

Create handlers:

- `ListDispatchOrders`
- `GetDispatchOrderDetail`
- `CancelDispatchOrder`
- `ReassignDispatchOrder`
- `ListDispatchConfigs`
- `SaveDispatchConfig`
- `PublishDispatchConfig`
- `RollbackDispatchConfig`
- `ListDispatchAudits`
- `ReplayTrack`
- `ExportDispatch`

- [x] **Step 2: Add router**

Routes:

- `GET order/list`
- `GET order/:id`
- `POST order/:id/cancel`
- `POST order/:id/reassign`
- `GET config/list`
- `POST config`
- `POST config/:id/publish`
- `POST config/:id/rollback`
- `GET audit/list`
- `GET track/replay`
- `POST export`

- [x] **Step 3: Wire and verify compile**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool -run "TestAIService|TestDispatchService" -count=1 -v
go test ./api/v1/carpool ./router/carpool ./initialize -count=1 -v
```

Expected: PASS.

---

### Task 6: Admin Web WO-10 and WO-11 Pages

**Files:**
- Create: `admin-platform/web/src/api/rideHailing/workorder10.js`
- Create: `admin-platform/web/src/api/rideHailing/workorder11.js`
- Create: `admin-platform/web/src/view/rideHailing/workorder10/ai/index.vue`
- Create: `admin-platform/web/src/view/rideHailing/workorder11/dispatch/index.vue`
- Modify: `admin-platform/web/src/router/index.js`
- Modify: `admin-platform/web/src/pinia/modules/router.js`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/index.vue`
- Modify: `admin-platform/web/src/view/rideHailing/workorders/components/WorkorderFilter.vue`

**Interfaces:**
- Consumes admin endpoints from Tasks 3 and 5.

- [x] **Step 1: Add API wrappers**

`workorder10.js` exports functions for `/carpool/ai/*`.

`workorder11.js` exports functions for `/carpool/dispatch/*`.

- [x] **Step 2: Add WO-10 AI page**

Page requirements:

- KPI cards: total calls, success, fallback, avg latency.
- Tabs: conversation logs, route plans, flood reports, fallback templates.
- Action panel: test chat and test rain route.
- Do not display token or Authorization.
- Empty state must say no records, not mock success.

- [x] **Step 3: Add WO-11 dispatch page**

Page requirements:

- Tabs: orders, dispatch config, dispatch audit, track replay.
- Filters: order source, status, created time, plate, phone.
- Actions: cancel, reassign, publish config, rollback config.
- Score detail drawer displays decision reason and AI context.

- [x] **Step 4: Wire routes, menus, workorder center**

Add static routes:

- `workorder10/ai`
- `workorder11/dispatch`

Add dynamic menus:

- `AI助手`
- `派单中心`

Update workorder cards after backend verification:

- WO-10 status `已完成`, link `/ride-hailing/workorder10/ai`
- WO-11 status `已完成`, link `/ride-hailing/workorder11/dispatch`

- [x] **Step 5: Build admin web**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\apps\admin-web\web
npm run build
```

Expected: exit code 0. Existing npm config warnings, tolerated BigInt transform warnings, and chunk-size warnings are non-blocking.

---

### Task 7: Passenger Uni-App WO-10/WO-11 Integration

**Files:**
- Create: `apps/passenger-uni-app/uni-app/src/api/workorder10.js`
- Create: `apps/passenger-uni-app/uni-app/src/api/workorder11.js`
- Modify: `apps/passenger-uni-app/uni-app/src/pages/aiAssistant/aiAssistant.vue`
- Modify: `apps/passenger-uni-app/uni-app/src/pages/floodReport/floodReport.vue`
- Modify: `apps/passenger-uni-app/uni-app/src/pages/tracking/tracking.vue`

**Interfaces:**
- Consumes mobile gateway endpoints. If gateway endpoint implementation is deferred, API wrapper still targets final path and page must show service unavailable instead of mock success.

- [x] **Step 1: Add API wrappers**

Use existing `src/utils/request.js`.

Functions:

- `chatAI`
- `planRainRoute`
- `chatWithRoute`
- `submitFloodReport`
- `listPassengerOrders`
- `getPassengerOrderDetail`
- `getPassengerTrack`

- [x] **Step 2: Update AI assistant page**

Requirements:

- Send button calls `chatAI` for plain text.
- If origin/destination form is filled, call `chatWithRoute`.
- Show loading, success, fallback, and failure states.
- Do not append fake “正在查询” success when API fails.

- [x] **Step 3: Update flood report page**

Requirements:

- Submit button calls `submitFloodReport`.
- Photo selection updates local state.
- API failure shows “上报暂未提交，请稍后重试”.
- Low-confidence response shows “待人工确认”.

- [x] **Step 4: Update tracking page**

Requirements:

- Load order track through `getPassengerTrack`.
- Render last known location and status messages.
- WebSocket can be feature-flagged; when unavailable, show polling/refresh state.

- [x] **Step 5: Build passenger H5**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\apps\passenger-uni-app\uni-app
npm run build:h5
```

Expected: exit code 0.

---

### Task 8: Driver Uni-App WO-10/WO-11 Integration

**Files:**
- Create: `apps/driver-uni-app/uni-app/src/api/workorder10.js`
- Create: `apps/driver-uni-app/uni-app/src/api/workorder11.js`
- Modify: `apps/driver-uni-app/uni-app/src/pages/aiAlerts/aiAlerts.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/orders/orders.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/location/location.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/track/track.vue`

**Interfaces:**
- Consumes mobile gateway endpoints for AI alerts, available orders, order accept, location report, and track replay.

- [x] **Step 1: Add API wrappers**

Functions:

- `getDriverAIAlerts`
- `listAvailableOrders`
- `acceptOrder`
- `reportDriverLocation`
- `getDriverTrackReplay`

- [x] **Step 2: Update AI alerts page**

Requirements:

- Load alerts from backend.
- Show fallback flag when AI service degrades.
- No hard-coded “台风预警/积水提醒” as success data after integration.

- [x] **Step 3: Update orders page**

Requirements:

- Load available orders.
- Accept action calls `acceptOrder` with idempotency key.
- If already accepted by another driver, show unavailable state.

- [x] **Step 4: Update location and track pages**

Requirements:

- Location report page calls `reportDriverLocation`.
- Failed reports remain in local pending list for retry display.
- Track page loads replay points and status.

- [x] **Step 5: Build driver H5**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\apps\driver-uni-app\uni-app
npm run build:h5
```

Expected: exit code 0.

---

### Task 9: Gateway/Microservice Alignment

**Files:**
- Modify or create only after inspecting current Kratos service boundaries:
  - `services/gateway-service/internal/service/*.go`
  - `services/order-service/internal/biz/*.go`
  - `services/order-service/internal/data/*.go`
  - `services/driver-service/internal/biz/*.go`
  - `services/driver-service/internal/data/*.go`
  - New `services/ai-service` only if implementing a real Kratos service is still required after admin backend completion.

**Interfaces:**
- Mobile endpoints listed in design must map to service methods. If full `ai-service` scaffolding would exceed safe scope, implement gateway-compatible stubs that call admin/backend service only through stable HTTP/gRPC boundaries.

- [x] **Step 1: Inspect existing gateway route pattern**

Read current gateway HTTP server and service files, then decide whether to add endpoints directly or scaffold `services/ai-service`.

- [x] **Step 2: Add route tests or compile checks**

At minimum run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing
go test ./services/gateway-service/... ./services/order-service/... ./services/driver-service/... -count=1
```

Expected before edits may expose existing failures; document unrelated failures before proceeding.

- [x] **Step 3: Implement mobile gateway endpoints**

Add endpoints matching Tasks 7 and 8:

- `/api/v1/ai/chat`
- `/api/v1/ai/rain-route`
- `/api/v1/ai/chat-with-route`
- `/api/v1/ai/flood-report`
- `/api/v1/driver/ai-alerts`
- `/api/v1/driver/location/report`
- `/api/v1/driver/orders/available`
- `/api/v1/driver/orders/{id}/accept`
- `/api/v1/passenger/orders/{id}/track`

- [x] **Step 4: Verify service builds**

Run targeted package tests for touched services. Do not claim full microservice verification if unrelated existing failures remain.

---

### Task 10: Documentation and Final Verification

**Files:**
- Modify: `网约车系统（管理端）需求分析文档.md`
- Modify: `网约车系统（乘客端）需求分析文档.md`
- Modify: `网约车系统（司机端）需求分析文档.md`
- Modify: `网约车系统技术评审文档.md`
- Modify: `docs/文档修订差异清单.md`
- Modify: `docs/superpowers/plans/2026-07-30-workorder10-11-ai-dispatch.md`

- [x] **Step 1: Update requirements documents**

Mark WO-10 and WO-11 as connected only after backend and frontend verification pass. Keep DOCX read-only.

- [x] **Step 2: Update technical review**

Add verification rows:

- Coze private interface request-shape tests.
- WO-10 AI admin backend tests.
- WO-11 dispatch backend tests.
- Admin-web build.
- Passenger H5 build.
- Driver H5 build.

- [x] **Step 3: Update difference list**

Append a section for WO-10/WO-11 including:

- Coze client redline enforcement.
- AI management and mobile pages.
- Dispatch center and mobile pages.
- AutoMigrate model additions.
- Remaining production hardening, if any.

- [x] **Step 4: Run final backend verification**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing\pkg
go test ./cozex -count=1 -v

cd C:\Users\李小龙\Desktop\RideHailing\admin-platform\server
go test ./service/carpool -run "TestAIService|TestDispatchService" -count=1 -v
go test ./api/v1/carpool ./router/carpool ./initialize -count=1 -v
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

Expected: all exit code 0.

- [x] **Step 6: Search stale copy and leaked token patterns**

Run:

```powershell
cd C:\Users\李小龙\Desktop\RideHailing
rg -n --glob '!**/node_modules/**' --glob '!admin-platform/web/dist/**' "工单10 AI助手.*待开发|工单11 派单中心.*待开发|api\\.coze\\.cn|Authorization: Bearer eyJ|eyJhbGci" apps admin-platform pkg services docs "网约车系统（管理端）需求分析文档.md" "网约车系统（乘客端）需求分析文档.md" "网约车系统（司机端）需求分析文档.md" "网约车系统技术评审文档.md"
```

Expected:

- No stale WO-10/WO-11 pending-copy in source/docs after verified integration.
- No long token committed into source/docs.
- `api.coze.cn` may appear only in explicit “禁止使用” documentation if needed.

- [x] **Step 7: Mark plan checklist complete**

Tick each completed step only after its verification command has passed.
