import { request } from '@/utils/request'
import { mapOkData, normalizeCertification, normalizeDriver, unwrap } from '@/utils/apiData'

export const getDriverInfo = () =>
  mapOkData(
    request({ url: '/carpool/drivers/me', method: 'GET', silent: true }),
    (data) => normalizeDriver(unwrap(data, ['driver']))
  )

export const submitCertification = (data) =>
  mapOkData(
    request({
      url: '/carpool/drivers/certification',
      method: 'POST',
      data: {
        real_name: data.real_name,
        id_card_no: data.id_card_no ?? data.id_card,
        license_no: data.license_no,
        license_type: data.license_type,
        city: data.city,
      },
    }),
    (payload) => normalizeCertification(unwrap(payload, ['certification']))
  )

export const getCertificationStatus = () =>
  mapOkData(
    request({ url: '/carpool/drivers/certification', method: 'GET', silent: true }),
    (data) => normalizeCertification(unwrap(data, ['certification']))
  )

export const updateDriverStatus = (status) =>
  mapOkData(
    request({ url: '/carpool/drivers/me', method: 'PUT', data: { service_status: status } }),
    (data) => normalizeDriver(unwrap(data, ['driver']))
  )

export const getDriverStats = (params = {}) =>
  request({ url: '/api/v1/driver/stats', method: 'GET', params, silent: true })

export const getDriverIncome = (params = {}) =>
  request({ url: '/api/v1/driver/income', method: 'GET', params, silent: true })
