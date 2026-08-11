<template>
  <view class="page">
    <view class="head">
      <text class="title">订单详情</text>
    </view>

    <view v-if="loading" class="skeleton">
      <view class="skeleton-item big" />
      <view class="skeleton-item" />
    </view>
    <template v-else-if="order">
      <view class="status-card" :class="order.status">
        <text class="status-text">{{ statusText(order.status) }}</text>
        <text class="status-tip">{{ statusTip(order.status) }}</text>
      </view>

      <view class="map-card">
        <map
          v-if="hasMapData"
          class="route-map"
          :latitude="center.latitude"
          :longitude="center.longitude"
          :scale="mapScale"
          :markers="markers"
          :polyline="polylines"
          :include-points="includePoints"
          show-location
        />
        <view v-else class="map-empty">
          <text class="map-empty-title">暂无路线数据</text>
          <text class="map-empty-copy">订单起终点会在这里展示，接单后也会显示司机实时位置</text>
        </view>
      </view>

      <view class="route-card">
        <view class="route-row">
          <view class="dot start" />
          <text>{{ order.origin }}</text>
        </view>
        <view class="route-line" />
        <view class="route-row">
          <view class="dot end" />
          <text>{{ order.destination }}</text>
        </view>
      </view>

      <view class="info-card">
        <view class="info-row"><text class="info-label">订单号</text><text class="info-value">{{ order.orderNo || order.id }}</text></view>
        <view class="info-row"><text class="info-label">乘客</text><text class="info-value">{{ order.passengerName || order.passenger_name || '-' }}</text></view>
        <view class="info-row"><text class="info-label">座位</text><text class="info-value">{{ order.seats || 1 }}</text></view>
        <view class="info-row"><text class="info-label">金额</text><text class="info-value price">¥{{ order.amount ?? '-' }}</text></view>
        <view class="info-row"><text class="info-label">当前位置</text><text class="info-value">{{ currentPointText }}</text></view>
      </view>

      <view class="action-row">
        <u-button :plain="true" @click="goChat">联系乘客</u-button>
        <u-button :plain="true" :disabled="!passengerPhone" @click="callPassenger">电话</u-button>
        <u-button v-if="isActiveOrder" type="primary" @click="goLocation">去上报位置</u-button>
        <u-button v-if="order.status === 'accepted'" type="success" :loading="startingPickup" @click="startPickup">去接乘客</u-button>
        <u-button v-if="order.status === 'picking_up'" type="success" :loading="startingDelivery" @click="startDelivery">送乘客</u-button>
        <u-button v-if="order.status === 'delivering'" type="success" :loading="completing" @click="complete">完成订单</u-button>
      </view>

      <view class="section-title">实时轨迹</view>
      <u-steps :current="currentStep" direction="column">
        <u-steps-item
          v-for="(event, idx) in events"
          :key="`${event.text}-${idx}`"
          :title="event.text"
          :desc="event.time"
        />
      </u-steps>
    </template>
    <u-empty v-else text="订单不存在" />
  </view>
</template>

<script setup>
import { computed, onUnmounted, ref } from 'vue'
import { onLoad, onHide, onShow, onUnload } from '@dcloudio/uni-app'
import { completeOrder, getDriverOrderDetail, startDeliveryOrder, startPickupOrder } from '@/api/order'
import { getDriverLocationHistory } from '@/api/location'
import { getRoutePreview } from '@/api/map'

const orderId = ref('')
const order = ref(null)
const loading = ref(true)
const completing = ref(false)
const startingPickup = ref(false)
const startingDelivery = ref(false)
const historyPoints = ref([])
const routeData = ref(null)
const currentStep = ref(0)
const mapScale = ref(14)
const center = ref({ latitude: 31.2304, longitude: 121.4737 })
let timer = null

const statusMap = {
  pending: '待接单',
  accepted: '已接单',
  picking_up: '去接乘客',
  delivering: '送乘客',
  ongoing: '进行中',
  completed: '已完成',
  cancelled: '已取消',
}
const statusText = (s) => statusMap[s] || s || '未知'
const statusTip = (s) =>
  ({
    pending: '等待司机开始接乘客',
    accepted: '订单已接单，前往起点',
    picking_up: '司机正在前往乘客上车点',
    delivering: '乘客已上车，正在送达目的地',
    ongoing: '司机正在接送中',
    completed: '订单已完成',
    cancelled: '订单已取消',
  }[s] || '')

const isActiveOrder = computed(() => ['accepted', 'picking_up', 'delivering'].includes(order.value?.status))
const passengerPhone = computed(() => order.value?.passengerMobile || order.value?.passenger_mobile || order.value?.passengerPhone || order.value?.passenger_phone || '')
const currentPoint = computed(() => historyPoints.value[historyPoints.value.length - 1] || null)
const currentPointText = computed(() => {
  const point = currentPoint.value
  if (!point) return '暂无司机位置'
  const lat = point.latitude ?? point.lat ?? '-'
  const lng = point.longitude ?? point.lng ?? '-'
  return `${lat}, ${lng}`
})

const events = computed(() => {
  const list = []
  if (order.value?.origin) {
    list.push({ text: `起点：${order.value.origin}`, time: order.value.created_at || order.value.createdAt || '' })
  }
  historyPoints.value.slice(-10).forEach((point, index) => {
    list.push({
      text: `轨迹点 ${index + 1}：${point.latitude}, ${point.longitude}`,
      time: point.reportedAt || point.createdAt || '',
    })
  })
  if (order.value?.destination) {
    list.push({ text: `终点：${order.value.destination}`, time: '' })
  }
  return list
})

const routePoints = computed(() => {
  const points = routeData.value?.polyline || routeData.value?.points || []
  return points
    .map((point) => ({
      latitude: Number(point.latitude ?? point.lat ?? 0),
      longitude: Number(point.longitude ?? point.lng ?? 0),
    }))
    .filter((point) => point.latitude !== 0 || point.longitude !== 0)
})

const historyLinePoints = computed(() =>
  historyPoints.value
    .map((point) => ({
      latitude: Number(point.latitude ?? point.lat ?? 0),
      longitude: Number(point.longitude ?? point.lng ?? 0),
    }))
    .filter((point) => point.latitude !== 0 || point.longitude !== 0)
)

const markers = computed(() => {
  const result = []
  const route = routeData.value
  if (route?.origin?.latitude && route?.origin?.longitude) {
    result.push({
      id: 1,
      latitude: route.origin.latitude,
      longitude: route.origin.longitude,
      title: '起点',
      iconPath: '/static/tab/home.png',
      width: 28,
      height: 28,
    })
  }
  if (route?.destination?.latitude && route?.destination?.longitude) {
    result.push({
      id: 2,
      latitude: route.destination.latitude,
      longitude: route.destination.longitude,
      title: '终点',
      iconPath: '/static/tab/orders.png',
      width: 28,
      height: 28,
    })
  }
  if (currentPoint.value) {
    result.push({
      id: 3,
      latitude: Number(currentPoint.value.latitude ?? currentPoint.value.lat ?? 0),
      longitude: Number(currentPoint.value.longitude ?? currentPoint.value.lng ?? 0),
      title: '当前司机',
      iconPath: '/static/tab/location.png',
      width: 32,
      height: 32,
    })
  }
  return result
})

const polylines = computed(() => {
  const lines = []
  if (routePoints.value.length > 1) {
    lines.push({
      points: routePoints.value,
      color: '#CBD5E1',
      width: 6,
      dottedLine: true,
    })
  }
  if (historyLinePoints.value.length > 1) {
    lines.push({
      points: historyLinePoints.value,
      color: '#07C160',
      width: 6,
    })
  }
  return lines
})

const includePoints = computed(() => {
  const points = []
  markers.value.forEach((marker) => {
    points.push({ latitude: marker.latitude, longitude: marker.longitude })
  })
  routePoints.value.forEach((point) => points.push(point))
  historyLinePoints.value.forEach((point) => points.push(point))
  return points
})

const hasMapData = computed(() => markers.value.length > 0 || routePoints.value.length > 0 || historyLinePoints.value.length > 0)

const updateCenter = () => {
  const point = currentPoint.value || routeData.value?.origin || routeData.value?.destination
  if (!point) return
  center.value = {
    latitude: Number(point.latitude ?? point.lat ?? 0) || center.value.latitude,
    longitude: Number(point.longitude ?? point.lng ?? 0) || center.value.longitude,
  }
}

const normalizeTrackPoints = (items = []) =>
  items
    .map((item) => ({
      ...item,
      latitude: Number(item.latitude ?? item.lat ?? 0),
      longitude: Number(item.longitude ?? item.lng ?? 0),
    }))
    .filter((item) => item.latitude !== 0 || item.longitude !== 0)

const loadHistory = async () => {
  if (!orderId.value) return
  const res = await getDriverLocationHistory({ orderId: orderId.value, page: 1, pageSize: 100 })
  if (res?.code === 0) {
    const data = res.data || {}
    historyPoints.value = normalizeTrackPoints(data.points || data.list || [])
    updateCenter()
    currentStep.value = Math.max(0, events.value.length - 1)
  }
}

const load = async () => {
  if (!orderId.value) return
  loading.value = true
  routeData.value = null
  historyPoints.value = []
  try {
    const res = await getDriverOrderDetail(orderId.value)
    if (res?.code === 0) {
      order.value = res.data
      if (isActiveOrder.value) {
        uni.setStorageSync('driverActiveOrderId', orderId.value)
      } else if (String(uni.getStorageSync('driverActiveOrderId') || '') === orderId.value) {
        uni.removeStorageSync('driverActiveOrderId')
      }
      if (order.value?.origin && order.value?.destination) {
        const routeRes = await getRoutePreview({
          origin: order.value.origin,
          destination: order.value.destination,
          city: order.value.city || '',
          strategy: '0',
        })
        if (routeRes?.code === 0) {
          routeData.value = routeRes.data
        }
      }
      await loadHistory()
      currentStep.value = Math.max(0, events.value.length - 1)
    }
  } finally {
    loading.value = false
    if (isActiveOrder.value) {
      startPolling()
    } else {
      stopPolling()
    }
  }
}

const callPassenger = () => {
  const phone = passengerPhone.value
  if (phone) {
    uni.makePhoneCall({ phoneNumber: String(phone) })
  } else {
    uni.showToast({ title: '暂无乘客电话', icon: 'none' })
  }
}

const goChat = () => {
  if (!orderId.value) return
  uni.navigateTo({
    url: `/pages/orderChat/orderChat?orderId=${encodeURIComponent(orderId.value)}&mobile=${encodeURIComponent(passengerPhone.value || '')}`,
  })
}

const goLocation = () => {
  uni.setStorageSync('driverActiveOrderId', orderId.value)
  uni.switchTab({ url: '/pages/locationReport/locationReport' })
}

const actionPayload = (action) => ({ idempotency_key: `d-${orderId.value}-${action}-${Date.now()}` })

const startPickup = async () => {
  startingPickup.value = true
  const res = await startPickupOrder(orderId.value, actionPayload('pickup'))
  startingPickup.value = false
  if (res?.code === 0) {
    uni.showToast({ title: '已开始接乘客', icon: 'success' })
    await load()
  } else {
    uni.showToast({ title: res?.msg || '开始接乘客失败', icon: 'none' })
  }
}

const startDelivery = async () => {
  startingDelivery.value = true
  const res = await startDeliveryOrder(orderId.value, actionPayload('delivery'))
  startingDelivery.value = false
  if (res?.code === 0) {
    uni.showToast({ title: '已开始送乘客', icon: 'success' })
    await load()
  } else {
    uni.showToast({ title: res?.msg || '开始送乘客失败', icon: 'none' })
  }
}

const complete = () => {
  uni.showModal({
    title: '完成订单',
    content: '确认乘客已送达并完成该订单吗？',
    success: async (m) => {
      if (!m.confirm) return
      completing.value = true
      const res = await completeOrder(orderId.value, actionPayload('complete'))
      completing.value = false
      if (res?.code === 0) {
        uni.showToast({ title: '订单已完成', icon: 'success' })
        await load()
      } else {
        uni.showToast({ title: res?.msg || '完成订单失败', icon: 'none' })
      }
    },
  })
}

const startPolling = () => {
  stopPolling()
  if (!orderId.value || !isActiveOrder.value) return
  timer = setInterval(() => {
    loadHistory()
  }, 5000)
}

const stopPolling = () => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

onLoad((options) => {
  orderId.value = String(options.id || '')
  load()
})

onShow(() => {
  if (orderId.value) {
    startPolling()
  }
})

onHide(() => stopPolling())
onUnload(() => stopPolling())
onUnmounted(() => stopPolling())
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 24rpx 24rpx 40rpx;
  background: #f4f7fb;
}
.head {
  margin-bottom: 20rpx;
}
.title {
  font-size: 36rpx;
  font-weight: 700;
  color: #1f2937;
}
.status-card {
  border-radius: 24rpx;
  padding: 32rpx;
  background: #1677ff;
}
.status-card.completed {
  background: #07c160;
}
.status-card.cancelled {
  background: #98a2b3;
}
.status-text {
  display: block;
  font-size: 38rpx;
  font-weight: 700;
  color: #fff;
}
.status-tip {
  display: block;
  margin-top: 8rpx;
  font-size: 26rpx;
  color: rgba(255, 255, 255, 0.85);
}
.map-card {
  margin-top: 20rpx;
  overflow: hidden;
  border-radius: 24rpx;
  background: #fff;
  box-shadow: 0 8rpx 24rpx rgba(16, 24, 40, 0.06);
}
.route-map {
  width: 100%;
  height: 460rpx;
}
.map-empty {
  height: 460rpx;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 40rpx;
  text-align: center;
}
.map-empty-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #1f2937;
}
.map-empty-copy {
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #98a2b3;
  line-height: 1.6;
}
.route-card {
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-top: 20rpx;
  box-shadow: 0 8rpx 24rpx rgba(16, 24, 40, 0.06);
}
.route-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  font-size: 30rpx;
  font-weight: 600;
  color: #1f2937;
}
.route-line {
  margin-left: 7rpx;
  height: 56rpx;
  border-left: 2rpx dashed #cbd2dc;
}
.dot {
  width: 18rpx;
  height: 18rpx;
  border-radius: 50%;
}
.dot.start {
  background: #1677ff;
}
.dot.end {
  background: #ee0a24;
}
.info-card {
  background: #fff;
  border-radius: 24rpx;
  padding: 16rpx 32rpx;
  margin-top: 20rpx;
  box-shadow: 0 8rpx 24rpx rgba(16, 24, 40, 0.06);
}
.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 0;
}
.info-row + .info-row {
  border-top: 1rpx solid #f0f2f5;
}
.info-label {
  font-size: 26rpx;
  color: #8a93a6;
}
.info-value {
  font-size: 28rpx;
  font-weight: 600;
  color: #1f2937;
  text-align: right;
}
.info-value.price {
  color: #ee0a24;
}
.action-row {
  display: flex;
  gap: 20rpx;
  margin-top: 32rpx;
}
.section-title {
  margin: 28rpx 4rpx 16rpx;
  font-size: 28rpx;
  font-weight: 700;
  color: #1f2937;
}
.skeleton {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.skeleton-item {
  height: 200rpx;
  border-radius: 24rpx;
  background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%);
  background-size: 400% 100%;
  animation: shimmer 1.4s infinite;
}
.skeleton-item.big {
  height: 320rpx;
}
@keyframes shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: 0 0;
  }
}
</style>
