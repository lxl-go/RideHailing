<template>
  <div>
    <el-tabs v-model="currentWorkorder" class="workorder-tabs" @tab-change="emitSwitch">
      <el-tab-pane v-for="item in workorders" :key="item.id" :label="`${item.id} ${item.name}`" :name="item.id" />
    </el-tabs>

    <el-alert :title="currentSummary" type="info" show-icon :closable="false" class="module-alert" />

    <el-form :inline="true" :model="searchForm" class="search-form">
      <el-form-item label="关键词">
        <el-input v-model="searchForm.keyword" clearable placeholder="订单号/标题/车牌" style="width: 220px" @keyup.enter="$emit('search')" />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="searchForm.status" clearable placeholder="全部状态" style="width: 150px" @change="$emit('search')">
          <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="currentWorkorder === '01'" label="审核类型">
        <el-segmented v-model="reviewTypeModel" :options="reviewTypeOptions" @change="emitReviewType" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="Search" @click="$emit('search')">查询</el-button>
        <el-button icon="RefreshLeft" @click="$emit('reset')">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const workorders = [
  { id: '01', name: '顺风车', summary: '认证审核、车辆审核与基础总览已完成。' },
  { id: '02', name: '班车', summary: '班车购票、司机排班与管理端班线页已补齐。' },
  { id: '03', name: '财务', summary: '退款、流水、对账与结算待开发。' },
  { id: '04', name: '订单', summary: '订单列表、详情、状态流转待开发。' },
  { id: '05', name: '人员', summary: '司机、乘客、审核与禁用流转待开发。' },
  { id: '06', name: '数据分析', summary: 'KPI、趋势、漏斗和导出待开发。' },
  { id: '07', name: '营销', summary: '优惠券模板、发放、核销、活动和推荐奖励已接入。' },
  { id: '08', name: 'Go性能', summary: '性能指标、压测报告和验收目标已接入。' },
  { id: '09', name: 'GVA框架', summary: '动态路由、权限审计、多数据源健康和定时任务治理已接入。' },
  { id: '10', name: 'AI助手', summary: '行程问答、推荐与日志审计待开发。' },
  { id: '11', name: '派单中心', summary: '派单规则、抢单、监控和轨迹待开发。' },
]

const reviewTypeOptions = [
  { label: '实名认证', value: 'cert' },
  { label: '车辆审核', value: 'vehicle' },
]

const props = defineProps({
  activeWorkorder: { type: String, default: '01' },
  reviewType: { type: String, default: 'cert' },
  searchForm: { type: Object, default: () => ({ keyword: '', status: '' }) },
})

const emit = defineEmits(['switch', 'search', 'reset', 'update:activeWorkorder', 'update:reviewType', 'reviewTypeChange'])

const currentWorkorder = computed({
  get: () => props.activeWorkorder,
  set: (value) => emit('update:activeWorkorder', value),
})

const reviewTypeModel = computed({
  get: () => props.reviewType,
  set: (value) => emit('update:reviewType', value),
})

const currentSummary = computed(() => {
  return workorders.find((item) => item.id === props.activeWorkorder)?.summary || ''
})

const statusOptions = computed(() => {
  return props.activeWorkorder === '01'
    ? ['待审核', '已通过', '已驳回', '补充材料']
    : ['待处理', '已启用', '处理中', '已完成']
})

const emitSwitch = (value) => emit('switch', value)
const emitReviewType = () => emit('reviewTypeChange', props.reviewType)
</script>
