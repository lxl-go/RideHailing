import { request } from '@/utils/request'
import { mapOkData, normalizePassenger, unwrap } from '@/utils/apiData'

export const getProfile = () =>
  mapOkData(
    request({ url: '/carpool/passengers/me', method: 'GET', silent: true }),
    (data) => normalizePassenger(unwrap(data, ['passenger']))
  )

export const updateProfile = (data) =>
  mapOkData(
    request({ url: '/carpool/passengers/me', method: 'PUT', data }),
    (payload) => normalizePassenger(unwrap(payload, ['passenger']))
  )

export const listCoupons = (params = {}) =>
  request({ url: '/carpool/coupons', method: 'GET', params })

export const claimCoupon = (couponId) =>
  request({ url: '/carpool/coupons/claim', method: 'POST', data: { coupon_id: couponId } })
