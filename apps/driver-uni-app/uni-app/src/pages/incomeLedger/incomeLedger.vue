<template>
  <view class="page">
    <view class="summary-card">
      <view class="summary-item"><text class="num">{{ moneyText(todayIncome) }}</text><text class="label">今日收入</text></view>
      <view class="summary-item"><text class="num">{{ orderCount }}</text><text class="label">今日接单</text></view>
      <view class="summary-item"><text class="num">{{ moneyText(pendingWithdraw) }}</text><text class="label">可提现</text></view>
    </view>

    <view v-if="loading" class="skeleton">
      <view v-for="n in 4" :key="n" class="skeleton-item" />
    </view>
    <u-empty v-else-if="records.length === 0" text="暂无收入明细" />
    <view v-else class="record-list">
      <view v-for="item in records" :key="item.id" class="record-card">
        <view class="record-head">
          <text class="record-title">{{ item.title }}</text>
          <text class="record-amount">+¥{{ moneyText(item.amount) }}</text>
        </view>
        <text class="record-sub">{{ item.origin || '-' }} → {{ item.destination || '-' }}</text>
        <text class="record-time">{{ formatTime(item.acceptedAt) }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onPullDownRefresh } from '@dcloudio/uni-app'
import { getDriverIncome } from '@/api/driver'

const loading = ref(false)
const records = ref([])
const todayIncome = ref(0)
const orderCount = ref(0)
const pendingWithdraw = ref(0)

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '-')
const moneyText = (value) => {
  const n = Number(value)
  return Number.isFinite(n) ? n.toFixed(2) : '0.00'
}

const load = async () => {
  loading.value = true
  const res = await getDriverIncome({ page: 1, page_size: 100 })
  loading.value = false
  if (res?.code !== 0) {
    records.value = []
    todayIncome.value = 0
    orderCount.value = 0
    pendingWithdraw.value = 0
    return
  }

  const payload = res.data || {}
  todayIncome.value = Number(payload.todayIncome ?? payload.today_income ?? 0) || 0
  orderCount.value = Number(payload.todayOrders ?? payload.today_orders ?? 0) || 0
  pendingWithdraw.value = Number(payload.pendingWithdraw ?? payload.pending_withdraw ?? 0) || 0
  records.value = (payload.records || payload.list || []).map((item, index) => ({
    id: item.id || item.orderId || `${index}`,
    title: item.orderNo || item.order_no || `订单 ${item.orderId || item.order_id || index + 1}`,
    amount: Number(item.amount ?? item.total_price ?? item.totalPrice ?? 0) || 0,
    origin: item.origin,
    destination: item.destination,
    acceptedAt: item.acceptedAt || item.accepted_at || item.createdAt || item.created_at,
  }))
}

onShow(load)
onPullDownRefresh(async () => { await load(); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.summary-card { display: flex; background: linear-gradient(135deg, #1677ff, #4a9bff); color: #fff; border-radius: 8px; padding: 28rpx 0; }
.summary-item { flex: 1; text-align: center; }
.num { display: block; font-size: 38rpx; font-weight: 700; }
.label { display: block; margin-top: 8rpx; font-size: 24rpx; opacity: .85; }
.record-list { display: flex; flex-direction: column; gap: 20rpx; margin-top: 20rpx; }
.record-card { background: #fff; border-radius: 8px; padding: 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.record-head { display: flex; align-items: center; justify-content: space-between; }
.record-title { font-size: 30rpx; font-weight: 600; color: #1f2937; }
.record-amount { font-size: 32rpx; font-weight: 700; color: #07c160; }
.record-sub { display: block; margin-top: 12rpx; font-size: 26rpx; color: #475467; }
.record-time { display: block; margin-top: 10rpx; font-size: 22rpx; color: #98a2b3; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; margin-top: 20rpx; }
.skeleton-item { height: 160rpx; border-radius: 8px; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
