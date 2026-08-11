<template>
  <view class="page">
    <view class="filter-bar">
      <u-search v-model="keyword" placeholder="搜索出发地 / 目的地" shape="square" :showAction="false" @search="reload" @clear="reload" />
    </view>

    <view v-if="loading" class="skeleton">
      <view v-for="n in 4" :key="n" class="skeleton-item" />
    </view>
    <u-empty v-else-if="list.length === 0" text="没有匹配的行程" />
    <view v-else class="trip-list">
      <view v-for="trip in list" :key="trip.id" class="trip-card" @click="goDetail(trip.id)">
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

    <view v-if="list.length > 0 && !hasMore" class="end-tip">没有更多行程了</view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { searchTrips } from '@/api/trip'

const keyword = ref('')
const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = 10
const hasMore = ref(true)

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(5, 16) : '-')
const statusText = (s) => (s === 1 ? '已确认' : '可预约')
const statusType = (s) => (s === 1 ? 'success' : 'warning')

const load = async (reset = false) => {
  if (reset) { page.value = 1; hasMore.value = true }
  if (!hasMore.value) return
  loading.value = true
  const res = await searchTrips({ keyword: keyword.value, page: page.value, page_size: pageSize })
  loading.value = false
  if (res.code === 0) {
    const items = res.data?.items || res.data || []
    list.value = reset ? items : [...list.value, ...items]
    hasMore.value = items.length >= pageSize
  } else {
    if (reset) list.value = []
  }
}

const reload = () => load(true)
const goDetail = (id) => uni.navigateTo({ url: `/pages/tripDetail/tripDetail?id=${id}` })

onShow(() => load(true))
onReachBottom(() => { if (!loading.value && hasMore.value) { page.value += 1; load() } })
onPullDownRefresh(async () => { await reload(); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.filter-bar { margin-bottom: 20rpx; }
.trip-list { display: flex; flex-direction: column; gap: 20rpx; }
.trip-card { background: #fff; border-radius: 20rpx; padding: 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.trip-route { display: flex; align-items: center; gap: 12rpx; }
.trip-point { font-size: 28rpx; font-weight: 600; color: #1f2937; }
.trip-line { flex: 1; height: 2rpx; background: #d0d7e2; position: relative; }
.trip-line::after { content: ''; position: absolute; right: -6rpx; top: -6rpx; border: 8rpx solid transparent; border-left-color: #d0d7e2; }
.trip-meta { display: flex; align-items: center; gap: 12rpx; margin-top: 16rpx; font-size: 24rpx; color: #8a93a6; }
.dot { color: #cbd2dc; }
.trip-foot { display: flex; align-items: center; justify-content: space-between; margin-top: 20rpx; }
.trip-price { font-size: 34rpx; font-weight: 700; color: #ee0a24; }
.end-tip { text-align: center; padding: 32rpx 0; font-size: 24rpx; color: #aab4c2; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; }
.skeleton-item { height: 150rpx; border-radius: 20rpx; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
