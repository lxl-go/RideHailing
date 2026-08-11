import { request } from '@/utils/request'
import { mapOkData, mobilePageParams, normalizeOrder, unwrap } from '@/utils/apiData'

function orderPathId(id) {
  return encodeURIComponent(String(id ?? ''))
}

function actionHeaders(data = {}) {
  const key = data.idempotency_key || data.idempotencyKey
  return key ? { 'Idempotency-Key': String(key) } : {}
}

function normalizeOrderList(data) {
  const payload = data || {}
  return {
    ...payload,
    items: (payload.items || payload.list || []).map(normalizeOrder),
    list: (payload.list || payload.items || []).map(normalizeOrder),
  }
}

export const listAvailableOrders = (params = {}) =>
  mapOkData(
    request({ url: '/api/v1/driver/orders/available', method: 'GET', params: mobilePageParams(params), silent: true }),
    normalizeOrderList
  )

export const acceptOrder = (id, data = {}) =>
  request({ url: `/api/v1/driver/orders/${orderPathId(id)}/accept`, method: 'POST', data, header: actionHeaders(data) })

export const getDriverOrderDetail = (id) =>
  mapOkData(
    request({ url: `/api/v1/driver/orders/${orderPathId(id)}`, method: 'GET', silent: true }),
    (data) => normalizeOrder(unwrap(data, ['order']))
  )

export const rejectOrder = (id, data = {}) =>
  request({ url: `/api/v1/driver/orders/${orderPathId(id)}/reject`, method: 'POST', data, header: actionHeaders(data) })

export const startPickupOrder = (id, data = {}) =>
  request({ url: `/api/v1/driver/orders/${orderPathId(id)}/start-pickup`, method: 'POST', data, header: actionHeaders(data) })

export const startDeliveryOrder = (id, data = {}) =>
  request({ url: `/api/v1/driver/orders/${orderPathId(id)}/start-delivery`, method: 'POST', data, header: actionHeaders(data) })

export const completeOrder = (id, data = {}) =>
  request({ url: `/api/v1/driver/orders/${orderPathId(id)}/complete`, method: 'POST', data, header: actionHeaders(data) })

export const listDriverOrders = (params = {}) =>
  mapOkData(
    request({ url: '/api/v1/driver/orders', method: 'GET', params: mobilePageParams(params) }),
    normalizeOrderList
  )
