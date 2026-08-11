import assert from 'node:assert/strict'
import test from 'node:test'

import { getOrderStatusText, getOrderStatusType } from './orderStatus.mjs'

test('pending order is reserved but unpaid', () => {
  assert.equal(getOrderStatusText('pending'), '已预约')
  assert.equal(getOrderStatusType('pending'), 'primary')
})

test('paid order is waiting for travel', () => {
  assert.equal(getOrderStatusText('paid'), '待出行')
  assert.equal(getOrderStatusType('paid'), 'warning')
})

test('cancelled order with reject reason is rejected', () => {
  assert.equal(getOrderStatusText('cancelled', { reject_reason: '临时有事' }), '已拒绝')
  assert.equal(getOrderStatusType('cancelled', { reject_reason: '临时有事' }), 'error')
})
