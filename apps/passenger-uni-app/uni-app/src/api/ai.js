import { request } from '@/utils/request'

// AI 对话
export const chatAI = (data) =>
  request({ url: '/api/v1/ai/chat', method: 'POST', data, silent: true })

// 雨天路线规划
export const planRainRoute = (data) =>
  request({ url: '/api/v1/ai/rain-route', method: 'POST', data, silent: true })

// 带路线的 AI 对话
export const chatWithRoute = (data) =>
  request({ url: '/api/v1/ai/chat-with-route', method: 'POST', data, silent: true })

// 积水上报
export const submitFloodReport = (data) =>
  request({ url: '/api/v1/ai/flood-report', method: 'POST', data })
