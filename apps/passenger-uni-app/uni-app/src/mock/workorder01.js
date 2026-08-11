export const mockRegisterList = Array.from({ length: 20 }, (_, i) => ({
  id: i + 1, name: `用户${i + 1}`, phone: `1380000${String(i).padStart(4, '0')}`,
  idCard: `1101011990${String(i).padStart(2, '0')}${String(Math.floor(Math.random() * 99)).padStart(2, '0')}1234`,
  type: i % 2 === 0 ? '司机' : '乘客', status: i < 12 ? '已认证' : '待审核',
  registerTime: `2025-08-${String(10 + (i % 20)).padStart(2, '0')}`
}))

export const mockTripList = Array.from({ length: 20 }, (_, i) => ({
  id: i + 1, orderNo: `WO2025${String(i + 1).padStart(6, '0')}`, passenger: `乘客${i + 1}`,
  from: ['北京国贸', '上海陆家嘴', '广州天河', '深圳南山', '杭州西湖'][i % 5],
  to: ['北京望京', '上海静安寺', '深圳福田', '广州珠江新城', '杭州滨江'][i % 5],
  time: `2025-09-${String(10 + (i % 20)).padStart(2, '0')} ${String(8 + (i % 10)).padStart(2, '0')}:00`,
  seats: (i % 4) + 1, status: ['待出发', '进行中', '已完成', '已取消'][i % 4]
}))

export const mockReviewList = Array.from({ length: 10 }, (_, i) => ({
  id: i + 1, orderNo: `WO2025${String(i + 1).padStart(6, '0')}`,
  from: i % 2 === 0 ? '乘客A' : '司机B', target: i % 2 === 0 ? '司机B' : '乘客A',
  score: 3 + (i % 3), content: ['服务态度很好', '准时到达', '车内整洁'][i % 3],
  time: `2025-09-${String(10 + i).padStart(2, '0')} 14:${String(30 + i).padStart(2, '0')}`
}))

export const mockEmergencyList = Array.from({ length: 5 }, (_, i) => ({
  id: i + 1, orderNo: `WO2025${String(i + 1).padStart(6, '0')}`, passenger: `乘客${i + 1}`,
  time: `2025-09-${String(10 + i).padStart(2, '0')} ${15 + i}:00`,
  type: ['交通事故', '身体不适', '路线纠纷', '车辆故障', '其他'][i],
  status: i < 3 ? '已处理' : '处理中'
}))
