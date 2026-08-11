import test from 'node:test'
import assert from 'node:assert/strict'
import { getOrderStatusText, getOrderStatusType } from '../src/utils/orderStatus.mjs'

test('reserved unpaid orders render blue booked label', () => {
  assert.equal(getOrderStatusText('pending'), '已预约')
  assert.equal(getOrderStatusType('pending'), 'primary')
})

test('paid orders render yellow pending trip label', () => {
  assert.equal(getOrderStatusText('paid'), '待出行')
  assert.equal(getOrderStatusType('paid'), 'warning')
})

test('accepted orders render Chinese text and success color', () => {
  assert.equal(getOrderStatusText('accepted'), '已接单')
  assert.equal(getOrderStatusType('accepted'), 'success')
})

test('cancelled orders with reject reason render red reject label', () => {
  assert.equal(getOrderStatusText('cancelled', { reject_reason: '司机拒单' }), '已拒绝')
  assert.equal(getOrderStatusType('cancelled', { reject_reason: '司机拒单' }), 'error')
})
