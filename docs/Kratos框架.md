# Kratos框架

Kratos 可以理解为一个“面向微服务工程化”的 Go 框架，不只是 Web 框架。它更像是把 **API 定义、服务分层、RPC/HTTP、配置、注册发现、日志、链路追踪、中间件、错误码、代码生成** 这些微服务常用能力整合成一套规范。

四块讲：**优势、框架结构、微服务搭建方式、代码流转链路**。

------

## **一、Kratos 的核心优势**

Kratos 最大的优势不是“写接口快”，而是 **工程规范强、微服务边界清晰、适合中大型 Go 项目长期维护**。

它的主要优势有：

1. **Proto First**

   Kratos 推荐先定义 `.proto` 文件，再生成 HTTP/gRPC 代码。

   这带来几个好处：

   - API 契约清晰
   - 前后端、服务间接口统一
   - HTTP 和 gRPC 可以共用一套接口定义
   - 文档、校验、客户端 SDK 都可以基于 proto 生成

2. **同时支持 HTTP 和 gRPC**

   一个服务可以同时暴露：

   - HTTP：给前端、uni-app、管理后台、开放接口使用
   - gRPC：给内部微服务之间调用使用

   比如打车系统里：

   - uni-app 调用 `passenger-service` 的 HTTP 接口
   - `order-service` 内部通过 gRPC 调用 `driver-service`、`payment-service`

3. **分层清晰，接近 DDD / Clean Architecture**

   Kratos 默认项目结构一般会分成：

   - `api`：接口契约
   - `service`：接口适配层
   - `biz`：业务逻辑层
   - `data`：数据访问层
   - `server`：HTTP/gRPC 服务启动
   - `conf`：配置定义
   - `cmd`：程序入口

   这样可以避免 Go 项目常见的“大 handler、大 service、大 dao”问题。

4. **微服务基础设施完整**

   Kratos 对这些能力都有比较自然的集成：

   - 服务注册与发现
   - 配置中心
   - 日志
   - 链路追踪
   - 指标监控
   - 中间件
   - 错误码
   - 参数校验
   - 服务间调用
   - 依赖注入 Wire

5. **代码生成能力强**

   Kratos 很依赖代码生成，这对中大型项目是优点：

   - 根据 proto 生成接口代码
   - 生成 HTTP/gRPC 路由绑定
   - 生成 client
   - 生成配置结构体
   - 配合 Wire 生成依赖注入代码

   虽然前期看起来“模板多”，但项目大了之后，它能保证统一性。

------

## **二、Kratos 典型框架结构**

一个 Kratos 服务通常长这样：

```
order-service/
├── api/
│   └── order/v1/
│       ├── order.proto
│       ├── order.pb.go
│       ├── order_grpc.pb.go
│       └── order_http.pb.go
├── cmd/
│   └── order/
│       ├── main.go
│       ├── wire.go
│       └── wire_gen.go
├── internal/
│   ├── conf/
│   │   ├── conf.proto
│   │   └── conf.pb.go
│   ├── server/
│   │   ├── http.go
│   │   └── grpc.go
│   ├── service/
│   │   └── order.go
│   ├── biz/
│   │   └── order.go
│   └── data/
│       ├── data.go
│       └── order.go
├── configs/
│   └── config.yaml
├── go.mod
└── Makefile
```

这几个目录的职责非常重要。

**api**

定义对外接口，例如：

```
service Order {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderReply) {
    option (google.api.http) = {
      post: "/v1/orders"
      body: "*"
    };
  }
}
```

这个接口可以同时生成：

- HTTP 路由：`POST /v1/orders`
- gRPC 方法：`Order/CreateOrder`

**service**

接口实现层，也可以叫 application service。

它接收请求参数，调用业务用例。

```
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderReply, error) {
    order, err := s.uc.CreateOrder(ctx, req.PassengerId, req.StartPoint, req.EndPoint)
    if err != nil {
        return nil, err
    }

    return &v1.CreateOrderReply{
        OrderId: order.ID,
    }, nil
}
```

这一层不应该写太重的业务逻辑。

**biz**

核心业务层。

比如订单创建、派单、取消订单、司机接单、价格计算，都应该在这里。

```
type OrderUsecase struct {
    repo OrderRepo
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, passengerID string, start string, end string) (*Order, error) {
    // 业务校验
    // 计算预估价格
    // 创建订单领域对象
    // 调用 repo 保存
}
```

`biz` 层不关心 MySQL、Redis、MongoDB、HTTP、gRPC 这些细节。

它只依赖接口：

```
type OrderRepo interface {
    Save(context.Context, *Order) error
    FindByID(context.Context, string) (*Order, error)
}
```

**data**

数据访问层，实现 `biz` 里定义的接口。

```
type orderRepo struct {
    data *Data
}

func (r *orderRepo) Save(ctx context.Context, order *biz.Order) error {
    // 写 MySQL
    // 写 Redis
    // 发消息
    return nil
}
```

这里才处理：

- MySQL
- Redis
- MongoDB
- Kafka
- RabbitMQ
- Elasticsearch
- 外部 RPC client

**server**

启动 HTTP/gRPC 服务，把 service 注册进去。

```
func NewHTTPServer(c *conf.Server, order *service.OrderService) *http.Server {
    srv := http.NewServer()
    v1.RegisterOrderHTTPServer(srv, order)
    return srv
}
```

**cmd**

程序入口，加载配置、初始化日志、启动 app。

------

## **三、微服务如何在 Kratos 中搭建**

以网约车系统为例，可以拆成这些服务：

```
user-service        用户服务
passenger-service   乘客服务
driver-service      司机服务
order-service       订单服务
dispatch-service    派单服务
payment-service     支付服务
map-service         地图位置服务
message-service     消息通知服务
admin-service       后台管理服务
gateway/bff         前端网关层
```

推荐架构大概是：

```
uni-app
  |
  | HTTP
  v
API Gateway / BFF
  |
  | gRPC / HTTP
  v
order-service
  |       |        |
  v       v        v
driver  dispatch  payment
service service   service
  |
  v
MySQL / Redis / MQ / Config / Registry
```

在 Kratos 中搭建一个微服务，通常流程是：

1. 创建服务项目

```
kratos new order-service
```

1. 定义 proto API

```
service Order {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderReply);
  rpc CancelOrder(CancelOrderRequest) returns (CancelOrderReply);
  rpc GetOrder(GetOrderRequest) returns (GetOrderReply);
}
```

1. 生成代码

```
make api
```

或者使用 Kratos CLI 生成。

1. 在 `service` 层实现接口

负责把请求转给 `biz`。

1. 在 `biz` 层写业务逻辑

比如：

- 创建订单
- 计算价格
- 判断乘客状态
- 查找附近司机
- 发起派单
- 修改订单状态

1. 在 `data` 层接数据库、缓存、消息队列、其他服务 client

比如订单服务需要：

- MySQL 保存订单
- Redis 缓存订单状态
- gRPC 调用司机服务
- MQ 发布订单创建事件

1. 接入注册中心

比如使用：

- etcd
- Consul
- Nacos
- Kubernetes Service

服务启动后注册自己，其他服务通过服务发现调用它。

1. 接入配置中心和可观测能力

生产项目里通常还要接：

- 配置中心
- 日志
- tracing
- metrics
- health check
- pprof
- graceful shutdown

------

## **四、Kratos 中一次请求的代码流转**

以 `uni-app 用户发起打车订单` 为例。

请求链路大概是：

```
uni-app
  |
  | POST /v1/orders
  v
HTTP Server
  |
  v
Kratos Middleware
  |
  v
Generated HTTP Handler
  |
  v
service.OrderService.CreateOrder
  |
  v
biz.OrderUsecase.CreateOrder
  |
  v
data.orderRepo.Save
  |
  v
MySQL / Redis / MQ / RPC
```

展开讲就是：

**1. uni-app 发起 HTTP 请求**

```
POST /v1/orders
{
  "passenger_id": "P1001",
  "start_point": "A",
  "end_point": "B"
}
```

**2. Kratos HTTP Server 接收请求**

`server/http.go` 里注册了路由：

```
v1.RegisterOrderHTTPServer(srv, orderService)
```

这个路由不是你手写的，是根据 proto 生成的。

**3. 中间件先执行**

比如：

- 日志
- recovery
- tracing
- auth
- rate limit
- metrics
- timeout

请求进入真正业务代码前，会先经过这些 middleware。

**4. 进入 service 层**

```
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderReply, error) {
    return s.uc.CreateOrder(ctx, req)
}
```

这里负责协议适配，不要堆业务。

**5. 进入 biz 层**

```
func (uc *OrderUsecase) CreateOrder(ctx context.Context, req *CreateOrderReq) (*Order, error) {
    // 1. 校验乘客是否可下单
    // 2. 创建订单
    // 3. 计算价格
    // 4. 查找附近司机
    // 5. 保存订单
    // 6. 发布派单事件
}
```

这是核心业务逻辑所在。

**6. 进入 data 层**

```
func (r *orderRepo) Save(ctx context.Context, order *biz.Order) error {
    // INSERT INTO orders ...
}
```

这里负责基础设施细节。

**7. 返回结果**

```
data -> biz -> service -> generated handler -> HTTP response -> uni-app
```

最终前端拿到：

```
{
  "order_id": "O202607290001"
}
```

------

## **五、Kratos 分层的关键思想**

我认为理解 Kratos 最重要的是这句话：

> `service` 负责接请求，`biz` 负责做业务，`data` 负责拿数据，`api` 负责定契约。

不要把所有逻辑都塞进 `service`。

比较好的职责划分是：

```
api      定义接口，不写业务
service  参数转换、调用 usecase、返回 response
biz      核心业务规则、领域逻辑、用例编排
data     数据库、缓存、RPC、MQ、外部系统
server   HTTP/gRPC 服务装配
cmd      程序启动入口
```

------

## **六、Kratos 适合什么场景**

Kratos 特别适合：

- 中大型 Go 后端项目
- 微服务系统
- 需要 HTTP + gRPC 双协议
- API 契约要求强
- 多团队协作
- 需要服务注册发现、配置、链路追踪
- 长期演进的业务系统

这两个点非常关键，属于 Kratos 项目“为什么看起来这么工程化”的核心。

## 七、按服务模块搭建Kratos框架

**每个后端服务都是一个独立的标准 Kratos 服务工程**。它们结构基本一致，只是业务域不同。

例如 `order-service`、`trip-service`、`driver-service` 都长这样：

```
services/
  order-service/
    api/
      order/v1/
        order.proto
        error_reason.proto
        order.pb.go
        order_grpc.pb.go
        order_http.pb.go
        error_reason.pb.go
        error_reason_errors.pb.go

    cmd/
      order/
        main.go
        wire.go
        wire_gen.go

    configs/
      config.yaml

    internal/
      conf/
        conf.proto
        conf.pb.go

      server/
        http.go
        grpc.go
        server.go

      service/
        order.go

      biz/
        order.go
        repo.go
        errors.go

      data/
        data.go
        order.go

    go.mod
    Makefile
```

核心分层职责是：

```
api       接口契约，proto 定义 HTTP/gRPC API
cmd       服务启动入口，加载配置，调用 Wire 初始化 App
conf      配置结构定义，由 conf.proto 生成 Go 配置类型
server    HTTP/gRPC Server 创建与路由注册
service   接口实现层，接收请求，调用 biz
biz       业务逻辑层，订单状态、规则、流程编排
data      数据访问层，MySQL、Redis、MQ、RPC client
configs   YAML 配置文件
```

以 `order-service` 为例：

```
order-service
  |
  | 对外提供：
  |   POST /v1/orders
  |   GET  /v1/orders/{id}
  |   POST /v1/orders/{id}/cancel
  |   POST /v1/orders/{id}/accept
  |
  | 对内提供：
  |   gRPC OrderService
  |
  | 依赖：
  |   user-service
  |   driver-service
  |   trip-service
  |   payment-service
  |   message-service
  |
  | 存储：
  |   MySQL orders
  |   Redis order status
  |   MQ order_created/order_cancelled
```

一次请求流转是：

```
前端 / Gateway
  |
  v
api/order/v1/order_http.pb.go
  |
  v
internal/service/order.go
  |
  v
internal/biz/order.go
  |
  v
internal/data/order.go
  |
  v
MySQL / Redis / MQ / 其他服务
```

比如乘客创建订单：

```
POST /v1/orders
  |
  v
OrderService.CreateOrder
  |
  v
OrderUsecase.CreateOrder
  |
  |-- 校验乘客状态
  |-- 校验行程余座
  |-- 创建订单
  |-- 锁定座位
  |-- 发布订单创建事件
  |
  v
OrderRepo.Save
```

------

每个服务都应该是这种形态：

```
user-service
  api/user/v1
  internal/service
  internal/biz
  internal/data

driver-service
  api/driver/v1
  internal/service
  internal/biz
  internal/data

trip-service
  api/trip/v1
  internal/service
  internal/biz
  internal/data

order-service
  api/order/v1
  internal/service
  internal/biz
  internal/data

payment-service
  api/payment/v1
  internal/service
  internal/biz
  internal/data
```

区别只在业务内容，不在框架结构。

------

`gateway-service` 稍微特殊一点。它也可以是标准 Kratos 工程，但它主要做 **BFF/API 聚合**：

```
services/
  gateway-service/
    api/
      gateway/v1/
        passenger.proto
        driver.proto
        admin.proto

    cmd/
      gateway/
        main.go
        wire.go
        wire_gen.go

    internal/
      conf/
      server/
        http.go
      service/
        passenger.go
        driver.go
        admin.go
      biz/
        passenger.go
        driver.go
        admin.go
      data/
        clients.go
```

Gateway 不直接拥有核心业务表，它主要调用其他服务：

```
gateway-service
  |
  |-- user-service
  |-- driver-service
  |-- trip-service
  |-- order-service
  |-- payment-service
  |-- review-service
```

所以 Gateway 的 `data` 层更多是 gRPC client，而不是数据库 repo。

------

最终整体结构可以是：

```
RideHailing/
  services/
    gateway-service/
    user-service/
    passenger-service/
    driver-service/
    trip-service/
    order-service/
    dispatch-service/
    payment-service/
    review-service/
    message-service/

  apps/
    passenger-uni-app/
    driver-uni-app/
    admin-web/

  pkg/
    logger/
    middleware/
    errors/
    registry/
    tracing/
    validator/

  api/
    common/
      pagination.proto
      error.proto
      money.proto

  deployments/
    docker-compose/
    k8s/

  docs/
```

一句话总结：

> 每个业务服务都是一个标准 Kratos 小工程；乘客端、司机端、管理端只是调用入口，不再决定后端服务边界。

## 八、按端拆服务搭建Kratos框架

按端拆服务，后端框架就是以“乘客端、司机端、管理端”作为服务边界。

也就是说后端仍然保留：

```
passenger-service
driver-service
admin-service
```

每个服务内部再按模块组织订单、行程、评价、财务等能力。

整体结构大概是：

```
RideHailing/
  services/
    passenger-service/
    driver-service/
    admin-service/

  apps/
    passenger-uni-app/
    driver-uni-app/
    admin-web/

  pkg/
    logger/
    middleware/
    errors/
    registry/
    tracing/
    validator/
```

------

**每个端服务内部结构**

以 `passenger-service` 为例：

```
services/
  passenger-service/
    api/
      passenger/v1/
        trip.proto
        order.proto
        review.proto
        finance.proto
        passenger.proto

        trip.pb.go
        trip_grpc.pb.go
        trip_http.pb.go
        order.pb.go
        order_grpc.pb.go
        order_http.pb.go

    cmd/
      passenger/
        main.go
        wire.go
        wire_gen.go

    configs/
      config.yaml

    internal/
      conf/
        conf.proto
        conf.pb.go

      server/
        http.go
        grpc.go
        server.go

      service/
        trip.go
        order.go
        review.go
        finance.go
        passenger.go

      biz/
        trip.go
        order.go
        review.go
        finance.go
        passenger.go
        repo.go
        errors.go

      data/
        data.go
        trip.go
        order.go
        review.go
        finance.go
        passenger.go

    go.mod
    Makefile
```

`driver-service` 类似：

```
services/
  driver-service/
    api/
      driver/v1/
        trip.proto
        order.proto
        review.proto
        income.proto
        driver.proto

    cmd/
      driver/
        main.go
        wire.go
        wire_gen.go

    internal/
      conf/
      server/
      service/
      biz/
      data/
```

`admin-service` 类似：

```
services/
  admin-service/
    api/
      admin/v1/
        user.proto
        driver.proto
        order.proto
        finance.proto
        audit.proto

    cmd/
      admin/
        main.go
        wire.go
        wire_gen.go

    internal/
      conf/
      server/
      service/
      biz/
      data/
```

------

**调用关系**

前端调用会比较直接：

```
乘客端 uni-app
  |
  v
passenger-service HTTP

司机端 uni-app
  |
  v
driver-service HTTP

管理端 web
  |
  v
admin-service HTTP
```

服务内部也可以暴露 gRPC：

```
passenger-service
  |
  | gRPC
  v
driver-service

admin-service
  |
  | gRPC
  v
passenger-service / driver-service
```

但是核心边界仍然是“端”。

------

**请求流转**

以乘客创建订单为例：

```
乘客端 uni-app
  |
  | POST /passenger/v1/orders
  v
passenger-service HTTP Server
  |
  v
internal/service/order.go
  |
  v
internal/biz/order.go
  |
  v
internal/data/order.go
  |
  v
MySQL / Redis / MQ
```

司机接单则在另一个服务里：

```
司机端 uni-app
  |
  | POST /driver/v1/orders/{id}/accept
  v
driver-service HTTP Server
  |
  v
internal/service/order.go
  |
  v
internal/biz/order.go
  |
  v
internal/data/order.go
```

------

**该方案的问题**

A 方案迁移成本低，但会带来明显长期问题。

比如订单能力会被拆散：

```
passenger-service/internal/biz/order.go   乘客下单、取消订单
driver-service/internal/biz/order.go      司机接单、拒单
admin-service/internal/biz/order.go       管理端查看、介入、退款
```

这会导致：

```
订单状态规则分散
重复维护 Order 模型
订单表访问逻辑重复
接口语义不统一
跨端一致性难保证
后续加派单/支付会更复杂
```

例如一个订单状态流转：

```
待支付 -> 已支付 -> 待接单 -> 已接单 -> 行程中 -> 已完成 -> 已评价
```

如果乘客端、司机端、管理端各维护一部分，后面非常容易出现状态不一致。

------

**适合什么情况**

```
项目规模中小
业务还不稳定
主要目标是快速交付
团队按前端端口划分
暂时不做复杂微服务治理
```

它像是“多个垂直应用服务”。

------

**不太适合什么情况**

```
生产级微服务
长期演进
复杂订单状态
复杂派单逻辑
支付退款
跨端一致性要求强
服务独立扩缩容
团队按业务域负责
```

一句话对比：

```
A：乘客端、司机端、管理端各自一个后端，端决定服务边界。
B：订单、行程、司机、支付、派单各自一个后端，业务域决定服务边界。
```

# Kratos解答层级

## **1. `internal/conf/` 是干什么的**

`internal/conf/` 负责定义和生成“配置结构体”。

简单说：

> `configs/config.yaml` 负责放配置值，`internal/conf/` 负责定义这些配置长什么样。

典型结构：

```
internal/conf/
├── conf.proto
└── conf.pb.go
```

`conf.proto` 里会定义配置模型，例如：

```
message Bootstrap {
  Server server = 1;
  Data data = 2;
}

message Server {
  message HTTP {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }

  message GRPC {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }

  HTTP http = 1;
  GRPC grpc = 2;
}

message Data {
  message Database {
    string driver = 1;
    string source = 2;
  }

  Database database = 1;
}
```

然后生成 `conf.pb.go`，Go 代码里就可以拿到强类型配置：

```
func NewHTTPServer(c *conf.Server, svc *service.OrderService) *http.Server {
    srv := http.NewServer(
        http.Address(c.Http.Addr),
        http.Timeout(c.Http.Timeout.AsDuration()),
    )
    return srv
}
```

所以 `internal/conf/` 的职责是：

- 定义配置结构
- 生成 Go 配置类型
- 让 YAML 配置变成强类型对象
- 给 `server`、`data`、`registry`、`trace` 等模块提供配置
- 避免代码里到处散落字符串配置读取

它不负责读取配置文件本身，读取通常在 `cmd/order/main.go` 里完成。

大概关系是：

```
configs/config.yaml
        |
        v
main.go 读取配置
        |
        v
internal/conf/*.pb.go 强类型结构
        |
        v
server / data / biz 等模块使用配置
```

------

## **2. `cmd/order/wire.go` 和 `wire_gen.go` 是干什么的**

这两个文件和 Google Wire 有关。

Wire 是 Go 的“编译期依赖注入工具”。

Kratos 项目里对象很多：

```
App
HTTP Server
gRPC Server
OrderService
OrderUsecase
OrderRepo
Data
DB
Redis
Logger
Config
```

如果你全部手写初始化，会变成这样：

```
data := data.NewData(...)
repo := data.NewOrderRepo(data)
uc := biz.NewOrderUsecase(repo)
svc := service.NewOrderService(uc)
httpSrv := server.NewHTTPServer(conf, svc)
grpcSrv := server.NewGRPCServer(conf, svc)
app := kratos.New(...)
```

项目一大，这种初始化代码会非常长、非常容易乱。

Wire 的作用就是帮你生成这段初始化代码。

------

`wire.go` 是“依赖装配声明文件”。

它告诉 Wire：

> 我要一个 `*kratos.App`，你可以从这些 ProviderSet 里找构造函数，把对象组装出来。

类似：

```
//go:build wireinject

func initApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
    panic(wire.Build(
        server.ProviderSet,
        data.ProviderSet,
        biz.ProviderSet,
        service.ProviderSet,
        newApp,
    ))
}
```

这里的重点是 `wire.Build(...)`。

它不是业务代码，而是告诉 Wire：

```
server.ProviderSet  提供 HTTP/gRPC Server
data.ProviderSet    提供 Data、Repo、DB
biz.ProviderSet     提供 Usecase
service.ProviderSet 提供 Service
newApp              提供最终 App
```

`wire.go` 一般带有构建标签：

```
//go:build wireinject
```

所以它不会参与正常编译，只给 Wire 生成代码用。

------

`wire_gen.go` 是 Wire 自动生成的真实初始化代码。

它里面会把依赖按顺序组装好，例如：

```
func initApp(serverConf *conf.Server, dataConf *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
    dataData, cleanup, err := data.NewData(dataConf, logger)
    if err != nil {
        return nil, nil, err
    }

    orderRepo := data.NewOrderRepo(dataData, logger)
    orderUsecase := biz.NewOrderUsecase(orderRepo, logger)
    orderService := service.NewOrderService(orderUsecase, logger)

    httpServer := server.NewHTTPServer(serverConf, orderService, logger)
    grpcServer := server.NewGRPCServer(serverConf, orderService, logger)

    app := newApp(logger, httpServer, grpcServer)

    return app, func() {
        cleanup()
    }, nil
}
```

这份代码是编译时真正使用的。

所以职责区别是：

```
wire.go      你写的，声明依赖如何组装
wire_gen.go  Wire 生成的，真正执行依赖初始化
```

一句话总结：

> `wire.go` 是设计图，`wire_gen.go` 是施工结果。

日常开发里：

- 新增构造函数、ProviderSet 时，改 `wire.go`
- 执行 `wire` 或 `make generate`
- 自动刷新 `wire_gen.go`
- 不建议手动改 `wire_gen.go`

------

在 Kratos 里它们和项目启动的关系是：

```
main.go
  |
  | 调用 initApp(...)
  v
wire_gen.go
  |
  | 组装依赖
  v
data -> repo -> biz/usecase -> service -> server -> kratos.App
  |
  v
app.Run()
```

所以你可以这样记：

```
internal/conf/  负责“配置模型”
wire.go         负责“依赖装配规则”
wire_gen.go     负责“自动生成的初始化代码”
```

这三个东西都不是业务核心，但它们决定了 Kratos 项目能不能长期保持清晰。

# Kratos常用命令

可以把 Kratos 里的代码分成两类：

**一类是脚手架生成的**：项目刚创建时由 `kratos new` 生成。
**一类是协议/依赖注入生成的**：后续你改了 `.proto`、`wire.go` 后重新生成。

我结合官方文档里 Kratos Layout 和 CLI 的说明整理一下。Kratos 官方也明确说，`api` 目录里会维护 proto 文件以及由它们生成的 Go 文件，`internal/conf` 里的 `.pb.go` 通常由 `make config` 生成，Wire 相关代码由 `go generate ./...` / `wire` 生成。参考：[Usage](https://go-kratos.dev/docs/getting-started/usage/)、[Layout](https://go-kratos.dev/docs/intro/layout/)、[Configuration](https://go-kratos.dev/docs/component/config/)、[Dependency Injection](https://go-kratos.dev/docs/guide/wire/)。

## **1. 哪些文件通常是命令生成的**

常见生成文件如下：

```
api/order/v1/order.pb.go
api/order/v1/order_grpc.pb.go
api/order/v1/order_http.pb.go
api/order/v1/order_errors.pb.go        可选
api/order/v1/order.swagger.json        可选，旧版常见
openapi.yaml                           可选，新版 layout 常见

internal/conf/conf.pb.go

cmd/order/wire_gen.go

internal/service/order.go              可由命令生成骨架，但后续通常手写维护
```

重点看这几个。

**`api/\**/\*.pb.go`**

由 `.proto` 生成。

例如：

```
api/order/v1/order.proto
        |
        v
api/order/v1/order.pb.go
```

`order.pb.go` 里主要是：

- request struct
- response struct
- message 序列化逻辑
- proto 反射信息

例如：

```
type CreateOrderRequest struct {
    PassengerId string
    StartPoint  string
    EndPoint    string
}
```

这个文件不要手改，改了也会被下一次生成覆盖。

------

**`api/\**/\*_grpc.pb.go`**

由 proto 的 `service` 定义生成。

```
order.proto
        |
        v
order_grpc.pb.go
```

里面主要是：

- gRPC Client
- gRPC Server interface
- Register 方法

例如：

```
type OrderServer interface {
    CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderReply, error)
}

func RegisterOrderServer(s grpc.ServiceRegistrar, srv OrderServer)
```

服务之间 RPC 调用时会用到这个文件。

------

**`api/\**/\*_http.pb.go`**

当 `.proto` 里写了 HTTP 注解时生成。

例如：

```
rpc CreateOrder(CreateOrderRequest) returns (CreateOrderReply) {
  option (google.api.http) = {
    post: "/v1/orders"
    body: "*"
  };
}
```

会生成：

```
order_http.pb.go
```

里面主要是 HTTP 路由绑定逻辑：

```
func RegisterOrderHTTPServer(s *http.Server, srv OrderHTTPServer)
```

uni-app 通过 HTTP 调用时，最终就是这部分生成代码把 HTTP 请求转成 Go 方法调用。

------

**`api/\**/\*_errors.pb.go`**

如果你使用 Kratos errors 插件，错误码 proto 会生成类似文件。

例如：

```
api/order/v1/error_reason.proto
        |
        v
api/order/v1/error_reason.pb.go
api/order/v1/error_reason_errors.pb.go
```

它通常用于生成标准业务错误：

```
return v1.ErrorOrderNotFound("订单不存在")
```

------

**`openapi.yaml` / `\*.swagger.json`**

由 proto + HTTP 注解生成。

用途是给：

- Swagger
- Apifox
- YApi
- Postman
- 前端接口文档

Kratos 官方 OpenAPI 文档里也提到，使用 kratos-layout 时通常可以通过 `make api` 生成 OpenAPI 文件。

------

**`internal/conf/conf.pb.go`**

由：

```
internal/conf/conf.proto
```

生成。

命令通常是：

```
make config
```

职责是生成配置结构体，例如：

```
type Bootstrap struct {
    Server *Server
    Data   *Data
}
```

也不要手改。

------

**`cmd/order/wire_gen.go`**

由 Wire 生成。

来源是：

```
cmd/order/wire.go
```

生成命令通常是：

```
go generate ./...
```

或者进入目录后执行：

```
wire
```

`wire_gen.go` 是真正被编译使用的依赖初始化代码。不要手动改它。

------

**`internal/service/order.go`**

这个比较特殊。

它可以通过命令生成初始骨架：

```
kratos proto server api/order/v1/order.proto -t internal/service
```

会生成类似：

```
type OrderService struct {
    v1.UnimplementedOrderServer
}

func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderReply, error) {
    return &v1.CreateOrderReply{}, nil
}
```

但是这个文件生成之后，通常就进入手写维护阶段。因为你会在里面注入 `biz.OrderUsecase`，并调用业务逻辑。

所以它是：

```
第一次可以生成，后续主要手写
```

------

## **2. Kratos 常用命令**

安装 Kratos CLI：

```
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
```

注意：目前 Kratos 生态里有 v2/v3 文档差异，实际项目里要以你项目 `go.mod` 和 CLI 版本为准。看版本：

```
kratos -v
```

------

**创建项目**

```
kratos new order-service
```

如果国内 GitHub 拉模板慢，可以指定 Gitee 模板源：

```
kratos new order-service -r https://gitee.com/go-kratos/kratos-layout.git
```

创建多服务单仓项目时常见：

```
kratos new app/order --nomod
kratos new app/driver --nomod
kratos new app/payment --nomod
```

------

**新增 proto 文件**

```
kratos proto add api/order/v1/order.proto
```

生成一个 proto 模板文件。

------

**根据 proto 生成 client / pb / grpc / http 代码**

```
kratos proto client api/order/v1/order.proto
```

常见输出：

```
api/order/v1/order.pb.go
api/order/v1/order_grpc.pb.go
api/order/v1/order_http.pb.go
```

如果 proto 没有 HTTP 注解，就不会生成 `*_http.pb.go`。

------

**根据 proto 生成 service 骨架**

```
kratos proto server api/order/v1/order.proto -t internal/service
```

输出：

```
internal/service/order.go
```

这个文件后面通常需要你自己补业务调用。

------

**生成所有代码**

Kratos 项目里最常用的是：

```
go generate ./...
```

官方快速开始文档也把它用于生成 proto、wire 等代码。

在 kratos-layout 项目中，也经常使用：

```
make api
make config
make generate
make all
```

一般含义是：

```
make api       生成 api 相关代码：pb、grpc、http、openapi 等
make config    生成 internal/conf/*.pb.go
make generate  执行 go generate ./...，通常生成 wire_gen.go
make all       api + config + generate 的组合
```

具体以你项目里的 `Makefile` 为准。

------

**运行项目**

```
kratos run
```

也可以直接 Go 原生命令运行：

```
go run ./cmd/order -conf ./configs
```

或者旧/默认 layout 可能是：

```
go run ./cmd/server -conf ./configs
```

------

**升级工具**

```
kratos upgrade
```

它通常会升级：

- Kratos CLI
- protoc 相关生成插件

------

**查看帮助**

```
kratos -h
kratos new -h
kratos proto -h
kratos proto client -h
kratos proto server -h
```

------

一句话记忆：

```
kratos new             生成项目骨架
kratos proto add       生成 proto 模板
kratos proto client    根据 proto 生成 pb/grpc/http 代码
kratos proto server    生成 service 层接口骨架
make api               批量生成 API 相关代码
make config            生成配置结构体代码
go generate ./...      生成 wire_gen.go 等代码
kratos run             本地运行服务
```

实际开发里最常见的节奏是：

```
# 1. 改 proto
make api

# 2. 改 internal/conf/conf.proto
make config

# 3. 新增 Provider / 修改 wire.go
go generate ./...

# 4. 跑服务
kratos run
```

# Kratos 完整开发流程 + 全套命令

## 一、先装前置工具（只执行一次）

### 1. 安装 Kratos 脚手架

```powershell
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
# 升级插件
kratos upgrade
```

验证：`kratos -v` 输出版本即成功

### 2. 安装 wire 依赖注入工具

```powershell
go get github.com/google/wire/cmd/wire
```

## 二、第一步：新建微服务（比如 order-service 订单服务）

### 国内拉取模板（必用，否则超时）

```powershell
# 创建订单服务 官方 Gitee 国内源
kratos new order-service -r shturl.cc/KkK3yPRwuittZdo7O6mXG2GaFjfttyGPD9x
# 不加 -r
kratos new order-service 
# 默认会去 GitHub 拉官方模板：
https://github.com/go-kratos/kratos-layout
# 进入服务目录，后续所有命令都在这里执行
cd order-service
# 拉取所有go依赖
go mod tidy
```

## 三、第二步：定义接口 Proto（核心 Proto First 流程）

### 1. 新建 proto 文件

```powershell
# 创建订单接口proto模板
kratos proto add api/order/v1/order.proto
```

打开`api/order/v1/order.proto`，写请求、响应、HTTP 路由注解（uni-app 走 HTTP）

### 2. 根据 proto 生成 pb/grpc/http 代码

```powershell
# 方式1：单文件生成
kratos proto client api/order/v1/order.proto
# 方式2：项目批量生成（推荐，Makefile封装）
make api
```

生成文件：

- `order.pb.go`：结构体
- `order_grpc.pb.go`：内部微服务 gRPC 调用
- `order_http.pb.go`：前端 uni-app HTTP 路由绑定

### 3. 生成 service 层骨架（接口入口）

```powershell
kratos proto server api/order/v1/order.proto -t internal/service
```

自动生成`internal/service/order.go`，接收前端请求，调用 biz 业务层

## 四、第三步：修改配置文件后生成配置代码

修改`internal/conf/conf.proto`（新增 MySQL、Redis、Nacos 地址） 执行生成配置结构体：

```powershell
make config
```

生成`internal/conf/conf.pb.go`，代码中强类型读取 yaml 配置

## 五、第四步：分层写业务代码（固定顺序）

1. **data 层**：写 MySQL/Redis/MQ 操作，实现 biz 定义的 Repo 接口
2. **biz 层**：写核心业务（下单、派单、价格计算），定义 Repo 接口
3. **service 层**：调用 biz 的 Usecase，做参数转换
4. 修改`wire.go`，把 biz、data、service 加入 ProviderSet

## 六、第五步：生成 wire 依赖注入代码（每次改 wire.go 必执行）

```powershell
# 全局生成所有generate标记文件（含wire_gen.go）
go generate ./...
# 等价Makefile命令
make generate
```

自动生成`cmd/order/wire_gen.go`，自动组装 DB、Redis、service、http 服务

## 七、第六步：本地启动服务（3 种方式任选）

### 方式 1：kratos 官方快捷启动（最简单）

```powershell
kratos run
```

### 方式 2：原生 go run（指定配置文件）

```powershell
go run ./cmd/order -conf ./configs
```

### 方式 3：先编译二进制再运行

```powershell
# 编译
make build
# 运行二进制
./bin/order -conf ./configs
```

启动成功输出：HTTP 8000、gRPC 9000 监听端口

## 八、日常开发标准循环（90% 开发只用这套）

1. 改接口 proto → `make api`
2. 改数据库配置 conf.proto → `make config`
3. 新增 biz/data/service、修改 wire.go → `make generate`
4. 启动服务调试 → `kratos run`

### 一键全量生成（改了多处代码时）

```powershell
make all
# 等价：make api + make config + make generate
```

## 九、常用辅助命令

1. 格式化代码、代码检查

```powershell
make lint
```

1. 单元测试

```powershell
go test ./...
```

1. 清理依赖缓存

```powershell
go mod tidy
```

## 十、完整新建服务实操顺序（复制直接跑）

```powershell
# 1. 创建服务 科学上网
kratos new order-service
cd order-service
go mod tidy

# 2. 创建并编辑proto
kratos proto add api/order/v1/order.proto
# 手动写完proto后执行
make api
# 生成service骨架
kratos proto server api/order/v1/order.proto -t internal/service

# 3. 写biz、data、service业务代码，修改wire.go

# 4. 生成依赖注入
make generate

# 5. 启动调试
kratos run
```

## 关键命令记忆总结

| 命令                                  | 作用                         | 使用时机                          |
| ------------------------------------- | ---------------------------- | --------------------------------- |
| `kratos new`                          | 创建微服务项目               | 第一次新建服务                    |
| `kratos proto add`                    | 新建 proto 模板              | 新增接口                          |
| `make api`                            | 生成 http/grpc/pb 代码       | 修改 proto 后                     |
| `make config`                         | 生成配置结构体               | 修改 conf.proto                   |
| `make generate` / `go generate ./...` | 生成 wire 依赖注入           | 新增 biz/data/service、改 wire.go |
| `make all`                            | 一次性生成 api + 配置 + wire | 批量修改代码后                    |
| `kratos run`                          | 本地启动服务                 | 日常调试                          |
| `make build`                          | 编译二进制                   | 打包部署                          |