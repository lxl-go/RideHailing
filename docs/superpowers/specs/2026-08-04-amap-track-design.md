# 高德地图与实时轨迹闭环设计

**Goal:** 让乘客端和司机端在接人、送人过程中都能真实显示地图、路线、当前位置和轨迹，不再依赖静态假图。

**Scope:** 仅覆盖高德地图代理、地理编码、路线规划、实时位置、轨迹回放和前端地图展示。不改订单金额、不改支付、不做性能治理。

## Current State

- 司机端已有 `POST /api/v1/driver/location/report` 和 `GET /api/v1/driver/track/replay`。
- 乘客端已有 `GET /api/v1/passenger/orders/{id}/track`，但页面仍是静态轨迹占位。
- 网关已有部分 AI 和订单路由，但没有统一的地图服务封装。
- 配置里已预留 `AMAP_WEB_KEY`，但没有对应的后端封装和统一前端地图数据接口。

## Root Problems

1. 地图能力分散，前端直接靠假数据渲染。
2. 订单和行程只有文本地址，没有统一的地理编码输出。
3. 轨迹接口能返回点，但前端没有把它画成地图。
4. 司机定位和乘客追踪没有统一的数据契约。
5. 高德 key 不能下发到客户端，必须由后端代理。

## Target Architecture

### Backend

- 新增 `pkg/amapx`，封装高德 Web 服务调用。
- 网关新增地图代理路由，前端只访问网关，不直接访问高德。
- 复用现有订单/轨迹接口，不新增独立地图服务。

### Frontend

- 乘客端 `tracking` 页改成真实地图展示：起点、终点、司机当前位置、轨迹线、ETA 文案。
- 司机端 `locationReport` 页改成真实地图展示：当前位置、上报状态、最近轨迹点。
- 司机端 `orderDetail` 页展示乘客上车点、目的地和当前司机位置。

## Backend Contracts

### `pkg/amapx`

提供以下能力：

- `Geocode(address, city)` -> 地址转坐标
- `Regeo(lat, lng)` -> 坐标转地址
- `DrivingRoute(origin, destination, strategy)` -> 路线和 ETA
- `Distance(origins, destination)` -> 距离计算
- `Weather(city)` -> 天气信息

输出统一结构，至少包含：

- `latitude`
- `longitude`
- `formattedAddress`
- `distanceMeters`
- `durationSeconds`
- `polyline`

### Gateway Routes

新增地图代理接口：

- `GET /api/v1/maps/geocode?address=&city=`
- `GET /api/v1/maps/regeo?lat=&lng=`
- `GET /api/v1/maps/route?origin=&destination=&strategy=`
- `GET /api/v1/maps/weather?city=`

已有轨迹接口继续保留：

- `GET /api/v1/passenger/orders/{id}/track`
- `POST /api/v1/driver/location/report`
- `GET /api/v1/driver/track/replay`

## Data Flow

### Passenger Tracking

1. 前端拿订单详情。
2. 前端请求轨迹数据。
3. 网关返回司机轨迹点、司机当前点、订单起终点文本和可选路线信息。
4. 前端把点渲染到地图上，形成移动轨迹和折线。

### Driver Location

1. 司机端获取当前位置。
2. 调用 `POST /api/v1/driver/location/report` 上报。
3. 网关转发到 driver-service 保存轨迹点。
4. 司机端地图显示当前点和最近状态。

### Route Preview

1. 前端传入起点/终点文本。
2. 网关调用 `pkg/amapx.Geocode` 获得坐标。
3. 网关调用 `pkg/amapx.DrivingRoute` 获得路线。
4. 前端只消费统一 DTO，不自己拼接高德参数。

## Files To Change

- Create: `pkg/amapx/amap.go`
- Create: `pkg/amapx/amap_test.go`
- Modify: `services/gateway-service/internal/conf/conf.go`
- Modify: `services/gateway-service/configs/config.yaml`
- Create: `services/gateway-service/internal/server/map_routes.go`
- Modify: `services/gateway-service/internal/server/http.go`
- Modify: `services/gateway-service/internal/server/mobile_ai_dispatch.go`
- Modify: `services/gateway-service/internal/server/mobile_ai_dispatch_test.go`
- Modify: `apps/passenger-uni-app/uni-app/src/pages/tracking/tracking.vue`
- Modify: `apps/passenger-uni-app/uni-app/src/api/order.js`
- Modify: `apps/passenger-uni-app/uni-app/src/utils/apiData.js`
- Modify: `apps/driver-uni-app/uni-app/src/pages/locationReport/locationReport.vue`
- Modify: `apps/driver-uni-app/uni-app/src/pages/orderDetail/orderDetail.vue`
- Modify: `apps/driver-uni-app/uni-app/src/api/location.js`
- Modify: `apps/driver-uni-app/uni-app/src/api/order.js`
- Modify: `apps/driver-uni-app/uni-app/src/utils/apiData.js`

## Error Handling

- 高德不可用时，网关返回明确错误，不伪造成功数据。
- 轨迹点为空时，前端展示“暂无轨迹”，不画假线。
- 司机未登录或无当前用户时，直接返回未授权。
- 地理编码失败时，保留地址文本，不阻塞订单详情打开。

## Testing

- `go test ./pkg/amapx`
- `go test ./services/gateway-service/internal/server -count=1`
- `go test ./services/gateway-service/... -count=1`
- `npm run build:h5` in passenger app
- `npm run build:h5` in driver app

## Acceptance Criteria

1. 乘客打开行程追踪页能看到真实地图、司机轨迹和最新位置。
2. 司机上报位置后，乘客侧能在追踪页看到更新后的轨迹点。
3. 司机订单详情页能看到接人和送达路线的真实地图表达。
4. 所有地图请求都经过后端代理，不把高德服务端 key 暴露给前端。
5. 无新增假成功接口，失败就明确失败。
