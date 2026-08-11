import { request } from '@/utils/request'
import { mapOkData, normalizeTrip, unwrap } from '@/utils/apiData'

function publishTripPayload(data = {}) {
  const seatsTotal = Number(data.seats_total ?? data.seatsTotal ?? data.seats ?? 4)
  return {
    ...data,
    request_id: data.request_id || `trip-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    seats_total: Math.min(6, Math.max(1, seatsTotal || 4)),
  }
}

function normalizeTripList(data) {
  const payload = data || {}
  return {
    ...payload,
    items: (payload.items || payload.list || []).map(normalizeTrip),
    list: (payload.list || payload.items || []).map(normalizeTrip),
  }
}

export const validateLocation = (data) =>
  request({ url: '/carpool/trips/locations/validate', method: 'POST', data })

export const suggestLocations = (data) =>
  request({ url: '/carpool/trips/locations/suggest', method: 'POST', data, silent: true })

export const previewTripPrice = (data) =>
  request({ url: '/carpool/trips/price-preview', method: 'POST', data })

export const publishTrip = (data) =>
  request({ url: '/carpool/trips', method: 'POST', data: publishTripPayload(data) })

export const listMyTrips = (params) =>
  mapOkData(
    request({ url: '/carpool/trips/mine', method: 'GET', params }),
    normalizeTripList
  )

export const getTripDetail = (id) =>
  mapOkData(
    request({ url: `/carpool/trips/${id}`, method: 'GET' }),
    (data) => normalizeTrip(unwrap(data, ['trip']))
  )

export const updateTripStatus = (id, status) =>
  request({ url: `/carpool/trips/${id}/status`, method: 'PUT', data: { status } })

export const deleteTrip = (id) =>
  request({ url: `/carpool/trips/${id}`, method: 'DELETE' })
