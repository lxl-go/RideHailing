import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const source = fs.readFileSync(path.resolve('src/api/order.js'), 'utf8')

assert.match(
  source,
  /request\(\{\s*url:\s*['"]\/carpool\/orders['"],\s*method:\s*['"]POST['"],\s*data:\s*createOrderPayload\(data\),\s*header:\s*actionHeaders\(/s,
  'createOrder should send Idempotency-Key through actionHeaders'
)

assert.match(
  source,
  /function\s+actionHeaders\([^)]*data\s*=\s*\{\}[^)]*\)[\s\S]*['"]Idempotency-Key['"]/,
  'actionHeaders should build the Idempotency-Key header'
)

assert.match(
  source,
  /export\s+const\s+syncPayment\s*=\s*\(id,\s*data\s*=\s*\{\}\)\s*=>[\s\S]*\/carpool\/orders\/\$\{orderPathId\(id\)\}\/payment\/sync[\s\S]*method:\s*['"]POST['"]/,
  'syncPayment should post to the active payment sync endpoint'
)

console.log('order API contract tests passed')
