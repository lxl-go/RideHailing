<template>
  <view class="page">
    <view class="hero">
      <view class="hero-copy">
        <text class="eyebrow">司机出行智能体</text>
        <text class="title">AI预警</text>
        <text class="sub">
          {{ activeOrder ? '已识别进行中订单，正在提供订单护航建议' : '当前空闲待接单，分析当前位置周边 5 公里天气与路况' }}
        </text>
      </view>
      <view class="mode-badge" :class="{ order: activeOrder }">
        <text>{{ modeText }}</text>
      </view>
    </view>

    <view class="status-panel">
      <view class="status-row">
        <text class="label">当前状态</text>
        <text class="value">{{ activeOrder ? statusText(activeOrder.status) : '空闲待接单' }}</text>
      </view>
      <view v-if="activeOrder" class="route-box">
        <text class="route-point">{{ activeOrder.origin || '-' }}</text>
        <u-icon name="arrow-right" size="18" color="#9aa3b2" />
        <text class="route-point">{{ activeOrder.destination || '-' }}</text>
      </view>
      <view v-else class="status-row">
        <text class="label">预警范围</text>
        <text class="value">当前位置周边 4999 米</text>
      </view>
      <text v-if="locationFallback" class="hint">定位不可用，已使用演示坐标生成预警</text>
    </view>

    <u-button type="primary" :loading="loading" @click="generateWarning">
      {{ activeOrder ? '生成订单护航建议' : '生成周边预警' }}
    </u-button>

    <view v-if="advice" class="advice-panel" :class="advice.riskLevel">
      <view class="advice-head">
        <text class="advice-title">{{ activeOrder ? '订单护航建议' : '周边接单预警' }}</text>
        <u-tag :text="riskText(advice.riskLevel)" :type="riskTag(advice.riskLevel)" />
      </view>
      <text class="summary">{{ advice.displayText || advice.summary }}</text>

      <view class="advice-section">
        <text class="section-title">天气提醒</text>
        <text v-for="item in listOf(advice.weatherAdvice)" :key="item" class="advice-item">{{ item }}</text>
      </view>
      <view class="advice-section">
        <text class="section-title">{{ activeOrder ? '路线建议' : '路况提醒' }}</text>
        <text v-for="item in routeItems" :key="item" class="advice-item">{{ item }}</text>
      </view>
      <view class="advice-section">
        <text class="section-title">安全与接单建议</text>
        <text v-for="item in safetyItems" :key="item" class="advice-item">{{ item }}</text>
      </view>
    </view>

    <view v-else class="empty-panel">
      <u-icon name="bell" size="34" color="#1677ff" />
      <text>点击按钮后，AI 将根据订单或周边路况生成预警</text>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { getDriverAIWarning } from '@/api/ai'
import { listDriverOrders } from '@/api/order'

const ACTIVE_STATUSES = ['accepted', 'picking_up', 'pickup', 'in_progress', 'delivering']
const STATUS_QUERY = ['accepted', 'picking_up', 'delivering']
const DEFAULT_LOCATION = { latitude: 31.2304, longitude: 121.4737 }

const loading = ref(false)
const activeOrder = ref(null)
const advice = ref(null)
const locationFallback = ref(false)

const modeText = computed(() => (activeOrder.value ? '订单护航' : '周边预警'))
const routeItems = computed(() => {
  const traffic = listOf(advice.value?.trafficAdvice)
  const route = listOf(advice.value?.routeAdvice)
  return traffic.length ? traffic : route
})
const safetyItems = computed(() => {
  const dispatch = listOf(advice.value?.dispatchAdvice)
  const safety = listOf(advice.value?.safetyAdvice)
  return [...safety, ...dispatch].filter(Boolean)
})

const listOf = (value) => {
  if (Array.isArray(value)) return value.filter(Boolean)
  if (value) return [String(value)]
  return []
}

const riskText = (level) => ({ low: '低风险', medium: '中风险', high: '高风险' }[level] || '中风险')
const riskTag = (level) => ({ low: 'success', medium: 'warning', high: 'error' }[level] || 'warning')
const statusText = (status) => ({
  accepted: '已接单',
  picking_up: '去接乘客',
  pickup: '去接乘客',
  in_progress: '行程中',
  delivering: '送乘客'
}[status] || '进行中')

const firstOrder = (res) => {
  if (res?.code !== 0) return null
  return (res.data?.items || res.data?.list || res.data || [])[0] || null
}

const resolveActiveOrder = async () => {
  for (const status of STATUS_QUERY) {
    const res = await listDriverOrders({ status, page: 1, page_size: 1 })
    const order = firstOrder(res)
    if (order?.id) {
      activeOrder.value = order
      return
    }
  }
  const all = await listDriverOrders({ page: 1, page_size: 10 })
  const order = (all?.data?.items || all?.data?.list || []).find((item) => ACTIVE_STATUSES.includes(item?.status))
  activeOrder.value = order || null
}

const getDriverLocation = () =>
  new Promise((resolve) => {
    uni.getLocation({
      type: 'gcj02',
      success: (res) => {
        locationFallback.value = false
        resolve({ latitude: Number(res.latitude), longitude: Number(res.longitude) })
      },
      fail: () => {
        locationFallback.value = true
        resolve(DEFAULT_LOCATION)
      }
    })
  })

const advicePayload = async () => {
  const point = await getDriverLocation()
  if (!activeOrder.value) {
    return {
      mode: 'nearby',
      driverLat: point.latitude,
      driverLng: point.longitude,
      scene: 'idle_warning'
    }
  }
  return {
    mode: 'order',
    orderId: String(activeOrder.value.id),
    orderStatus: activeOrder.value.status,
    startAddress: activeOrder.value.origin || '',
    endAddress: activeOrder.value.destination || '',
    driverLat: point.latitude,
    driverLng: point.longitude,
    scene: 'before_departure'
  }
}

const generateWarning = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getDriverAIWarning(await advicePayload())
    if (res?.code === 0) {
      advice.value = res.data || null
    } else {
      advice.value = null
      uni.showToast({ title: res?.msg || 'AI预警生成失败', icon: 'none' })
    }
  } finally {
    loading.value = false
  }
}

const refresh = async () => {
  advice.value = null
  await resolveActiveOrder()
}

onShow(refresh)
onPullDownRefresh(async () => { await refresh(); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.hero { display: flex; justify-content: space-between; align-items: flex-start; gap: 20rpx; padding: 26rpx 8rpx 24rpx; }
.hero-copy { flex: 1; min-width: 0; }
.eyebrow { display: block; font-size: 22rpx; color: #1677ff; font-weight: 600; }
.title { display: block; margin-top: 8rpx; font-size: 42rpx; font-weight: 800; color: #172033; }
.sub { display: block; margin-top: 10rpx; font-size: 24rpx; color: #667085; line-height: 1.5; }
.mode-badge { flex: 0 0 auto; min-width: 132rpx; height: 56rpx; padding: 0 22rpx; border-radius: 10rpx; display: flex; align-items: center; justify-content: center; background: #52c41a; box-sizing: border-box; }
.mode-badge.order { background: #faad14; }
.mode-badge text { display: block; font-size: 24rpx; line-height: 1; font-weight: 700; color: #fff; white-space: nowrap; word-break: keep-all; }
.status-panel, .advice-panel, .empty-panel { margin-bottom: 24rpx; background: #fff; border-radius: 20rpx; padding: 26rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.status-row { display: flex; align-items: center; justify-content: space-between; gap: 20rpx; }
.label { font-size: 24rpx; color: #8a93a6; }
.value { font-size: 28rpx; color: #1f2937; font-weight: 700; }
.route-box { display: flex; align-items: center; gap: 14rpx; margin-top: 20rpx; padding: 22rpx; border-radius: 16rpx; background: #f8fafc; }
.route-point { flex: 1; font-size: 26rpx; color: #1f2937; font-weight: 600; line-height: 1.45; }
.hint { display: block; margin-top: 18rpx; font-size: 22rpx; color: #f59e0b; }
.advice-panel { margin-top: 24rpx; border-left: 8rpx solid #f59e0b; }
.advice-panel.low { border-left-color: #10b981; }
.advice-panel.high { border-left-color: #ef4444; }
.advice-head { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; }
.advice-title { font-size: 32rpx; font-weight: 800; color: #172033; }
.summary { display: block; margin-top: 18rpx; font-size: 28rpx; line-height: 1.6; color: #344054; }
.advice-section { margin-top: 24rpx; padding-top: 20rpx; border-top: 1rpx solid #eef2f7; }
.section-title { display: block; margin-bottom: 12rpx; font-size: 26rpx; font-weight: 700; color: #1f2937; }
.advice-item { display: block; position: relative; padding-left: 22rpx; margin-top: 10rpx; font-size: 25rpx; line-height: 1.55; color: #475467; }
.advice-item::before { content: ''; position: absolute; left: 0; top: 16rpx; width: 8rpx; height: 8rpx; border-radius: 50%; background: #1677ff; }
.empty-panel { display: flex; align-items: center; gap: 16rpx; margin-top: 24rpx; color: #667085; font-size: 25rpx; line-height: 1.5; }
</style>
