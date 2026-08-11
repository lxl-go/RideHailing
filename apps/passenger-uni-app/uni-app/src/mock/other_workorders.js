export const mockRefundRecords = Array.from({ length: 10 }, (_, i) => ({
  id: i + 1, orderNo: `WO2025${String(i + 1).padStart(6, '0')}`, amount: (i + 1) * 15 + 5,
  reason: ['行程变更', '重复支付', '司机取消', '其他'][i % 4],
  status: ['处理中', '已完成', '已驳回'][i % 3],
  time: `2026-07-${String(15 + i).padStart(2, '0')}`
}))

export const mockCoupons = Array.from({ length: 8 }, (_, i) => ({
  id: i + 1, title: ['新用户立减', '周末特惠', '老用户回馈', '邀请有礼'][i % 4],
  amount: [10, 20, 5, 15][i % 4], rule: ['满20可用', '满50可用', '无门槛', '满30可用'][i % 4],
  expire: `2026-08-${String(15 + i * 2).padStart(2, '0')}`, status: ['available', 'used', 'expired'][i % 3]
}))
