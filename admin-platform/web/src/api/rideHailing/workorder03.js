import service from '@/utils/request'

export const getFinanceSummary = () => {
  return service({
    url: '/carpool/finance/summary',
    method: 'get'
  })
}

export const listFinanceTransactions = (params = {}) => {
  return service({
    url: '/carpool/finance/transaction/list',
    method: 'get',
    params
  })
}

export const listFinanceRefunds = (params = {}) => {
  return service({
    url: '/carpool/finance/refund/list',
    method: 'get',
    params
  })
}

export const listAbnormalTransactions = () => {
  return service({
    url: '/carpool/finance/abnormal/list',
    method: 'get'
  })
}

export const exportFinance = () => {
  return service({
    url: '/carpool/finance/export',
    method: 'post'
  })
}
