import assert from 'node:assert/strict'

import { normalizeVehicleForView } from '../src/utils/vehicleViewModel.mjs'

const largeSnowflakeID = '2084282472140513377'

const vehicle = normalizeVehicleForView({
  id: largeSnowflakeID,
  auditId: largeSnowflakeID,
  source: 'vehicle',
  plateNo: '京A12345',
  status: 1,
})

assert.equal(vehicle.id, largeSnowflakeID)
assert.equal(vehicle.auditId, largeSnowflakeID)
assert.equal(typeof vehicle.id, 'string')
assert.equal(vehicle.id, largeSnowflakeID, 'vehicle id must not be converted to a JS Number')
