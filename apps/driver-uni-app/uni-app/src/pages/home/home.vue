<template>
  <view class="page">
    <view class="hero">
      <view class="hero-info">
        <text class="hero-title">司机工作台</text>
        <text class="hero-sub">{{ userStore.online ? '当前在线，可接单' : '当前离线，开启接单赚收益' }}</text>
      </view>
      <view class="online-switch">
        <text class="switch-label">{{ userStore.online ? '在线' : '离线' }}</text>
        <u-switch v-model="userStore.online" active-color="#07c160" inactive-color="#c0c4cc" @change="toggleOnline" />
      </view>
    </view>

    <view class="stats">
      <view class="stat-item">
        <text class="stat-num">{{ stats.todayOrders }}</text>
        <text class="stat-label">今日完成</text>
      </view>
      <view class="stat-item">
        <text class="stat-num">¥{{ stats.todayIncome }}</text>
        <text class="stat-label">今日收入</text>
      </view>
      <view class="stat-item">
        <text class="stat-num">{{ stats.pending }}</text>
        <text class="stat-label">待接单</text>
      </view>
    </view>

    <view class="grid">
      <view class="grid-item" @click="navigate('/pages/pendingOrders/pendingOrders')">
        <view class="grid-icon blue"><u-icon name="list" color="#fff" size="24" /></view>
        <text>待接订单</text>
      </view>
      <view class="grid-item" @click="navigate('/pages/publishTrip/publishTrip')">
        <view class="grid-icon green"><u-icon name="edit-pen" color="#fff" size="24" /></view>
        <text>发布行程</text>
      </view>
      <view class="grid-item" @click="navigate('/pages/vehicle/vehicle')">
        <view class="grid-icon orange"><u-icon name="car" color="#fff" size="24" /></view>
        <text>我的车辆</text>
      </view>
      <view class="grid-item" @click="navigate('/pages/certification/certification')">
        <view class="grid-icon purple"><u-icon name="account" color="#fff" size="24" /></view>
        <text>司机认证</text>
      </view>
      <view class="grid-item" @click="navigate('/pages/locationReport/locationReport')">
        <view class="grid-icon red"><u-icon name="map" color="#fff" size="24" /></view>
        <text>位置上报</text>
      </view>
      <view class="grid-item" @click="navigate('/pages/trackReplay/trackReplay')">
        <view class="grid-icon cyan"><u-icon name="reload" color="#fff" size="24" /></view>
        <text>轨迹回放</text>
      </view>
      <view class="grid-item" @click="navigate('/pages/aiAlerts/aiAlerts')">
        <view class="grid-icon indigo"><u-icon name="bell" color="#fff" size="24" /></view>
        <text>AI 预警</text>
      </view>
      <view class="grid-item" @click="navigate('/pages/incomeLedger/incomeLedger')">
        <view class="grid-icon gold"><u-icon name="rmb" color="#fff" size="24" /></view>
        <text>收入明细</text>
      </view>
    </view>

    <view class="section">
      <view class="section-head">
        <text class="section-title">最新订单</text>
        <text class="section-more" @click="navigate('/pages/pendingOrders/pendingOrders')">查看全部</text>
      </view>
      <view v-if="recentLoading" class="skeleton">
        <view v-for="n in 3" :key="n" class="skeleton-item" />
      </view>
      <u-empty v-else-if="recentOrders.length === 0" text="暂无待接订单，开启接单后刷新" />
      <view v-else class="order-list">
        <view v-for="order in recentOrders" :key="order.id" class="order-card" @click="goDetail(order.id)">
          <view class="order-head">
            <text class="order-no">订单 {{ order.orderNo || order.id }}</text>
            <text class="order-price">¥{{ order.amount ?? '-' }}</text>
          </view>
          <view class="order-route">
            <text>{{ order.origin }}</text>
            <u-icon name="arrow-right" size="18" color="#c0c4cc" />
            <text>{{ order.destination }}</text>
          </view>
          <view class="order-meta">
            <text>{{ formatTime(order.departTime) }}</text>
            <text>座位 {{ order.seats || 1 }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { onShow, onHide, onUnload, onPullDownRefresh } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { getDriverStats } from '@/api/driver'
import { listAvailableOrders } from '@/api/order'
import { navigate, goDetail as goDetailNav } from '@/utils/nav'
import { createAutoRefresh } from '@/utils/autoRefresh.mjs'

const userStore = useUserStore()
const stats = reactive({ todayOrders: 0, todayIncome: 0, pending: 0 })
const recentOrders = ref([])
const recentLoading = ref(false)

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(5, 16) : '-')

const loadStats = async () => {
  const res = await getDriverStats()
  if (res.code === 0) {
    stats.todayOrders = res.data?.today_orders ?? res.data?.todayOrders ?? res.data?.orders ?? 0
    stats.todayIncome = res.data?.today_income ?? res.data?.todayIncome ?? res.data?.earnings ?? 0
    stats.pending = res.data?.pending ?? 0
  }
}

const loadRecent = async (showLoading = true) => {
  if (showLoading) recentLoading.value = true
  const res = await listAvailableOrders({ page: 1, page_size: 3 })
  if (showLoading) recentLoading.value = false
  if (res?.code === 0) recentOrders.value = res.data?.items || res.data?.list || res.data || []
  else recentOrders.value = []
}

const toggleOnline = async (val) => {
  const res = await import('@/api/driver').then((m) => m.updateDriverStatus(val ? 2 : 1))
  if (res.code !== 0) userStore.setOnline(!val)
}

const goDetail = (id) => goDetailNav(`/pages/orderDetail/orderDetail?id=${id}`)

const loadDriverOnline = async () => {
  const res = await userStore.loadDriverInfo()
  if (res?.code === 0) {
    const st = res.data?.service_status
    userStore.setOnline(st === 2)
  }
}

const refreshHomeOrders = (showLoading = false) => Promise.all([loadStats(), loadRecent(showLoading)])
const autoRefresh = createAutoRefresh(() => refreshHomeOrders(false), { intervalMs: 5000, runImmediately: false })

onShow(() => { autoRefresh.start(); loadDriverOnline(); refreshHomeOrders(true) })
onHide(() => autoRefresh.stop())
onUnload(() => autoRefresh.stop())
onPullDownRefresh(async () => { await Promise.all([loadDriverOnline(), refreshHomeOrders(true)]); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 0 24rpx 40rpx; background: #f4f7fb; }
.hero { display: flex; align-items: center; justify-content: space-between; padding: 32rpx 8rpx 28rpx; }
.hero-title { font-size: 38rpx; font-weight: 700; color: #172033; }
.hero-sub { display: block; margin-top: 8rpx; font-size: 24rpx; color: #8a93a6; }
.online-switch { display: flex; align-items: center; gap: 12rpx; }
.switch-label { font-size: 24rpx; color: #475467; }
.stats { display: flex; gap: 16rpx; }
.stat-item { flex: 1; background: #fff; border-radius: 20rpx; padding: 28rpx 0; text-align: center; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.stat-num { display: block; font-size: 38rpx; font-weight: 700; color: #1f2937; }
.stat-label { display: block; margin-top: 6rpx; font-size: 22rpx; color: #8a93a6; }
.grid { display: flex; flex-wrap: wrap; margin-top: 24rpx; background: #fff; border-radius: 24rpx; padding: 24rpx 16rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.grid-item { width: 25%; display: flex; flex-direction: column; align-items: center; gap: 12rpx; padding: 16rpx 0; }
.grid-item text { font-size: 22rpx; color: #475467; }
.grid-icon { width: 80rpx; height: 80rpx; border-radius: 22rpx; display: flex; align-items: center; justify-content: center; }
.grid-icon.blue { background: #1677ff; }
.grid-icon.green { background: #07c160; }
.grid-icon.orange { background: #f5a623; }
.grid-icon.purple { background: #7c5cff; }
.grid-icon.red { background: #f5222d; }
.grid-icon.cyan { background: #13c2c2; }
.grid-icon.indigo { background: #2f54eb; }
.grid-icon.gold { background: #fa8c16; }
.section { margin-top: 24rpx; background: #fff; border-radius: 24rpx; padding: 28rpx 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.section-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20rpx; }
.section-title { font-size: 30rpx; font-weight: 700; color: #1f2937; }
.section-more { font-size: 24rpx; color: #1677ff; }
.order-list { display: flex; flex-direction: column; gap: 20rpx; }
.order-card { background: #f8fafc; border-radius: 20rpx; padding: 24rpx; }
.order-head { display: flex; align-items: center; justify-content: space-between; }
.order-no { font-size: 24rpx; color: #8a93a6; }
.order-price { font-size: 30rpx; font-weight: 700; color: #ee0a24; }
.order-route { display: flex; align-items: center; gap: 12rpx; margin-top: 16rpx; font-size: 28rpx; font-weight: 600; color: #1f2937; }
.order-meta { display: flex; gap: 24rpx; margin-top: 14rpx; font-size: 24rpx; color: #8a93a6; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; }
.skeleton-item { height: 160rpx; border-radius: 20rpx; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
