import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('../src/pages/tracking/tracking.vue', import.meta.url), 'utf8')

assert.match(source, /passenger-track-order-id/, '轨迹页必须支持从 tabBar 缓存读取订单 ID')
assert.match(source, /const\s+isDeliveryStage\s*=/, '轨迹页必须显式判断已接到乘客后的送达阶段')
assert.match(source, /const\s+routeOriginText\s*=/, '轨迹页必须封装当前路线起点')
assert.match(source, /const\s+routeDestinationText\s*=/, '轨迹页必须封装当前路线终点')
assert.match(source, /currentPoint\.value[\s\S]*order\.value\?\.origin/, '接乘客阶段路线起点必须优先使用司机当前位置，终点为乘客起点')
assert.match(source, /order\.value\?\.origin[\s\S]*order\.value\?\.destination/, '送达阶段路线必须从乘客起点切换到目的地')
assert.match(source, /routeSummaryText/, '轨迹页必须展示距离和预计时间')
assert.match(source, /formatDuration/, '预计时间必须格式化为中文分钟/小时')
assert.match(source, /routeStageText/, '轨迹页必须展示当前路线阶段文案')
assert.match(source, /const\s+driverId\s*=/, '轨迹页必须从订单详情保存司机 ID')
assert.match(source, /getOrderTrack\(orderId\.value,\s*\{\s*driverId:\s*driverId\.value\s*\}\)/, '轨迹页查询司机轨迹时必须把司机 ID 传给网关')
assert.match(source, /routeOriginText/, '轨迹页必须提供路线起点文案兜底')
assert.match(source, /formatPointForRoute\(currentPoint\.value\)\s*\|\|\s*order\.value\?\.origin/, '没有司机位置点时，接乘客阶段必须兜底展示订单起点到终点路线')
assert.match(source, /getRoutePreview\(\{[\s\S]*origin:\s*routeOriginText\.value[\s\S]*destination:\s*routeDestinationText\.value/, '高德路线请求必须使用已兜底的路线起终点')

assert.match(source, /const\s+normalizeMapPoint\s*=/, 'track page must normalize map points before passing them to H5 map')
assert.match(source, /Number\.isFinite\(latitude\)[\s\S]*latitude\s*>=\s*-90[\s\S]*longitude\s*<=\s*180/, 'track page must filter invalid coordinate values')
assert.match(source, /const\s+mapIncludePoints\s*=/, 'track page must use sanitized include-points to avoid AMap Bounds errors')
assert.match(source, /:include-points\s*=\s*['"]mapIncludePoints['"]/, 'map must bind sanitized include-points')

console.log('passenger track route stage contract ok')
