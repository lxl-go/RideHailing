<template>
  <view class="page">
    <view class="hero">
      <view class="hero-info">
        <text class="hero-greeting">{{ greeting }}</text>
        <text class="hero-sub">{{ currentAddress || '正在获取当前位置' }}</text>
      </view>
      <view class="hero-avatar" @click="goProfile">
        <u-icon name="account-fill" size="22" color="#fff" />
      </view>
    </view>

    <view class="search-card">
      <view class="search-row" @click="goPublish">
        <u-icon name="map-fill" color="#1677ff" size="22" />
        <view class="search-text">
          <text class="search-title">发布出行需求</text>
          <text class="search-tip">填写起终点与时间，等待司机接单</text>
        </view>
        <u-icon name="arrow-right" color="#c0c4cc" />
      </view>
      <view class="search-row" @click="goTrips">
        <u-icon name="search" color="#07c160" size="22" />
        <view class="search-text">
          <text class="search-title">查找顺风车</text>
          <text class="search-tip">按出发地、目的地搜索可预约行程</text>
        </view>
        <u-icon name="arrow-right" color="#c0c4cc" />
      </view>
    </view>

    <view class="quick">
      <view class="quick-item" @click="goPublish">
        <view class="quick-icon blue"><u-icon name="edit-pen" color="#fff" size="22" /></view>
        <text>发布需求</text>
      </view>
      <view class="quick-item" @click="goTrips">
        <view class="quick-icon green"><u-icon name="list" color="#fff" size="22" /></view>
        <text>行程列表</text>
      </view>
      <view class="quick-item" @click="goTracking">
        <view class="quick-icon orange"><u-icon name="map" color="#fff" size="22" /></view>
        <text>行程追踪</text>
      </view>
      <view class="quick-item" @click="goFlood">
        <view class="quick-icon red"><u-icon name="warning" color="#fff" size="22" /></view>
        <text>积水上报</text>
      </view>
    </view>

    <view class="section">
      <view class="section-head">
        <text class="section-title">推荐行程</text>
        <text class="section-more" @click="goTrips">查看全部</text>
      </view>
      <view v-if="loading" class="skeleton">
        <view v-for="n in 3" :key="n" class="skeleton-item" />
      </view>
      <u-empty v-else-if="trips.length === 0" text="暂无可预约行程，去发布需求试试" />
      <view v-else class="trip-list">
        <view v-for="trip in trips" :key="trip.id" class="trip-card" @click="goDetail(trip.id)">
          <view class="trip-route">
            <text class="trip-point">{{ trip.origin }}</text>
            <view class="trip-line" />
            <text class="trip-point">{{ trip.destination }}</text>
          </view>
          <view class="trip-meta">
            <text>{{ formatTime(trip.depart_time) }}</text>
            <text class="dot">·</text>
            <text>余座 {{ trip.seats_available ?? '-' }}</text>
          </view>
          <view class="trip-foot">
            <u-tag :text="statusText(trip.status)" :type="statusType(trip.status)" size="mini" />
            <text class="trip-price">¥{{ trip.price }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { recommendTrips } from '@/api/trip'
import { useLocationStore } from '@/store/location'

const locationStore = useLocationStore()
const loading = ref(false)
const trips = ref([])

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '深夜好'
  if (h < 12) return '早上好'
  if (h < 18) return '下午好'
  return '晚上好'
})
const currentAddress = computed(() => locationStore.address)

const loadTrips = async () => {
  loading.value = true
  const res = await recommendTrips({ page: 1, page_size: 5 })
  loading.value = false
  if (res.code === 0) trips.value = res.data?.items || res.data || []
  else trips.value = []
}

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(5, 16) : '-')
const statusText = (s) => (s === 1 ? '已确认' : '可预约')
const statusType = (s) => (s === 1 ? 'success' : 'warning')

const goPublish = () => uni.navigateTo({ url: '/pages/publishDemand/publishDemand' })
const goTrips = () => uni.navigateTo({ url: '/pages/trips/trips' })
const goTracking = () => uni.switchTab({ url: '/pages/tracking/tracking' })
const goFlood = () => uni.navigateTo({ url: '/pages/floodReport/floodReport' })
const goProfile = () => uni.switchTab({ url: '/pages/profile/profile' })
const goDetail = (id) => uni.navigateTo({ url: `/pages/tripDetail/tripDetail?id=${id}` })

const tryLocate = () => {
  uni.getLocation({
    type: 'gcj02',
    success: (res) => locationStore.setLocation({ longitude: res.longitude, latitude: res.latitude, address: '当前位置' }),
    fail: () => locationStore.setLocation({ address: '定位未开启' })
  })
}

onMounted(() => { tryLocate(); loadTrips() })
onShow(loadTrips)
onPullDownRefresh(async () => { await loadTrips(); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 0 24rpx 40rpx; background: #f4f7fb; }
.hero { display: flex; align-items: center; justify-content: space-between; padding: 32rpx 8rpx 24rpx; }
.hero-greeting { font-size: 36rpx; font-weight: 700; color: #172033; }
.hero-sub { display: block; margin-top: 8rpx; font-size: 24rpx; color: #8a93a6; }
.hero-avatar { width: 72rpx; height: 72rpx; border-radius: 50%; background: #1677ff; display: flex; align-items: center; justify-content: center; }
.search-card { background: #fff; border-radius: 24rpx; padding: 8rpx 24rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.search-row { display: flex; align-items: center; gap: 20rpx; padding: 24rpx 0; }
.search-row + .search-row { border-top: 1rpx solid #f0f2f5; }
.search-text { flex: 1; }
.search-title { display: block; font-size: 30rpx; font-weight: 600; color: #1f2937; }
.search-tip { display: block; margin-top: 6rpx; font-size: 22rpx; color: #98a2b3; }
.quick { display: flex; justify-content: space-between; margin: 32rpx 0; }
.quick-item { display: flex; flex-direction: column; align-items: center; gap: 12rpx; width: 25%; }
.quick-item text { font-size: 22rpx; color: #475467; }
.quick-icon { width: 88rpx; height: 88rpx; border-radius: 24rpx; display: flex; align-items: center; justify-content: center; }
.quick-icon.blue { background: #1677ff; }
.quick-icon.green { background: #07c160; }
.quick-icon.orange { background: #f5a623; }
.quick-icon.red { background: #f5222d; }
.section { background: #fff; border-radius: 24rpx; padding: 28rpx 24rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.section-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20rpx; }
.section-title { font-size: 30rpx; font-weight: 700; color: #1f2937; }
.section-more { font-size: 24rpx; color: #1677ff; }
.trip-list { display: flex; flex-direction: column; gap: 20rpx; }
.trip-card { background: #f8fafc; border-radius: 20rpx; padding: 24rpx; }
.trip-route { display: flex; align-items: center; gap: 12rpx; }
.trip-point { font-size: 28rpx; font-weight: 600; color: #1f2937; }
.trip-line { flex: 1; height: 2rpx; background: #d0d7e2; position: relative; }
.trip-line::after { content: ''; position: absolute; right: -6rpx; top: -6rpx; border: 8rpx solid transparent; border-left-color: #d0d7e2; }
.trip-meta { display: flex; align-items: center; gap: 12rpx; margin-top: 16rpx; font-size: 24rpx; color: #8a93a6; }
.dot { color: #cbd2dc; }
.trip-foot { display: flex; align-items: center; justify-content: space-between; margin-top: 20rpx; }
.trip-price { font-size: 34rpx; font-weight: 700; color: #ee0a24; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; }
.skeleton-item { height: 140rpx; border-radius: 20rpx; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
