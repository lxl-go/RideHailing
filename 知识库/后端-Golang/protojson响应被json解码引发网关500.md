# Web 服务「服务开小差了」系统性故障：HTTP 客户端用 json.Decode 解析 protojson 响应

## 适用场景 / 问题背景

网关（gateway-service）负责聚合各下游微服务（trip / passenger / order / review / driver / auth）。
某批次 App 端（乘客端：首页、订单、我的、发布需求、优惠券；司机端：发布行程）出现同一症状——所有需要调用下游数据的功能都返回 HTTP 500，提示「服务开小差了，请稍后重试」。排查时发现：

- 绕过网关直调下游服务（如 `GET http://127.0.0.1:9040/v1/trips?page=1&page_size=5`）返回 200 且数据正常；
- 经网关转发同一请求却返回 500 / 空 body；
- 只有鉴权（auth）相关接口行为正常（权限不足时能正确返回 403），说明鉴权链路没问题。

## 核心结论

**根因不是下游服务故障，而是网关的 HTTP 上行客户端用 `encoding/json` 解码了 protojson 编码格式的响应。**

- 上游微服务（基于 go-kratos + protobuf）的 HTTP 响应体遵循 protojson 规范：**int64 类型的字段以 JSON 字符串形式序列化**（例如 `"total":"0"`、`"id":"999991"`），不是数字。
- 网关多个上行客户端（trip / passenger / order / review）用 `json.NewDecoder(resp.Body).Decode(out)` 直接解码，遇到字符串形式的 int64 立即报错：
  `json: cannot unmarshal string into Go struct field ... of type int64`。
- 该错误向上抛，被网关 `returnGatewayError` 捕获，最终走默认分支返回 HTTP 500「服务开小差了，请稍后重试」。

对照：`auth_client.go` 与 `driver_client.go` 已经正确使用 `protojson.UnmarshalOptions{DiscardUnknown: true}`，所以鉴权与司机模块的调用正常，这也让故障定位看起来「很散」。

## 关键原理

### protojson 的 int64 序列化规则

protobuf 的 `int64`/`uint64` 在 JSON 映射（ETag/JSON mapping）中必须用字符串，以保证 JS 端大整数精度不丢失。因此 protojson 会输出：

```json
{ "total": "0", "id": "999998244" }
```

而 Go 标准库 `encoding/json` 不会把 string 自动转成 `int64`，一遇到就报错（不会降级为 `0`）。

### 正确的解码方式

当目标结构体是 protobuf 生成的消息（实现了 `proto.Message`）时，应使用 protojson：

```go
if msg, ok := out.(proto.Message); ok {
    body, _ := io.ReadAll(resp.Body)
    return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(body, msg)
}
return json.Unmarshal(body, out) // 普通结构体仍走标准库
```

`DiscardUnknown: true` 让对方新增字段不会因未知字段而失败，提高前后端迭代兼容性。

## 示例（错误 vs 正确）

请求：`GET /v1/trips?page=1&page_size=5`

上游实际响应体：

```json
{"total":"0","items":[]}
```

- 错误解码（`json.Unmarshal` 到 struct，total 为 int64）：报 `cannot unmarshal string into Go struct field ... total of type int64`，进而 500。
- 正确解码（`protojson.Unmarshal`）：正常得到 `Total=0`。

## 常见误区 / 注意事项

- 误以为 500 是下游 DB 挂了：直连下游 200，网关 500，说明问题在网关的「解码」或「转发」，而不是数据层。可先做「绕网关直调 + 对比响应体」的二分定位。
- 不要对所有接口一律改用 `protojson`——只有目标是 `proto.Message` 时才用；普通业务结构体仍用 `json`。
- 同一个 HTTP 客户端里 `execute` 解码和「错误响应解析」是两处逻辑；错误分支用到的 `json.Unmarshal(&payload)` 解析错误体本身是普通 struct，保持标准库即可。
- 排查时先看所有上行子客户（各 service 的 `*_client.go`）解码是否统一，容易发现「auth 用了 protojson、其他没统一」的历史遗留。

## 延伸方向

- 统一网关上行客户端解码基类/工具函数，避免各子客户端各自实现、各自踩坑。
- 在网关层增加对上游 500 的处理：上游返回 `internal error` 时，上游日志中打印可观察的错误栈，而网关层负责「服务开小差」的业务提示，两者职责分离。
- 可沉淀为代码评审 check：网关聚合层的 HTTP 上行客户端遇到「请求体/响应体含 protobuf 结构」时，解码一律用 protojson。

## 文档信息

- 创建日期：2026-08-06
- 最近更新：2026-08-06
- 修复状态：已按上述方案落地于 gateway-service `internal/data/trip_client.go` / `passenger_client.go` / `order_client.go` / `review_client.go`，`go build`、`go vet`、`internal/data` 单测通过，重启网关后 5 个典型接口实测返回 200。
- 标签：#go #kratos #protojson #encoding_json #HTTP客户端 #gateway #int64序列化 #线上故障排查