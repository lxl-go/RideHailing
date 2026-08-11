import { request } from '@/utils/request'

export const sendPassengerLoginCode = (mobile) =>
  request({
    url: '/carpool/auth/sms/send',
    method: 'POST',
    data: {
      mobile,
      role: 'passenger'
    },
    silent: false,
    skipAuthRefresh: true
  })

export const loginPassenger = (principal, code = '') =>
  request({
    url: '/carpool/auth/login',
    method: 'POST',
    data: {
      principal,
      role: 'passenger',
      code
    },
    silent: true,
    skipAuthRefresh: true
  })

export const refreshPassengerToken = (refreshToken) =>
  request({
    url: '/carpool/auth/refresh',
    method: 'POST',
    data: {
      refresh_token: refreshToken
    },
    silent: true,
    skipAuthRefresh: true
  })

export const logoutPassenger = (refreshToken) =>
  request({
    url: '/carpool/auth/logout',
    method: 'POST',
    data: {
      refresh_token: refreshToken
    },
    silent: true,
    skipAuthRefresh: true
  })
