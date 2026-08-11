# kratos 路由注册顺序：静态路径被通配段 `{id}` 抢占

- 创建日期：2026-08-10
- 最近更新：2026-08-10
- 标签：kratos, httprouter, 路由冲突, 网关, 排障

## 适用场景

使用 kratos（底层 httprouter 风格路由树）的网关在新增「集合类」REST 接口（如 `/carpool/trips/mine`、`/carpool/orders/pending`）后，请求本应命中新路由，却总是落到既有 `/xxx/{id}` 动态路由上，响应也跟着变成「invalid trip」「invalid order」等语义错误。

## 问题背景

网关注册了两条路由（原 http.go 顺序）：

```go
router.GET("/carpool/trips/{id}", ...)   // 注册在前（约第 120 行）
router.GET("/carpool/trips/mine", ...)   // 注册在后（约第 139 行）
```

请求 `GET /carpool/trips/mine` 后网关路由树把 `mine` 当作 `{id}` 参数：

- `id = "mine"` 被解析成 0
- 转发到下游 `GetTripDetail(0)`
- trip-service 返回 `ErrInvalidTrip` → 网关返回 **400「invalid trip」**，而非该行程列表

## 核心结论

- kratos/httprouter 按**注册顺序**匹配路由；`{id}` 这类通配段会抢占同一层级下注册在其**后**的静态路径。
- 解决方案：**把静态路径注册在通配路径之前**（`mine` 提到 `{id}` 前）。
- 排查时先做二分验证：`/carpool/trips/0` 与 `/carpool/trips/mine` 返回完全一致的错误，即可断定「误命中动态路由」，而非 handler 或下游逻辑问题。
- 即使两个语义都正常走通，也会带来返回值类型混乱；必须保证同类路由中「静态 > 动态」的注册顺序。

## 关键原理

httprouter 路由树按段建立前缀树，参数段（kratos 语法是通配 **`{id}`**）匹配任意非空段。同层静态段与参数段并存时，匹配优先级取决于**插入顺序**：先插入参数段的路径也会先匹配到参数段，后插入的静态段不会覆盖先插入的参数段匹配。

这与大多数框架「静态段永远优先于参数段」的实现策略不同（例如 Express 5 即静态优先）。因此在 kratos 中**顺序即优先级**，是开发该类 API 最常见的一个坑。

## 示例：正确注册顺序

```go
router.GET("/carpool/trips/mine", listMyTripsHandler)        // ① 静态在前
router.GET("/carpool/trips/{id}", tripDetailHandler)          // ② 动态在后
```

同样检查其余集合接口，例如 `pending` 必须注册在 `/carpool/orders/{id}` 之前。

## 常见误区或注意事项

- 只改 handler 内容无效：即使让 mine handler 直接返回成功，上文路由树仍把它当 `{id}` 分发给 detail handler，响应不会变化。这是「排除 handler 执行」的强证据。
- 新增任何静态集合路径时都要意识到顺序问题，最好的办法是在路由注册处维护注释：静态路径统一放最前。
- 遗留调试代码（重复注册、调试响应头）在定位后必须清理，避免二次踩坑。

## 延伸方向

- httprouter 的 `SpecialRouter` / 中间件内手动 `r.Lookup` 做静态优先的兼容方案。
- 若无法调整注册顺序（某些框架固定），可在业务层对 `id` 做白名单解析（`mine/pending` → 走列表逻辑）。