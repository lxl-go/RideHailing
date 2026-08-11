# Identity Services Boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `passenger-service` and `driver-service` as separate Kratos identity profile services, then expose them through `gateway-service`.

**Architecture:** Add two domain services under `services/`, following the existing `trip/order/review-service` layout. `gateway-service` keeps uni-app-compatible HTTP routes and calls the new services through Nacos discovery with local HTTP fallback.

**Tech Stack:** Go 1.26, Kratos v2, protobuf, gRPC, Kratos HTTP transport, GORM, MySQL `01ride`, Nacos config/registry, uni-app Vue.

## Global Constraints

- `admin-platform` remains unchanged.
- `auth-service + JWT middleware` is not part of this phase.
- Temporary identity continues to use `X-User-Id`.
- New backend services live under `services/`.
- Nacos is both configuration center and service registry/discovery.
- Each service has `configs/nacos.yaml` and complete fallback `configs/config.yaml`.
- MySQL database name is `01ride`.
- Business services must not listen on or register `8080` or `8848`.
- Generated files are not edited manually.
- No git commit is required.

---

## File Structure

Create:

- `services/passenger-service/...`: standard Kratos service for passenger profiles.
- `services/driver-service/...`: standard Kratos service for driver profiles, certification, and vehicles.

Modify:

- `go.work`: add both new service modules.
- `services/gateway-service`: add passenger/driver clients, usecases, services, config, and HTTP routes.
- `driver-platform/uni-app/src/api/workorder01.js`: add certification/vehicle API functions.
- `driver-platform/uni-app/src/pages/auth/auth.vue`: submit real gateway requests.

---

### Task 1: Passenger Service

**Files:**
- Create: `services/passenger-service/go.mod`
- Create: `services/passenger-service/Makefile`
- Create: `services/passenger-service/api/passenger/v1/passenger.proto`
- Create: `services/passenger-service/configs/nacos.yaml`
- Create: `services/passenger-service/configs/config.yaml`
- Create: `services/passenger-service/internal/conf/conf.go`
- Create: `services/passenger-service/internal/biz/*.go`
- Create: `services/passenger-service/internal/data/*.go`
- Create: `services/passenger-service/internal/service/*.go`
- Create: `services/passenger-service/internal/server/*.go`
- Create: `services/passenger-service/cmd/passenger/*.go`
- Modify: `go.work`

**Interfaces:**
- Produces proto service `passenger.v1.PassengerService`.
- Produces gateway-consumable gRPC and HTTP fallback APIs.

- [ ] Write failing biz tests for `EnsurePassenger` default creation and `UpdatePassenger` field trimming.
- [ ] Implement passenger biz domain, repo interface, and errors.
- [ ] Write failing data test for create/read/update of `passenger_profile`.
- [ ] Implement GORM repo and DB provider.
- [ ] Add proto, generate pb/grpc/http files.
- [ ] Add service adapter, server setup, config, and main entry.
- [ ] Run `go test ./...` in `services/passenger-service`.

### Task 2: Driver Service

**Files:**
- Create: `services/driver-service/go.mod`
- Create: `services/driver-service/Makefile`
- Create: `services/driver-service/api/driver/v1/driver.proto`
- Create: `services/driver-service/configs/nacos.yaml`
- Create: `services/driver-service/configs/config.yaml`
- Create: `services/driver-service/internal/conf/conf.go`
- Create: `services/driver-service/internal/biz/*.go`
- Create: `services/driver-service/internal/data/*.go`
- Create: `services/driver-service/internal/service/*.go`
- Create: `services/driver-service/internal/server/*.go`
- Create: `services/driver-service/cmd/driver/*.go`
- Modify: `go.work`

**Interfaces:**
- Produces proto service `driver.v1.DriverService`.
- Produces driver profile, certification, and vehicle APIs.

- [ ] Write failing biz tests for certification pending status and vehicle default active status.
- [ ] Implement driver biz domain, repo interface, and errors.
- [ ] Write failing data test for profile, certification, and vehicle persistence.
- [ ] Implement GORM repo and DB provider.
- [ ] Add proto, generate pb/grpc/http files.
- [ ] Add service adapter, server setup, config, and main entry.
- [ ] Run `go test ./...` in `services/driver-service`.

### Task 3: Gateway Aggregation

**Files:**
- Modify: `services/gateway-service/go.mod`
- Modify: `services/gateway-service/configs/config.yaml`
- Modify: `services/gateway-service/internal/conf/conf.go`
- Create: `services/gateway-service/internal/data/passenger_client.go`
- Create: `services/gateway-service/internal/data/driver_client.go`
- Create: `services/gateway-service/internal/data/passenger_client_test.go`
- Create: `services/gateway-service/internal/data/driver_client_test.go`
- Create: `services/gateway-service/internal/biz/passenger.go`
- Create: `services/gateway-service/internal/biz/driver.go`
- Create: `services/gateway-service/internal/service/passenger.go`
- Create: `services/gateway-service/internal/service/driver.go`
- Modify: `services/gateway-service/internal/server/http.go`
- Modify: `services/gateway-service/cmd/gateway/init.go`

**Interfaces:**
- Produces uni-app-compatible routes:
  - `GET /carpool/passengers/me`
  - `PUT /carpool/passengers/me`
  - `GET /carpool/drivers/me`
  - `PUT /carpool/drivers/me`
  - `POST /carpool/drivers/certification`
  - `GET /carpool/drivers/certification`
  - `POST /carpool/drivers/vehicles`
  - `GET /carpool/drivers/vehicles`

- [ ] Write gateway HTTP fallback client tests for passenger and driver paths.
- [ ] Implement passenger and driver gateway clients with discovery-first and HTTP fallback.
- [ ] Add gateway biz/service adapters.
- [ ] Register compatible HTTP routes using `X-User-Id`.
- [ ] Run `go test ./...` in `services/gateway-service`.

### Task 4: Driver Uni-App Certification Wiring

**Files:**
- Modify: `driver-platform/uni-app/src/api/workorder01.js`
- Modify: `driver-platform/uni-app/src/pages/auth/auth.vue`

**Interfaces:**
- Consumes gateway routes:
  - `POST /carpool/drivers/certification`
  - `POST /carpool/drivers/vehicles`

- [ ] Add API wrapper functions.
- [ ] Change certification page submit to call real gateway APIs.
- [ ] Keep local image selection as URL placeholder until upload service exists.

### Task 5: Verification

**Files:**
- Read only unless failures require fixes.

- [ ] Run `go test ./configx ./registry` in `pkg`.
- [ ] Run `go test ./...` in `services/passenger-service`.
- [ ] Run `go test ./...` in `services/driver-service`.
- [ ] Run `go test ./...` in `services/gateway-service`.
- [ ] Run existing migrated service tests for trip/order/review.
- [ ] Scan for forbidden service listeners `8080` and `8848`.
- [ ] Scan for `01ride` in service fallback configs.
- [ ] Scan uni-app source for `localhost:8081`, `localhost:8082`, and `X-Driver-Id`.
