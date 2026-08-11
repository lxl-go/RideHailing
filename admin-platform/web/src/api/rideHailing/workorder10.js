import service from '@/utils/request'

export const getAISummary = () => service({ url: '/carpool/ai/summary', method: 'get' })

export const chatAI = (data) => service({ url: '/carpool/ai/chat', method: 'post', data })

export const planRainRoute = (data) => service({ url: '/carpool/ai/rain-route', method: 'post', data })

export const chatWithRainRoute = (data) => service({ url: '/carpool/ai/chat-with-route', method: 'post', data })

export const reportFlooding = (data) => service({ url: '/carpool/ai/flood-report', method: 'post', data })

export const listAIConversationLogs = (params = {}) => service({ url: '/carpool/ai/conversation/list', method: 'get', params })

export const listAIRoutePlanLogs = (params = {}) => service({ url: '/carpool/ai/route-plan/list', method: 'get', params })

export const listAIFloodReports = (params = {}) => service({ url: '/carpool/ai/flood-report/list', method: 'get', params })

export const auditFloodReport = (data) => service({ url: '/carpool/ai/flood-report/audit', method: 'post', data })

export const exportAI = () => service({ url: '/carpool/ai/export', method: 'post' })
