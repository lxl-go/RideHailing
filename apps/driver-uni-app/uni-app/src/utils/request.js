const DEFAULT_BASE_URL = 'http://localhost:9000'
const BASE_URL = (import.meta.env?.VITE_API_BASE_URL || DEFAULT_BASE_URL).replace(/\/$/, '')
export const API_BASE_URL = BASE_URL

function createTraceId() {
  return `driver-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function readStorage(key, fallback = '') {
  try {
    return uni.getStorageSync(key) || fallback
  } catch {
    return fallback
  }
}

function saveSession(session = {}) {
  const userId = String(session.user_id || session.userId || session.userID || '')
  const accessToken = session.access_token || session.accessToken || ''
  const refreshToken = session.refresh_token || session.refreshToken || ''
  if (userId) uni.setStorageSync('driverUserId', userId)
  if (accessToken) uni.setStorageSync('driverAccessToken', accessToken)
  if (refreshToken) uni.setStorageSync('driverRefreshToken', refreshToken)
}

function clearSession() {
  uni.removeStorageSync('driverAccessToken')
  uni.removeStorageSync('driverRefreshToken')
}

function buildHeaders(extraHeaders = {}, traceId = '') {
  const headers = { 'Content-Type': 'application/json', ...extraHeaders }
  const token = readStorage('driverAccessToken')
  if (token) headers.Authorization = `Bearer ${token}`
  if (traceId) headers['X-Trace-Id'] = traceId
  return headers
}

function normalizeErrorMessage(message, statusCode) {
  const raw = String(message || '').trim()
  const lower = raw.toLowerCase()
  if (lower.includes('invalid principal')) return '请输入正确的手机号'
  if (lower.includes('invalid role')) return '登录身份类型不正确，请刷新后重试'
  if (lower.includes('please do not request sms code frequently')) return '验证码发送太频繁，请稍后再试'
  if (lower.includes('sms login locked')) return '验证码错误次数过多，请稍后再试'
  if (lower.includes('invalid sms code') || lower.includes('sms code not found')) return '验证码错误或已过期，请重新获取'
  if (lower.includes('sms send failed') || lower.includes('ihuyi sms') || lower.includes('internal error')) return '验证码发送失败，请检查短信服务配置或稍后重试'
  if (raw) return raw
  if (statusCode === 401) return '登录已过期，请重新登录'
  if (statusCode === 404) return '订单不存在或已被处理，请刷新后重试'
  return `请求失败(${statusCode})`
}

export function buildSocketUrl(path) {
  const normalizedPath = String(path || '').startsWith('/') ? String(path || '') : `/${path || ''}`
  return `${BASE_URL.replace(/^http/i, 'ws')}${normalizedPath}`
}

export function socketHeaders(extraHeaders = {}) {
  return buildHeaders(extraHeaders, createTraceId())
}

export function request(config) {
  return doRequest(config, false)
}

function doRequest(config, retried) {
  const traceId = config.traceId || createTraceId()
  config.traceId = traceId
  return new Promise((resolve) => {
    uni.request({
      url: `${BASE_URL}${config.url}`,
      method: config.method || 'GET',
      data: config.method === 'GET' ? config.params : config.data,
      header: buildHeaders(config.header, traceId),
      timeout: 15000,
      success: (res) => {
        const data = res.data ?? {}
        console.info('[driver-api]', traceId, config.method || 'GET', config.url, res.statusCode, data?.code ?? '')
        if (res.statusCode === 401 && !retried && !config.skipAuthRefresh) {
          refreshToken().then((ok) => {
            if (ok) doRequest(config, true).then(resolve)
            else resolve({ code: 401, data: null, msg: '登录已过期，请重新登录' })
          })
          return
        }
        if (res.statusCode >= 400) {
          const msg = normalizeErrorMessage(data?.msg || data?.message || data?.reason, res.statusCode)
          console.warn('[driver-api:error]', traceId, config.method || 'GET', config.url, res.statusCode, msg)
          if (!config.silent) uni.showToast({ title: msg, icon: 'none' })
          resolve({ code: res.statusCode, data: null, msg })
          return
        }
        if (data?.code === 0) {
          resolve(data)
          return
        }
        const msg = normalizeErrorMessage(data?.msg || data?.message || data?.reason || '请求失败', data?.code)
        console.warn('[driver-api:business]', traceId, config.method || 'GET', config.url, data?.code, msg)
        if (!config.silent) uni.showToast({ title: msg, icon: 'none' })
        resolve({ ...data, msg })
      },
      fail: () => {
        console.warn('[driver-api:network]', traceId, config.method || 'GET', config.url)
        if (!config.silent) uni.showToast({ title: '网络异常，请稍后重试', icon: 'none' })
        resolve({ code: -1, data: null, msg: '网络异常' })
      },
    })
  })
}

function refreshToken() {
  const refreshTokenValue = readStorage('driverRefreshToken')
  if (!refreshTokenValue) return Promise.resolve(false)
  return new Promise((resolve) => {
    uni.request({
      url: `${BASE_URL}/carpool/auth/refresh`,
      method: 'POST',
      data: { refresh_token: refreshTokenValue },
      header: { 'Content-Type': 'application/json' },
      timeout: 15000,
      success: (res) => {
        const session = res.data?.data || {}
        const accessToken = session.access_token || session.accessToken
        if (res.statusCode === 200 && res.data?.code === 0 && accessToken) {
          saveSession(session)
          resolve(true)
          return
        }
        clearSession()
        resolve(false)
      },
      fail: () => resolve(false),
    })
  })
}
