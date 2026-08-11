<template>
  <view class="page">
    <view class="tabs">
      <view v-for="tab in tabs" :key="tab.value" class="tab" :class="{ active: active === tab.value }" @click="switchTab(tab.value)">
        <text>{{ tab.label }}</text>
      </view>
    </view>

    <view v-if="loading" class="skeleton">
      <view v-for="n in 4" :key="n" class="skeleton-item" />
    </view>
    <u-empty v-else-if="list.length === 0" text="暂无订单" />
    <view v-else class="order-list">
      <view v-for="order in list" :key="order.id" class="order-card" @click="goDetail(order.id)">
        <view class="order-head">
          <text class="order-no">订单号 {{ order.order_no || order.id }}</text>
          <text class="status-badge" :class="statusType(order.status, order)">
            {{ statusText(order.status, order) }}
          </text>
        </view>
        <view class="order-route">
          <text>{{ order.origin }}</text>
          <u-icon name="arrow-right" size="20" color="#c0c4cc" />
          <text>{{ order.destination }}</text>
        </view>
        <view class="order-foot">
          <text class="order-time">{{ formatTime(order.created_at) }}</text>
          <text class="order-price">¥{{ order.amount ?? '-' }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { listOrders } from '@/api/order'
import { getOrderStatusText, getOrderStatusType } from '@/utils/orderStatus.mjs'

const tabs = [
  { label: '全部', value: '' },
  { label: '待出行', value: 'pending' },
  { label: '进行中', value: 'ongoing' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' }
]
const active = ref('')
const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = 10
const hasMore = ref(true)

const statusText = (s, order = {}) => getOrderStatusText(s, order)
const statusType = (s, order = {}) => getOrderStatusType(s, order)
const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(5, 16) : '-')

const load = async (reset = false) => {
  if (reset) { page.value = 1; hasMore.value = true }
  if (!hasMore.value) return
  loading.value = true
  const res = await listOrders({ status: active.value, page: page.value, page_size: pageSize })
  loading.value = false
  if (res.code === 0) {
    const items = res.data?.items || res.data?.list || res.data || []
    list.value = reset ? items : [...list.value, ...items]
    hasMore.value = items.length >= pageSize
  } else if (reset) list.value = []
}

const switchTab = (val) => { active.value = val; load(true) }
const goDetail = (id) => uni.navigateTo({ url: `/pages/orderDetail/orderDetail?id=${id}` })

onShow(() => load(true))
onReachBottom(() => { if (!loading.value && hasMore.value) { page.value += 1; load() } })
onPullDownRefresh(async () => { await load(true); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 0 24rpx 40rpx; background: #f4f7fb; }
.tabs { display: flex; gap: 12rpx; padding: 20rpx 0; overflow-x: auto; }
.tab { flex: none; padding: 12rpx 28rpx; border-radius: 32rpx; background: #fff; font-size: 26rpx; color: #475467; }
.tab.active { background: #1677ff; color: #fff; }
.order-list { display: flex; flex-direction: column; gap: 20rpx; }
.order-card { background: #fff; border-radius: 20rpx; padding: 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.order-head { display: flex; align-items: center; justify-content: space-between; }
.order-no { font-size: 24rpx; color: #8a93a6; }
.status-badge { min-width: 112rpx; height: 44rpx; line-height: 44rpx; padding: 0 16rpx; border-radius: 8rpx; box-sizing: border-box; text-align: center; font-size: 24rpx; font-weight: 600; color: #fff; }
.status-badge.success { background: #12b76a; }
.status-badge.error { background: #f04438; }
.status-badge.warning { background: #f79009; }
.status-badge.primary { background: #1677ff; }
.status-badge.info { background: #98a2b3; }
.order-route { display: flex; align-items: center; gap: 12rpx; margin-top: 20rpx; font-size: 30rpx; font-weight: 600; color: #1f2937; }
.order-foot { display: flex; align-items: center; justify-content: space-between; margin-top: 20rpx; }
.order-time { font-size: 24rpx; color: #98a2b3; }
.order-price { font-size: 32rpx; font-weight: 700; color: #ee0a24; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; }
.skeleton-item { height: 180rpx; border-radius: 20rpx; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
