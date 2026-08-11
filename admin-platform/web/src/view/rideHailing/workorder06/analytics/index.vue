<template>
  <div class="analytics-page">
    <RidePageHeader title="数据分析" subtitle="订单趋势、转化漏斗、复购与分类统计">
      <el-button icon="Download" @click="handleExport">导出</el-button>
    </RidePageHeader>
    <div class="toolbar">
      <el-form :inline="true" :model="filters">
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            value-format="YYYY-MM-DD"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item label="订单来源">
          <el-select v-model="filters.serviceType" clearable placeholder="全部" style="width: 140px">
            <el-option label="顺风车" value="carpool" />
            <el-option label="班车" value="shuttle" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="loadData">查询</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-row :gutter="12" class="kpi-grid">
      <el-col v-for="item in kpis" :key="item.key" :xs="24" :sm="12" :md="8" :lg="4">
        <div class="kpi-item">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </el-col>
    </el-row>

    <el-tabs v-model="activeTab" class="analysis-tabs">
      <el-tab-pane label="订单量分析" name="volume">
        <div ref="volumeChartRef" class="chart"></div>
        <el-table :data="volumeRows" border>
          <el-table-column prop="date" label="日期" />
          <el-table-column prop="total" label="订单数" />
          <el-table-column prop="valid" label="有效订单" />
          <el-table-column prop="growth" label="环比" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="用户转化分析" name="conversion">
        <div ref="conversionChartRef" class="chart compact"></div>
      </el-tab-pane>
      <el-tab-pane label="复购分析" name="repurchase">
        <div ref="repurchaseChartRef" class="chart compact"></div>
      </el-tab-pane>
      <el-tab-pane label="订单分类统计" name="classification">
        <div ref="classificationChartRef" class="chart compact"></div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import {
  exportAnalytics,
  getAnalyticsDashboard,
  getConversion,
  getOrderClassification,
  getOrderVolume,
  getRepurchase,
} from '@/api/rideHailing/workorder06'
import RidePageHeader from '@/components/RidePageHeader/index.vue'

defineOptions({ name: 'RideHailingAnalytics' })

const today = new Date()
const startOfMonth = new Date(today.getFullYear(), today.getMonth(), 1)
const formatDate = (date) => date.toISOString().slice(0, 10)

const dateRange = ref([formatDate(startOfMonth), formatDate(today)])
const filters = ref({ serviceType: '' })
const activeTab = ref('volume')
const dashboard = ref({})
const volume = ref({ categories: [], totalOrders: [], validOrders: [], growthRates: [] })
const classification = ref({})
const conversion = ref({})
const repurchase = ref({ repurchaseBuckets: [] })
const loading = ref(false)

const volumeChartRef = ref()
const conversionChartRef = ref()
const repurchaseChartRef = ref()
const classificationChartRef = ref()

const queryParams = computed(() => ({
  period: 'day',
  startDate: dateRange.value?.[0],
  endDate: dateRange.value?.[1],
  serviceType: filters.value.serviceType,
}))

const kpis = computed(() => [
  { key: 'today', label: '今日订单量', value: dashboard.value.todayOrderCount ?? 0 },
  { key: 'month', label: '本月订单量', value: dashboard.value.monthOrderCount ?? 0 },
  { key: 'revenue', label: '本月营收', value: `¥${dashboard.value.monthRevenue ?? 0}` },
  { key: 'drivers', label: '活跃司机数', value: dashboard.value.activeDrivers ?? 0 },
  { key: 'passengers', label: '活跃乘客数', value: dashboard.value.activePassengers ?? 0 },
  { key: 'conversion', label: '转化率', value: `${dashboard.value.conversionRate ?? 0}%` },
])

const volumeRows = computed(() => {
  return (volume.value.categories || []).map((date, index) => ({
    date,
    total: volume.value.totalOrders?.[index] ?? 0,
    valid: volume.value.validOrders?.[index] ?? 0,
    growth: `${volume.value.growthRates?.[index] ?? 0}%`,
  }))
})

const loadData = async () => {
  loading.value = true
  try {
    const [dashboardRes, volumeRes, classificationRes, conversionRes, repurchaseRes] = await Promise.all([
      getAnalyticsDashboard(queryParams.value),
      getOrderVolume(queryParams.value),
      getOrderClassification(queryParams.value),
      getConversion(queryParams.value),
      getRepurchase(queryParams.value),
    ])
    dashboard.value = dashboardRes.data || {}
    volume.value = volumeRes.data || volume.value
    classification.value = classificationRes.data || {}
    conversion.value = conversionRes.data || {}
    repurchase.value = repurchaseRes.data || { repurchaseBuckets: [] }
    await nextTick()
    renderCharts()
  } finally {
    loading.value = false
  }
}

const renderCharts = () => {
  renderVolumeChart()
  renderConversionChart()
  renderRepurchaseChart()
  renderClassificationChart()
}

const renderVolumeChart = () => {
  if (!volumeChartRef.value) return
  echarts.init(volumeChartRef.value).setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['订单量', '有效订单'] },
    xAxis: { type: 'category', data: volume.value.categories },
    yAxis: { type: 'value' },
    series: [
      { name: '订单量', type: 'line', smooth: true, data: volume.value.totalOrders },
      { name: '有效订单', type: 'bar', data: volume.value.validOrders },
    ],
  })
}

const renderConversionChart = () => {
  if (!conversionChartRef.value) return
  echarts.init(conversionChartRef.value).setOption({
    tooltip: { trigger: 'item' },
    series: [{
      type: 'funnel',
      data: [
        { name: '注册用户', value: conversion.value.registeredUsers || 0 },
        { name: '产生购买', value: conversion.value.purchasedUsers || 0 },
      ],
    }],
  })
}

const renderRepurchaseChart = () => {
  if (!repurchaseChartRef.value) return
  echarts.init(repurchaseChartRef.value).setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: (repurchase.value.repurchaseBuckets || []).map((item) => item.name) },
    yAxis: { type: 'value' },
    series: [{ type: 'bar', data: (repurchase.value.repurchaseBuckets || []).map((item) => item.count) }],
  })
}

const renderClassificationChart = () => {
  if (!classificationChartRef.value) return
  echarts.init(classificationChartRef.value).setOption({
    tooltip: { trigger: 'item' },
    series: [{
      type: 'pie',
      radius: ['45%', '70%'],
      data: [
        { name: '有效订单', value: classification.value.validOrders || 0 },
        { name: '无效订单', value: classification.value.invalidOrders || 0 },
        { name: '优惠券订单', value: classification.value.couponOrders || 0 },
      ],
    }],
  })
}

const handleExport = async () => {
  const res = await exportAnalytics()
  ElMessage.success(`导出任务已创建：${res.data?.taskId || '-'}`)
}

watch(activeTab, () => nextTick(renderCharts))
onMounted(loadData)
</script>

<style scoped>
.analytics-page {
  padding: 8px 0 24px;
}

.toolbar {
  margin-bottom: 12px;
}

.kpi-grid {
  margin-bottom: 12px;
}

.kpi-item {
  min-height: 86px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  padding: 16px;
  background: var(--el-bg-color);
}

.kpi-item span,
.kpi-item strong {
  display: block;
}

.kpi-item span {
  color: var(--el-text-color-secondary);
}

.kpi-item strong {
  margin-top: 10px;
  font-size: 24px;
}

.analysis-tabs {
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  padding: 12px;
  background: var(--el-bg-color);
}

.chart {
  width: 100%;
  height: 360px;
  margin-bottom: 12px;
}

.chart.compact {
  height: 420px;
}
</style>
