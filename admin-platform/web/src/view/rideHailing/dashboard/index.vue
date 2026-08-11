<template>
  <div class="ride-dashboard">
    <RidePageHeader title="运营概览" subtitle="实时掌握订单、司机、乘客与营收概况" />
    <el-row :gutter="16" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.totalOrders }}</div>
          <div class="stat-label">总订单数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.activeDrivers }}</div>
          <div class="stat-label">活跃司机</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.totalPassengers }}</div>
          <div class="stat-label">注册乘客</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card revenue">
          <div class="stat-value">￥{{ formatMoney(stats.revenue) }}</div>
          <div class="stat-label">今日营收</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :span="16">
        <el-card class="recent-card" v-loading="loading">
          <template #header>
            <div class="card-header">
              <span>最近订单</span>
              <el-button link type="primary" @click="goOrders">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentOrders" style="width: 100%">
            <el-table-column prop="orderNo" label="订单号" width="170" />
            <el-table-column prop="passengerName" label="乘客" width="100" />
            <el-table-column prop="driverName" label="司机" width="100" />
            <el-table-column prop="routeName" label="路线" min-width="160" />
            <el-table-column label="金额" width="100">
              <template #default="{ row }">￥{{ row.payAmount ?? row.amount ?? '-' }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="quick-card">
          <template #header><span>运营入口</span></template>
          <div class="quick-grid">
            <div class="quick-item" @click="go('/ride-hailing/audit')">
              <el-icon class="quick-icon blue"><CircleCheck /></el-icon>
              <span>认证审核</span>
            </div>
            <div class="quick-item" @click="go('/ride-hailing/finance')">
              <el-icon class="quick-icon green"><Money /></el-icon>
              <span>财务管理</span>
            </div>
            <div class="quick-item" @click="go('/ride-hailing/analytics')">
              <el-icon class="quick-icon orange"><DataAnalysis /></el-icon>
              <span>数据分析</span>
            </div>
            <div class="quick-item" @click="go('/ride-hailing/dispatch')">
              <el-icon class="quick-icon purple"><Guide /></el-icon>
              <span>派单中心</span>
            </div>
            <div class="quick-item" @click="go('/ride-hailing/marketing')">
              <el-icon class="quick-icon red"><Present /></el-icon>
              <span>营销管理</span>
            </div>
            <div class="quick-item" @click="go('/ride-hailing/ai')">
              <el-icon class="quick-icon cyan"><ChatDotRound /></el-icon>
              <span>AI 助手</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { CircleCheck, Money, DataAnalysis, Guide, Present, ChatDotRound } from '@element-plus/icons-vue'
import { getDashboardOverview } from '@/api/rideHailing/dashboard'
import { listOrders } from '@/api/rideHailing/rideOrders'
import RidePageHeader from '@/components/RidePageHeader/index.vue'

defineOptions({ name: 'RideHailingDashboard' })

const router = useRouter()
const loading = ref(false)
const stats = reactive({ totalOrders: 0, activeDrivers: 0, totalPassengers: 0, revenue: 0 })
const recentOrders = ref([])

const statusOptions = [
  { label: '待支付', value: 'pending' },
  { label: '已支付', value: 'paid' },
  { label: '进行中', value: 'ongoing' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' }
]
const statusText = (value) => (statusOptions.find(item => item.value === value)?.label || value || '-')
const statusType = (value) => ({ paid: 'success', ongoing: 'warning', completed: 'info', cancelled: 'danger' }[value] || '')
const formatMoney = (v) => Number(v || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

const loadOverview = async () => {
  const res = await getDashboardOverview()
  if (res.code === 0) {
    stats.totalOrders = res.data.totalOrders ?? res.data.total_orders ?? res.data.todayOrderCount ?? res.data.monthOrderCount ?? 0
    stats.activeDrivers = res.data.activeDrivers ?? res.data.active_drivers ?? 0
    stats.totalPassengers = res.data.totalPassengers ?? res.data.total_passengers ?? res.data.activePassengers ?? 0
    stats.revenue = res.data.revenue ?? res.data.todayRevenue ?? res.data.monthRevenue ?? 0
  }
}

const loadRecent = async () => {
  loading.value = true
  try {
    const res = await listOrders({ page: 1, pageSize: 6 })
    if (res.code === 0) recentOrders.value = res.data.list || []
  } finally {
    loading.value = false
  }
}

const go = (path) => router.push(path)
const goOrders = () => router.push('/ride-hailing/orders')

loadOverview()
loadRecent()
</script>

<style scoped>
.ride-dashboard { padding: 16px; }
.stat-row { margin-bottom: 16px; }
.stat-card { text-align: center; }
.stat-card.revenue .stat-value { color: #ee0a24; }
.stat-value { font-size: 28px; font-weight: 700; color: #409eff; }
.stat-label { margin-top: 8px; font-size: 14px; color: #909399; }
.recent-card { margin-top: 8px; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.quick-card { margin-top: 8px; }
.quick-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.quick-item { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 16px 0; cursor: pointer; border-radius: 8px; transition: background 0.2s; }
.quick-item:hover { background: #f4f7fb; }
.quick-item span { font-size: 13px; color: #475467; }
.quick-icon { font-size: 28px; }
.quick-icon.blue { color: #1677ff; }
.quick-icon.green { color: #07c160; }
.quick-icon.orange { color: #f5a623; }
.quick-icon.purple { color: #7c5cff; }
.quick-icon.red { color: #f5222d; }
.quick-icon.cyan { color: #13c2c2; }
</style>
