<template>
  <view class="page">
    <view class="title-block">
      <text class="title">AI 运营预警</text>
      <text class="sub">根据接单、定位和历史数据给出风险提示</text>
    </view>
    <view v-if="loading" class="skeleton">
      <view v-for="n in 3" :key="n" class="skeleton-item" />
    </view>
    <u-empty v-else-if="alerts.length === 0" text="暂无预警" />
    <view v-else class="alert-list">
      <view v-for="alert in alerts" :key="alert.id" class="alert-card">
        <view class="alert-head">
          <text class="alert-title">{{ alert.title || alert.type || '预警' }}</text>
          <u-tag :text="alert.level || 'info'" size="mini" :type="levelType(alert.level)" />
        </view>
        <text class="alert-desc">{{ alert.description || alert.content || '无详细说明' }}</text>
        <text class="alert-time">{{ formatTime(alert.created_at || alert.time) }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { getDriverAIAlerts } from '@/api/ai'

const loading = ref(false)
const alerts = ref([])
const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(5, 19) : '-')
const levelType = (l) => ({ low: 'success', medium: 'warning', high: 'error', info: 'info' }[l] || 'info')

const load = async () => {
  loading.value = true
  const res = await getDriverAIAlerts()
  loading.value = false
  if (res?.code === 0) alerts.value = res.data?.items || res.data || []
  else alerts.value = []
}

onShow(load)
onPullDownRefresh(async () => { await load(); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.title-block { margin-bottom: 20rpx; }
.title { display: block; font-size: 36rpx; font-weight: 700; color: #1f2937; }
.sub { display: block; margin-top: 8rpx; font-size: 24rpx; color: #8a93a6; }
.alert-list { display: flex; flex-direction: column; gap: 20rpx; }
.alert-card { background: #fff; border-radius: 20rpx; padding: 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.alert-head { display: flex; justify-content: space-between; align-items: center; }
.alert-title { font-size: 30rpx; font-weight: 600; color: #1f2937; }
.alert-desc { display: block; margin-top: 14rpx; font-size: 26rpx; color: #475467; line-height: 1.6; }
.alert-time { display: block; margin-top: 12rpx; font-size: 22rpx; color: #98a2b3; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; }
.skeleton-item { height: 160rpx; border-radius: 20rpx; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
