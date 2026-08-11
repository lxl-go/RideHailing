# Auth Session and Ihuyi SMS Design

## Goal

Upgrade app authentication from a single access token to a production-oriented session flow with real Ihuyi SMS verification, refresh tokens, logout, and JWT-only gateway protection.

## Scope

- Keep `admin-platform` unchanged.
- Keep `auth-service` as the owner of app login identity and token/session lifecycle.
- Add `pkg/smsx` as the shared Ihuyi SMS client package.
- Store local fallback SMS credentials in `services/auth-service/configs/config.yaml`; production should override through Nacos.
- Disable gateway temporary `X-User-Id` compatibility.
- Do not add RBAC, captcha, device fingerprinting, or admin login integration in this phase.

## SMS Boundary

`pkg/smsx` sends verification SMS through Ihuyi Submit.json:

- Endpoint: `https://api.ihuyi.com/sms/Submit.json`
- Request method: `POST`
- Content type: `application/x-www-form-urlencoded`
- Fields: `account`, `password`, `mobile`, `content`
- Success code: `2`

`auth-service` generates a six-digit login code, stores its hash in `auth_sms_code`, sends the plaintext code through `pkg/smsx`, and verifies the code during login.

## Session Boundary

`auth-service` stores refresh sessions in `auth_session`. Access tokens remain short-lived JWTs. Refresh tokens are opaque random strings returned only to the client; the database stores their SHA-256 hash.

Session statuses:

- `1`: active
- `2`: revoked

## API Contract

`auth-service` exposes:

- `POST /v1/auth/sms/send`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `POST /v1/auth/verify`

`gateway-service` exposes:

- `POST /carpool/auth/sms/send`
- `POST /carpool/auth/login`
- `POST /carpool/auth/refresh`
- `POST /carpool/auth/logout`

All other `/carpool/**` business routes require `Authorization: Bearer <access_token>`.

## Frontend Flow

Passenger and driver uni-app request utilities store:

- `access_token`
- `refresh_token`
- `expires_in`

Requests attach `Authorization: Bearer <access_token>`. On HTTP `401`, the request utility calls `/carpool/auth/refresh` once, stores the new tokens, and retries the original request.

## Security Notes

- Refresh tokens are never stored in plaintext on the backend.
- SMS verification codes are stored as hashes and expire after five minutes.
- Login code verification consumes the code, preventing reuse.
- The local YAML file contains development fallback credentials only; Nacos should own production secrets.
