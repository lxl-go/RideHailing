<template>
  <view class="page">
    <view class="status-card">
      <text class="status-title">位置上报</text>
      <text class="status-desc">{{ reportText }}</text>
      <text v-if="activeOrderId" class="status-extra">{{ stageText }}</text>
      <text v-if="estimateText" class="status-extra">{{ estimateText }}</text>
    </view>

    <view class="map-card">
      <map
        v-if="latitude !== null && longitude !== null"
        class="report-map"
        :latitude="latitude"
        :longitude="longitude"
        :scale="mapScale"
        :markers="markers"
        :polyline="routePolyline"
        show-location
      />
      <view v-else class="map-empty">
        <text class="map-empty-title">等待定位</text>
        <text class="map-empty-copy">授权后会显示司机当前位置和实时上报状态</text>
      </view>
    </view>

    <view class="loc-card">
      <view class="loc-row">
        <text class="loc-label">订单</text>
        <text class="loc-val">{{ activeOrderId || '暂无进行中订单' }}</text>
      </view>
      <view class="loc-row">
        <text class="loc-label">经度</text>
        <text class="loc-val">{{ longitude ?? '-' }}</text>
      </view>
      <view class="loc-row">
        <text class="loc-label">纬度</text>
        <text class="loc-val">{{ latitude ?? '-' }}</text>
      </view>
      <view class="loc-row">
        <text class="loc-label">地址</text>
        <text class="loc-val">{{ address || '等待定位' }}</text>
      </view>
      <view class="loc-row">
        <text class="loc-label">上报状态</text>
        <u-tag :text="autoReport ? '自动上报中' : '等待接单'" :type="autoReport ? 'success' : 'warning'" size="mini" />
      </view>
    </view>

    <view class="ctrl">
      <u-button type="primary" :disabled="!activeOrderId" @click="goChat">联系乘客</u-button>
      <u-button :plain="true" :disabled="!passengerPhone" @click="callPassenger">电话联系</u-button>
    </view>

    <view class="tip">接单后自动开启位置上报，每 15 秒同步一次当前位置。系统会按订单阶段显示去接乘客或去往目的地的预计路程。</view>
  </view>
</template>

<script setup>
import { computed, onUnmounted, ref } from 'vue'
import { onHide, onShow, onUnload } from '@dcloudio/uni-app'
import { getDriverOrderDetail, listDriverOrders } from '@/api/order'
import { getDriverLocationHistory, reportDriverLocation } from '@/api/location'
import { getRoutePreview, reverseGeocode } from '@/api/map'

const longitude = ref(null)
const latitude = ref(null)
const address = ref('')
const reporting = ref(false)
const autoReport = ref(false)
const reportText = ref('请先授权定位权限')
const mapScale = ref(16)
const activeOrderId = ref('')
const activeOrder = ref(null)
const routeEstimate = ref(null)
const routePolyline = ref([])
let timer = null

const isActiveOrder = computed(() => ['accepted', 'picking_up', 'delivering'].includes(activeOrder.value?.status))
const passengerPhone = computed(() => activeOrder.value?.passengerMobile || activeOrder.value?.passenger_mobile || activeOrder.value?.passengerPhone || activeOrder.value?.passenger_phone || '')

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

const markers = computed(() => {
  const current = normalizeMapPoint({ latitude: latitude.value, longitude: longitude.value })
  if (!current) return []
  const items = [
    {
      id: 1,
      latitude: current.latitude,
      longitude: current.longitude,
      title: '司机当前位置',
      iconPath: '/static/tab/location.png',
      width: 32,
      height: 32,
    },
  ]
  const destination = normalizeMapPoint(routeEstimate.value?.destination)
  if (destination) {
    items.push({
      id: 2,
      latitude: destination.latitude,
      longitude: destination.longitude,
      title: targetAddress() || '目的位置',
      iconPath: '/static/tab/orders.png',
      width: 28,
      height: 28,
    })
  }
  return items
})

const stageText = computed(() => {
  const status = activeOrder.value?.status
  if (status === 'accepted' || status === 'picking_up') return '当前阶段：去接乘客'
  if (status === 'delivering') return '当前阶段：去往目的地'
  if (activeOrderId.value) return '当前订单暂未进入位置上报阶段'
  return ''
})

const estimateText = computed(() => {
  if (!routeEstimate.value) return ''
  const distance = Number(routeEstimate.value.distance_meters ?? routeEstimate.value.distanceMeters ?? routeEstimate.value.distance ?? 0)
  const duration = Number(routeEstimate.value.duration_seconds ?? routeEstimate.value.durationSeconds ?? routeEstimate.value.duration ?? 0)
  const distanceText = distance > 0 ? `${(distance / 1000).toFixed(1)} 公里` : ''
  const durationText = duration > 0 ? `${Math.max(1, Math.round(duration / 60))} 分钟` : ''
  if (!distanceText && !durationText) return ''
  return `路线预计：${[distanceText, durationText].filter(Boolean).join('，')}`
})

const getCurrentOrderId = () => {
  const stored = uni.getStorageSync('driverActiveOrderId')
  activeOrderId.value = stored ? String(stored) : ''
  if (!activeOrderId.value) {
    activeOrder.value = null
    routeEstimate.value = null
    routePolyline.value = []
    reportText.value = '请先从进行中的订单进入位置上报'
  }
}

const cacheActiveOrder = (order) => {
  if (!order?.id) return false
  activeOrder.value = order
  activeOrderId.value = String(order.id)
  uni.setStorageSync('driverActiveOrderId', activeOrderId.value)
  return true
}

const resolveActiveOrder = async () => {
  getCurrentOrderId()
  if (activeOrderId.value) return true
  const accepted = await listDriverOrders({ status: 'accepted', page: 1, page_size: 1 })
  if (accepted?.code === 0 && cacheActiveOrder((accepted.data?.items || accepted.data?.list || [])[0])) return true
  const pickingUp = await listDriverOrders({ status: 'picking_up', page: 1, page_size: 1 })
  if (pickingUp?.code === 0 && cacheActiveOrder((pickingUp.data?.items || pickingUp.data?.list || [])[0])) return true
  const delivering = await listDriverOrders({ status: 'delivering', page: 1, page_size: 1 })
  if (delivering?.code === 0 && cacheActiveOrder((delivering.data?.items || delivering.data?.list || [])[0])) return true
  reportText.value = '暂无进行中订单，接单后自动开启位置上报'
  return false
}

const loadActiveOrder = async () => {
  if (!activeOrderId.value) return
  const res = await getDriverOrderDetail(activeOrderId.value)
  if (res?.code === 0) {
    activeOrder.value = res.data
    await refreshEstimate()
  }
}

const targetAddress = () => {
  const status = activeOrder.value?.status
  if (status === 'accepted' || status === 'picking_up') return activeOrder.value?.origin
  if (status === 'delivering') return activeOrder.value?.destination
  return ''
}

const refreshEstimate = async () => {
  routeEstimate.value = null
  routePolyline.value = []
  const target = targetAddress()
  const current = normalizeMapPoint({ latitude: latitude.value, longitude: longitude.value })
  if (!target || !current) return
  const origin = address.value || `${latitude.value},${longitude.value}`
  const res = await getRoutePreview({
    origin,
    destination: target,
    city: activeOrder.value?.city || '',
    strategy: '0',
  }).catch(() => null)
  if (res?.code === 0) {
    routeEstimate.value = res.data
    routePolyline.value = buildRoutePolyline(res.data)
  }
}

const buildRoutePolyline = (route = {}) => {
  const points = route.points || route.polyline || []
  if (!Array.isArray(points) || points.length < 2) return []
  const normalizedPoints = points.map(normalizeMapPoint).filter(Boolean)
  if (normalizedPoints.length < 2) return []
  return [{
    points: normalizedPoints,
    color: '#07c160',
    width: 6,
    dottedLine: false,
    arrowLine: true,
  }]
}

const getLoc = () =>
  new Promise((resolve) => {
    uni.getLocation({
      type: 'gcj02',
      success: async (res) => {
        const point = normalizeMapPoint(res)
        if (!point) {
          reportText.value = '定位坐标异常，请稍后重试'
          resolve(false)
          return
        }
        longitude.value = point.longitude
        latitude.value = point.latitude
        const geoRes = await reverseGeocode({ lat: point.latitude, lng: point.longitude }).catch(() => null)
        address.value = geoRes?.code === 0 ? geoRes.data?.formattedAddress || geoRes.data?.formatted_address || '' : ''
        reportText.value = address.value ? `当前位置：${address.value}` : '已获取当前位置'
        await refreshEstimate()
        resolve(true)
      },
      fail: () => {
        reportText.value = '定位失败，请检查权限'
        resolve(false)
      },
    })
  })

const reportOnce = async (silent = false) => {
  await resolveActiveOrder()
  if (!activeOrderId.value) {
    stopAuto()
    if (!silent) uni.showToast({ title: '请先选择进行中的订单', icon: 'none' })
    return
  }
  if (!activeOrder.value) await loadActiveOrder()
  const ok = await getLoc()
  const point = normalizeMapPoint({ latitude: latitude.value, longitude: longitude.value })
  if (!ok || !point) {
    if (!silent) uni.showToast({ title: '获取定位失败', icon: 'none' })
    return
  }
  reporting.value = true
  const res = await reportDriverLocation({
    orderId: activeOrderId.value,
    latitude: latitude.value,
    longitude: longitude.value,
    lat: latitude.value,
    lng: longitude.value,
    address: address.value,
    timestamp: Date.now(),
  })
  reporting.value = false
  if (res?.code === 0) {
    reportText.value = address.value ? `已上报：${address.value}` : '位置已上报'
    if (!silent) uni.showToast({ title: '已上报', icon: 'success' })
  } else {
    reportText.value = res?.msg || '上报失败'
    if (!silent) uni.showToast({ title: res?.msg || '上报失败', icon: 'none' })
  }
}

const startAuto = () => {
  if (autoReport.value) return
  if (!activeOrderId.value) {
    return
  }
  if (!isActiveOrder.value) return
  autoReport.value = true
  reportText.value = '自动上报中，每 15 秒同步一次'
  timer = setInterval(() => reportOnce(true), 15000)
  reportOnce(true)
}

const stopAuto = () => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  autoReport.value = false
}

const goChat = () => {
  if (!activeOrderId.value) {
    uni.showToast({ title: '暂无进行中订单', icon: 'none' })
    return
  }
  uni.navigateTo({
    url: `/pages/orderChat/orderChat?orderId=${encodeURIComponent(activeOrderId.value)}&mobile=${encodeURIComponent(passengerPhone.value || '')}`,
  })
}

const callPassenger = () => {
  if (!passengerPhone.value) {
    uni.showToast({ title: '暂无乘客电话', icon: 'none' })
    return
  }
  uni.makePhoneCall({ phoneNumber: String(passengerPhone.value) })
}

onShow(async () => {
  await resolveActiveOrder()
  await getLoc()
  await loadActiveOrder()
  if (activeOrderId.value) {
    getDriverLocationHistory({ orderId: activeOrderId.value, page: 1, pageSize: 1 }).catch(() => null)
  }
  startAuto()
})
onHide(() => stopAuto())
onUnload(() => stopAuto())
onUnmounted(() => stopAuto())
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 24rpx;
  background: #f4f7fb;
}
.status-card {
  border-radius: 8px;
  padding: 32rpx;
  background: linear-gradient(135deg, #1677ff, #4a9bff);
  color: #fff;
}
.status-title {
  display: block;
  font-size: 36rpx;
  font-weight: 700;
}
.status-desc,
.status-extra {
  display: block;
  margin-top: 10rpx;
  font-size: 26rpx;
  opacity: 0.92;
}
.map-card {
  margin-top: 20rpx;
  overflow: hidden;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.05);
}
.report-map {
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
.loc-card {
  background: #fff;
  border-radius: 8px;
  padding: 8rpx 32rpx;
  margin-top: 20rpx;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.05);
}
.loc-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24rpx;
  padding: 28rpx 0;
}
.loc-row + .loc-row {
  border-top: 1rpx solid #f0f2f5;
}
.loc-label {
  font-size: 26rpx;
  color: #8a93a6;
}
.loc-val {
  flex: 1;
  font-size: 28rpx;
  font-weight: 600;
  color: #1f2937;
  text-align: right;
  word-break: break-all;
}
.ctrl {
  display: flex;
  gap: 20rpx;
  margin-top: 28rpx;
}
.tip {
  margin-top: 28rpx;
  font-size: 24rpx;
  color: #667085;
  line-height: 1.6;
}
</style>
