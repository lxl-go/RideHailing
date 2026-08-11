# RBAC and Gateway Role Isolation Design

## Goal

Add standard RBAC persistence to `auth-service`, standardize gateway authenticated user context, and enforce passenger/driver role isolation before business handlers call downstream services.

## Scope

- Keep `admin-platform` unchanged.
- Add five RBAC tables in `auth-service`:
  - `auth_user`
  - `auth_role`
  - `auth_user_role`
  - `auth_permission`
  - `auth_role_permission`
- Keep SMS login, refresh token, and logout behavior from the previous phase.
- Keep JWT as the gateway authentication carrier.
- Do not add admin RBAC screens or dynamic admin-managed permission editing in this phase.

## Auth Service RBAC

`auth_user` replaces the earlier `auth_account` naming at the persistence layer. The domain type can remain `Account` during low-risk migration, but its table is `auth_user`.

When a passenger or driver logs in:

- `auth_user` is created or reused by `principal + login_role`.
- `auth_role` ensures built-in roles `passenger` and `driver` exist.
- `auth_user_role` ensures the user is bound to the requested login role.
- JWT `role` claim remains the selected app role.

`auth_permission` and `auth_role_permission` are migrated now and can be seeded later with fine-grained permission codes. Gateway role isolation in this phase uses JWT role directly.

## Gateway Current User

The gateway auth filter parses JWT and writes a `CurrentUser` into request context:

- `UserID`
- `Role`
- `JTI`

Business handlers read `CurrentUserFromRequest(r)` instead of directly reading client-provided `X-User-Id`.

For downstream compatibility, the filter may still write internal `X-User-Id` after JWT validation, but clients cannot bypass auth with their own `X-User-Id`.

## Role Isolation

Passenger-only route groups require `role=passenger`:

- passenger profile routes
- passenger order creation, cancellation, listing, and detail

Driver-only route groups require `role=driver`:

- driver profile routes
- driver certification and vehicle routes
- trip publishing, driver trip list, trip status update
- pending orders, accept, reject

Shared authenticated routes:

- review submission remains authenticated only for now.
- trip search and trip detail remain authenticated in this phase because the gateway filter protects all non-auth `/carpool/**` routes.

## Frontend Login Entry

Passenger and driver uni-apps add a login page that calls:

- `send*LoginCode(mobile)`
- `login*(mobile, code)`

The login page stores session data through the existing user store and redirects back to the app home.
