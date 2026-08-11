export const workorders = [
  { id: '01', name: '顺风车', moduleName: '顺风车管理', summary: '实名审核、车辆审核和顺风车三端真实接口联调入口。' },
  { id: '02', name: '班车', moduleName: '班车管理', summary: '班线、站点、排班、购票、今日排班占位页面与接口。' },
  { id: '03', name: '财务', moduleName: '财务管理', summary: '交易流水、对账、结算、退款进度和收入流水。' },
  { id: '04', name: '订单', moduleName: '订单管理', summary: '订单列表、详情、退票审核、批量退票和状态机约束。' },
  { id: '05', name: '人员', moduleName: '人员管理', summary: '角色、司机、乘客、导入、禁用申诉链路。' },
  { id: '06', name: '数据分析', moduleName: '数据分析', summary: 'KPI、趋势、漏斗、用户司机分析和异步报表导出。' },
  { id: '07', name: '营销', moduleName: '营销管理', summary: '优惠券模板、发放规则、活动、推荐奖励和退券结果。' },
  { id: '08', name: '性能', moduleName: '性能专项', summary: 'QPS、P99、pprof、压测报告、WS/地图/位置上报验收。' },
  { id: '09', name: 'GVA', moduleName: 'GVA专项', summary: '动态路由、多数据源、权限审计、XxlJob 参数治理。' },
  { id: '10', name: 'AI助手', moduleName: 'AI智能出行助手', summary: 'AI配置、对话日志、积水审核、降级和隐私删除。' },
  { id: '11', name: '派单', moduleName: '订单派单中心', summary: '规则配置、派单日志、监控、位置上报和行程追踪。' },
]

export const statusTextMap = {
  0: '待审核',
  1: '通过',
  2: '驳回',
  3: '补充',
}

export const reviewTypeOptions = [
  { label: '实名认证', value: 'cert' },
  { label: '车辆审核', value: 'vehicle' },
]
