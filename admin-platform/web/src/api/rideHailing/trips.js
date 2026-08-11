import service from '@/utils/request'

export const listTrips = (params = {}) => service({ url: '/carpool/trip/list', method: 'get', params })
export const getTrip = (id) => service({ url: `/carpool/trip/${id}`, method: 'get' })
export const reviewTrip = (id, data) => service({ url: `/carpool/trip/${id}/review`, method: 'post', data })
export const deactivateTrip = (id, data) => service({ url: `/carpool/trip/${id}/deactivate`, method: 'post', data })
