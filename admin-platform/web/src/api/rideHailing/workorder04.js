import service from '@/utils/request'

export const listOrders = (params = {}) => service({ url: '/carpool/order/list', method: 'get', params })

export const getOrderDetail = (orderNo) => service({ url: `/carpool/order/${orderNo}`, method: 'get' })

export const listRefunds = (params = {}) => service({ url: '/carpool/order/refund/list', method: 'get', params })

export const applyRefund = (data) => service({ url: '/carpool/order/refund/apply', method: 'post', data })

export const reviewRefund = (data) => service({ url: '/carpool/order/refund/review', method: 'post', data })

export const batchRefund = (data) => service({ url: '/carpool/order/refund/batch', method: 'post', data })

export const exportOrders = (params = {}) => service({ url: '/carpool/order/export', method: 'post', params })
