<template>
  <view class="page">
    <view class="search-bar">
      <u-input v-model="orderIdInput" type="text" placeholder="请输入订单号" border />
      <u-button type="primary" :plain="true" :loading="loading" @click="queryTrack">查询</u-button>
    </view>

    <view class="status-card">
      <view class="status-row">
        <text class="status-label">{{ routeStageText }}</text>
        <u-tag :text="connected ? '已连接' : '等待数据'" :type="connected ? 'success' : 'warning'" />
      </view>
      <text class="status-copy">{{ statusText }}</text>
      <text v-if="routeSummaryText" class="route-summary">{{ routeSummaryText }}</text>
    </view>

    <view class="map-card">
      <map
        v-if="hasMapData"
        class="track-map"
        :latitude="center.latitude"
        :longitude="center.longitude"
        :scale="mapScale"
        :markers="markers"
        :polyline="polylines"
        :include-points="mapIncludePoints"
        show-location
      />
      <view v-else class="map-empty">
        <text class="map-empty-title">暂无轨迹</text>
        <text class="map-empty-copy">请输入订单号后，系统会拉取订单、路线和司机实时位置</text>
      </view>
    </view>

    <view class="summary-card" v-if="order">
      <view class="summary-row">
        <text class="summary-label">起点</text>
        <text class="summary-value">{{ order.origin || '-' }}</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">终点</text>
        <text class="summary-value">{{ order.destination || '-' }}</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">当前</text>
        <text class="summary-value">{{ currentPointText }}</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">路线</text>
        <text class="summary-value">{{ routeStageText }}</text>
      </view>
    </view>

    <view class="contact-row" v-if="canContact">
      <u-button type="primary" @click="goChat">联系司机</u-button>
      <u-button :plain="true" :disabled="!driverPhone" @click="callDriver">电话</u-button>
    </view>

    <view class="section-title">轨迹节点</view>
    <u-steps :current="currentStep" direction="column">
      <u-steps-item
        v-for="(event, idx) in events"
        :key="`${event.text}-${idx}`"
        :title="event.text"
        :desc="event.time"
      />
    </u-steps>

    <u-empty v-if="!loading && events.length === 0" text="暂无轨迹数据" />
  </view>
</template>

<script setup>
import { computed, onUnmounted, ref } from 'vue'
import { onHide, onLoad, onPullDownRefresh, onShow, onUnload } from '@dcloudio/uni-app'
import { getOrderDetail, getOrderTrack } from '@/api/order'
import { getRoutePreview } from '@/api/map'

const TRACK_ORDER_STORAGE_KEY = 'passenger-track-order-id'
const orderIdInput = ref('')
const orderId = ref('')
const driverId = ref('')
const order = ref(null)
const loading = ref(false)
const connected = ref(false)
const currentStep = ref(0)
const trackPoints = ref([])
const routeData = ref(null)
const statusText = ref('请输入订单号后查看真实行程轨迹')
const mapScale = ref(15)
const center = ref({ latitude: 31.2304, longitude: 121.4737 })
let timer = null

const currentPoint = computed(() => trackPoints.value[trackPoints.value.length - 1] || null)
const isDeliveryStage = computed(() => ['delivering', 'completed'].includes(order.value?.status))
const formatPointForRoute = (point) => {
  const latitude = Number(point?.latitude ?? point?.lat ?? 0)
  const longitude = Number(point?.longitude ?? point?.lng ?? 0)
  if (!latitude || !longitude) return ''
  return `${longitude},${latitude}`
}
const routeOriginText = computed(() => {
  if (isDeliveryStage.value) return order.value?.origin || ''
  return formatPointForRoute(currentPoint.value) || order.value?.origin || ''
})
const routeDestinationText = computed(() => {
  if (isDeliveryStage.value) return order.value?.destination || ''
  return currentPoint.value ? order.value?.origin || '' : order.value?.destination || ''
})
const routeStageText = computed(() => (isDeliveryStage.value ? '前往目的地' : '司机来接您'))
const currentPointText = computed(() => {
  const point = currentPoint.value
  if (!point) return '暂无当前位置'
  const lat = point.latitude ?? point.lat ?? '-'
  const lng = point.longitude ?? point.lng ?? '-'
  return `${lat}, ${lng}`
})
const formatDistance = (meters) => {
  const value = Number(meters || 0)
  if (!value) return ''
  if (value < 1000) return `${Math.round(value)}米`
  return `${(value / 1000).toFixed(1)}公里`
}
const formatDuration = (seconds) => {
  const value = Number(seconds || 0)
  if (!value) return ''
  const minutes = Math.max(1, Math.round(value / 60))
  if (minutes < 60) return `${minutes}分钟`
  const hours = Math.floor(minutes / 60)
  const remain = minutes % 60
  return remain ? `${hours}小时${remain}分钟` : `${hours}小时`
}
const routeSummaryText = computed(() => {
  if (!routeData.value) return ''
  const distance = formatDistance(routeData.value.distanceMeters ?? routeData.value.distance_meters)
  const duration = formatDuration(routeData.value.durationSeconds ?? routeData.value.duration_seconds)
  if (!distance && !duration) return ''
  return `${routeStageText.value}：${distance || '-'}，预计${duration || '-'}`
})

const normalizeMapPoint = (point) => {
  const latitude = Number(point?.latitude ?? point?.lat)
  const longitude = Number(point?.longitude ?? point?.lng)
  const valid =
    Number.isFinite(latitude) &&
    Number.isFinite(longitude) &&
    latitude >= -90 &&
    latitude <= 90 &&
    longitude >= -180 &&
    longitude <= 180
  if (!valid) {
    return null
  }
  return { latitude, longitude }
}

const events = computed(() => {
  const list = []
  if (order.value?.origin) {
    list.push({ text: `起点：${order.value.origin}`, time: order.value.created_at || order.value.createdAt || '' })
  }
  trackPoints.value.slice(-10).forEach((point, index) => {
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
    .map(normalizeMapPoint)
    .filter(Boolean)
})

const trackLinePoints = computed(() =>
  trackPoints.value
    .map(normalizeMapPoint)
    .filter(Boolean)
)

const markers = computed(() => {
  const result = []
  const route = routeData.value
  const origin = normalizeMapPoint(route?.origin)
  const destination = normalizeMapPoint(route?.destination)
  const current = normalizeMapPoint(currentPoint.value)
  if (origin) {
    result.push({
      id: 1,
      latitude: origin.latitude,
      longitude: origin.longitude,
      title: '起点',
      iconPath: '/static/tab/home.png',
      width: 28,
      height: 28,
    })
  }
  if (destination) {
    result.push({
      id: 2,
      latitude: destination.latitude,
      longitude: destination.longitude,
      title: '终点',
      iconPath: '/static/tab/orders.png',
      width: 28,
      height: 28,
    })
  }
  if (current) {
    result.push({
      id: 3,
      latitude: current.latitude,
      longitude: current.longitude,
      title: '当前车辆',
      iconPath: '/static/tab/tracking.png',
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
  if (trackLinePoints.value.length > 1) {
    lines.push({
      points: trackLinePoints.value,
      color: '#1677FF',
      width: 6,
    })
  }
  return lines
})

const mapIncludePoints = computed(() => {
  const points = []
  markers.value.forEach((marker) => {
    const point = normalizeMapPoint(marker)
    if (point) points.push(point)
  })
  routePoints.value.forEach((point) => points.push(point))
  trackLinePoints.value.forEach((point) => points.push(point))
  return points
})

const hasMapData = computed(() => markers.value.length > 0 || routePoints.value.length > 0 || trackLinePoints.value.length > 0)
const canContact = computed(() => ['accepted', 'picking_up', 'delivering'].includes(order.value?.status))
const driverPhone = computed(() => order.value?.driverMobile || order.value?.driver_mobile || order.value?.driverPhone || order.value?.driver_phone || '')

const updateCenter = () => {
  const point = normalizeMapPoint(currentPoint.value) || normalizeMapPoint(routeData.value?.origin) || normalizeMapPoint(routeData.value?.destination)
  if (!point) return
  center.value = {
    latitude: point.latitude,
    longitude: point.longitude,
  }
}

const normalizeTrackPoints = (items = []) =>
  items
    .map((item) => ({
      ...item,
      ...normalizeMapPoint(item),
    }))
    .filter((item) => normalizeMapPoint(item))

const loadTrack = async () => {
  if (!orderId.value) return
  loading.value = true
  routeData.value = null
  trackPoints.value = []
  statusText.value = '正在拉取订单、路线和实时轨迹'
  try {
    const orderResult = await getOrderDetail(orderId.value)
    if (orderResult.status === 'fulfilled' && orderResult.value?.code === 0) {
      order.value = orderResult.value.data
    } else if (orderResult?.code === 0) {
      order.value = orderResult.data
    }

    driverId.value = String(order.value?.driver_id || order.value?.driverId || '')
    const trackResult = await getOrderTrack(orderId.value, { driverId: driverId.value })
    if (trackResult?.code === 0) {
      const data = trackResult.data || {}
      trackPoints.value = normalizeTrackPoints(data.points || data.list || [])
    } else {
      trackPoints.value = []
    }

    if (routeOriginText.value && routeDestinationText.value) {
      const routeResult = await getRoutePreview({
        origin: routeOriginText.value,
        destination: routeDestinationText.value,
        city: order.value.city || '',
        strategy: '0',
      })
      if (routeResult?.code === 0) {
        routeData.value = routeResult.data
      }
    }

    connected.value = trackPoints.value.length > 0 || !!routeData.value
    statusText.value = routeData.value
      ? `${routeStageText.value}路线已更新`
      : (trackPoints.value.length > 0 ? '已获取司机实时轨迹' : '司机暂未上报位置')
    updateCenter()
    currentStep.value = Math.max(0, events.value.length - 1)
  } catch (error) {
    connected.value = false
    statusText.value = error?.msg || error?.message || '轨迹接口暂不可用'
  } finally {
    loading.value = false
    uni.stopPullDownRefresh?.()
  }
}

const startPolling = () => {
  stopPolling()
  if (!orderId.value) return
  timer = setInterval(() => {
    loadTrack()
  }, 5000)
}

const stopPolling = () => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

const queryTrack = () => {
  if (!orderIdInput.value) {
    uni.showToast({ title: '请输入订单号', icon: 'none' })
    return
  }
  orderId.value = String(orderIdInput.value).trim()
  uni.setStorageSync(TRACK_ORDER_STORAGE_KEY, orderId.value)
  loadTrack()
  startPolling()
}

const goChat = () => {
  if (!orderId.value) return
  uni.navigateTo({
    url: `/pages/orderChat/orderChat?orderId=${encodeURIComponent(orderId.value)}&mobile=${encodeURIComponent(driverPhone.value || '')}`,
  })
}

const callDriver = () => {
  if (!driverPhone.value) {
    uni.showToast({ title: '暂无司机电话', icon: 'none' })
    return
  }
  uni.makePhoneCall({ phoneNumber: String(driverPhone.value) })
}

onLoad((query) => {
  if (query?.orderId) {
    orderId.value = String(query.orderId)
    orderIdInput.value = String(query.orderId)
    uni.setStorageSync(TRACK_ORDER_STORAGE_KEY, orderId.value)
    loadTrack()
    startPolling()
  }
})

onShow(() => {
  const storedOrderId = String(uni.getStorageSync(TRACK_ORDER_STORAGE_KEY) || '')
  if (storedOrderId && storedOrderId !== orderId.value) {
    orderId.value = storedOrderId
    orderIdInput.value = storedOrderId
    loadTrack()
  }
  if (orderId.value) {
    startPolling()
  }
})

onHide(() => stopPolling())
onUnload(() => stopPolling())
onPullDownRefresh(() => loadTrack())
onUnmounted(() => stopPolling())
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 24rpx;
  background: #f4f7fb;
}
.search-bar {
  display: flex;
  gap: 16rpx;
  align-items: center;
}
.status-card {
  margin-top: 20rpx;
  padding: 24rpx;
  border-radius: 20rpx;
  background: #fff;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.05);
}
.status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}
.status-label {
  font-size: 28rpx;
  font-weight: 700;
  color: #1f2937;
}
.status-copy {
  display: block;
  margin-top: 12rpx;
  font-size: 24rpx;
  color: #667085;
}
.route-summary {
  display: block;
  margin-top: 12rpx;
  font-size: 28rpx;
  font-weight: 700;
  color: #0f9f6e;
}
.map-card {
  margin-top: 20rpx;
  overflow: hidden;
  border-radius: 20rpx;
  background: #fff;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.05);
}
.track-map {
  width: 100%;
  height: 520rpx;
}
.map-empty {
  height: 520rpx;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  padding: 40rpx;
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
.summary-card {
  margin-top: 20rpx;
  padding: 8rpx 24rpx;
  border-radius: 20rpx;
  background: #fff;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.05);
}
.summary-row {
  display: flex;
  justify-content: space-between;
  gap: 20rpx;
  padding: 22rpx 0;
}
.summary-row + .summary-row {
  border-top: 1rpx solid #f0f2f5;
}
.summary-label {
  min-width: 90rpx;
  font-size: 26rpx;
  color: #8a93a6;
}
.summary-value {
  flex: 1;
  font-size: 26rpx;
  color: #1f2937;
  text-align: right;
}
.section-title {
  margin: 28rpx 8rpx 16rpx;
  font-size: 28rpx;
  font-weight: 700;
  color: #1f2937;
}
.contact-row {
  display: flex;
  gap: 20rpx;
  margin-top: 24rpx;
}
</style>
