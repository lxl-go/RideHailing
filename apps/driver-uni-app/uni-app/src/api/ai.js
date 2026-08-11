import { request } from '@/utils/request'
import { mapOkData, mobilePageParams } from '@/utils/apiData'

export const getDriverAIAlerts = (params = {}) =>
  mapOkData(
    request({ url: '/api/v1/driver/ai-alerts', method: 'GET', params: mobilePageParams(params), silent: true }),
    (data) => {
      const payload = data || {}
      const list = payload.items || payload.list || []
      return { ...payload, items: list, list }
    }
  )

export const getDriverTrackReplay = (params = {}) =>
  request({ url: '/api/v1/driver/track/replay', method: 'GET', params: mobilePageParams(params), silent: true })

export const getDriverTravelAdvice = (data = {}) =>
  request({ url: '/api/v1/driver/ai/travel-advice', method: 'POST', data, silent: true })

export const getDriverAIWarning = (data = {}) =>
  request({ url: '/api/v1/driver/ai/travel-advice', method: 'POST', data, silent: true })
