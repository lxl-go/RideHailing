import { request } from '@/utils/request'

export const sendDriverLoginCode = (mobile) =>
  request({
    url: '/carpool/auth/sms/send',
    method: 'POST',
    data: {
      mobile,
      role: 'driver'
    },
    silent: false,
    skipAuthRefresh: true
  })

export const loginDriver = (principal, code = '') =>
  request({
    url: '/carpool/auth/login',
    method: 'POST',
    data: {
      principal,
      role: 'driver',
      code
    },
    silent: true,
    skipAuthRefresh: true
  })

export const refreshDriverToken = (refreshToken) =>
  request({
    url: '/carpool/auth/refresh',
    method: 'POST',
    data: {
      refresh_token: refreshToken
    },
    silent: true,
    skipAuthRefresh: true
  })

export const logoutDriver = (refreshToken) =>
  request({
    url: '/carpool/auth/logout',
    method: 'POST',
    data: {
      refresh_token: refreshToken
    },
    silent: true,
    skipAuthRefresh: true
  })
