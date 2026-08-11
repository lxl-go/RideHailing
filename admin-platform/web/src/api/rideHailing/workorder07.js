import service from '@/utils/request'

export const createCouponTemplate = (data) => service({ url: '/carpool/marketing/coupon/template', method: 'post', data })

export const listCouponTemplates = (params = {}) => service({ url: '/carpool/marketing/coupon/template/list', method: 'get', params })

export const deleteCouponTemplate = (couponNo) => service({ url: `/carpool/marketing/coupon/template/${couponNo}`, method: 'delete' })

export const issueCoupon = (data) => service({ url: '/carpool/marketing/coupon/issue', method: 'post', data })

export const redeemCoupon = (data) => service({ url: '/carpool/marketing/coupon/redeem', method: 'post', data })

export const listUserCoupons = (params = {}) => service({ url: '/carpool/marketing/coupon/user/list', method: 'get', params })

export const listCampaigns = (params = {}) => service({ url: '/carpool/marketing/campaign/list', method: 'get', params })

export const getReferralSummary = () => service({ url: '/carpool/marketing/referral/summary', method: 'get' })

export const exportMarketing = () => service({ url: '/carpool/marketing/export', method: 'post' })
