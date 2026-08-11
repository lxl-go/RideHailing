import service from '@/utils/request'

export const getPerformanceSummary = () => service({ url: '/carpool/performance/summary', method: 'get' })

export const listPerformanceReports = (params = {}) => service({ url: '/carpool/performance/report/list', method: 'get', params })

export const createPerformanceReport = (data) => service({ url: '/carpool/performance/report', method: 'post', data })

export const listPerformanceScenarios = (params = {}) => service({ url: '/carpool/performance/scenario/list', method: 'get', params })

export const getRuntimeSnapshot = () => service({ url: '/carpool/performance/runtime', method: 'get' })

export const exportPerformanceReports = () => service({ url: '/carpool/performance/export', method: 'post' })
