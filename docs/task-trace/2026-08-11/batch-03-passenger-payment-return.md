# Batch 03 Passenger Payment Return

## 1. Task Goal

Passenger payment keeps a single Alipay sandbox flow. After successful payment, the success page should provide clear return actions to the order list and home page.

## 2. Current Batch

Batch 3 only. Do not change driver-side, map, publish-demand, review, income, or database schema behavior in this batch.

## 3. Allowed Modification Scope

- `services/gateway-service/internal/server/payment_routes.go`
- `services/gateway-service/internal/server/payment_success_test.go`
- `apps/passenger-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue`

## 4. Verification

- `go test ./internal/server -run TestAlipaySuccessRouteReturnsSuccessPage -count=1`
- `UNI_INPUT_DIR=src npm run build:h5` in `apps/passenger-uni-app/uni-app`

## 5. Result

- Payment success page now returns to `/#/pages/orders/orders` or `/#/pages/home/home`
- Passenger payment confirm text now states the single Alipay sandbox flow and test buyer credentials
- H5 build completed successfully
