# Kratos Standardization Design

## Goal

Refactor the passenger and driver backend capabilities into production-oriented Kratos microservices. The admin platform stays on gin-vue-admin and is not part of this migration.

The target architecture is domain-service based, not client-end based. Passenger and driver applications become clients of a gateway; backend services are split by business capability.

## Scope

In scope:

- Migrate passenger and driver backend capabilities toward standard Kratos services.
- Keep `admin-platform` unchanged.
- Introduce `services/` as the target backend service root.
- Use Nacos as both configuration center and service registry/discovery.
- Use `configs/nacos.yaml` as the Nacos connection file.
- Use `configs/config.yaml` as the complete local fallback file.
- Use MySQL database name `01ride`.
- Avoid business service ports `8080` and `8848`.
- Replace hard-coded gRPC targets with service discovery.

Out of scope for the first migration:

- Rewriting gin-vue-admin.
- Rebuilding all future services such as payment, dispatch, and message in the first batch.
- Full Kubernetes deployment manifests, except for leaving the structure ready for them.

## Target Repository Shape

```text
RideHailing/
  admin-platform/              # unchanged gin-vue-admin

  apps/
    passenger-uni-app/
    driver-uni-app/

  services/
    gateway-service/
    passenger-service/
    driver-service/
    trip-service/
    order-service/
    review-service/
    payment-service/           # later
    dispatch-service/          # later
    message-service/           # later

  pkg/
    configx/
    registry/
    logger/
    middleware/
    errors/
    tracing/
    validator/

  deployments/
    docker-compose/
    k8s/

  docs/
```

During migration, the existing `passenger-platform/service` and `driver-platform/service` may temporarily remain, but the long-term backend source of truth is `services/`.

## Service Boundary

Backend services are split by business domain:

- `gateway-service`: frontend HTTP entry and BFF aggregation for passenger and driver apps.
- `passenger-service`: passenger profile and passenger-side account data.
- `driver-service`: driver profile, certification, vehicle, and driver-side account data.
- `trip-service`: carpool trips, shuttle trips, route/trip publishing, trip search, seat availability.
- `order-service`: order creation, cancellation, acceptance/rejection, status flow, order query.
- `review-service`: passenger and driver reviews.
- `payment-service`: payment and refund, introduced later.
- `dispatch-service`: matching, dispatching, and location-driven assignment, introduced later.

The passenger and driver uni-apps do not own backend service boundaries. They call `gateway-service`, and `gateway-service` calls domain services.

## Standard Kratos Service Layout

Each service follows the same Kratos layout. Example: `order-service`.

```text
services/order-service/
  api/order/v1/
    order.proto
    error_reason.proto
    order.pb.go
    order_grpc.pb.go
    order_http.pb.go
    error_reason.pb.go
    error_reason_errors.pb.go

  cmd/order/
    main.go
    wire.go
    wire_gen.go

  configs/
    nacos.yaml
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
      clients.go

  Makefile
  go.mod
```

Layer responsibilities:

- `api`: proto contract for HTTP and gRPC.
- `cmd`: process entry, config loading, logger setup, registry setup, Wire initialization.
- `internal/conf`: strong typed config generated from `conf.proto`.
- `internal/server`: Kratos HTTP and gRPC servers.
- `internal/service`: proto service implementation and request/response mapping.
- `internal/biz`: business rules, use cases, domain models, repo interfaces.
- `internal/data`: database, Redis, MQ, external RPC clients, repo implementations.

## API Contract

Proto is the source of truth.

Each public service method that must be called through HTTP gets `google.api.http` annotations. Kratos-generated HTTP code replaces handwritten Gin gateway routes for migrated APIs.

Example:

```proto
service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderReply) {
    option (google.api.http) = {
      post: "/v1/orders"
      body: "*"
    };
  }
}
```

Generated files are not edited manually:

- `*.pb.go`
- `*_grpc.pb.go`
- `*_http.pb.go`
- `*_errors.pb.go`
- `internal/conf/*.pb.go`
- `cmd/*/wire_gen.go`

## Gateway Design

`gateway-service` is also a Kratos service. It exposes HTTP APIs to passenger and driver apps and calls domain services through gRPC clients discovered by Nacos.

```text
passenger uni-app
  -> gateway-service HTTP
  -> trip-service/order-service/review-service/passenger-service gRPC

driver uni-app
  -> gateway-service HTTP
  -> driver-service/trip-service/order-service/review-service gRPC
```

The gateway should not directly own core domain tables. Its `internal/data` layer mainly provides service clients.

## Nacos Configuration And Discovery

Nacos is used for both:

- configuration center
- service registry and discovery

Nacos console values:

```text
Namespace: public
Data ID: ride-car
Group: DEFAULT_GROUP
Format: YAML
```

Each service keeps:

```text
configs/nacos.yaml
configs/config.yaml
```

`configs/nacos.yaml` contains only Nacos connection information:

```yaml
nacos:
  enabled: true
  server_addr: 127.0.0.1
  server_port: 8848
  namespace_id: public
  group: DEFAULT_GROUP
  data_id: ride-car
  username: nacos
  password: nacos
  timeout_ms: 5000
  fail_fast: false
```

`configs/config.yaml` contains complete local fallback configuration. It is used when:

- `nacos.enabled = false`
- or Nacos loading fails and `fail_fast = false`

Production should normally use `fail_fast = true`. Development and local demos may use `fail_fast = false`.

Loading order:

1. Read `configs/nacos.yaml`.
2. If Nacos is enabled, load YAML config from Nacos `ride-car`.
3. If Nacos succeeds, use remote config.
4. If Nacos fails and `fail_fast = true`, fail startup.
5. If Nacos fails and `fail_fast = false`, load `configs/config.yaml`.
6. If Nacos is disabled, load `configs/config.yaml`.

## Database

The shared MySQL database name is:

```text
01ride
```

Local fallback DSN example:

```yaml
data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/01ride?charset=utf8mb4&parseTime=True&loc=Local
```

Domain services may share the same database in the first phase to reduce migration risk. Later, service-owned schemas or databases can be introduced if isolation is required.

## Ports

Business services must not listen on or register:

```text
8080
8848
```

`8848` is reserved for Nacos. `8080` is avoided to prevent conflicts with common defaults.

Recommended local ports:

```text
gateway-service HTTP:     9000
passenger-service HTTP:   9020
passenger-service gRPC:   9120
driver-service HTTP:      9030
driver-service gRPC:      9130
trip-service HTTP:        9040
trip-service gRPC:        9140
order-service HTTP:       9050
order-service gRPC:       9150
review-service HTTP:      9060
review-service gRPC:      9160
payment-service HTTP:     9070
payment-service gRPC:     9170
dispatch-service HTTP:    9080
dispatch-service gRPC:    9180
```

Service-to-service calls use Nacos discovery names, not fixed local ports:

```text
discovery:///trip-service
discovery:///order-service
discovery:///review-service
discovery:///driver-service
discovery:///passenger-service
```

## Migration Strategy

Use an incremental migration into the new standard structure.

Phase 1:

- Create shared `pkg/configx` and `pkg/registry` support for Nacos config and discovery.
- Create Kratos service skeletons for:
  - `gateway-service`
  - `trip-service`
  - `order-service`
  - `review-service`
  - `passenger-service`
  - `driver-service`
- Add `nacos.yaml` and fallback `config.yaml` to each service.
- Move or adapt existing proto contracts into service-specific `api/*/v1`.
- Generate gRPC and HTTP code from proto.

Phase 2:

- Migrate trip, order, and review business logic from existing passenger/driver services into domain services.
- Fix current passenger-side broken entries by replacing them with domain services, not by reviving old `internal/order` packages.
- Change passenger and driver apps to call `gateway-service`.
- Remove hard-coded `127.0.0.1:port` gRPC targets.

Phase 3:

- Add production-grade middleware: request ID, recovery, logging, tracing, auth, validation, metrics.
- Add service health checks and graceful shutdown.
- Add integration tests around gateway-to-service flows.
- Mark old `passenger-platform/service` and `driver-platform/service` as deprecated, then remove after migration is verified.

## Compatibility

Admin stays unchanged:

```text
admin-platform/server
admin-platform/web
```

If admin later needs data from the microservices, the gin-vue-admin backend can call the Kratos services through HTTP or gRPC, but this is not part of the first migration.

Passenger and driver uni-apps should use configurable gateway base URLs. They should not hard-code `localhost` for production. The driver app and passenger app should use a consistent identity propagation strategy, preferably authorization tokens, with temporary local fallback only for development.

## Testing

Minimum verification for each migrated service:

- `go test ./...`
- proto generation succeeds
- Wire generation succeeds
- service starts with Nacos config
- service starts with fallback `config.yaml` when allowed
- service registers to Nacos
- gateway can call domain services through discovery
- HTTP endpoints match frontend expectations

Migration acceptance criteria:

- Admin platform still runs unchanged.
- Passenger app can search trips, create/cancel/list orders, and submit reviews through `gateway-service`.
- Driver app can publish trips, manage trips, accept/reject orders, and submit reviews through `gateway-service`.
- No migrated business service listens on or registers `8080` or `8848`.
- MySQL DSN uses database `01ride`.
- No gateway or service client uses hard-coded `127.0.0.1:<grpc-port>` for service calls after migration.
