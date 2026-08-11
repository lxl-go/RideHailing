# Identity Services Boundary Design

## Goal

Build the next Kratos migration slice for passenger and driver identity profile boundaries.

This phase creates independent `passenger-service` and `driver-service` domain services, then exposes frontend-compatible HTTP routes through `gateway-service`. It does not introduce the final login, JWT, refresh token, or permission model; those are reserved for a later `auth-service + JWT middleware` phase.

## Scope

In scope:

- Create `services/passenger-service` as the passenger identity profile service.
- Create `services/driver-service` as the driver identity, certification, and vehicle service.
- Keep `gateway-service` as the only HTTP entry for passenger and driver uni-apps.
- Continue using temporary `X-User-Id` identity propagation.
- Use Nacos for config center and service registry/discovery.
- Keep `configs/nacos.yaml` and `configs/config.yaml` in each new service.
- Use MySQL database `01ride` in fallback config.
- Avoid business service listener ports `8080` and `8848`.
- Leave `admin-platform` unchanged.

Out of scope:

- Full login/register flow.
- JWT, refresh tokens, RBAC, or service-to-service auth middleware.
- Replacing gin-vue-admin.
- Removing old `passenger-platform/service` and `driver-platform/service`.
- Payment, dispatch, finance, shuttle, message, or marketing services.

## Architecture Decision

Use two domain services, not a single mixed user service:

- `passenger-service` owns passenger profile data and passenger-facing preferences.
- `driver-service` owns driver profile data, driver certification state, and vehicle records.
- `gateway-service` aggregates both services and keeps frontend routes stable.

This keeps boundaries clear for later auth and operations work. `auth-service` can later issue tokens and map accounts to passenger or driver profile IDs without forcing profile data into the auth domain.

## Repository Shape

```text
services/
  gateway-service/
  passenger-service/
  driver-service/
  trip-service/
  order-service/
  review-service/
```

Each new service follows the current standardized Kratos layout:

```text
services/passenger-service/
  api/passenger/v1/passenger.proto
  cmd/passenger/
  configs/nacos.yaml
  configs/config.yaml
  internal/conf/
  internal/server/
  internal/service/
  internal/biz/
  internal/data/
  Makefile
  go.mod

services/driver-service/
  api/driver/v1/driver.proto
  cmd/driver/
  configs/nacos.yaml
  configs/config.yaml
  internal/conf/
  internal/server/
  internal/service/
  internal/biz/
  internal/data/
  Makefile
  go.mod
```

## Service Responsibilities

### passenger-service

Owns passenger profile state:

- Ensure a passenger profile exists for a given user ID.
- Get passenger profile.
- Update passenger profile.
- Store passenger preference fields that are useful to passenger-side flows.

It does not own orders, reviews, payments, login, or token validation.

### driver-service

Owns driver profile and qualification state:

- Ensure a driver profile exists for a given user ID.
- Get driver profile.
- Update driver profile.
- Submit or update certification application.
- Read certification application.
- Create or update vehicle records.
- List driver vehicle records.

It does not own trips, orders, reviews, payments, login, or dispatch.

### gateway-service

Owns frontend-compatible HTTP routes:

- Extract temporary identity from `X-User-Id`.
- Call `passenger-service` or `driver-service` through Nacos discovery.
- Return current uni-app response shape:

```json
{
  "code": 0,
  "data": {},
  "msg": "success"
}
```

## API Contracts

Proto remains the source of truth.

### passenger-service proto

Service: `passenger.v1.PassengerService`

- `EnsurePassenger(EnsurePassengerRequest) returns (PassengerProfileReply)`
- `GetPassenger(GetPassengerRequest) returns (PassengerProfileReply)`
- `UpdatePassenger(UpdatePassengerRequest) returns (PassengerProfileReply)`

Messages:

- `PassengerProfile`
  - `id int64`
  - `nickname string`
  - `phone string`
  - `avatar_url string`
  - `common_address string`
  - `payment_preference string`
  - `status int32`
  - `created_at string`
  - `updated_at string`

HTTP annotations:

- `POST /v1/passengers/ensure`
- `GET /v1/passengers/{id}`
- `PUT /v1/passengers/{id}`

### driver-service proto

Service: `driver.v1.DriverService`

- `EnsureDriver(EnsureDriverRequest) returns (DriverProfileReply)`
- `GetDriver(GetDriverRequest) returns (DriverProfileReply)`
- `UpdateDriver(UpdateDriverRequest) returns (DriverProfileReply)`
- `SubmitCertification(SubmitCertificationRequest) returns (CertificationReply)`
- `GetCertification(GetCertificationRequest) returns (CertificationReply)`
- `SaveVehicle(SaveVehicleRequest) returns (VehicleReply)`
- `ListVehicles(ListVehiclesRequest) returns (ListVehiclesReply)`

Messages:

- `DriverProfile`
  - `id int64`
  - `name string`
  - `phone string`
  - `avatar_url string`
  - `service_status int32`
  - `certification_status int32`
  - `created_at string`
  - `updated_at string`
- `DriverCertification`
  - `id int64`
  - `driver_id int64`
  - `real_name string`
  - `license_no string`
  - `vehicle_license_no string`
  - `vehicle_photo_url string`
  - `face_photo_url string`
  - `status int32`
  - `reject_reason string`
  - `created_at string`
  - `updated_at string`
- `DriverVehicle`
  - `id int64`
  - `driver_id int64`
  - `plate_no string`
  - `brand string`
  - `model string`
  - `color string`
  - `vehicle_type string`
  - `seats int32`
  - `status int32`
  - `created_at string`
  - `updated_at string`

HTTP annotations:

- `POST /v1/drivers/ensure`
- `GET /v1/drivers/{id}`
- `PUT /v1/drivers/{id}`
- `POST /v1/drivers/{id}/certification`
- `GET /v1/drivers/{id}/certification`
- `POST /v1/drivers/{id}/vehicles`
- `GET /v1/drivers/{id}/vehicles`

## Gateway Compatibility Routes

Add frontend routes in `gateway-service`:

- `GET /carpool/passengers/me`
- `PUT /carpool/passengers/me`
- `GET /carpool/drivers/me`
- `PUT /carpool/drivers/me`
- `POST /carpool/drivers/certification`
- `GET /carpool/drivers/certification`
- `POST /carpool/drivers/vehicles`
- `GET /carpool/drivers/vehicles`

Driver uni-app certification page will submit to:

```text
POST /carpool/drivers/certification
```

Vehicle form data will submit to:

```text
POST /carpool/drivers/vehicles
```

## Data Model

Use the shared `01ride` database for this migration phase.

### passenger_profile

- `id bigint primary key`
- `nickname varchar(64)`
- `phone varchar(32)`
- `avatar_url varchar(255)`
- `common_address varchar(255)`
- `payment_preference varchar(64)`
- `status tinyint`
- `created_at datetime`
- `updated_at datetime`

### driver_profile

- `id bigint primary key`
- `name varchar(64)`
- `phone varchar(32)`
- `avatar_url varchar(255)`
- `service_status tinyint`
- `certification_status tinyint`
- `created_at datetime`
- `updated_at datetime`

### driver_certification

- `id bigint primary key`
- `driver_id bigint unique`
- `real_name varchar(64)`
- `license_no varchar(64)`
- `vehicle_license_no varchar(64)`
- `vehicle_photo_url varchar(255)`
- `face_photo_url varchar(255)`
- `status tinyint`
- `reject_reason varchar(255)`
- `created_at datetime`
- `updated_at datetime`

### driver_vehicle

- `id bigint primary key`
- `driver_id bigint index`
- `plate_no varchar(32)`
- `brand varchar(64)`
- `model varchar(64)`
- `color varchar(32)`
- `vehicle_type varchar(32)`
- `seats int`
- `status tinyint`
- `created_at datetime`
- `updated_at datetime`

## Status Values

Keep simple integer states in this phase.

Passenger status:

- `1`: enabled
- `2`: disabled

Driver service status:

- `1`: offline
- `2`: online
- `3`: disabled

Certification status:

- `1`: draft
- `2`: pending
- `3`: approved
- `4`: rejected

Vehicle status:

- `1`: active
- `2`: inactive

## Config And Ports

New service ports:

- `passenger-service` HTTP `9020`, gRPC `9120`
- `driver-service` HTTP `9030`, gRPC `9130`

Nacos config file:

```text
configs/nacos.yaml
```

Fallback config file:

```text
configs/config.yaml
```

Nacos values:

- Namespace: `public`
- Data ID: `ride-car`
- Group: `DEFAULT_GROUP`
- Format: YAML

Gateway service discovery endpoints:

```yaml
clients:
  passenger:
    endpoint: discovery:///passenger-service
    http_base_url: http://127.0.0.1:9020
  driver:
    endpoint: discovery:///driver-service
    http_base_url: http://127.0.0.1:9030
```

HTTP fallback URLs are allowed only for local development. Discovery remains the target production path.

## Error Handling

Domain errors map to gRPC status codes:

- invalid request: `InvalidArgument`
- profile not found: `NotFound`
- duplicate or forbidden state transition: `FailedPrecondition`
- internal storage failure: `Internal`

Gateway returns transport errors directly for now, matching the existing gateway style. Later auth middleware should centralize response mapping and request IDs.

## Testing

Minimum tests for this phase:

- `passenger-service/internal/biz`: ensure creates default profile; update trims and persists profile fields.
- `passenger-service/internal/data`: GORM repo creates, reads, and updates `passenger_profile`.
- `driver-service/internal/biz`: certification submission sets pending status; vehicle save defaults active status and validates required fields.
- `driver-service/internal/data`: GORM repo creates/updates driver profile, certification, and vehicle records.
- `gateway-service/internal/data`: passenger and driver HTTP fallback clients forward expected paths and payloads.
- Full verification:
  - `go test ./...` in `passenger-service`
  - `go test ./...` in `driver-service`
  - `go test ./...` in `gateway-service`
  - `go test ./configx ./registry` in `pkg`
  - scan services for forbidden listeners `8080` and `8848`
  - scan configs for database `01ride`

## Acceptance Criteria

- `services/passenger-service` exists and follows the standard Kratos layout.
- `services/driver-service` exists and follows the standard Kratos layout.
- Both services load Nacos config first and use `config.yaml` as fallback.
- Both services use MySQL database `01ride` in fallback config.
- Neither service listens on `8080` or `8848`.
- `gateway-service` can call both services through Nacos discovery when available.
- `gateway-service` keeps HTTP fallback for local development.
- Driver uni-app certification flow has real gateway APIs to call.
- Passenger and driver profile routes use `X-User-Id` only as a temporary bridge.
- `auth-service + JWT middleware` remains a separate future phase.

## Notes

The workspace root is not a git repository, so design commit is skipped.
