import service from '@/utils/request'

export const listPersons = (params = {}) => service({ url: '/carpool/person/list', method: 'get', params })

export const getPerson = (id) => service({ url: `/carpool/person/${id}`, method: 'get' })

export const createPerson = (data) => service({ url: '/carpool/person', method: 'post', data })

export const updatePerson = (id, data) => service({ url: `/carpool/person/${id}`, method: 'put', data })

export const assignPersonRoles = (data) => service({ url: '/carpool/person/roles', method: 'post', data })

export const batchPersonStatus = (data) => service({ url: '/carpool/person/batch/status', method: 'post', data })

export const batchDeleteDrivers = (data) => service({ url: '/carpool/person/driver/batch/delete', method: 'post', data })

export const previewPersonImport = (data) => service({ url: '/carpool/person/import/preview', method: 'post', data })

export const commitPersonImport = (data) => service({ url: '/carpool/person/import/commit', method: 'post', data })

export const listPersonImportErrors = (params = {}) => service({ url: '/carpool/person/import/errors', method: 'get', params })

export const exportPersons = () => service({ url: '/carpool/person/export', method: 'post' })
