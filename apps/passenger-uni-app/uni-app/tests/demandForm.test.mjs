import assert from 'node:assert/strict'

import {
  buildDemandPayload,
  buildDateOptions,
  buildTimeOptions,
  estimateDemandBudget,
  formatDateLabel,
  formatTimeLabel,
} from '../src/utils/demandForm.mjs'

const dates = buildDateOptions(new Date('2026-08-11T08:00:00+08:00'), 3)
assert.deepEqual(dates.map((item) => item.value), ['2026-08-11', '2026-08-12', '2026-08-13'])
assert.equal(formatDateLabel(dates[0]), '今天 08月11日')
assert.equal(formatDateLabel(dates[1]), '明天 08月12日')

const times = buildTimeOptions(30)
assert.equal(times[0].value, '00:00')
assert.equal(times[1].value, '00:30')
assert.equal(formatTimeLabel(times[18]), '09:00')

assert.equal(estimateDemandBudget({ distanceMeters: 0, durationSeconds: 0, seats: 1 }), '')
assert.equal(estimateDemandBudget({ distanceMeters: 12500, durationSeconds: 1800, seats: 2 }), '88.00')

const payload = buildDemandPayload({
  origin: { name: '北京站', latitude: 39.9042, longitude: 116.4074 },
  destination: { name: '北京南站', latitude: 39.865, longitude: 116.378 },
  departDate: '2026-08-12',
  departTime: '09:30',
  seats: 2,
  estimatedBudget: '88.00',
  remark: '靠近地铁口',
})

assert.deepEqual(payload, {
  origin: '北京站',
  destination: '北京南站',
  depart_time: '2026-08-12T09:30:00+08:00',
  seats: 2,
  budget: 88,
  remark: '靠近地铁口',
})
