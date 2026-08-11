import { request } from '@/utils/request'
import { mapOkData, normalizeVehicle, unwrap } from '@/utils/apiData'

function vehiclePayload(data = {}) {
  return {
    plateNo: data.plateNo ?? data.plate_no,
    brand: data.brand,
    model: data.model,
    color: data.color,
    vehicleType: data.vehicleType ?? data.vehicle_type,
    seats: Number(data.seats) || 4,
  }
}

export const listVehicles = () =>
  mapOkData(
    request({ url: '/carpool/drivers/vehicles', method: 'GET', silent: true }),
    (data) => {
      const payload = data || {}
      return {
        ...payload,
        items: (payload.items || payload.list || []).map(normalizeVehicle),
        list: (payload.list || payload.items || []).map(normalizeVehicle),
      }
    }
  )

export const saveVehicle = (data) =>
  mapOkData(
    request({ url: '/carpool/drivers/vehicles', method: 'POST', data: vehiclePayload(data) }),
    (payload) => normalizeVehicle(unwrap(payload, ['vehicle']))
  )

export const updateVehicle = (id, data) =>
  mapOkData(
    request({ url: `/carpool/drivers/vehicles/${id}`, method: 'PUT', data: vehiclePayload(data) }),
    (payload) => normalizeVehicle(unwrap(payload, ['vehicle']))
  )

export const deleteVehicle = (id) =>
  request({ url: `/carpool/drivers/vehicles/${id}`, method: 'DELETE' })
