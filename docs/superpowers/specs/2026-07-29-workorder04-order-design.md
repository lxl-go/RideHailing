# Workorder04 Order Management Design

## Goal

Build workorder04 as a real, testable order management module covering order query/detail, ticket purchase records, refund application, automatic review, manual review, batch refund, status history, audit tracing, and admin UI delivery.

## Scope

- Admin platform is the primary delivery surface for workorder04.
- Passenger platform adds real refund application and refund status APIs for user-facing ticket refund flows.
- Driver platform adds read-only order list/detail support for driver order visibility.
- Dispatch center rules stay out of scope until workorder11.
- Finance reconciliation stays decoupled. Workorder04 returns `refundNo`, `orderNo`, and `refundAmount`; workorder03 can consume these records later.

## Architecture

Use a new isolated admin carpool order module under the existing GVA structure. The module owns `order_main`, `order_refund`, and `order_status_history` models. It follows the current `model -> request -> service -> api -> router` pattern used by shuttle and finance modules.

Passenger and driver services get lightweight `workorder04` modules, not invasive rewrites of the verified workorder01 order flow. This keeps existing順風車 APIs stable while providing the 04-specific refund and query endpoints required for acceptance.

## Data Model

### `order_main`

- `id`: uint64 snowflake-style primary key.
- `order_no`: unique business order number.
- `service_type`: `carpool` or `shuttle`.
- `passenger_id`, `passenger_name`, `passenger_phone`.
- `driver_id`, `driver_name`, `driver_phone`, `vehicle_no`.
- `route_name`, `depart_time`, `arrival_time`.
- `status`: `pending`, `paid`, `ongoing`, `completed`, `cancelled`, `refunding`, `refunded`.
- `pay_amount`, `refund_amount`, `cancel_fee`.
- `cancel_reason`, `version`, `created_at`, `updated_at`.

### `order_refund`

- `id`: uint64 primary key.
- `refund_no`: unique refund number.
- `order_no`, `service_type`, `passenger_id`.
- `refund_amount`, `cancel_fee`, `reason`.
- `review_type`: `auto` or `manual`.
- `status`: `pending`, `approved`, `rejected`, `refunding`, `refunded`.
- `idempotent_key`.
- `reviewer`, `review_remark`, `estimated_finish_at`, `created_at`, `updated_at`.

### `order_status_history`

- `id`: uint64 primary key.
- `order_no`, `from_status`, `to_status`.
- `operator`, `reason`, `created_at`.

## Status Machine

- Forward: `pending -> paid -> ongoing -> completed`.
- Cancel: `pending -> cancelled`, `paid -> cancelled -> refunded`.
- Exception: `ongoing -> cancelled`, `completed -> refunding -> refunded` only through manual review.
- Completed orders cannot use automatic refund. They may only enter manual dispute review.
- All status changes create an `order_status_history` record.
- State updates use `version` as optimistic lock input.

## Refund Rules

- Same `idempotent_key` returns the existing refund result.
- Departure more than 120 minutes away: full refund, no cancel fee.
- Departure 30 to 120 minutes away: 10% cancel fee.
- Departure less than 30 minutes away: manual review.
- Completed order: manual review only.
- Batch refund accepts selected order numbers, applies the same rule engine, returns per-order success/failure rows.

## API Design

### Admin

- `GET /carpool/order/list`
- `GET /carpool/order/:orderNo`
- `GET /carpool/order/:orderNo/history`
- `GET /carpool/order/refund/list`
- `POST /carpool/order/refund/apply`
- `POST /carpool/order/refund/review`
- `POST /carpool/order/refund/batch`
- `POST /carpool/order/export`

### Passenger Gateway

- `POST /carpool/orders/:id/refund`
- `GET /carpool/orders/:id/refund`

### Driver Gateway

- `GET /carpool/orders/mine`
- `GET /carpool/orders/:id`

## Validation

- Admin request structs use Gin binding and existing `utils.Verify` where page validation is needed.
- Refund apply requires `orderNo`, `reason`, and `idempotentKey`.
- Manual review requires `refundNo`, `decision`, and `reviewer`.
- Batch refund requires at least one order number and at most 100 order numbers.

## Frontend

Add `admin-platform/web/src/view/admin/workorder04/order.vue` following the current GVA + Element Plus page style:

- Filter by order number, service type, status, date range.
- List page with pagination.
- Detail drawer with passenger, driver, vehicle, route, amount, refund, and status history sections.
- Refund tab for refund records.
- Single refund dialog.
- Manual review dialog.
- Batch refund action using table selection.
- Export button.

Passenger and driver mobile pages can remain minimal API-ready surfaces because the current workorder04 acceptance is primarily admin-side.

## Testing

- Admin service tests cover summary query, status history, idempotent refund, auto refund fee calculation, completed-order manual review, and batch refund mixed results.
- Passenger service tests cover refund apply idempotency and refund status.
- Driver service tests cover order list/detail visibility.
- Verification commands:
  - `go test -c ./service/carpool -o C:\Users\李小龙\Desktop\RideHailing\.tmp\admin-carpool-order.test.exe`
  - `go build ./...`
  - passenger and driver service finance/shuttle/order tests as applicable.
  - three frontend `npm.cmd run build`.

## Notes

The workspace root is not currently recognized as a git repository, so design commit is skipped. The design file itself is saved for review and traceability.
