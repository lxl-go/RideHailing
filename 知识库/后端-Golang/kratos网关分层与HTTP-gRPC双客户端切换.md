# kratos 网关分层与 HTTP/gRPC 双客户端切换

- 创建日期：2026-08-10
- 最近更新：2026-08-10
- 标签：kratos, 网关, 微服务, gRPC, HTTP, wire, nacos

## 适用场景

阅读或扩展 kratos 微服务网关（gateway-service）时，理解一次「客户端请求 → 下游服务」请求如何流转；以及「为什么配置里写了 gRPC 却实际走 HTTP」。

## 核心结论

- 网关不是"没路由、没调其他服务"，而是标准 kratos BFF 分层：**server（路由）→ service → biz（usecase）→ data（client）→ 下游**。
- 每个下游服务都实现了**两个 client**（gRPC 与 HTTP），由 `NewXxxClient` 根据是否存在服务发现自动选择。
- 当前本地 `registry.nacos.enabled: false` → `discovery == nil` → 实际走 **HTTP over localhost 端口**；开启 nacos 后同一套代码自动切到 gRPC。

## 关键原理：五层调用链

以「发布行程 POST /carpool/trips」为例：

| 层 | 文件 | 作用 |
|---|---|---|
| server | internal/server/http.go | kratos 路由，绑定参数、鉴权（requirePermission）、统一响应 returnData |
| service | internal/service/trip.go | 薄透传层，`TripService.PublishTrip` → `uc.PublishTrip` |
| biz | internal/biz/trip.go | usecase，持有 `data.TripClient` 接口 |
| data | internal/data/trip_client.go | 真正发 HTTP/gRPC 的 client 实现 |
| 下游 | trip-service 9040 | 处理 `/v1/driver/trips`，写库 |

典型路由 handler 骨架（http.go）：
```go
router.POST("/carpool/trips", func(ctx khttp.Context) error {
    if !requirePermission(...) { return nil }
    req := new(tripv1.PublishTripRequest)
    ctx.Bind(req)
    reply, err := tripSvc.PublishTrip(ctx, req)
    return returnData(ctx, reply, err)
})
```

## 双客户端切换原理

data.go 的工厂函数：
```go
func NewTripClient(c *conf.Clients, discovery registry.Discovery) (TripClient, error) {
    baseURL := "http://127.0.0.1:9040"                 // 兜底
    if c.Trip.HTTPBaseURL != "" { baseURL = c.Trip.HTTPBaseURL }
    endpoint := c.Trip.Endpoint                        // "discovery:///trip-service"
    if discovery != nil && strings.HasPrefix(endpoint, "discovery:///") {
        conn, _ := grpcx.DialInsecure(ctx, endpoint, discovery, ...)
        return NewTripGRPCClient(tripv1.NewTripServiceClient(conn)), nil  // gRPC
    }
    return NewTripHTTPClient(baseURL), nil              // HTTP
}
```

决定因素：
- `main.go` 只在 `registry.nacos.enabled == true` 时创建 `discovery`，否则为 nil → 必然走 HTTP 分支。
- HTTP client 用 `doJSON`/`get`，把请求发到下游 HTTP 端口（如 9040），并用 protojson 反序列化响应。
- gRPC client 直接调用 `tripv1.TripServiceClient` 的方法（下游 9140 gRPC 端口）。

## 常见误区或注意事项

- 「配置里写了 endpoint: discovery:///xxx」≠ 走 gRPC，必须 nacos 开启且 discovery 非 nil 才生效。本地联调默认走 http_base_url。
- 网关对客户端暴露 `/carpool/*`，对下游是 `/v1/*`——两套路径，别混用。
- biz 层部分方法用「类型断言探测能力」：`uc.client.(interface{...})`，万一 client 未实现该方法会返回 `Unimplemented`——新增 RPC 时必须同时补 gRPC 与 HTTP 两个 client 的方法。

## 延伸方向

- 开启 nacos：`configs/nacos.yaml` + `registry.nacos.enabled: true`，各微服务注册 discovery:/// 服务名。
- 网关聚合编排：一个 /carpool/* 路由可串行调用多个下游（例如登录后 seed passenger/driver profile）。