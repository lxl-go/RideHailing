# Batch 04 Passenger Order Status Badges

## 1. Task Goal

Passenger order list should show Chinese status labels and color badges. Accepted orders must render green. Rejected/cancelled orders with a reject reason must render red.

## 2. Current Batch

Batch 4 only. Do not change driver-side, payment flow, income, map, publish-demand, review, or database behavior in this batch.

## 3. Allowed Modification Scope

- `apps/passenger-uni-app/uni-app/src/pages/orders/orders.vue`
- `apps/passenger-uni-app/uni-app/src/utils/orderStatus.mjs`
- `apps/passenger-uni-app/uni-app/tests/orderStatus.test.mjs`

## 4. Verification

- `node --test tests/orderStatus.test.mjs`
- `UNI_INPUT_DIR=src npm run build:h5` in `apps/passenger-uni-app/uni-app`

## 5. Result

- `accepted` now renders as `已接单` with a green badge
- cancelled orders with reject reasons now render as `已拒绝` with a red badge
- H5 build completed successfully
