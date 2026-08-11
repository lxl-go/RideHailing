import service from '@/utils/request'
import { listPersons, getPerson, updatePerson, batchPersonStatus } from '@/api/rideHailing/workorder05'

export const listDriverRecords = (params = {}) =>
  listPersons({ ...params, personType: 'driver', roleCode: 'carpool_driver' })

export const getDriverRecord = (id) => getPerson(id)

export const updateDriverRecord = (id, data) => updatePerson(id, data)

export const batchDriverStatus = (data) => batchPersonStatus(data)

export const getDriverStats = () =>
  service({ url: '/carpool/driver/stats', method: 'get' })
