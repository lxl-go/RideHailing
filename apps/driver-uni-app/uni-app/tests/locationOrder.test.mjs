import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const locationSource = fs.readFileSync(path.resolve('src/pages/locationReport/locationReport.vue'), 'utf8')
const orderApiSource = fs.readFileSync(path.resolve('src/api/order.js'), 'utf8')

assert.match(
  orderApiSource,
  /export\s+const\s+listDriverOrders\s*=\s*\(params\s*=\s*\{\}\)\s*=>[\s\S]*url:\s*['"]\/api\/v1\/driver\/orders['"][\s\S]*method:\s*['"]GET['"]/,
  'driver order API should expose /api/v1/driver/orders'
)

assert.match(
  locationSource,
  /import\s+\{\s*getDriverOrderDetail,\s*listDriverOrders\s*\}\s+from\s+['"]@\/api\/order['"]/,
  'location page should import listDriverOrders for active order fallback'
)

assert.match(
  locationSource,
  /await\s+listDriverOrders\(\{\s*status:\s*['"]accepted['"],\s*page:\s*1,\s*page_size:\s*1\s*\}\)/,
  'location page should query accepted driver orders when no cached active order exists'
)

assert.match(
  locationSource,
  /:polyline\s*=\s*['"]routePolyline['"]/,
  'map should render route polyline'
)

assert.match(
  locationSource,
  /routePolyline\.value\s*=\s*buildRoutePolyline\(/,
  'route preview should update map polyline'
)

assert.match(
  locationSource,
  /title:\s*targetAddress\(\)\s*\|\|\s*['"][\s\S]*iconPath:\s*['"]\/static\/tab\/orders\.png['"]/,
  'destination marker must include iconPath on H5 map'
)

assert.match(
  locationSource,
  /Number\.isFinite\(latitude\)[\s\S]*latitude\s*>=\s*-90[\s\S]*longitude\s*<=\s*180/,
  'location page must validate map coordinates before binding markers and route'
)

assert.match(
  locationSource,
  /latitude:\s*latitude\.value[\s\S]*longitude:\s*longitude\.value[\s\S]*lat:\s*latitude\.value[\s\S]*lng:\s*longitude\.value/,
  'location report payload should send both canonical and compatibility coordinate fields'
)

console.log('driver location order sync contract tests passed')
