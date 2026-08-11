const rejectedReasonKeys = ['reject_reason', 'rejectReason', 'refund_reason', 'refundReason']

const statusTextMap = {
  pending: '已预约',
  waiting: '已预约',
  waiting_pay: '待支付',
  paid: '待出行',
  accepted: '已接单',
  picking_up: '司机来接您',
  delivering: '前往目的地',
  ongoing: '进行中',
  in_progress: '进行中',
  completed: '已完成',
  cancelled: '已取消',
  rejected: '已拒绝',
  unknown: '未知状态',
}

const statusTypeMap = {
  pending: 'primary',
  waiting: 'primary',
  waiting_pay: 'warning',
  paid: 'warning',
  accepted: 'success',
  picking_up: 'primary',
  delivering: 'primary',
  ongoing: 'primary',
  in_progress: 'primary',
  completed: 'success',
  cancelled: 'error',
  rejected: 'error',
  unknown: 'info',
}

function hasRejectReason(order = {}) {
  return rejectedReasonKeys.some((key) => String(order?.[key] || '').trim())
}

export function getOrderStatusText(status, order = {}) {
  const normalized = String(status || 'unknown')
  if (normalized === 'cancelled' && hasRejectReason(order)) return statusTextMap.rejected
  return statusTextMap[normalized] || statusTextMap.unknown
}

export function getOrderStatusType(status, order = {}) {
  const normalized = String(status || 'unknown')
  if (normalized === 'cancelled' && hasRejectReason(order)) return statusTypeMap.rejected
  return statusTypeMap[normalized] || statusTypeMap.unknown
}
