<template>
  <view class="page">
    <view class="filters">
      <u-tabs :list="tabs" :current="tabIndex" @change="changeTab" />
    </view>
    <view v-if="loading" class="state">加载中...</view>
    <view v-else-if="errorMessage" class="state state-error">{{ errorMessage }}</view>
    <view v-else-if="!trips.length" class="state">暂无行程</view>
    <view v-else class="list">
      <view v-for="trip in trips" :key="trip.id" class="trip-card">
        <view class="row">
          <text class="route">{{ trip.origin }} → {{ trip.destination }}</text>
          <text :class="['status', statusClass(trip.status)]">{{ statusText(trip.status) }}</text>
        </view>
        <text class="meta">出发：{{ formatTime(trip.depart_time) }}</text>
        <text class="meta">{{ trip.seats_available }}/{{ trip.seats_total }} 座 · ￥{{ Number(trip.price || 0).toFixed(2) }}</text>
        <text v-if="trip.reject_reason" class="reject">驳回原因：{{ trip.reject_reason }}</text>
        <view class="actions">
          <u-button size="small" plain type="error" @click="removeTrip(trip)">删除</u-button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { deleteTrip, listMyTrips } from '@/api/trip'

const tabs = [{ name: '全部', status: null }, { name: '待审核', status: 10 }, { name: '已通过', status: 20 }, { name: '已驳回', status: 30 }]
const tabIndex = ref(0)
const trips = ref([])
const loading = ref(false)
const errorMessage = ref('')

const statusBucket = (status) => {
  const normalized = typeof status === 'string' ? status.trim().toLowerCase() : status
  const numeric = Number(normalized)
  if (numeric === 10 || numeric === 1) return 10
  if (numeric === 20 || numeric === 2) return 20
  if (numeric === 30 || numeric === 3) return 30
  const textMap = {
    pending: 10,
    wait_review: 10,
    waiting_review: 10,
    reviewing: 10,
    approved: 20,
    passed: 20,
    pass: 20,
    rejected: 30,
    reject: 30,
  }
  return textMap[normalized] || numeric
}
const statusText = (status) => ({ 10: '待审核', 20: '已通过', 30: '已驳回' }[statusBucket(status)] || '未知')
const statusClass = (status) => ({ 10: 'pending', 20: 'approved', 30: 'rejected' }[statusBucket(status)] || '')
const formatTime = (value) => value ? String(value).replace('T', ' ').slice(0, 16) : '--'

const buildListParams = () => {
  const params = { page: 1, page_size: 100 }
  const status = tabs[tabIndex.value]?.status
  if (status !== null && status !== undefined && status !== '') params.status = status
  return params
}

const extractTripItems = (result = {}) => {
  const data = result.data
  if (Array.isArray(data)) return data
  const candidates = [
    data?.items,
    data?.list,
    data?.records,
    data?.rows,
    data?.trips,
    data?.data?.items,
    data?.data?.list,
    data?.data?.records,
    data?.data?.rows,
    data?.data?.trips,
    data?.data,
    result.items,
    result.list,
    result.records,
    result.rows,
    result.trips,
  ]
  return candidates.find((item) => Array.isArray(item)) || []
}

const filterByCurrentTab = (items) => {
  const status = tabs[tabIndex.value]?.status
  if (status === null || status === undefined || status === '') return items
  return items.filter((trip) => statusBucket(trip.status) === status)
}

const load = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await listMyTrips(buildListParams())
    if (result?.code !== 0) {
      trips.value = []
      errorMessage.value = result?.msg || '行程加载失败，请稍后重试'
      return
    }
    trips.value = filterByCurrentTab(extractTripItems(result))
  } catch (e) {
    trips.value = []
    errorMessage.value = '行程加载失败，请稍后重试'
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}
const resolveTabIndex = (payload) => {
  const rawIndex = typeof payload === 'number' ? payload : payload?.index ?? payload?.detail?.index ?? payload?.value
  const index = Number(rawIndex)
  if (Number.isInteger(index) && index >= 0 && index < tabs.length) return index
  const statusIndex = tabs.findIndex((item) => item.status === Number(payload?.status))
  if (statusIndex >= 0) return statusIndex
  const nameIndex = tabs.findIndex((item) => item.name === payload?.name)
  return nameIndex >= 0 ? nameIndex : tabIndex.value
}
const changeTab = (payload) => { tabIndex.value = resolveTabIndex(payload); load() }
const removeTrip = (trip) => {
  uni.showModal({
    title: '删除行程',
    content: '删除后不可恢复，确认继续吗？',
    success: async ({ confirm }) => {
      if (!confirm) return
      const res = await deleteTrip(trip.id)
      if (res.code === 0) {
        uni.showToast({ title: '已删除', icon: 'success' })
        load()
      } else {
        uni.showToast({ title: res.msg || '删除失败', icon: 'none' })
      }
    }
  })
}
onShow(load)
onPullDownRefresh(load)
</script>

<style scoped>
.page { min-height: 100vh; background: #f4f7fb; padding: 20rpx; }
.filters, .trip-card { background: #fff; border-radius: 20rpx; }
.filters { margin-bottom: 20rpx; overflow: hidden; }
.state { padding: 120rpx 0; text-align: center; color: #8c96a5; }
.state-error { color: #dc2626; }
.trip-card { padding: 24rpx; margin-bottom: 18rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,.05); }
.row { display: flex; align-items: flex-start; justify-content: space-between; gap: 16rpx; }
.route { flex: 1; color: #172033; font-size: 31rpx; font-weight: 600; }
.status { padding: 4rpx 12rpx; border-radius: 8rpx; font-size: 22rpx; }
.pending { color: #b26a00; background: #fff5dd; }.approved { color: #15803d; background: #eaf8ef; }.rejected { color: #dc2626; background: #fff0f0; }
.meta, .reject { display: block; margin-top: 14rpx; color: #697386; font-size: 25rpx; }.reject { color: #dc2626; }
.actions { display: flex; justify-content: flex-end; margin-top: 18rpx; }
</style>
