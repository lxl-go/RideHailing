import service from '@/utils/request'

export const getGvaGovernanceSummary = () => service({ url: '/system/gva-governance/summary', method: 'get' })

export const getGvaRouteSnapshot = () => service({ url: '/system/gva-governance/routes', method: 'get' })

export const getGvaAuditSnapshot = () => service({ url: '/system/gva-governance/audit', method: 'get' })

export const getGvaDatasourceSnapshot = () => service({ url: '/system/gva-governance/datasource', method: 'get' })

export const getGvaTimedTaskSnapshot = () => service({ url: '/system/gva-governance/timed-task', method: 'get' })

export const exportGvaGovernance = () => service({ url: '/system/gva-governance/export', method: 'post' })
