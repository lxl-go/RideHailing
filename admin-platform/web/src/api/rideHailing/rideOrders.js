import service from '@/utils/request'
import { listOrders, getOrderDetail, exportOrders } from '@/api/rideHailing/workorder04'

export const getOrderOverview = (params = {}) =>
  service({ url: '/carpool/order/overview', method: 'get', params })

export { listOrders, getOrderDetail, exportOrders }
