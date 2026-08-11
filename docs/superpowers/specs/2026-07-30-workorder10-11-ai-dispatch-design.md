# WO-10/WO-11 AI 智能出行与订单派单设计

## 范围

本设计覆盖两份后续工单的完整开发边界：

- WO-10：AI 智能出行助手 V1.1，完成雨天预警、智能对话、雨天路线规划 Function Calling、积水上报、降级审计和三端页面接入。
- WO-11：网约车订单中心与派单规则配置，完成管理端订单中心、订单详情、实时跟踪、轨迹回放、派单配置版本、派单评分引擎和移动端订单/位置/轨迹联动。

本轮必须严格整合用户提供的两套 Coze 私有自定义接口。所有 HTTP 请求按给定 curl 结构 1:1 复刻，不使用扣子官方 `api.coze.cn` OpenAPI，不改域名、Header 名称、JSON 固定字段和层级。

## 强制红线

### Coze 智能体对话接口

用途：用户自然语言闲聊、多轮出行咨询、雨天智能提示回复。

- 方法固定：`POST`
- URL 固定：`https://fff2xdtnzj.coze.site/stream_run`
- Header 固定：
  - `Authorization: Bearer <travel_bot_token>`
  - `Content-Type: application/json`
- `Bearer` 与 token 中间必须保留 1 个空格。
- `project_id` 固定为 `7668272524714786851`。
- 外层 `type` 固定为 `"query"`。
- JSON 层级固定为：

```json
{
  "content": {
    "query": {
      "prompt": [
        {
          "type": "text",
          "content": {
            "text": ""
          }
        }
      ]
    }
  },
  "type": "query",
  "session_id": "",
  "project_id": 7668272524714786851
}
```

仅允许变更：

- `content.query.prompt[0].content.text`
- `session_id`
- `Authorization` 中的智能体专属 token 值

后端函数名固定为：

```go
CallTravelBot(ctx context.Context, req TravelBotRequest) (*TravelBotResponse, error)
```

### Coze 路线工作流接口

用途：用户明确需要规划路线，或已提供起点/终点后由后端组装结构化参数计算避雨路线。

- 方法固定：`POST`
- URL 固定：`https://xchnkhx636.coze.site/run`
- Header 固定：
  - `Authorization: Bearer <rain_route_workflow_token>`
  - `Content-Type: application/json`
- `Bearer` 与 token 中间必须保留 1 个空格。
- JSON 固定 7 个 key，不能新增、删除、改名：

```json
{
  "origin": "",
  "destination": "",
  "city": "",
  "weather": "",
  "avoid": "",
  "preference": "",
  "user_role": ""
}
```

后端函数名固定为：

```go
CallRainRouteWorkflow(ctx context.Context, req RainRouteWorkflowRequest) (*RainRouteWorkflowResponse, error)
```

### Token 与配置

Token 不写死进代码，不写入设计文档和测试日志。配置项：

- `COZE_TRAVEL_BOT_TOKEN`
- `COZE_RAIN_ROUTE_WORKFLOW_TOKEN`
- `COZE_TRAVEL_BOT_URL=https://fff2xdtnzj.coze.site/stream_run`
- `COZE_RAIN_ROUTE_WORKFLOW_URL=https://xchnkhx636.coze.site/run`
- `COZE_TRAVEL_BOT_PROJECT_ID=7668272524714786851`
- `COZE_HTTP_TIMEOUT_MS=5000`

两套 token 绝对不能互换。单元测试必须覆盖“对话接口只使用智能体 token、路线接口只使用工作流 token”的请求构造。

## 总体架构

### 后端

后端分为三层，不把 AI、订单、派单写成一个大服务：

- `pkg/cozex`：通用 Coze 私有接口客户端，只负责 1:1 构造 HTTP 请求、超时、响应读取、错误包装和脱敏日志。
- Kratos 微服务层：
  - `services/ai-service`：WO-10 AI 业务服务，负责对话、路线规划、积水上报、降级、审计。
  - `services/order-service`：WO-11 订单主链路，负责订单创建、取消、改派、详情、状态流转、轨迹归档。
  - `services/driver-service`：司机位置、司机状态、候选司机能力。
  - `services/gateway-service`：乘客端/司机端 HTTP 聚合入口，调用 AI、订单、司机服务。
- `admin-platform/server`：Gin-Vue-Admin 管理端，继续使用现有 `carpool` 模块承载 WO-10/WO-11 管理接口和页面。

### 前端

- `apps/passenger-uni-app/uni-app`：乘客端 AI 助手、积水上报、订单跟踪页面接真实接口。
- `apps/driver-uni-app/uni-app`：司机端 AI 风险提醒、待接订单、位置上报、轨迹页面接真实接口。
- `admin-platform/web`：管理端新增 WO-10 AI 管理页和 WO-11 派单中心页，使用 GVA 路由、菜单和 `@/utils/request`。

## WO-10 设计

### 后端能力

AI 服务提供这些业务方法：

- `Chat(ctx, req)`：自然语言提问，调用 `CallTravelBot`。
- `PlanRainRoute(ctx, req)`：结构化路线规划，调用 `CallRainRouteWorkflow`。
- `ChatWithRainRoute(ctx, req)`：先调用路线工作流，再把路线 JSON 拼入智能体 text，由 `CallTravelBot` 润色为可读话术。
- `ReportFlooding(ctx, req)`：保存积水上报，调用 AI 或规则识别，低置信度进入人工审核。
- `ListConversationLogs(ctx, search)`：管理端查看对话日志。
- `ListRouteLogs(ctx, search)`：管理端查看路线工作流调用日志。
- `GetAISummary(ctx)`：统计调用量、成功量、失败量、降级量、平均耗时。

### 数据模型

管理端 `admin-platform/server/model/carpool` 新增表，继续注册到 `AutoMigrate`：

- `AiConversationLog`
  - `session_id`
  - `user_id`
  - `user_role`
  - `request_text`
  - `response_text`
  - `provider`
  - `latency_ms`
  - `success`
  - `fallback`
  - `error_message`
  - `trace_id`
- `AiRoutePlanLog`
  - `route_plan_no`
  - `session_id`
  - `origin`
  - `destination`
  - `city`
  - `weather`
  - `avoid`
  - `preference`
  - `user_role`
  - `raw_response`
  - `risk_level`
  - `fallback`
  - `latency_ms`
  - `trace_id`
- `AiFloodReport`
  - `report_no`
  - `reporter_id`
  - `reporter_role`
  - `city`
  - `location_text`
  - `photo_url`
  - `recognized_depth_cm`
  - `confidence`
  - `risk_level`
  - `audit_status`
  - `coupon_status`
  - `trace_id`
- `AiFallbackTemplate`
  - `scene`
  - `template`
  - `enabled`

敏感字段写入前脱敏或加密：手机号、车牌号、用户输入中的手机号片段、Token、Authorization Header。

### API

管理端私有接口：

- `GET /carpool/ai/summary`
- `POST /carpool/ai/chat`
- `POST /carpool/ai/rain-route`
- `POST /carpool/ai/chat-with-route`
- `GET /carpool/ai/conversation/list`
- `GET /carpool/ai/route-plan/list`
- `GET /carpool/ai/flood-report/list`
- `POST /carpool/ai/flood-report/audit`
- `POST /carpool/ai/export`

移动端网关接口：

- `POST /api/v1/ai/chat`
- `POST /api/v1/ai/rain-route`
- `POST /api/v1/ai/chat-with-route`
- `POST /api/v1/ai/flood-report`
- `GET /api/v1/ai/weather-alert`
- `GET /api/v1/driver/ai-alerts`

### 降级

- Coze 对话接口超时或非 2xx：记录失败日志，返回固定模板，不显示“调用成功”。
- 路线工作流超时或非 2xx：返回基础路线建议，`fallback=true`。
- 路线工作流字段不足：不得补造 Coze 返回，按降级处理。
- 积水识别置信度低于 80%：仅上报，不自动影响订单和路线。

## WO-11 设计

### 后端能力

订单中心服务：

- 管理端订单列表：来源、状态、创建时间、车牌号、手机号组合筛选，多字段分页排序。
- 订单详情：订单基础、乘客、司机、费用明细、状态变更记录、风控标记。
- 取消订单：幂等校验、事务更新、审计记录。
- 改派订单：幂等校验、事务更新、触发派单引擎重新计算。
- 轨迹回放：按订单号和时间范围查询位置点。

派单引擎：

- 按城市配置派单时长、可服务司机时间、白天/夜间权重、可用车队。
- 版本记录、灰度发布、一键回滚。
- 4 条硬性规则必须实现：
  - 优先分配用车时间更近的乘客订单。
  - 校验司机时间窗无订单冲突。
  - 区分白天 7:00-22:00 与夜间权重、超时惩罚。
  - 候选司机池按城市、车队、热区筛选。
- 综合评分：
  - GEO 距离
  - 司机评分
  - 时间罚分
  - 历史接单率
  - AI 风险参数：`ai_context_id`、风险等级、推荐车型、避雨路线
- 并发安全：
  - 订单锁
  - 司机锁
  - 幂等键
  - 派单决策审计

实时跟踪：

- 司机位置上报。
- WebSocket 推送订单状态和实时定位。
- 心跳保活。
- 断线重连后补发最近消息。
- 推送节流。
- 热轨迹写 Redis GEO，历史轨迹落 MySQL。

### 数据模型

复用并扩展已有订单相关模型：

- `OrderMain`
  - 增加 AI 上下文字段：`ai_context_id`、`ai_risk_level`、`ai_route_summary`、`recommended_vehicle_type`。
- `OrderDispatchAudit`
  - `dispatch_no`
  - `order_no`
  - `driver_id`
  - `rule_version`
  - `score_detail`
  - `decision_reason`
  - `idempotency_key`
  - `trace_id`
- `DispatchConfig`
  - `city`
  - `day_weight_json`
  - `night_weight_json`
  - `timeout_penalty`
  - `service_window_minutes`
  - `fleet_scope`
  - `version`
  - `gray_percent`
  - `enabled`
- `DispatchConfigVersion`
  - `config_id`
  - `version`
  - `snapshot_json`
  - `operator_id`
  - `action`
- `DriverLocationPoint`
  - `driver_id`
  - `order_no`
  - `lat`
  - `lng`
  - `heading`
  - `speed`
  - `reported_at`
- `RealtimeMessage`
  - `message_no`
  - `order_no`
  - `receiver_id`
  - `receiver_role`
  - `message_type`
  - `payload`
  - `acked`
  - `trace_id`

继续使用 `AutoMigrate`，不引入 SQL migration。

### API

管理端私有接口：

- `GET /carpool/dispatch/order/list`
- `GET /carpool/dispatch/order/:id`
- `POST /carpool/dispatch/order/:id/cancel`
- `POST /carpool/dispatch/order/:id/reassign`
- `GET /carpool/dispatch/config/list`
- `POST /carpool/dispatch/config`
- `POST /carpool/dispatch/config/:id/publish`
- `POST /carpool/dispatch/config/:id/rollback`
- `GET /carpool/dispatch/audit/list`
- `GET /carpool/dispatch/track/replay`
- `POST /carpool/dispatch/export`

移动端网关接口：

- `GET /api/v1/passenger/orders`
- `GET /api/v1/passenger/orders/:id`
- `POST /api/v1/passenger/orders`
- `POST /api/v1/passenger/orders/:id/cancel`
- `GET /api/v1/passenger/orders/:id/track`
- `POST /api/v1/driver/location/report`
- `GET /api/v1/driver/orders/available`
- `POST /api/v1/driver/orders/:id/accept`
- `GET /api/v1/driver/track/replay`
- `GET /ws/orders/:orderNo`

## WO-10 与 WO-11 联动

1. 乘客在 AI 助手输入出行需求。
2. 后端判断为普通聊天或路线规划。
3. 普通聊天调用 `CallTravelBot`。
4. 路线规划调用 `CallRainRouteWorkflow`，再把路线 JSON 拼入 `CallTravelBot` 做自然语言润色。
5. 用户确认下单时，订单请求携带 AI 上下文：
   - `ai_context_id`
   - `ai_risk_level`
   - `ai_route_summary`
   - `recommended_vehicle_type`
   - `recommended_departure_time`
6. WO-11 创建订单并触发派单引擎。
7. 派单引擎把 AI 风险参数作为加权因子，但 AI 失败不得阻塞普通下单。
8. 订单状态和轨迹通过 WebSocket 推送到乘客端与司机端。

## 前端设计

### 乘客端

`apps/passenger-uni-app/uni-app` 页面更新：

- `pages/aiAssistant/aiAssistant.vue`
  - 聊天气泡接 `POST /api/v1/ai/chat`。
  - 起点终点明确时接 `POST /api/v1/ai/chat-with-route`。
  - 展示风险等级、推荐路线、费用预估和确认叫车入口。
- `pages/floodReport/floodReport.vue`
  - 拍照后提交 `POST /api/v1/ai/flood-report`。
  - 展示识别中、已识别、待人工确认、补偿券状态。
- `pages/tracking/tracking.vue`
  - 接入订单状态、位置点和 WebSocket 推送。

### 司机端

`apps/driver-uni-app/uni-app` 页面更新：

- `pages/aiAlerts/aiAlerts.vue`
  - 接 `GET /api/v1/driver/ai-alerts`。
  - 展示天气风险、积水风险、路线风险和降级标识。
- `pages/orders/orders.vue`
  - 接待接订单与抢单/接单接口。
- `pages/location/location.vue`
  - 上报司机位置，失败时保留待补传状态。
- `pages/track/track.vue`
  - 展示轨迹回放和当前行程风险提示。

### 管理端

`admin-platform/web` 新增：

- `src/api/rideHailing/workorder10.js`
- `src/api/rideHailing/workorder11.js`
- `src/view/rideHailing/workorder10/ai/index.vue`
- `src/view/rideHailing/workorder11/dispatch/index.vue`

路由：

- `/ride-hailing/workorder10/ai`
- `/ride-hailing/workorder11/dispatch`

工单总览在验证通过后标记 WO-10、WO-11 为已接入真实页面与接口。

## 可观测与安全

- 所有写操作记录操作审计。
- 所有 AI 调用记录 trace id、session id、耗时、是否降级。
- 日志使用结构化字段，不输出 token、Authorization、完整手机号、完整车牌号。
- 管理端列表接口对手机号、车牌做脱敏。
- 订单取消、改派、接单、派单发布、派单回滚全部使用事务与幂等键。
- gRPC/HTTP 调用继续沿用已铺开的 trace id、metadata、错误码和超时规范。

## 测试策略

### Coze 客户端

- 使用 `httptest.Server` 校验请求方法、URL path、Header、Bearer 空格、JSON 层级。
- 对话接口测试必须断言：
  - `project_id=7668272524714786851`
  - 外层 `type=query`
  - 仅 text/session_id 为业务变量
- 路线接口测试必须断言：
  - JSON 只有 7 个固定 key
  - token 使用工作流 token
  - 不出现 `project_id`

### WO-10

- AI 对话成功落库日志。
- Coze 超时触发降级模板。
- 路线工作流成功返回后可生成 AI 上下文。
- 积水上报低置信度进入人工审核。
- 管理端 AI 接口编译与服务测试通过。

### WO-11

- 派单 4 条硬规则分别有单元测试。
- 综合评分输出最优司机和决策理由。
- 并发接单/改派不产生重复派单。
- 订单取消和改派幂等。
- 轨迹回放按订单和时间范围过滤。
- WebSocket 鉴权、心跳、断线补发有可复现测试或本地集成验证。

### 前端

- `admin-platform/web` 执行 `npm run build`。
- 乘客端、司机端 uni-app 执行 H5 构建。
- 搜索确认 WO-10/WO-11 不再以“待开发”出现在已接入入口文案中。

## 验收标准

- 两套 Coze 私有接口封装函数存在，且测试证明没有混用 token、域名、Header、JSON 结构。
- WO-10 后端、管理端、乘客端、司机端页面均走真实接口或明确降级接口，不再只展示 mock 成功状态。
- WO-11 管理端订单中心、派单配置、派单审计、轨迹回放可验收。
- 派单引擎实现 4 条硬规则和综合评分模型。
- AI 路线结果可以写入订单 AI 上下文，派单引擎能读取但不依赖 AI 成功。
- `AutoMigrate` 保持为当前表结构同步方式。
- 不修改 DOCX 工单原文，需求分析文档、技术评审文档和差异清单记录真实实现状态。
