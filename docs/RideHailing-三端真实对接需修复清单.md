# RideHailing 三端真实对接需修复清单

生成时间：2026-08-04  
核查批次：第 0 批，全盘真实对接核查清单批  
核查性质：只读扫描 + 文档固化，不修改业务代码、不修改配置、不修改数据库。

## 1. 执行门禁表

【执行门禁表】

当前批次：第 0 批，全盘真实对接核查清单批

允许修改范围：仅允许新建/覆盖 `C:\Users\李小龙\Desktop\RideHailing\docs\RideHailing-三端真实对接需修复清单.md`

禁止修改范围：禁止修改业务代码、接口逻辑、配置密钥、数据库数据、前端页面逻辑、后端服务逻辑

本轮核心验收标准：基于真实项目代码列出三端页面/按钮是否真实对接后端、失败根因、影响链路、修复优先级，并重新生成需修复清单

发现非本批次问题处理规则：仅登记留存，不插入当前批次修复

## 2. 核查结论

整体结论：三端不是完全没有对接后端，但目前只能算“部分链路接通”。乘客端/司机端订单查询、下单、接单、位置上报、地图路线预览有真实接口；但支付回调、优惠券、需求发布、司机完成订单、司机删除行程、管理端真实订单/财务/营销联动、状态统一、金额精度、异常中文提示、日志闭环仍未闭环。

| 标签 | 结论 |
| --- | --- |
| 文档已确认 | 顺风车项目只做支付宝沙箱支付，需要高德地图 API，司机接乘客和送乘客过程需要地图实时显示。 |
| 代码已存在 | `pkg/alipayx`、`pkg/amapx`、gateway 支付路由、gateway 地图路由、移动端订单/地图/位置 API 均已存在。 |
| 已测试通过 | `go test ./services/gateway-service/internal/server` 通过；`admin-platform/web` 执行 `npm run build` 通过；乘客端/司机端 Vue SFC 语法解析通过。 |
| 仍未闭环 | 移动端多个按钮打到占位接口；支付通知只验签不改订单；订单状态枚举前后端不统一；金额仍使用 `float64/double`；管理端多数功能使用独立管理表而不是移动端真实业务表。 |

## 3. 全套链路契约清单

| 契约项 | 当前核查结果 |
| --- | --- |
| 前端触发入口/页面按钮 | 乘客端：首页查行程、发布需求、行程详情下单、订单详情支付/取消/轨迹、优惠券页；司机端：待接单详情/接单、订单详情位置上报/完成订单、发布/删除行程、位置上报；管理端：订单、调度、营销、财务、评价/投诉、看板。 |
| 前端封装 API 请求方法 | 乘客端：`apps/passenger-uni-app/uni-app/src/api/*.js`；司机端：`apps/driver-uni-app/uni-app/src/api/*.js`；管理端：`admin-platform/web/src/api/**` 与 `admin-platform/web/src/view/rideHailing/**`。 |
| 后端网关/接口路由地址 | `services/gateway-service/internal/server/http.go`、`mobile_ai_dispatch.go`、`mobile_compat.go`、`payment_routes.go`、`map_routes.go`。 |
| 请求入参 DTO 完整结构 | 移动端主要走 JSON + gateway handler + proto request；管理端走 GVA request model。当前存在 string/int64、float64/decimal、状态 string/int 混用。 |
| 响应出参 DTO 完整结构 | gateway 已对部分移动端 ID 做 string 化；但金额仍为 double/float64，状态映射会丢失 `waiting_pay/paid/in_progress/refunding/refunded`。 |
| 后端对应业务服务方法 | `order-service`、`trip-service`、`driver-service`、gateway data client、admin-platform server service。 |
| 涉及数据库表 | 移动端核心：`carpool_order`、`carpool_trip`、`driver_location_point`；管理端：`order_main`、`order_refund`、`finance_transaction`、`marketing_coupon_template`、`marketing_user_coupon`、`order_dispatch_audit`。 |
| 订单状态流转规则 | 文档目标应为 `waiting_pay -> paid -> accepted -> in_progress -> completed/cancelled/refunding/refunded`；当前 order-service 只有 `0 pending / 1 confirmed / 2 completed / 3 cancelled`。 |
| 事务、锁、幂等点位 | 下单扣座有事务和条件扣减；取消/拒单/完成缺少同事务、期望状态更新、幂等键；支付回调缺少幂等、订单更新、金额校验。 |
| 本次改动文件 | 仅本文档。 |

## 4. 三端按钮/页面问题清单

### 4.1 乘客端

| 优先级 | 页面/按钮 | 前端入口 | 调用接口 | 后端入口 | 当前状态 | 根因 | 影响 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| P0 | 订单详情“去支付” | `apps/passenger-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue` | `POST /carpool/orders/{id}/pay` | `services/gateway-service/internal/server/http.go` | 仍未闭环 | gateway 能生成支付宝沙箱 WAP 表单，但 `POST /api/v1/pay/notify` 只验签和记录日志，不更新订单状态、不写支付流水、不校验金额/交易状态、不幂等。 | 用户支付成功后订单仍可能停留在待支付/待出行状态，前后端状态不一致。 |
| P0 | 订单详情“取消订单” | `pages/orderDetail/orderDetail.vue` | `POST /carpool/orders/{id}/cancel` | order-service `CancelOrder` | 仍未闭环 | `UpdateStatus` 与 `IncrementTripSeats` 分两步执行，不在同一个事务；`UpdateStatus` 只按 `id` 更新，不校验期望状态。 | 并发取消/司机处理时可能重复退座、状态漂移。 |
| P0 | 行程详情“立即预约/下单” | `pages/tripDetail/tripDetail.vue` -> `createOrder` | `POST /carpool/orders` | order-service `CreateAtomic` | 代码已存在 | 下单扣座有事务和 `seats_available >= ?` 条件，创建时超卖已被控制；但金额使用 `trip.Price * seats` 的 `float64`。 | 创建链路基本可用，但金额精度和后续支付对账不稳。 |
| P0 | 发布需求按钮 | `apps/passenger-uni-app/uni-app/src/pages/publishDemand/publishDemand.vue` | `POST /carpool/trips/demands` | `mobile_compat.go` | 仍未闭环 | 后端是兼容占位，固定返回 `trip demand service unavailable`；没有 demand DTO、service、表、状态流转。 | 乘客发布需求必然无法形成真实后端数据。 |
| P1 | 我的需求/取消需求 | `api/trip.js` `listMyDemands/cancelDemand` | `GET /carpool/trips/demands/mine`、`POST /carpool/trips/demands/{id}/cancel` | `mobile_compat.go` | 仍未闭环 | 查询固定空列表，取消固定不可用。 | 前端显示为空或取消失败，业务闭环不存在。 |
| P1 | 优惠券列表 | `pages/coupons/coupons.vue`、`api/profile.js` | `GET /carpool/coupons` | `mobile_compat.go` | 仍未闭环 | 后端固定返回空列表；未接管理端营销表 `marketing_user_coupon`。 | 乘客永远看不到真实优惠券。 |
| P1 | 领取优惠券 | `api/profile.js` `claimCoupon` | `POST /carpool/coupons/claim` | `mobile_compat.go` | 仍未闭环 | 后端固定返回 `coupon claim service unavailable`，没有库存锁、领取幂等、防重复。 | 领取按钮无法使用；后续接入时存在超领风险。 |
| P1 | 查看轨迹/地图实时显示 | `pages/tracking/tracking.vue` | `GET /api/v1/passenger/orders/{id}/track` + `GET /api/v1/maps/route` | `mobile_ai_dispatch.go` + `map_routes.go` | 部分代码已存在，仍未闭环 | 地图 route 和司机位置回放存在，但轨迹依赖订单 detail 的 driverId；订单状态没有 `in_progress`，接驾/送达阶段未强绑定地图上报权限和订单阶段。页面大量中文文案乱码。 | 地图能显示部分路线/点位，但接乘客、送乘客全过程状态语义不稳定，用户提示乱码。 |
| P2 | 订单列表状态筛选 | `pages/orders/orders.vue` | `GET /carpool/orders?status=...` | `order_response.go` | 仍未闭环 | 前端有 `pending/ongoing/completed/cancelled/paid/waiting`，gateway 将 `waiting_pay` 和 `paid` 都映射到 `0`，`in_progress` 映射到 `1`，后端没有 paid/in_progress。 | 筛选结果和页面标签可能不一致。 |
| P2 | 异常提示 | `utils/request.js` | 所有请求 | gateway/order-service | 仍未闭环 | 前端 fallback 文案乱码；gateway 非 UpstreamHTTPError 直接 return err，可能变 500/英文错误。 | 用户看到乱码、英文或 500，不知道如何处理。 |

### 4.2 司机端

| 优先级 | 页面/按钮 | 前端入口 | 调用接口 | 后端入口 | 当前状态 | 根因 | 影响 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| P0 | 订单详情“完成订单” | `apps/driver-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue` | `POST /api/v1/driver/orders/{id}/complete` | gateway -> order client | 仍未闭环 | `OrderUsecase.CompleteOrder` 存在，但 `order.proto` 没有 `CompleteOrder` RPC，order-service HTTP 生成路由也没有 `/v1/driver/orders/{id}/complete`；gateway gRPC client 直接返回 `complete order is not available over grpc yet`，HTTP client 会打到不存在的 route。 | 司机完成订单按钮大概率失败，订单无法闭环到 completed。 |
| P0 | 待接单“接单” | `pages/pendingOrders/pendingOrders.vue` | `POST /api/v1/driver/orders/{id}/accept` | order-service `AcceptOrder` | 仍未闭环 | 接单只先查再更新，`UpdateStatus WHERE id = ?` 不带 `status = pending` 或 version 条件；传了 `idempotency_key` 但后端没有使用。 | 并发接单/取消可能状态漂移，重复点击不幂等。 |
| P0 | 拒单 | `api/order.js` `rejectOrder` | `POST /api/v1/driver/orders/{id}/reject` | order-service `RejectOrder` | 仍未闭环 | 拒单状态更新和退座不在同一事务；无期望状态原子更新。 | 拒单后座位数与订单状态可能不一致。 |
| P1 | 位置上报“立即上报/自动上报” | `pages/locationReport/locationReport.vue` | `POST /api/v1/driver/location/report` | driver-service track | 代码已存在，仍未闭环 | 位置上报接口存在，但 `orderId` 可为空时会保存到 0；页面从 `driverActiveOrderId` 读当前订单，缺少与 accepted/in_progress 的状态约束。页面中文文案乱码。 | 轨迹可能与订单脱钩，乘客端查轨迹为空。 |
| P1 | 司机订单详情地图 | `pages/orderDetail/orderDetail.vue` | `GET /api/v1/driver/location/history` + `GET /api/v1/maps/route` | gateway map/track | 部分代码已存在，仍未闭环 | 地图查询存在，但司机端没有“已到达乘客上车点/开始送达”的状态接口，只有 accepted/completed。 | 接乘客、送乘客两个阶段无法在地图和订单状态上真实区分。 |
| P1 | 删除行程 | `apps/driver-uni-app/uni-app/src/api/trip.js` | `DELETE /carpool/trips/{id}` | `mobile_compat.go` | 仍未闭环 | 后端固定返回 `trip delete service unavailable`。 | 删除按钮若被页面调用会失败。 |
| P2 | 司机订单列表 | `GET /api/v1/driver/orders` | `mobile_ai_dispatch.go` | `PendingOrders` | 仍未闭环 | 当前只返回 pending 订单，不是真正“司机全部订单”。 | 已接/进行中/已完成订单列表不完整。 |
| P2 | 异常提示 | `utils/request.js` | 所有请求 | gateway/order-service | 仍未闭环 | 错误转换文案乱码；后端部分错误仍为英文。 | 司机端故障提示不可读。 |

### 4.3 管理端

| 优先级 | 页面/按钮 | 前端入口 | 后端入口/表 | 当前状态 | 根因 | 影响 |
| --- | --- | --- | --- | --- | --- | --- |
| P0 | 订单中心查询/退款/导出 | `admin-platform/web/src/view/rideHailing/orders/index.vue` 与 workorder04 | `admin-platform/server/service/carpool/order.go` -> `order_main` | 仍未闭环 | 管理端订单模型是 `order_main`，移动端真实订单是 `carpool_order`；不是同一张表，也没有同步/视图/适配层。 | 管理端看到的订单不等于乘客/司机真实订单，退款和状态无法反哺移动端。 |
| P0 | 财务流水/退款 | `workorder03/finance` | `finance_transaction`、`finance_refund` | 仍未闭环 | 支付回调没有写财务流水，也没有订单支付状态同步；金额仍为 `float64`。 | 财务数据与真实支付不一致。 |
| P1 | 营销优惠券 | `workorder07/marketing` | `marketing_coupon_template`、`marketing_user_coupon` | 仍未闭环 | 管理端营销表未接乘客端 `/carpool/coupons`、`/carpool/coupons/claim`。 | 管理端发券后乘客端看不到/领不了。 |
| P1 | 智能调度/轨迹查询 | `workorder11/dispatch/index.vue` | `order_main`、`order_dispatch_audit`、`driver_location_point` | 仍未闭环 | 调度管理使用管理端订单模型；移动端司机位置表虽有交集，但订单 ID 来源不统一。前端对 `orderId/driverId/userId` 使用数字输入控件，存在大 ID 精度风险。 | 管理端调度结果和移动端司机/乘客订单可能对不上。 |
| P1 | 评价/投诉 | admin review API/page | review service / admin model | 仍未闭环 | 移动端 `GET /carpool/reviews/mine/{orderId}` 是占位，总返回 `review: nil`；管理端评价数据无法闭环到移动端“已评价”状态。 | 用户可能重复评价或看不到评价状态。 |
| P2 | 权限/菜单 | GVA 权限配置 | admin router/api | 仍未闭环 | 既有日志曾出现 `/carpool/analytics/dashboard`、`/carpool/order/list`、`/carpool/person/list` 权限错误，需要联调确认权限种子和按钮权限。 | 管理端页面可构建但运行时可能 403/无权限。 |

## 5. 根因分类清单

### 5.1 后端 ID 传入前端后的精度问题

| 标签 | 结论 |
| --- | --- |
| 代码已存在 | gateway 对移动端订单/轨迹/车辆部分响应 ID 做了 string 化，例如 `mobileOrderItemResponse`、`mobileTrackPointFromProto`。 |
| 仍未闭环 | gateway 多处 `parseInt` 忽略错误，非法 ID 会变成 0；管理端多个页面仍用 `el-input-number` 输入 `orderId/driverId/userId`，JS 对 64 位 ID 有精度风险。 |

治本方向：所有业务 ID 在 HTTP JSON 层统一为 string；后端 DTO 接收 string 后用 `strconv.ParseInt/ParseUint` 严格校验，错误返回中文 400；管理端大 ID 输入全部改文本输入并加格式校验。

### 5.2 金额精度问题

| 标签 | 结论 |
| --- | --- |
| 代码已存在 | order proto 使用 `double total_price`，trip proto 使用 `double price`，order/trip/admin finance/marketing 多处使用 `float64`。 |
| 仍未闭环 | 支付金额 `formatAlipayAmount(reply.GetOrder().GetTotalPrice())` 直接从 float64 格式化，订单、优惠券、退款、财务对账都没有根上统一为分或 decimal string。 |

治本方向：HTTP DTO 金额统一 decimal string，服务内部统一 `amount_cent int64` 或定点 decimal；支付宝请求只从分转换为两位字符串。

### 5.3 状态统一问题

| 标签 | 结论 |
| --- | --- |
| 文档已确认 | 目标状态应覆盖 `waiting_pay/paid/accepted/in_progress/completed/cancelled/refunding/refunded`。 |
| 代码已存在 | order-service 当前只有 `pending/confirmed/completed/cancelled` 四态；gateway 映射为 `pending/accepted/completed/cancelled`。 |
| 仍未闭环 | `waiting_pay` 与 `paid` 都被映射为 `0`，`in_progress` 被映射为 `1`，状态含义丢失。 |

治本方向：先定义唯一订单状态枚举和流转矩阵，再同步 proto、DB、gateway response、乘客端、司机端、管理端。

### 5.4 支付闭环问题

| 标签 | 结论 |
| --- | --- |
| 代码已存在 | `pkg/alipayx` 已封装支付宝沙箱 WAP 支付和通知验签；gateway 有 `/carpool/orders/{id}/pay` 和 `/api/v1/pay/notify`。 |
| 仍未闭环 | 通知验签成功后只返回 `success`，没有根据 `out_trade_no` 找订单、校验 `app_id/total_amount/trade_status`、更新订单 paid、写支付流水、处理重复通知。 |

治本方向：order-service 增加支付 RPC/DTO 和 payment 表；gateway notify 验签后调用 order-service `MarkPaid`，用支付单号做幂等。

### 5.5 事务、锁、超卖与状态漂移

| 标签 | 结论 |
| --- | --- |
| 代码已存在 | 创建订单 `CreateAtomic` 已用事务和条件扣座控制创建时超卖。 |
| 仍未闭环 | 取消/拒单退座、接单/完成状态更新缺少同事务、期望状态、version 或行锁；前端传幂等键但后端不消费。 |

治本方向：状态变更统一走 `WHERE id=? AND status=?` 原子更新；涉及退座的操作必须与订单状态变更同事务；接单/取消/拒单/支付回调使用幂等键或唯一业务单号。

### 5.6 乱码问题

| 标签 | 结论 |
| --- | --- |
| 代码已存在 | 乘客端/司机端多个 Vue 页面和 request 工具中存在可见中文乱码；gateway 支付成功页、AI 默认文案也有乱码。 |
| 仍未闭环 | 虽然 SFC 语法解析通过，但用户看到的提示、按钮、状态、错误文案不可读，部分模板文本也存在破损风险。 |

治本方向：统一 UTF-8 编码重写移动端中文文案；后端错误文案统一中文；提交前增加乱码扫描检查。

### 5.7 500 报错和无日志问题

| 标签 | 结论 |
| --- | --- |
| 代码已存在 | 部分支付、地图、位置接口有 zap 日志；移动端 request 也有 console 日志。 |
| 仍未闭环 | gateway `returnData` 只特殊处理 `UpstreamHTTPError`，其它错误直接 return，容易变成 Kratos 500；`parseInt` 静默吞错；多数业务错误没有带 trace/order/user 字段日志。 |

治本方向：gateway 增加统一错误映射和请求链路日志；参数错误返回中文 400；业务错误返回中文可读 code/msg；500 只用于未知系统错误并记录完整业务上下文。

### 5.8 占位接口假闭环

| 标签 | 结论 |
| --- | --- |
| 代码已存在 | `mobile_compat.go` 用占位方式兼容了优惠券、需求、评价查询、删除行程接口。 |
| 仍未闭环 | 前端按钮真实请求了这些地址，但后端没有真实 service/DTO/表/状态规则。 |

治本方向：删除“假成功/固定空列表”的错觉，按业务模块补真实服务；暂未实现的接口要返回明确中文不可用，并在前端隐藏或降级按钮。

### 5.9 管理端和移动端数据源不一致

| 标签 | 结论 |
| --- | --- |
| 代码已存在 | 管理端 order service 使用 `order_main`，移动端 order-service 使用 `carpool_order`。 |
| 仍未闭环 | 管理端订单、财务、营销、调度不是直接管理移动端真实订单闭环。 |

治本方向：建立统一业务表或只读视图/同步适配层；管理端操作必须落到同一订单状态机和同一支付/退款链路。

## 6. 建议修复批次

### 第 1 批：契约根修复

目标：先统一 ID、金额、状态、错误返回，不做业务扩展。  
范围：proto、gateway DTO、移动端/管理端 API 类型适配、统一错误映射。  
验收：订单列表/详情/下单/取消/接单接口入参出参类型一致，非法 ID 不再变 0，金额不再 float64 传递，错误中文可读。

### 第 2 批：支付闭环

目标：支付宝沙箱支付从发起到回调到订单 paid 真闭环。  
范围：`pkg/alipayx` 保留；新增/完善 order-service 支付 DTO/RPC/表；gateway notify 调 order-service。  
验收：支付成功回调后订单状态变 paid，重复通知不重复更新，金额/app_id/trade_status 校验失败不改状态。

### 第 3 批：订单状态与司机履约闭环

目标：司机接单、到达上车点、开始送达、完成订单完整闭环。  
范围：order proto 增加 complete/start pickup/start trip 等必要 RPC；状态机统一；位置上报绑定有效订单。  
验收：司机端“接单 -> 上报位置 -> 完成订单”能驱动乘客端地图和订单状态同步变化。

### 第 4 批：占位接口转真实服务

目标：优惠券、需求发布、评价查询、删除行程从占位接口变真实接口。  
范围：coupon/demand/review mine/trip delete 的 DTO、service、表、事务、幂等。  
验收：乘客能发布需求、领券、查询券、评价状态；司机删除行程有真实规则和中文错误。

### 第 5 批：管理端真实联动

目标：管理端订单、财务、营销、调度接入移动端真实数据源。  
范围：admin-platform server service/model/query adapter，必要时建立统一视图。  
验收：管理端看到的订单、支付、退款、优惠券、调度与乘客端/司机端同一业务数据一致。

### 第 6 批：乱码、日志、导航体验修复

目标：清理中文乱码，补齐业务日志，修复回退/登录导航体验。  
范围：移动端页面文案、request 工具、gateway/admin 错误文案与日志字段。  
验收：按钮报错中文可读；关键接口日志包含 user/order/payment/trace；页面回退回上一页，不误回登录页。

## 7. 自测记录

| 自测项 | 命令/方式 | 结果 |
| --- | --- | --- |
| gateway server 单元测试 | `go test ./services/gateway-service/internal/server` | 已测试通过 |
| 管理端前端构建 | 在 `admin-platform/web` 执行 `npm run build` | 已测试通过 |
| 乘客端/司机端 SFC 语法 | 使用 Vue SFC compiler 扫描 `src/pages/**/*.vue` | 已测试通过 |
| 乘客端/司机端 H5 构建 | `npm run build:h5` | 120 秒超时，未判定失败 |
| 本轮代码修改 | 不适用 | 本轮未修改业务代码 |

## 8. 后续待处理问题登记

| 编号 | 问题 | 标签 | 处理规则 |
| --- | --- | --- | --- |
| D-001 | 移动端中文乱码影响按钮、弹窗、错误提示、地图提示 | 仍未闭环 | 进入第 6 批，不插入当前批次修复 |
| D-002 | 支付密钥配置需确认只保留支付宝沙箱，不接其它支付渠道 | 文档已确认 | 进入第 2 批 |
| D-003 | 高德地图查询接口存在，但接驾/送达实时地图状态闭环不足 | 代码已存在，仍未闭环 | 进入第 3 批 |
| D-004 | 管理端权限种子和按钮权限需联调确认 | 仍未闭环 | 进入第 5 批 |
| D-005 | 回退直接进入登录页的问题需要结合路由栈和 401 刷新失败场景验证 | 仍未闭环 | 进入第 6 批 |

