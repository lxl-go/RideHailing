# 第 2 批任务追溯：司机/乘客联系与自动位置上报

## 范围
- 司机端：位置页取消“立即上报”，接单后自动上报；新增联系乘客聊天/电话入口
- 乘客端：新增联系司机聊天/电话入口
- 网关：新增订单 WebSocket 聊天房间；订单详情补齐联系人字段

## 已完成
- `services/gateway-service/internal/server/order_chat.go`
- `services/gateway-service/internal/server/order_chat_test.go`
- `services/gateway-service/internal/server/http.go`
- `services/gateway-service/internal/server/mobile_ai_dispatch.go`
- `services/gateway-service/internal/server/order_response.go`
- `apps/driver-uni-app/uni-app/src/pages/locationReport/locationReport.vue`
- `apps/driver-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue`
- `apps/driver-uni-app/uni-app/src/pages/pendingOrders/pendingOrders.vue`
- `apps/driver-uni-app/uni-app/src/pages/orderChat/orderChat.vue`
- `apps/driver-uni-app/uni-app/src/utils/request.js`
- `apps/driver-uni-app/uni-app/src/pages.json`
- `apps/passenger-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue`
- `apps/passenger-uni-app/uni-app/src/pages/tracking/tracking.vue`
- `apps/passenger-uni-app/uni-app/src/pages/orderChat/orderChat.vue`
- `apps/passenger-uni-app/uni-app/src/utils/request.js`
- `apps/passenger-uni-app/uni-app/src/pages.json`

## 验证
- `go test ./internal/server -run TestOrderChatHub -count=1`
- `go test ./internal/server -run '^$' -count=1`
- `go test ./... -run '^$' -count=1`
- `UNI_INPUT_DIR=src npm run build:h5` in driver app
- `UNI_INPUT_DIR=src npm run build:h5` in passenger app

## 说明
- 全量 `go test ./internal/server -count=1` 仍会被仓库既有 `TestBatch7GatewayFinalSandboxConfig` 失败打断，属于本批次外的 sandbox 配置问题
- 聊天消息本批次只做实时 WebSocket 广播，不做历史消息落库
