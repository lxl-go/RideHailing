import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const apiSource = fs.readFileSync(path.resolve('src/api/trip.js'), 'utf8')
const homeSource = fs.readFileSync(path.resolve('src/pages/home/home.vue'), 'utf8')

assert.match(
  apiSource,
  /export\s+const\s+recommendTrips\s*=\s*\(params\)\s*=>[\s\S]*url:\s*['"]\/carpool\/trips\/demands\/recommendations['"][\s\S]*method:\s*['"]GET['"]/,
  'recommendTrips should call demand-based recommendation endpoint'
)

assert.match(
  homeSource,
  /import\s+\{\s*recommendTrips\s*\}\s+from\s+['"]@\/api\/trip['"]/,
  'home page should import recommendTrips'
)

assert.match(
  homeSource,
  /await\s+recommendTrips\(\{\s*page:\s*1,\s*page_size:\s*5\s*\}\)/,
  'home page should load demand-based recommendations'
)

assert.doesNotMatch(
  homeSource,
  /await\s+searchTrips\(\{\s*page:\s*1,\s*page_size:\s*5\s*\}\)/,
  'home page should not load generic trips for recommendations'
)

console.log('trip API recommendation contract tests passed')
