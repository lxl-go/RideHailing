import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const manifest = JSON.parse(readFileSync(new URL('../src/manifest.json', import.meta.url), 'utf8'))
const amapKey = '22ba26c4d757d904aef8138acda60ab7'

assert.equal(manifest.aMapKey, amapKey, '乘客端业务高德 key 必须保留')
assert.equal(manifest.h5?.sdkConfigs?.maps?.amap?.key, amapKey, 'H5 map 组件必须配置 h5.sdkConfigs.maps.amap.key')

console.log('passenger track map key contract ok')
