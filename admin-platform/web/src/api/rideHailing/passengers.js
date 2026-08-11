import service from '@/utils/request'
import { listPersons, getPerson, updatePerson, batchPersonStatus } from '@/api/rideHailing/workorder05'

export const listPassengerRecords = (params = {}) =>
  listPersons({ ...params, personType: 'passenger', roleCode: 'passenger' })

export const getPassengerRecord = (id) => getPerson(id)

export const updatePassengerRecord = (id, data) => updatePerson(id, data)

export const batchPassengerStatus = (data) => batchPersonStatus(data)

export const getPassengerStats = () =>
  service({ url: '/carpool/passenger/stats', method: 'get' })
