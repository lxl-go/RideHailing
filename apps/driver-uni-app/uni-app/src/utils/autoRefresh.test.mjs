import assert from 'node:assert/strict'
import test from 'node:test'

import { createAutoRefresh } from './autoRefresh.mjs'

test('createAutoRefresh starts immediately and schedules one interval', () => {
  const calls = []
  const intervals = []
  const refresher = createAutoRefresh(() => calls.push('refresh'), {
    intervalMs: 3000,
    setIntervalFn: (fn, interval) => {
      intervals.push({ fn, interval })
      return 'timer-1'
    },
    clearIntervalFn: () => {},
  })

  refresher.start()

  assert.deepEqual(calls, ['refresh'])
  assert.equal(intervals.length, 1)
  assert.equal(intervals[0].interval, 3000)
})

test('createAutoRefresh stop clears the active interval once', () => {
  const cleared = []
  const refresher = createAutoRefresh(() => {}, {
    setIntervalFn: () => 'timer-2',
    clearIntervalFn: (timer) => cleared.push(timer),
  })

  refresher.start()
  refresher.stop()
  refresher.stop()

  assert.deepEqual(cleared, ['timer-2'])
})

test('createAutoRefresh restarting replaces the previous interval', () => {
  const cleared = []
  let timerIndex = 0
  const refresher = createAutoRefresh(() => {}, {
    setIntervalFn: () => `timer-${++timerIndex}`,
    clearIntervalFn: (timer) => cleared.push(timer),
  })

  refresher.start()
  refresher.start()

  assert.deepEqual(cleared, ['timer-1'])
})

test('createAutoRefresh skips overlapping refresh calls', async () => {
  let release
  let calls = 0
  const intervals = []
  const refresher = createAutoRefresh(() => {
    calls += 1
    return new Promise((resolve) => {
      release = resolve
    })
  }, {
    setIntervalFn: (fn) => {
      intervals.push(fn)
      return 'timer-3'
    },
    clearIntervalFn: () => {},
  })

  refresher.start()
  intervals[0]()
  assert.equal(calls, 1)

  release()
  await Promise.resolve()
  await Promise.resolve()
  intervals[0]()

  assert.equal(calls, 2)
})
