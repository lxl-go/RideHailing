import { request } from '@/utils/request'
import { mapOkData } from '@/utils/apiData'

const AMAP_KEY = '22ba26c4d757d904aef8138acda60ab7'

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

function requestAmap(url, data = {}) {
  return new Promise((resolve) => {
    uni.request({
      url,
      method: 'GET',
      data: { key: AMAP_KEY, ...data },
      timeout: 12000,
      success: (res) => resolve(res.data || {}),
      fail: () => resolve({ status: '0', tips: [], route: null }),
    })
  })
}

function parseLocation(location = '') {
  if (!location || typeof location !== 'string' || !location.includes(',')) return { latitude: 0, longitude: 0 }
  const [longitude, latitude] = location.split(',').map(toNumber)
  return { latitude, longitude, lat: latitude, lng: longitude }
}

function normalizeTip(item = {}) {
  const point = parseLocation(item.location)
  return {
    id: item.id || `${item.name || ''}-${item.adcode || ''}-${item.location || ''}`,
    name: item.name || '',
    district: item.district || '',
    address: item.address && typeof item.address === 'string' ? item.address : '',
    latitude: point.latitude,
    longitude: point.longitude,
    lat: point.latitude,
    lng: point.longitude,
  }
}

function parsePolyline(polyline = '') {
  return String(polyline || '')
    .split(';')
    .map(parseLocation)
    .filter((point) => point.latitude && point.longitude)
}

export async function getAmapInputTips(keyword, city = '全国') {
  const trimmed = String(keyword || '').trim()
  if (!trimmed) return []
  const data = await requestAmap('https://restapi.amap.com/v3/assistant/inputtips', {
    keywords: trimmed,
    city,
    citylimit: false,
    datatype: 'all',
  })
  if (data.status !== '1') return []
  return (data.tips || [])
    .map(normalizeTip)
    .filter((item) => item.name && item.latitude && item.longitude)
}

export async function getAmapDrivingRoute(origin, destination) {
  if (!origin?.longitude || !origin?.latitude || !destination?.longitude || !destination?.latitude) {
    return { distanceMeters: 0, durationSeconds: 0, points: [] }
  }
  const data = await requestAmap('https://restapi.amap.com/v3/direction/driving', {
    origin: `${origin.longitude},${origin.latitude}`,
    destination: `${destination.longitude},${destination.latitude}`,
    extensions: 'base',
  })
  const path = data.route?.paths?.[0] || {}
  const steps = path.steps || []
  const points = steps.flatMap((step) => parsePolyline(step.polyline))
  return {
    distanceMeters: toNumber(path.distance),
    durationSeconds: toNumber(path.duration),
    points,
    polyline: points,
  }
}
