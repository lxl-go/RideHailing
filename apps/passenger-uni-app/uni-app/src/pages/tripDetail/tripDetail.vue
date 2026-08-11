<template>
  <view class="page">
    <view v-if="loading" class="skeleton">
      <view class="skeleton-item big" />
      <view class="skeleton-item" />
      <view class="skeleton-item" />
    </view>
    <template v-else-if="trip">
      <view class="trip-card">
        <view class="trip-route">
          <view class="route-point">
            <view class="dot start" />
            <text class="route-text">{{ trip.origin }}</text>
          </view>
          <view class="route-line" />
          <view class="route-point">
            <view class="dot end" />
            <text class="route-text">{{ trip.destination }}</text>
          </view>
        </view>
        <view class="info-grid">
          <view class="info-cell"><text class="info-label">出发时间</text><text class="info-value">{{ formatTime(trip.depart_time) }}</text></view>
          <view class="info-cell"><text class="info-label">座位</text><text class="info-value">{{ trip.seats_available ?? '-' }}</text></view>
          <view class="info-cell"><text class="info-label">车型</text><text class="info-value">{{ trip.vehicle_type || '顺风车' }}</text></view>
          <view class="info-cell"><text class="info-label">状态</text><text class="info-value">{{ statusText(trip.status) }}</text></view>
        </view>
      </view>

      <view class="driver-card" v-if="trip.driver">
        <u-avatar :src="trip.driver.avatar" size="80" />
        <view class="driver-info">
          <text class="driver-name">{{ trip.driver.name || '司机师傅' }}</text>
          <text class="driver-meta">{{ trip.driver.vehicle_no || '' }} · 评分 {{ trip.driver.rating || '5.0' }}</text>
        </view>
      </view>

      <view class="price-card">
        <text class="price-label">行程价格</text>
        <text class="price-value">¥{{ trip.price }}</text>
      </view>

      <view class="bottom-bar">
        <view class="seat-row">
          <text class="seat-label">座位</text>
          <view class="seat-ctrl">
            <view class="seat-btn" @click="changeSeat(-1)">-</view>
            <text class="seat-num">{{ seatCount }}</text>
            <view class="seat-btn" @click="changeSeat(1)">+</view>
          </view>
        </view>
        <u-button type="primary" :loading="booking" @click="book">立即预约</u-button>
      </view>
    </template>
    <u-empty v-else text="行程不存在或已下架" />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getTripDetail } from '@/api/trip'
import { createOrder } from '@/api/order'

const tripId = ref('')
const trip = ref(null)
const loading = ref(true)
const seatCount = ref(1)
const booking = ref(false)

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 16) : '-')
const statusText = (s) => (s === 1 ? '已确认' : '可预约')

const loadDetail = async () => {
  if (!tripId.value) return
  loading.value = true
  const res = await getTripDetail(tripId.value)
  loading.value = false
  if (res.code === 0) trip.value = res.data
}

const changeSeat = (delta) => {
  const next = seatCount.value + delta
  if (next < 1) return
  const max = trip.value?.seats_available || 1
  if (next > max) return uni.showToast({ title: '超过剩余座位', icon: 'none' })
  seatCount.value = next
}

const book = async () => {
  booking.value = true
  const res = await createOrder({ trip_id: tripId.value, seats: seatCount.value })
  booking.value = false
  if (res.code === 0) {
    uni.showToast({ title: '预约成功', icon: 'success' })
    const orderId = res.data?.order_id || res.data?.id || res.data
    setTimeout(() => uni.redirectTo({ url: `/pages/orderDetail/orderDetail?id=${orderId}` }), 600)
  }
}

onLoad((options) => { tripId.value = String(options.id || ''); loadDetail() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx 24rpx 200rpx; background: #f4f7fb; }
.trip-card { background: #fff; border-radius: 24rpx; padding: 32rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.trip-route { display: flex; flex-direction: column; gap: 8rpx; }
.route-point { display: flex; align-items: center; gap: 16rpx; }
.dot { width: 18rpx; height: 18rpx; border-radius: 50%; }
.dot.start { background: #1677ff; }
.dot.end { background: #ee0a24; }
.route-text { font-size: 30rpx; font-weight: 600; color: #1f2937; }
.route-line { margin-left: 8rpx; height: 48rpx; border-left: 2rpx dashed #cbd2dc; }
.info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 24rpx 0; margin-top: 32rpx; }
.info-cell { display: flex; flex-direction: column; gap: 8rpx; }
.info-label { font-size: 24rpx; color: #8a93a6; }
.info-value { font-size: 28rpx; font-weight: 600; color: #1f2937; }
.driver-card { display: flex; align-items: center; gap: 20rpx; margin-top: 20rpx; background: #fff; border-radius: 24rpx; padding: 24rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.driver-info { flex: 1; }
.driver-name { display: block; font-size: 30rpx; font-weight: 600; color: #1f2937; }
.driver-meta { display: block; margin-top: 8rpx; font-size: 24rpx; color: #8a93a6; }
.price-card { display: flex; align-items: baseline; justify-content: space-between; margin-top: 20rpx; background: #fff; border-radius: 24rpx; padding: 28rpx 32rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.price-label { font-size: 28rpx; color: #475467; }
.price-value { font-size: 44rpx; font-weight: 700; color: #ee0a24; }
.bottom-bar { position: fixed; left: 0; right: 0; bottom: 0; padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom)); background: #fff; display: flex; align-items: center; gap: 24rpx; box-shadow: 0 -6rpx 20rpx rgba(16,24,40,0.08); }
.seat-row { display: flex; align-items: center; gap: 12rpx; }
.seat-label { font-size: 26rpx; color: #475467; }
.seat-ctrl { display: flex; align-items: center; gap: 16rpx; }
.seat-btn { width: 52rpx; height: 52rpx; border-radius: 50%; background: #f0f2f5; display: flex; align-items: center; justify-content: center; font-size: 32rpx; color: #475467; }
.seat-num { font-size: 32rpx; font-weight: 600; min-width: 40rpx; text-align: center; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; }
.skeleton-item { height: 200rpx; border-radius: 24rpx; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
.skeleton-item.big { height: 320rpx; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
