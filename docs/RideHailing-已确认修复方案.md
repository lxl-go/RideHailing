# RideHailing 已确认修复方案

> 日期：2026-08-04  
> 范围：Go/Kratos 微服务、Gin 管理端、司机端 uni-app、乘客端 uni-app  
> 目的：把已确认的支付、地图、精度、状态、事务、DTO 修复口径固化，避免后续修复时遗漏或降级。

## 1. 已确认原则

1. 支付只接支付宝沙箱，不做泛化支付占位。
2. 项目必须接入地图能力，司机接乘客、送乘客过程中司机端和乘客端都要实时显示地图、司机位置、路线和轨迹。
3. 64 位业务 ID 对前端统一输出 string，前端不得再把雪花 ID 转成 Number。
4. 金额不得使用 float64/double 表达，统一改为 int64 分或 decimal 字符串。
5. 订单状态只能由后端状态机流转，前端只消费统一 DTO 状态，不自行猜枚举。
6. 订单状态更新、座位扣减/回补、支付流水、财务入账必须在事务或可靠事件中同成同败。
7. 管理端不得直接返回 GORM Model，必须通过 DTO 层白名单输出并脱敏。
8. 配置密钥只允许进入本地配置或环境变量，文档和仓库只保留模板，不保存私钥明文。

## 2. P0 修复：支付宝沙箱支付闭环

### 2.1 当前问题

- `POST /carpool/orders/{id}/pay` 仍在 `services/gateway-service/internal/server/mobile_compat.go` 返回 `payment service unavailable`。
- 当前仓库没有 `pkg/alipayx` 或等价支付宝封装。
- `order-service` 只有 pending/confirmed/completed/cancelled 四个状态，缺少支付态、支付流水、支付回调幂等处理。
- 管理端财务读独立表，移动端支付成功后即使实现，也需要同步到管理端可见的数据口径。

### 2.2 包封装要求

新增 `pkg/alipayx`，参考现有 `pkg/cozex` 的第三方 API 封装风格。

必须提供：

- `Config`：`AppID`、`PrivateKey`、`AlipayPublicKey`、`Production`、`NotifyURL`、`ReturnURL`、`Timeout`
- `Client`
- `CreateWapPay`：H5/网页支付，适配 uni-app H5
- `CreateAppPay`：App 支付 order string，适配 uni-app App 打包
- `VerifyNotify`：验签支付宝异步通知
- `QueryTrade`：主动查单兜底
- `Refund`：后续退款闭环预留
- RSA2 签名、验签、参数排序、UTF-8 编码、错误体裁剪日志

配置模板如下，真实私钥和支付宝公钥不得提交到仓库：

```yaml
alipay:
  app_id: "9021000161685035"
  private_key: "${ALIPAY_PRIVATE_KEY}"
  alipay_public_key: "${ALIPAY_PUBLIC_KEY}"
  production: false
  notify_url: "http://3742778a.r22.cpolar.top/api/v1/pay/notify"
  return_url: "http://3742778a.r22.cpolar.top/pay/success"
```

本地环境变量模板：

```powershell
$env:ALIPAY_PRIVATE_KEY="你的支付宝应用私钥"
$env:ALIPAY_PUBLIC_KEY="支付宝公钥"
```

### 2.3 支付接口闭环

短期不强行新增 `payment-service`，先按当前仓库结构落地：

1. `gateway-service` 增加支付宝配置读取。
2. `gateway-service` 实现 `POST /carpool/orders/{id}/pay`：
   - 从当前登录乘客取 `passenger_id`
   - 查询订单详情，校验订单归属、金额、状态
   - 生成平台支付流水号 `payment_no`
   - 调 `pkg/alipayx` 生成支付表单或 app order string
   - 返回 DTO：`paymentNo`、`orderId`、`amountCent`、`payForm/payUrl/orderString`
3. 新增 `POST /api/v1/pay/notify`：
   - 用 `pkg/alipayx.VerifyNotify` 验签
   - 校验 `app_id`、`out_trade_no`、`total_amount`、交易状态
   - 幂等处理，同一 `trade_no/out_trade_no` 重复通知只能成功一次
   - 调 `order-service` 支付确认方法，在事务内写支付流水并更新订单支付态
4. `order-service` 新增支付相关 RPC：
   - `PreparePayment`
   - `MarkPaid`
   - `GetPayment`
   - 后续可补 `Refund`
5. 管理端订单/财务读取同一真实支付口径，不能继续只读独立 seed 表。

### 2.4 订单状态建议

统一订单状态，后端用 int/枚举，DTO 输出 string：

| 后端状态 | DTO 状态 | 含义 |
|---|---|---|
| 0 | `waiting_pay` | 待支付 |
| 1 | `paid` | 已支付待接单 |
| 2 | `accepted` | 司机已接单，去接乘客 |
| 3 | `in_progress` | 已接到乘客，送往目的地 |
| 4 | `completed` | 已完成 |
| 5 | `cancelled` | 已取消 |
| 6 | `refunding` | 退款中 |
| 7 | `refunded` | 已退款 |

前端不得直接判断数字状态，只判断 DTO 字符串状态。

## 3. P0 修复：高德地图与实时轨迹闭环

### 3.1 当前问题

- 司机端有位置上报和轨迹回放 API，但实时地图体验不完整。
- 乘客端跟踪页只能查轨迹，未形成司机实时位置、路线、接驾/送达两个阶段的完整地图体验。
- `mobile_ai_dispatch.go` 的积水报告是内存存储，和管理端 AI/地图数据不可靠互通。
- 仓库没有 `pkg/amapx` 封装。

### 3.2 包封装要求

新增 `pkg/amapx`，封装高德 Web 服务 API。前端展示地图可以使用平台地图组件或高德 JS SDK，但路线规划、地理编码、逆地理编码、天气等服务端能力优先通过后端代理，避免把服务端 key 滥用到客户端。

配置模板：

```yaml
amap:
  web_key: "${AMAP_WEB_KEY}"
  timeout: "5s"
```

本地环境变量模板：

```powershell
$env:AMAP_WEB_KEY="你的高德 Web 服务 Key"
```

`pkg/amapx` 必须提供：

- `Geocode(address, city)`
- `Regeo(lat, lng)`
- `DrivingRoute(origin, destination, strategy)`
- `Distance(origins, destination)`
- `Weather(city)`
- 统一错误处理、超时、响应码校验、请求日志脱敏

### 3.3 司机端地图流程

司机接单后进入订单详情：

1. 状态 `accepted`：地图显示司机当前位置、乘客上车点、去接乘客路线。
2. 司机点击或系统判定“已接到乘客”后，状态进入 `in_progress`。
3. 状态 `in_progress`：地图显示司机当前位置、目的地、送乘客路线。
4. 司机端持续位置上报：
   - 前台每 3-5 秒上报一次
   - 后台/弱网降级为 10-15 秒
   - 弱网缓存最近 10 个点，恢复后补传
5. 到达后调用 `POST /api/v1/driver/orders/{id}/complete`，后端事务内完成订单并触发财务入账。

### 3.4 乘客端地图流程

乘客支付成功后进入订单详情/跟踪页：

1. 状态 `paid`：展示等待司机接单。
2. 状态 `accepted`：地图显示司机实时位置、乘客上车点、预计到达时间。
3. 状态 `in_progress`：地图显示司机当前位置、目的地、剩余距离和预计到达时间。
4. 状态 `completed`：展示完整轨迹回放入口。
5. 实时位置获取优先 WebSocket；如果当前没有 WebSocket 网关，先用 3 秒轮询 `/api/v1/passenger/orders/{id}/track` 过渡。

### 3.5 后端轨迹要求

- `driver-service` 的位置上报写入持久化轨迹表或 Redis Stream，不再只依赖内存。
- 轨迹点使用 `order_id`、`driver_id`、`lat`、`lng`、`speed`、`heading`、`reported_at`。
- 乘客查询轨迹时必须校验订单归属。
- 司机查询轨迹时必须校验司机归属。
- 高德路线规划结果只做展示和 ETA 参考，不作为订单状态唯一判定依据。

## 4. P0/P1 修复：ID、金额、状态统一

### 4.1 ID 统一

后端内部继续使用 `int64` 雪花 ID，所有 HTTP DTO 对前端输出 string：

```json
{
  "id": "5001",
  "orderId": "5001",
  "tripId": "1001",
  "driverId": "2001",
  "passengerId": "3001"
}
```

前端规范：

- 禁止对业务 ID 使用 `Number()`、`parseInt()`、`parseFloat()`。
- 路由参数、请求参数、store、组件 props 中的业务 ID 全部保持 string。
- 数值类型只用于座位数、经纬度、速度、评分等非 ID 字段。

需要修复的典型点：

- `apps/passenger-uni-app/uni-app/src/api/order.js` 的 `trip_id: Number(tripId)`
- `apps/passenger-uni-app/uni-app/src/pages/tripDetail/tripDetail.vue` 的 `Number(options.id)`
- `apps/passenger-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue` 的 `Number(options.id)`
- `apps/driver-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue` 的 `Number(options.id)`
- `apps/passenger-uni-app/uni-app/src/pages/tracking/tracking.vue` 的订单号 Number 转换

### 4.2 金额统一

推荐统一为分：

```proto
int64 price_cent = 9;
int64 total_amount_cent = 2;
```

DTO 同时给前端展示字段：

```json
{
  "amountCent": 3980,
  "amountText": "39.80"
}
```

修复范围：

- `trip-service`：`price double/float64` 改为 `price_cent int64`
- `order-service`：`total_price double/float64` 改为 `total_amount_cent int64`
- `admin-platform`：订单、财务、营销优惠券等资金字段改为分或 decimal 类型 DTO
- 前端展示统一通过 `formatAmount(amountCent)`，不直接拼 `¥{{ amount }}`

### 4.3 状态统一

后端定义状态机，网关输出字符串 DTO。移动端 `normalizeOrder` 必须映射状态：

```js
const ORDER_STATUS = {
  0: 'waiting_pay',
  1: 'paid',
  2: 'accepted',
  3: 'in_progress',
  4: 'completed',
  5: 'cancelled',
  6: 'refunding',
  7: 'refunded'
}
```

状态流转必须白名单：

- `waiting_pay -> paid`
- `waiting_pay -> cancelled`
- `paid -> accepted`
- `paid -> cancelled/refunding`
- `accepted -> in_progress`
- `accepted -> cancelled/refunding`
- `in_progress -> completed`
- `completed -> refunding`
- `refunding -> refunded/cancelled`

## 5. P1 修复：事务、锁、幂等

### 5.1 订单创建

当前 `CreateAtomic` 已有条件扣座位：

- `WHERE id=? AND status=? AND seats_available>=?`
- 事务内扣减座位并创建订单

保留这个思路，但金额改为分，状态改为 `waiting_pay` 或 `paid` 取决于业务最终选择。

### 5.2 订单流转

新增 repo 层原子方法，替代业务层多次独立调用：

- `CancelOrderTx(orderID, actorID, expectedStatus, reason)`
- `AcceptOrderTx(orderID, driverID, expectedStatus)`
- `StartTripTx(orderID, driverID, expectedStatus)`
- `CompleteOrderTx(orderID, driverID, expectedStatus)`
- `MarkPaidTx(orderID, paymentNo, tradeNo, amountCent)`

## 6. 本轮已落地

- 乘客端订单详情页已接入 `POST /carpool/orders/{id}/pay`，可拉起支付宝沙箱 WAP 表单。
- 网关已新增 `POST /api/v1/pay/notify`，以及 `GET /pay/success` 回跳页，避免支付链路 404。
- 司机端“完成订单”已从前端按钮真正打到后端接口。
- 前端登录已改为 `reLaunch`，避免回退直接回到登录页。
- 前后端请求日志已补充，便于继续按按钮 -> 日志 -> 定位 -> 修复的流程排查。

所有状态更新必须带条件：

```sql
UPDATE carpool_order
SET status = ?, version = version + 1
WHERE id = ? AND status = ? AND version = ?
```

涉及座位回补、支付流水、财务流水、状态历史时，必须在一个事务中完成。

### 5.3 支付幂等

- `payment_no` 唯一索引
- `alipay_trade_no` 唯一索引
- `notify_id` 或原始通知摘要去重
- 支付金额必须等于订单金额
- 支付成功回调重复到达时返回 success，但不得重复改订单或重复入账

## 6. P1 修复：管理端 DTO 和三端协调

### 6.1 管理端订单数据源

管理端订单列表、详情、统计必须接入真实订单口径：

- 短期：管理端直接读 `carpool_order` + `carpool_trip` + 支付流水 DTO
- 中期：由订单/支付服务通过事件同步 `order_main`，但必须有同步任务和一致性校验

不能继续让移动端写 `carpool_order`，管理端读无关的 `order_main` seed 数据。

### 6.2 DTO 输出

订单、财务、营销、派单接口补 response DTO：

- 手机号脱敏：`138****0001`
- 金额输出 `amountCent/amountText`
- ID 输出 string
- 内部字段如 `version`、幂等键、内部流水备注默认不输出
- 只有管理端需要乐观锁时，单独输出 `version`，不要混在通用列表 DTO 中

## 7. P2 修复：乱码、stub 和接口收口

1. 修复 `admin-platform/server/api/v1/carpool/review.go` 的乱码：
   - `"ID鍙傛暟閿欒"` -> `"ID参数错误"`
2. 网关手写 JSON 响应统一：
   - `Content-Type: application/json; charset=utf-8`
3. 清理或实现 `mobile_compat.go` stub：
   - 优惠券列表/领取
   - 我的评价
   - 拼车需求发布/取消/我的需求
   - 支付
   - 删除行程
4. 未完成的接口前端必须置灰或隐藏，不能假成功。

## 8. 建议执行顺序

### 第 1 批：核心闭环

1. 修复乱码和 charset。
2. 新增统一订单 DTO，先解决 ID string 和状态 string。
3. 实现 `pkg/alipayx`。
4. 实现 `/carpool/orders/{id}/pay` 和 `/api/v1/pay/notify`。
5. `order-service` 增加支付态、支付流水、支付成功事务。
6. 实现 `/api/v1/driver/orders/{id}/complete`。

### 第 2 批：一致性和精度

1. 金额字段从 `float64/double` 改为分或 decimal 字符串。
2. 取消、接单、拒单、开始行程、完成行程全部改为事务方法。
3. 状态更新加 expected status、version 或行锁。
4. 前端全量清理业务 ID 的 `Number()/parseInt()`。
5. 移动端 `normalizeOrder/normalizeTrip` 统一 DTO。

### 第 3 批：地图实时体验

1. 新增 `pkg/amapx`。
2. 网关增加路线规划、逆地理编码、天气代理接口。
3. 司机端订单详情接入地图、路线、实时位置上报。
4. 乘客端 tracking 接入地图、司机位置、ETA、轨迹回放。
5. 位置轨迹持久化，内存 store 只作为测试替身。

### 第 4 批：管理端和运营能力

1. 管理端订单接入真实订单数据源。
2. 财务、退款、营销券和支付流水打通。
3. 管理端 DTO 脱敏。
4. 补接口契约测试、并发测试、支付回调幂等测试、移动端页面联调清单。

## 9. 验收标准

1. 乘客创建订单后能拉起支付宝沙箱支付。
2. 支付宝异步回调验签成功后，订单进入 `paid`，重复回调不重复入账。
3. 司机只能接 `paid` 订单，并发接同一订单只有一个成功。
4. 司机接单后，双方地图显示司机位置和去乘客路线。
5. 司机接到乘客后，双方地图切换为去目的地路线。
6. 司机完成订单后，订单进入 `completed`，管理端订单和财务可见。
7. 前端所有业务 ID 都是 string，不再出现雪花 ID 精度丢失。
8. 所有金额展示和计算基于分或 decimal，不再使用 float64/double 进行资金计算。
9. 管理端订单、财务接口不直接暴露 GORM Model。
10. Stub 接口要么真实闭环，要么前端隐藏入口并明确不可用。
