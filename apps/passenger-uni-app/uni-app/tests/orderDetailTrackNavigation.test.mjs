import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('../src/pages/orderDetail/orderDetail.vue', import.meta.url), 'utf8')

assert.match(source, /onLoad\(\(options\)\s*=>\s*{[\s\S]*options\.orderId[\s\S]*options\.id[\s\S]*}\)/, '订单详情页必须同时兼容 orderId 和 id 参数')
assert.match(source, /uni\.setStorageSync\('passenger-track-order-id',\s*String\(orderId\.value\)\)/, '跳转 tabBar 轨迹页前必须缓存订单 ID')
assert.match(source, /uni\.switchTab\(\{\s*url:\s*'\/pages\/tracking\/tracking'\s*}\)/, '查看轨迹必须使用 switchTab 打开 tabBar 页面')

console.log('orderDetail track navigation contract ok')
