# Auth Service and JWT Gateway Design

## Goal

Add a standard Kratos `auth-service` and gateway JWT middleware so passenger and driver identity can move from temporary `X-User-Id` headers to token-based identity without breaking existing uni-app flows.

## Scope

- Keep `admin-platform` unchanged.
- Add `services/auth-service` with Kratos-style layout, Nacos-first config, `config.yaml` fallback, Nacos registration, and non-reserved ports.
- Use database `01ride` for auth account persistence.
- Use JWT as the gateway-facing user credential.
- Keep temporary `X-User-Id` fallback during this phase for low-risk migration.
- Do not implement full refresh-token rotation, SMS provider integration, RBAC, or blacklist revocation in this phase.

## Service Boundary

`auth-service` owns app login identity only. It stores app accounts by `principal + role`, where role is `passenger` or `driver`. For development and early integration, login accepts a principal and optional verification code, creates the account if missing, and returns an access token.

Passenger and driver profile data remains in `passenger-service` and `driver-service`. The gateway uses the authenticated user id from JWT claims to call those profile services.

## API Contract

`auth-service` exposes:

- `POST /v1/auth/login`: login or auto-register by principal and role.
- `POST /v1/auth/verify`: validate an access token and return claims.

`gateway-service` exposes:

- `POST /carpool/auth/login`: uni-app compatible login route proxied to auth-service.
- Protected `/carpool/**` business routes: require `Authorization: Bearer <token>` when `auth.jwt.enabled` is true.

## JWT Claims

Tokens are HS256 signed with service config:

- `sub`: numeric user id as string.
- `user_id`: numeric user id as string.
- `role`: `passenger` or `driver`.
- `jti`: token id.
- `iss`: configured issuer.
- `exp`: configured expiration.

Gateway middleware writes the authenticated user id back into request context and `X-User-Id` header so current route handlers and downstream calls keep working.

## Compatibility

During this phase, gateway has `auth.jwt.compatible_header_enabled: true`. This allows existing local uni-app pages that still send `X-User-Id` to continue working while new login/token storage is adopted. The next phase can disable the fallback.

## Testing

- Unit-test token generation, parsing, expired token rejection, and malformed bearer handling in `pkg/authx`.
- Unit-test auth usecase login creates/reuses accounts and validates roles.
- Unit-test gateway middleware accepts JWT, rejects missing/invalid JWT, and honors temporary header compatibility when enabled.
- Run `go test ./...` for changed Go modules.
