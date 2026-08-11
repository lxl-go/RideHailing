import { request } from '@/utils/request'

// 提交评价
export const submitReview = (data) =>
  request({ url: '/carpool/reviews', method: 'POST', data })

// 查看我对某订单的评价
export const getMyReview = (orderId) =>
  request({ url: `/carpool/reviews/mine/${orderId}`, method: 'GET', silent: true })
