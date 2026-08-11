<template>
  <view class="page">
    <view class="head">
      <text class="title">轨迹回放</text>
      <u-button size="mini" type="primary" :plain="true" :loading="loading" @click="load">刷新</u-button>
    </view>
    <u-empty v-if="!loading && points.length === 0" text="暂无轨迹数据" />
    <view v-else class="track-list">
      <view v-for="(p, idx) in points" :key="idx" class="track-item">
        <view class="track-dot" :class="{ active: idx === points.length - 1 }" />
        <view class="track-info">
          <text class="track-time">{{ formatTime(p.timestamp || p.time) }}</text>
          <text class="track-coord">{{ p.longitude ?? p.lng }} , {{ p.latitude ?? p.lat }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { getDriverTrackReplay } from '@/api/ai'

const orderId = ref('')
const points = ref([])
const loading = ref(false)

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '-')

const load = async () => {
  loading.value = true
  const res = await getDriverTrackReplay({ order_id: orderId.value })
  loading.value = false
  if (res?.code === 0) points.value = res.data?.points || res.data?.list || []
  else points.value = []
}

onLoad((q) => { if (q?.orderId) orderId.value = String(q.orderId); load() })
onPullDownRefresh(async () => { await load(); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20rpx; }
.title { font-size: 36rpx; font-weight: 700; color: #1f2937; }
.track-list { background: #fff; border-radius: 20rpx; padding: 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.track-item { display: flex; gap: 20rpx; padding: 20rpx 0; }
.track-item + .track-item { border-top: 1rpx solid #f0f2f5; }
.track-dot { width: 20rpx; height: 20rpx; border-radius: 50%; background: #cbd2dc; margin-top: 8rpx; }
.track-dot.active { background: #1677ff; }
.track-info { display: flex; flex-direction: column; gap: 8rpx; }
.track-time { font-size: 26rpx; font-weight: 600; color: #1f2937; }
.track-coord { font-size: 24rpx; color: #8a93a6; }
</style>
