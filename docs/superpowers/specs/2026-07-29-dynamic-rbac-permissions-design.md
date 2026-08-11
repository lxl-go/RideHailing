# Dynamic RBAC Permission Check Design

## Goal

Upgrade gateway authorization from role-only checks to permission-code checks backed by the `auth-service` RBAC tables.

This phase makes `auth_permission` and `auth_role_permission` part of the real runtime authorization path while keeping the change low risk and compatible with the existing passenger and driver app flows.

## Scope

- Keep `admin-platform` unchanged.
- Keep passenger and driver uni-app API contracts unchanged.
- Keep JWT as the authentication carrier.
- Keep `gateway-service` as the external traffic entry.
- Keep `auth-service` as the owner of user, role, and permission data.
- Add built-in permission seed data for passenger and driver roles.
- Add an `auth-service` permission check API for gateway use.
- Replace gateway route `requireRole` checks with permission-code checks.
- Do not build RBAC management screens or admin APIs in this phase.
- Do not introduce tenant, department, data-scope, or field-level authorization in this phase.

## Permission Model

The existing five RBAC tables remain the authority:

- `auth_user`
- `auth_role`
- `auth_user_role`
- `auth_permission`
- `auth_role_permission`

`auth_permission.code` is the stable permission identifier used by code and configuration. Built-in codes use a domain-action style:

- `trip:publish`
- `trip:search`
- `trip:view_detail`
- `trip:list_driver_self`
- `trip:update_status_self`
- `order:create`
- `order:cancel_self`
- `order:list_passenger_self`
- `order:view_passenger_self`
- `order:list_driver_pending`
- `order:accept_driver_self`
- `order:reject_driver_self`
- `review:submit`
- `passenger:profile:view_self`
- `passenger:profile:update_self`
- `driver:profile:view_self`
- `driver:profile:update_self`
- `driver:certification:submit_self`
- `driver:certification:view_self`
- `driver:vehicle:manage_self`
- `driver:vehicle:list_self`

The passenger role receives trip search/detail, passenger profile, passenger order, and review permissions. The driver role receives trip search/detail, driver profile, certification, vehicle, trip management, driver order, and review permissions.

## Auth Service API

`auth-service` adds:

- `POST /v1/auth/permission/check`
- gRPC method `CheckPermission`

Request:

- `user_id`
- `permission_code`

Reply:

- `allowed`

The check returns `allowed=true` only when:

- the user exists and is active;
- the user has at least one active role;
- one active role is bound to the requested active permission.

Unknown users, inactive roles, missing permissions, or missing role-permission bindings return `allowed=false` without leaking internal lookup details.

## Gateway Enforcement

`gateway-service` keeps JWT parsing in the auth filter. Once authenticated, business handlers call `requirePermission` with a route-specific permission code.

Route mapping stays in gateway code for now because gateway already owns frontend-compatible `/carpool/**` routing. This keeps the first dynamic RBAC phase simple and avoids introducing remote route configuration before the service boundaries are stable.

Public routes remain unchanged:

- `/carpool/auth/sms/send`
- `/carpool/auth/login`
- `/carpool/auth/refresh`
- `/carpool/auth/logout`
- `OPTIONS` requests

Authenticated routes without a permission check should be treated as a defect. In this phase, review submission receives `review:submit` so it no longer relies on authentication-only access.

## Data Flow

1. Client calls a protected `/carpool/**` route with `Authorization: Bearer <access_token>`.
2. Gateway auth filter validates JWT and stores `CurrentUser` in request context.
3. Handler resolves the required permission code for that operation.
4. Gateway calls `auth-service.CheckPermission(user_id, permission_code)`.
5. If denied, gateway returns HTTP `403`.
6. If allowed, gateway forwards the business call to the downstream service and injects internal `X-User-Id` only after JWT validation.

## Seed Behavior

`auth-service` seeds built-in roles and permissions during startup or data initialization:

- ensure `passenger` and `driver` roles exist;
- ensure built-in permission rows exist;
- ensure role-permission bindings exist.

Seed operations must be idempotent, so repeated service starts do not duplicate rows.

## Error Handling

- Missing or invalid JWT remains HTTP `401`.
- Valid JWT with missing permission returns HTTP `403`.
- Auth-service lookup errors return an error to gateway. Gateway should fail closed and return HTTP `403` or a service error according to the existing error handling path.
- Empty permission codes in gateway should be considered programmer error and deny access.

## Testing

Auth-service tests cover:

- built-in permission seeding is idempotent;
- passenger role has passenger permissions and lacks driver permissions;
- driver role has driver permissions and lacks passenger permissions;
- inactive or missing permissions deny access.

Gateway tests cover:

- `requirePermission` returns `401` without current user;
- `requirePermission` returns `403` when auth-service denies;
- `requirePermission` allows when auth-service allows;
- representative passenger and driver route mappings use permission checks instead of role checks.

Verification commands:

- `go test ./... -count=1` in `services/auth-service`
- `go test ./... -count=1` in `services/gateway-service`
- targeted scan confirming no remaining gateway business `requireRole` checks

## Out of Scope Later Work

Future phases can add:

- admin-managed RBAC APIs and screens;
- permission cache with explicit TTL and invalidation strategy;
- resource ownership checks shared across services;
- richer policy models for data scope and audit-sensitive operations.
