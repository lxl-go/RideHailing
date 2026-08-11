import { request } from '@/utils/request'
import { mapOkData, normalizeOrder, unwrap } from '@/utils/apiData'

function orderPathId(id) {
  return encodeURIComponent(String(id ?? ''))
}

function generateIdempotencyKey(prefix) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function actionHeaders(data = {}, fallbackPrefix = '') {
  const key = data.idempotency_key || data.idempotencyKey || (fallbackPrefix ? generateIdempotencyKey(fallbackPrefix) : '')
  return key ? { 'Idempotency-Key': String(key) } : {}
}

function createOrderPayload(data = {}) {
  const tripId = data.trip_id ?? data.tripId
  const seatsBooked = data.seats_booked ?? data.seatsBooked ?? data.seats ?? 1
  return {
    ...data,
    trip_id: String(tripId || ''),
    seats_booked: Number(seatsBooked) || 1,
  }
}

export const createOrder = (data) =>
  mapOkData(
    request({
      url: '/carpool/orders',
      method: 'POST',
      data: createOrderPayload(data),
      header: actionHeaders(data, 'create-order'),
    }),
    (payload = {}) => ({
      ...payload,
      order_id: payload.order_id ?? payload.orderId,
      total_price: payload.total_price ?? payload.totalPrice,
    })
  )

export const cancelOrder = (id, data = {}) =>
  request({
    url: `/carpool/orders/${orderPathId(id)}/cancel`,
    method: 'POST',
    data,
    header: actionHeaders(data, 'cancel-order'),
  })

export const listOrders = (params) =>
  mapOkData(
    request({ url: '/carpool/orders', method: 'GET', params }),
    (data) => {
      const payload = data || {}
      return {
        ...payload,
        items: (payload.items || payload.list || []).map(normalizeOrder),
        list: (payload.list || payload.items || []).map(normalizeOrder),
      }
    }
  )

export const getOrderDetail = (id) =>
  mapOkData(
    request({ url: `/carpool/orders/${orderPathId(id)}`, method: 'GET' }),
    (data) => normalizeOrder(unwrap(data, ['order']))
  )

export const getOrderTrack = (id, params = {}) =>
  request({
    url: `/api/v1/passenger/orders/${orderPathId(id)}/track`,
    method: 'GET',
    params,
    silent: true,
  })

export const payOrder = (id, data = {}) =>
  request({ url: `/carpool/orders/${orderPathId(id)}/pay`, method: 'POST', data })

export const syncPayment = (id, data = {}) =>
  request({ url: `/carpool/orders/${orderPathId(id)}/payment/sync`, method: 'POST', data, silent: true })
