import { request } from '@/utils/request'
import { mapOkData } from '@/utils/apiData'

function toNumber(value) {
  const n = Number(value)
  return Number.isFinite(n) ? n : 0
}

function normalizePoint(item = {}) {
  const latitude = item.latitude ?? item.lat ?? 0
  const longitude = item.longitude ?? item.lng ?? 0
  return {
    ...item,
    latitude: toNumber(latitude),
    longitude: toNumber(longitude),
    lat: toNumber(latitude),
    lng: toNumber(longitude),
    formattedAddress: item.formattedAddress ?? item.formatted_address ?? '',
    formatted_address: item.formatted_address ?? item.formattedAddress ?? '',
  }
}

function normalizeRoute(data = {}) {
  const payload = data || {}
  const polyline = (payload.polyline || payload.points || []).map(normalizePoint)
  return {
    ...payload,
    origin: normalizePoint(payload.origin),
    destination: normalizePoint(payload.destination),
    polyline,
    points: polyline,
    distanceMeters: payload.distanceMeters ?? payload.distance_meters ?? 0,
    durationSeconds: payload.durationSeconds ?? payload.duration_seconds ?? 0,
  }
}

export const geocodeAddress = (params = {}) =>
  mapOkData(
    request({ url: '/api/v1/maps/geocode', method: 'GET', params, silent: true }),
    normalizePoint
  )

export const reverseGeocode = (params = {}) =>
  mapOkData(
    request({ url: '/api/v1/maps/regeo', method: 'GET', params, silent: true }),
    normalizePoint
  )

export const getRoutePreview = (params = {}) =>
  mapOkData(
    request({ url: '/api/v1/maps/route', method: 'GET', params, silent: true }),
    normalizeRoute
  )

export const getWeather = (params = {}) =>
  mapOkData(
    request({ url: '/api/v1/maps/weather', method: 'GET', params, silent: true }),
    (data) => data || {}
  )
