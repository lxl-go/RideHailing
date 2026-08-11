# Batch 01 Driver Income Reject Refund

## 1. Task Goal

Fix the closed loop for driver accept income, reject refund reason, passenger visibility, and admin driver-income summary.

## 2. Current Batch

Batch 1 only. Do not change map, address autocomplete, publish trip/request, payment page layout, or location auto-reporting flows in this batch.

## 3. Allowed Modification Scope

- order-service order domain, repo, data transaction, service adapter, proto, generated pb files
- gateway order client/usecase/service/server routing and order DTO response
- driver uni-app pending orders, home stats, income ledger, and related order/driver API wrappers
- passenger uni-app order normalization and order detail reject/refund display
- admin finance model/service/test/API/page files

## 4. Forbidden Modification Scope

Out-of-batch modules and behavior are only registered as follow-up issues.

## 5. Planned Files

| File | Reason | Solves | Status |
| --- | --- | --- | --- |
| services/order-service/internal/biz/order.go | Add accept time, reject reason, refund, and income query contract | Driver income and reject refund domain rules | still open |
| services/order-service/internal/data/order.go | Add transaction row lock and atomic refund/seat restore | State sync and idempotent concurrency safety | still open |
| services/order-service/api/order/v1/order.proto | Expose reject reason, refund fields, and driver income endpoint | Gateway/frontend contract | still open |
| services/gateway-service/internal/server/*.go | Forward reject reason and return real income stats | Mobile API closure | still open |
| apps/driver-uni-app/uni-app/src/pages/incomeLedger/incomeLedger.vue | Use backend income instead of completed-order local calculation | Today income reset by date | still open |
| apps/passenger-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue | Show driver reject reason/refund amount | Passenger receives reject reason | still open |
| admin-platform/server/service/carpool/finance.go | Add day/week/month/year driver income summary | Admin can see aggregate income | still open |

## 6. Solution

Use order accepted time as the income accounting time. Rejecting a paid order requires a non-empty reason, moves order to cancelled, restores trip seats, marks the payment refunded, and stores reject/refund fields in the same transaction.

## 7. Frontend Backend Chain

Driver pending-order reject sends reject_reason to gateway. Gateway passes it to order-service. Order-service writes order/payment/trip changes atomically. Passenger order detail receives reject/refund fields through gateway order DTO normalization. Driver income page and home stats use the driver income endpoint.

## 8. Database And Config Impact

Adds nullable accounting/reject/refund fields to carpool_order model usage. Payment refund is represented by carpool_payment.status = refunded numeric status in order-service.

## 9. Local Development Relaxations

None.

## 10. Release Corrections

Confirm production migrations contain the new carpool_order fields before release.

## 11. Self-Test Commands And Results
| Command | Result |
| --- | --- |
| `go test ./services/order-service/... -count=1` | PASS |
| `go test ./services/gateway-service/internal/server -run '^$' -count=1` | PASS, compile only |
| `go test ./admin-platform/server/service/carpool -run TestFinanceServiceUsesRealCarpoolPaymentsAndOrderRefunds -count=1` | PASS |
| `go test ./admin-platform/server/service/carpool -run 'TestOrderServiceListsRealCarpoolOrdersWithStringIDs|TestOrderServiceRefundRulesHistoryAndBatch|TestOrderServiceOverviewAggregatesOrders|TestFinanceServiceUsesRealCarpoolPaymentsAndOrderRefunds' -count=1` | PASS |
| `npm run build` in `admin-platform/web` | PASS |
| `$env:UNI_INPUT_DIR='src'; npm run build:h5` in `apps/driver-uni-app/uni-app` | PASS |
| `$env:UNI_INPUT_DIR='src'; npm run build:h5` in `apps/passenger-uni-app/uni-app` | PASS |

## 12. Uncovered Risks
- Gateway package still has an unrelated existing test failure: `TestBatch7GatewayFinalSandboxConfig` expects `notify_url` in sandbox config.
- `admin-platform/server/service/carpool` still has unrelated existing test failures in `person_test.go` because the test DB does not create `auth_user`.
- Frontend build emits existing Sass deprecation warnings and large-chunk warnings; they do not block this batch.

## 13. Out-Of-Scope Issues
- Gateway sandbox config test failure: `gateway config missing notify_url`.
- Admin person-service tests: `no such table: auth_user`.
