import service from '@/utils/request'
import { getAnalyticsDashboard } from '@/api/rideHailing/workorder06'

export { getAnalyticsDashboard }

export const getDashboardOverview = (params = {}) =>
  service({ url: '/carpool/analytics/dashboard', method: 'get', params })
