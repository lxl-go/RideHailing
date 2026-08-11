import assert from 'node:assert/strict'
import fs from 'node:fs'
import vm from 'node:vm'

const source = fs.readFileSync(new URL('../src/utils/nav.js', import.meta.url), 'utf8')
  .replace(/export function navigate/, 'function navigate')
  .replace(/export function goDetail/, 'function goDetail')

const calls = []
const context = {
  uni: {
    switchTab(payload) {
      calls.push(['switchTab', payload])
    },
    navigateTo(payload) {
      calls.push(['navigateTo', payload])
    },
  },
}

vm.createContext(context)
vm.runInContext(`${source}; this.navigate = navigate;`, context)

context.navigate('/pages/pendingOrders/pendingOrders')

assert.equal(calls.length, 1)
assert.equal(calls[0][0], 'switchTab')
assert.equal(calls[0][1].url, '/pages/pendingOrders/pendingOrders')
