# 知识库索引

教学知识点资料库，按「学科 / 技术方向 / 主题」组织，用于检索、复习、授课与知识沉淀。

## 后端 - Golang / ORM

- [GORM：定义外键 ID 复用作审计同步键导致新增冲突](后端-Golang/GORM定义外键ID复用作审计同步键导致新增冲突.md) — 审核流同步正式表时，用原 ID 做存在性匹配，避免车牌变更后误走 Create 触发主键冲突。
- [模型与库表字段漂移导致MySQL1364缺默认值](后端-Golang/模型与库表字段漂移导致MySQL1364缺默认值.md) — trip-service 模型缺 `publisher_id` 导致 INSERT 缺非空列 → `Field ... doesn't have a default value`。
- [AutoMigrate收紧NOTNULL因存量NULL行报1265](后端-Golang/AutoMigrate收紧NOTNULL因存量NULL行报1265.md) — 探针插入不带 `created_at` 留下 NULL 行，管理端启动收紧列时报 `Error 1265`。
- [protojson响应被json解码引发网关500](后端-Golang/protojson响应被json解码引发网关500.md) — 网关上行客户端用 json.Decode 解码 protojson（int64 为字符串）导致所有聚合调用 500「服务开小差」。
- [管理端与移动端同源：person 管理从 mock 表迁移到真实业务表](后端-Golang/管理端与移动端同源person管理迁移.md) — 人员管理改读写 auth_user/driver_profile 等真实表，状态/角色映射、多表关键词搜索与事务写入要点。
- [登录成功但资料手机号为空：懒创建 EnsureXxx 未回填登录手机号](后端-Golang/懒创建资料未回填登录手机号.md) — 账号与资料分服务时，登录成功后需把 principal 回填进资料表，EnsureXxx 幂等补 phone、网关非致命编排。
- [登录成功但首页/订单/我的报「服务开小差」：下游微服务未启动](后端-Golang/登录成功但页面报服务开小差-下游未启动.md) — 一句「服务开小差」对应多个根因；用 netstat 端口二分 + 直连下游确认，先起齐 7 个服务再测。
- [网关类型断言仅部分客户端实现导致误报「XX服务未配置」](后端-Golang/网关类型断言仅部分客户端实现导致服务未配置.md) — biz 用断言探测能力，但 HTTP 客户端漏实现时运行时退化；补全客户端方法并同步进接口契约。
- [司机发布行程链路排查：地点校验匹配逻辑、网关吞错与15分钟竞态](后端-Golang/司机发布行程链路排查-校验匹配与网关吞错.md) — 网关不透传下游错误会把 400 误报成 500；校验逻辑须匹配前端数据能力；15 分钟压线竞态。
- [司机发布行程：间歇503、发布400、行程审核入口缺失三连排查](后端-Golang/间歇503与发布无效的网关trip链路.md) — 服务超时<上游高德耗时→间歇503；幂等接口漏传 request_id 直连必 400；gva 管理端菜单与路由分源，页面存在≠菜单可见。
- [司机端地址联想搜索与地图渲染：proto→服务→网关→前端全链路](后端-Golang/司机端地址联想搜索与地图渲染全链路.md) — 新增一个 RPC 能力的五处同步改动；protoc 三件套重生成；protojson 输出 snake_case；联想请求须 silent。
- [kratos 路由注册顺序：静态路径被通配段 `{id}` 抢占](后端-Golang/kratos路由注册顺序静态路径被通配段抢占.md) — 网关静态集合路由（mine/pending）注册在 `{id}` 通配路径之后会被抢匹配，须静态在前；`/trips/0` 与 `/trips/mine` 同错即为路由抢占铁证。
- [kratos 网关分层与 HTTP/gRPC 双客户端切换](后端-Golang/kratos网关分层与HTTP-gRPC双客户端切换.md) — 网关五层调用链（server→service→biz→data→下游）；每个下游都有 gRPC/HTTP 双 client，由 discovery 是否为 nil 决定；本地 nacos 关闭时实际走 HTTP 端口。

## 前端 - uni-app

- [严格RFC3339时间与缺少失败分支导致发布静默失败](前端-uni-app/严格RFC3339时间与缺少失败分支导致发布静默失败.md) — 发布行程回填严格 RFC3339 的 `depart_time`/`arrive_time` 并补失败反馈，避免「点击无反应」。
- [uni-app H5 端地图：manifest 配置 sdkConfigs 与安全密钥](前端-uni-app/uni-app-H5地图sdkConfigs配置与安全密钥.md) — H5 `<map>` 需 `h5.sdkConfigs.maps.amap` 而非顶层 aMapKey；新 key 必须配 securityJsCode；地图需显式高度与中心点。