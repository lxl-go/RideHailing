import { request } from '@/utils/request'
import { mapOkData, normalizeTrip, unwrap } from '@/utils/apiData'

function normalizeTripList(data) {
  const payload = data || {}
  return {
    ...payload,
    items: (payload.items || payload.list || []).map(normalizeTrip),
    list: (payload.list || payload.items || []).map(normalizeTrip),
  }
}

export const searchTrips = (params) =>
  mapOkData(
    request({ url: '/carpool/trips', method: 'GET', params }),
    normalizeTripList
  )

export const recommendTrips = (params) =>
  mapOkData(
    request({ url: '/carpool/trips/demands/recommendations', method: 'GET', params }),
    normalizeTripList
  )

export const getTripDetail = (id) =>
  mapOkData(
    request({ url: `/carpool/trips/${id}`, method: 'GET' }),
    (data) => normalizeTrip(unwrap(data, ['trip']))
  )

export const publishDemand = (data) =>
  request({ url: '/carpool/trips/demands', method: 'POST', data })

export const listMyDemands = (params) =>
  request({ url: '/carpool/trips/demands/mine', method: 'GET', params })

export const cancelDemand = (id) =>
  request({ url: `/carpool/trips/demands/${id}/cancel`, method: 'POST' })
