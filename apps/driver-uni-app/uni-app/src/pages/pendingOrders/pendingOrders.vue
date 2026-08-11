<template>
  <view class="page">
    <view v-if="loading" class="skeleton">
      <view v-for="n in 4" :key="n" class="skeleton-item" />
    </view>
    <u-empty v-else-if="orders.length === 0" text="暂无待接订单" />
    <view v-else class="order-list">
      <view v-for="order in orders" :key="order.id" class="order-card">
        <view class="order-head">
          <text class="order-no">{{ order.orderNo || `#${order.id}` }}</text>
          <u-tag :text="order.serviceType || '顺风车'" type="primary" size="mini" />
        </view>
        <view class="order-route">
          <text>{{ order.origin || '-' }}</text>
          <u-icon name="arrow-right" />
          <text>{{ order.destination || '-' }}</text>
        </view>
        <view class="order-meta">
          <text>{{ formatTime(order.departTime || order.depart_time) }}</text>
          <text>座位 {{ order.seats || 1 }}</text>
          <text class="price">¥{{ order.amount ?? '-' }}</text>
        </view>
        <view class="order-action">
          <u-button size="mini" :plain="true" @click="goDetail(order.id)">详情</u-button>
          <u-button size="mini" :plain="true" type="error" :loading="rejectingId === order.id" @click="reject(order)">拒单</u-button>
          <u-button size="mini" type="success" :loading="acceptingId === order.id" @click="accept(order)">接单</u-button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow, onHide, onUnload, onPullDownRefresh } from '@dcloudio/uni-app'
import { listAvailableOrders, acceptOrder, rejectOrder } from '@/api/order'
import { createAutoRefresh } from '@/utils/autoRefresh.mjs'

const orders = ref([])
const loading = ref(false)
const acceptingId = ref('')
const rejectingId = ref('')

const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(5, 16) : '-')

const load = async (showLoading = true) => {
  if (showLoading) loading.value = true
  const res = await listAvailableOrders()
  if (showLoading) loading.value = false
  if (res?.code === 0) orders.value = res.data?.items || res.data?.list || res.data || []
  else orders.value = []
}

const accept = async (order) => {
  acceptingId.value = order.id
  const res = await acceptOrder(order.id, { idempotency_key: `d-${order.id}-accept-${Date.now()}` })
  acceptingId.value = ''
  if (res?.code === 0) {
    uni.setStorageSync('driverActiveOrderId', order.id)
    uni.showToast({ title: '接单成功' })
    orders.value = orders.value.filter((item) => item.id !== order.id)
    setTimeout(() => {
      uni.switchTab({ url: '/pages/locationReport/locationReport' })
    }, 500)
  } else {
    uni.showToast({ title: res?.msg || '订单暂不可接', icon: 'none' })
  }
}

const reject = (order) => {
  uni.showModal({
    title: '拒绝订单',
    content: '请输入拒单原因，乘客端将按订单状态查看变更。',
    editable: true,
    placeholderText: '填写拒单原因',
    success: async (modal) => {
      if (!modal.confirm) return
      rejectingId.value = order.id
      const reason = String(modal.content || '').trim()
      const res = await rejectOrder(order.id, {
        idempotency_key: `d-${order.id}-reject-${Date.now()}`,
        reject_reason: reason,
      })
      rejectingId.value = ''
      if (res?.code === 0) {
        uni.showToast({ title: '已拒绝订单', icon: 'success' })
        orders.value = orders.value.filter((item) => item.id !== order.id)
      } else {
        uni.showToast({ title: res?.msg || '拒单失败', icon: 'none' })
      }
    },
  })
}

const goDetail = (id) => uni.navigateTo({ url: `/pages/orderDetail/orderDetail?id=${id}` })

const autoRefresh = createAutoRefresh(() => load(false), { intervalMs: 5000, runImmediately: false })

onShow(() => { autoRefresh.start(); load(true) })
onHide(() => autoRefresh.stop())
onUnload(() => autoRefresh.stop())
onPullDownRefresh(async () => { await load(true); uni.stopPullDownRefresh() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.order-list { display: flex; flex-direction: column; gap: 20rpx; }
.order-card { background: #fff; border-radius: 8px; padding: 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.order-head { display: flex; align-items: center; justify-content: space-between; }
.order-no { font-size: 24rpx; color: #8a93a6; }
.order-route { display: flex; align-items: center; gap: 12rpx; margin-top: 20rpx; font-size: 30rpx; font-weight: 600; color: #1f2937; }
.order-meta { display: flex; gap: 20rpx; margin-top: 16rpx; font-size: 24rpx; color: #8a93a6; }
.order-meta .price { margin-left: auto; font-size: 30rpx; font-weight: 700; color: #ee0a24; }
.order-action { display: flex; gap: 16rpx; justify-content: flex-end; margin-top: 20rpx; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; }
.skeleton-item { height: 200rpx; border-radius: 8px; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
