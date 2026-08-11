import { request } from '@/utils/request'
import { mobilePageParams } from '@/utils/apiData'

function locationPayload(data = {}) {
  const reportedAt = data.reportedAt || data.reported_at || data.timestamp
  return {
    ...data,
    orderId: data.orderId ?? data.order_id,
    lat: data.lat ?? data.latitude,
    lng: data.lng ?? data.longitude,
    reportedAt: typeof reportedAt === 'number' ? new Date(reportedAt).toISOString() : reportedAt,
  }
}

export const reportDriverLocation = (data) =>
  request({ url: '/api/v1/driver/location/report', method: 'POST', data: locationPayload(data), silent: true })

export const getDriverLocationHistory = (params = {}) =>
  request({ url: '/api/v1/driver/location/history', method: 'GET', params: mobilePageParams(params), silent: true })
