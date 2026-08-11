import service from '@/utils/request'

export const listDispatchOrders = (params = {}) => service({ url: '/carpool/dispatch/order/list', method: 'get', params })

export const getDispatchOrderDetail = (id) => service({ url: `/carpool/dispatch/order/${id}`, method: 'get' })

export const cancelDispatchOrder = (id, data) => service({ url: `/carpool/dispatch/order/${id}/cancel`, method: 'post', data })

export const reassignDispatchOrder = (id, data) => service({ url: `/carpool/dispatch/order/${id}/reassign`, method: 'post', data })

export const scoreDrivers = (data) => service({ url: '/carpool/dispatch/score', method: 'post', data })

export const listDispatchConfigs = () => service({ url: '/carpool/dispatch/config/list', method: 'get' })

export const saveDispatchConfig = (data) => service({ url: '/carpool/dispatch/config', method: 'post', data })

export const publishDispatchConfig = (id) => service({ url: `/carpool/dispatch/config/${id}/publish`, method: 'post' })

export const rollbackDispatchConfig = (id) => service({ url: `/carpool/dispatch/config/${id}/rollback`, method: 'post' })

export const listDispatchAudits = (params = {}) => service({ url: '/carpool/dispatch/audit/list', method: 'get', params })

export const replayDispatchTrack = (params = {}) => service({ url: '/carpool/dispatch/track/replay', method: 'get', params })

export const exportDispatch = () => service({ url: '/carpool/dispatch/export', method: 'post' })
