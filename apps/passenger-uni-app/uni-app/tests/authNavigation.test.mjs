import assert from 'node:assert/strict'

import { buildAuthenticatedLoginRedirect } from '../src/utils/authNavigation.mjs'

const result = buildAuthenticatedLoginRedirect({
  readStorage: (key) => (key === 'passengerAccessToken' ? 'token-123' : ''),
  tokenKey: 'passengerAccessToken',
  homeUrl: '/pages/home/home',
})

assert.deepEqual(result, {
  shouldRedirect: true,
  url: '/pages/home/home',
})

const anonymous = buildAuthenticatedLoginRedirect({
  readStorage: () => '',
  tokenKey: 'passengerAccessToken',
  homeUrl: '/pages/home/home',
})

assert.equal(anonymous.shouldRedirect, false)
