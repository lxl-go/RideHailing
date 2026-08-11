---
创建日期: 2026-08-10
最近更新: 2026-08-10
标签: [网关, trip-service, 高德地图, 超时, 幂等, 行程审核, 管理端菜单]
---

# 司机发布行程：间歇 503、发布 400、行程审核入口缺失 三连排查

## 适用场景 / 问题背景

司机端发布行程后出现多种并发表象，需拆解为四个相互独立的问题：

1. 日志间歇出现 `503 route service unavailable, please retry`
2. 日志出现多次 `400 invalid trip`
3. 行程已入库但「我的行程」看不到 / 审核流程走不通
4. 管理端侧边栏没有「行程审核」菜单；司机端输入地址无地图渲染、无联想下拉

## 核心结论

- **间歇 503 根因不是高德挂了**：trip-service 的 `server.http.timeout: 1s` 小于高德接口的合理耗时（冷启动首个请求实测 2622ms），服务把超过 1s 的高德调用掐断 → 网关层映射成 `route service unavailable`。**修复：把 http/grpc timeout 提到 5s**（configs/config.yaml）。
- **400 invalid trip 的三处来源**：
  1. trip-service 数据层实现了发布幂等（`AcquirePublishRequest` 等），`request_id` 为空会被 `biz.PublishTrip` 直接判 `ErrInvalidTrip`（直连调试时最易踩）；
  2. `DriverID<=0 / Origin/Destination 为空 / Seats 超范围 / DepartTime < now+15min` 也判 `ErrInvalidTrip`；
  3. 网关 `/carpool/trips` 只注入 `DriverId=currentUserID`，不生成 request_id——`request_id` 由前端 `publishTripPayload()` 生成，因此**直连 trip-service 调试发布必须手动带上 `request_id` 和 `driver_id`**。
- **行程审核功能其实已存在**：管理端后端 `carpool/trip` 路由有 `list / :id / :id/review / :id/deactivate`，前端页面 `view/rideHailing/trips/index.vue` 完整（通过/驳回/归档）。**唯一缺口是侧边栏菜单**：`web/src/pinia/modules/router.js` 的 `buildRideHailingMenus()` 未注册 `trips` 入口，所以页面能直接输 URL 访问但菜单不显示。**修复：在该函数中补一条 `trips`（行程审核）菜单项**。
- 「我的行程」查询逻辑正常：`ListByDriver` 过滤 `is_deleted=false`，status=10 也显示；问题在发布侧失败，不是列表侧。

## 关键原理

### 1. 服务超时 vs 上游调用超时 是两套

trip-service 对高德有自己的 `amap.timeout: 3s`，但 kratos HTTP server 的 `server.http.timeout: 1s` 是**接收请求的整体超时**。冷启动/网络抖动时高德响应 2.6s，请求在服务端就被切了，客户端拿到的不是高德错误而是服务端超时/断连，网关再包装成 503。**服务器整体超时必须 ≥ 上游最慢调用耗时**。

### 2. 幂等接口的请求字段契约

`biz.PublishTrip` 开头的能力断言（repo 是否实现 `AcquirePublishRequest`）决定了 `request_id` 是**必填**。这类「接口能力 vs 实现」的隐性契约在调试/直连时最容易漏：测试 payload 少一个字段 → 400，看起来像业务校验不过。

### 3. gva 管理端菜单与路由是两套来源

admin 前端路由（`router/index.js`）静态定义了 `/ride-hailing/*` 全部子页，但**侧边栏菜单**由 `pinia/modules/router.js` 的 `buildRideHailingMenus()` 单独构造。页面文件存在 ≠ 菜单可见，两者必须都注册。

### 4. 验证完整发布链路的直连要点

trip-service 暴露 HTTP：`POST /v1/driver/trips`、`POST /v1/driver/trips/locations/validate`、`POST /v1/driver/trips/price-preview`、`DELETE /v1/driver/trips/{id}`。直连发布需同时带：

```json
{
  "driver_id": 3,
  "request_id": "trip-<ms>-<rand>",
  "origin": { "poiId": "B000A83AJN", "name": "北京南站", "formattedAddress": "永外大街车站路12号" },
  "destination": { "poiId": "B000A83A7X", "name": "北京西站", "formattedAddress": "莲花池东路118号" },
  "depart_time": "2026-08-10T11:00:00+08:00",
  "seats_total": 4
}
```

删除同理要带 `driver_id`（网关注入，直连需自己传）。

## 示例（PowerShell 无 BOM 直连）

PowerShell 5.1 `Set-Content -Encoding UTF8` 会写入 BOM，导致 protojson 解码报 `syntax error (line 1:1)`。必须用无 BOM 写入：

```powershell
[System.IO.File]::WriteAllText("D:\Temp\opencode\pub.json", $payload, (New-Object System.Text.UTF8Encoding($false)))
curl.exe -s -X POST -H "Content-Type: application/json" --data-binary "@D:\Temp\opencode\pub.json" http://localhost:9040/v1/driver/trips
```

## 常见误区 / 注意事项

- 别把「高德接口慢」当「高德挂了」：先看本服务 HTTP/上游整体超时设置是否覆盖上游最慢耗时。
- 直连 trip-service 调试发布不传 `request_id` → 400 invalid trip，容易误判为业务校验 bug。
- 管理端「模块不存在」先分清楚「页面/后端存在但菜单没注册」与「真的没实现」。
- Windows curl 测试含中文 JSON 必须无 BOM，否则报 CODEC syntax error。
- 发布成功后测完要软删清理：`DELETE /v1/driver/trips/{id}?driver_id=3`（`is_deleted=1`）。

## 延伸方向

- 司机端地址联想下拉：前端目前用纯文本 `u-input` + 静默 `validateLocation`，无地图渲染与候选列表，属新功能而非本次缺陷；可基于 amap `/v3/place/text` 扩展候选返回。
- 把 `request_id` 注入下沉到网关（`req.RequestId = newRequestID()`），使下游幂等契约对调用方透明。
