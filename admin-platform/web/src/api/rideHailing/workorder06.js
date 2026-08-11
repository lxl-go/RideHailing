import service from '@/utils/request'

export const getAnalyticsDashboard = (params = {}) => service({ url: '/carpool/analytics/dashboard', method: 'get', params })

export const getOrderVolume = (params = {}) => service({ url: '/carpool/analytics/order-volume', method: 'get', params })

export const getOrderClassification = (params = {}) => service({ url: '/carpool/analytics/classification', method: 'get', params })

export const getConversion = (params = {}) => service({ url: '/carpool/analytics/conversion', method: 'get', params })

export const getRepurchase = (params = {}) => service({ url: '/carpool/analytics/repurchase', method: 'get', params })

export const exportAnalytics = () => service({ url: '/carpool/analytics/export', method: 'post' })
