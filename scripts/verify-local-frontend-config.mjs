import fs from 'node:fs'
import path from 'node:path'

const root = path.resolve(import.meta.dirname, '..')

const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8')
const readJson = (relativePath) => JSON.parse(read(relativePath))
const readIfExists = (relativePath) => {
  const fullPath = path.join(root, relativePath)
  return fs.existsSync(fullPath) ? fs.readFileSync(fullPath, 'utf8') : ''
}

const failures = []

function check(condition, message) {
  if (!condition) failures.push(message)
}

const adminEnv = read('admin-platform/web/.env.development')
check(
  /^\s*VITE_CLI_PORT\s*=\s*8081\s*$/m.test(adminEnv),
  'admin-platform/web/.env.development should use VITE_CLI_PORT = 8081 to avoid local Nacos on 8080'
)

for (const appName of ['driver-uni-app', 'passenger-uni-app']) {
  const base = `apps/${appName}/uni-app`
  const pkg = readJson(`${base}/package.json`)
  const main = read(`${base}/src/main.js`)
  const uniScss = read(`${base}/src/uni.scss`)
  const loginPage = read(`${base}/src/pages/login/login.vue`)
  const childLockPath = path.join(root, `${base}/package-lock.json`)

  check(
    pkg.dependencies?.['uview-plus'],
    `${appName} should depend on uview-plus for Vue 3 compatibility`
  )
  check(
    !pkg.dependencies?.['uview-ui'],
    `${appName} should not depend on uview-ui because it calls Vue.filter under Vue 3`
  )
  check(
    !main.includes("from 'uview-ui'") && !main.includes('from "uview-ui"'),
    `${appName} main.js should not import uview-ui`
  )
  check(
    !uniScss.includes('uview-ui/theme.scss'),
    `${appName} uni.scss should not import uview-ui theme.scss`
  )
  check(
    !fs.existsSync(childLockPath),
    `${appName} should use the monorepo root package-lock.json instead of a stale nested package-lock.json`
  )
  check(
    /\.login-page\s*{[^}]*box-sizing:\s*border-box/s.test(loginPage) &&
      /\.login-page\s*{[^}]*overflow-x:\s*hidden/s.test(loginPage),
    `${appName} login page should prevent mobile horizontal overflow at the page boundary`
  )
  check(
    /\.card\s*{[^}]*width:\s*100%/s.test(loginPage) &&
      /\.card\s*{[^}]*box-sizing:\s*border-box/s.test(loginPage),
    `${appName} login card should fit within the padded mobile viewport`
  )
  check(
    /\.code-row\s*{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+180rpx/s.test(loginPage),
    `${appName} verification-code row should use a shrinkable input column and a mobile-safe button width`
  )
}

const rootLock = read('package-lock.json')
check(
  rootLock.includes('"uview-plus": "3.8.86"') && !rootLock.includes('"uview-ui": "2.0.36"'),
  'root package-lock.json should contain uview-plus and should not retain uview-ui'
)

const defaultRoute = readIfExists('admin-platform/web/src/utils/defaultRoute.js')
check(
  defaultRoute.includes("'dashboard': RIDE_HAILING_DASHBOARD_ROUTE") &&
    defaultRoute.includes('resolveDefaultRouteName'),
  'admin web should normalize legacy dashboard defaultRouter values to RideHailingDashboard'
)

const userStore = read('admin-platform/web/src/pinia/modules/user.js')
const permission = read('admin-platform/web/src/permission.js')
check(
  userStore.includes('normalizeUserInfoDefaultRoute') &&
    permission.includes('resolveDefaultRouteName'),
  'admin user and route guard should use default route normalization before navigation'
)

const dashboardApi = read('admin-platform/web/src/api/rideHailing/dashboard.js')
check(
  dashboardApi.includes("url: '/carpool/analytics/dashboard'") &&
    !dashboardApi.includes('/carpool/dashboard/overview'),
  'admin dashboard overview API should call the existing backend analytics dashboard endpoint'
)

if (failures.length) {
  console.error('Local frontend configuration check failed:')
  for (const failure of failures) {
    console.error(`- ${failure}`)
  }
  process.exit(1)
}

console.log('Local frontend configuration check passed.')
