<template>
  <view class="page">
    <view v-if="loading" class="skeleton">
      <view class="skeleton-item big" />
      <view class="skeleton-item" />
    </view>
    <template v-else-if="order">
      <view class="status-card" :class="statusType(order.status)">
        <text class="status-text">{{ statusText(order.status) }}</text>
        <text class="status-tip">{{ statusTip(order.status) }}</text>
      </view>

      <view class="route-card">
        <view class="route-row">
          <view class="dot start" />
          <text>{{ order.origin }}</text>
        </view>
        <view class="route-line" />
        <view class="route-row">
          <view class="dot end" />
          <text>{{ order.destination }}</text>
        </view>
      </view>

      <view class="info-card">
        <view class="info-row"><text class="info-label">订单号</text><text class="info-value">{{ order.order_no || order.id }}</text></view>
        <view class="info-row"><text class="info-label">下单时间</text><text class="info-value">{{ formatTime(order.created_at) }}</text></view>
        <view class="info-row"><text class="info-label">座位</text><text class="info-value">{{ order.seats ?? 1 }}</text></view>
        <view class="info-row"><text class="info-label">金额</text><text class="info-value price">¥{{ order.amount ?? '-' }}</text></view>
        <view v-if="rejectReason(order)" class="info-row"><text class="info-label">拒单原因</text><text class="info-value">{{ rejectReason(order) }}</text></view>
        <view v-if="refundAmount(order) > 0" class="info-row"><text class="info-label">退款金额</text><text class="info-value price">¥{{ refundAmount(order).toFixed(2) }}</text></view>
      </view>

      <view class="action-row">
        <u-button v-if="canPay(order.status)" type="primary" :loading="paying" @click="pay">去支付</u-button>
        <u-button v-if="canCancel(order.status)" :plain="true" @click="cancel">取消订单</u-button>
        <u-button v-if="canContact(order.status)" :plain="true" @click="goChat">联系司机</u-button>
        <u-button v-if="canContact(order.status)" :plain="true" :disabled="!driverPhone" @click="callDriver">电话</u-button>
        <u-button v-if="canViewTrack(order.status)" type="primary" @click="goTrack">查看轨迹</u-button>
      </view>
    </template>
    <u-empty v-else text="订单不存在" />
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { cancelOrder, getOrderDetail, payOrder, syncPayment } from '@/api/order'

const orderId = ref('')
const order = ref(null)
const loading = ref(true)
const paying = ref(false)
const syncingPayment = ref(false)

const statusMap = { pending: '已预约', ongoing: '进行中', completed: '已完成', cancelled: '已取消', paid: '待出行', waiting: '已预约', accepted: '已接单', picking_up: '司机来接您', delivering: '前往目的地' }
const statusText = (s) => statusMap[s] || s || '未知'
const statusType = (s) => ({ pending: 'primary', ongoing: 'primary', completed: 'success', cancelled: 'muted', paid: 'warn', waiting: 'primary', accepted: 'success', picking_up: 'primary', delivering: 'primary' }[s] || 'muted')
const statusTip = (s) => ({ pending: '已预约，请完成支付后等待司机接单', ongoing: '司机正在赶来', completed: '感谢您的使用', cancelled: '订单已取消', paid: '支付成功，等待司机接单', waiting: '已预约，请完成支付后等待司机接单', accepted: '司机已接单，可联系司机确认上车点', picking_up: '司机正在前往上车点', delivering: '您已上车，正在前往目的地' }[s] || '')
const formatTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 16) : '-')
const canPay = (s) => s === 'pending'
const canCancel = (s) => s === 'pending' || s === 'waiting' || s === 'paid'
const canViewTrack = (s) => ['accepted', 'picking_up', 'delivering', 'completed'].includes(s)
const canContact = (s) => ['accepted', 'picking_up', 'delivering'].includes(s)
const rejectReason = (item) => item?.rejectReason || item?.reject_reason || ''
const refundAmount = (item) => Number(item?.refundAmount ?? item?.refund_amount ?? 0) || 0
const driverPhone = computed(() => order.value?.driverMobile || order.value?.driver_mobile || order.value?.driverPhone || order.value?.driver_phone || '')

const paymentStorageKey = () => `passenger-payment-${orderId.value}`
const savePaymentNo = (paymentNo) => {
  if (!orderId.value || !paymentNo) return
  uni.setStorageSync(paymentStorageKey(), String(paymentNo))
}
const readPaymentNo = () => {
  if (!orderId.value) return ''
  return String(uni.getStorageSync(paymentStorageKey()) || '')
}
const clearPaymentNo = () => {
  if (!orderId.value) return
  uni.removeStorageSync(paymentStorageKey())
}

const submitAlipayForm = (formHtml) => {
  if (!formHtml) {
    uni.showToast({ title: '支付表单为空', icon: 'none' })
    return
  }
  if (typeof document === 'undefined') {
    uni.showToast({ title: '当前环境不支持支付宝跳转', icon: 'none' })
    return
  }
  const container = document.createElement('div')
  container.style.display = 'none'
  container.innerHTML = formHtml
  document.body.appendChild(container)
  const form = container.querySelector('form')
  if (!form) {
    uni.showToast({ title: '支付表单解析失败', icon: 'none' })
    container.remove()
    return
  }
  form.setAttribute('target', '_self')
  form.submit()
}

const load = async (showLoading = true) => {
  if (!orderId.value) return
  if (showLoading) loading.value = true
  const res = await getOrderDetail(orderId.value)
  if (showLoading) loading.value = false
  if (res.code === 0) order.value = res.data
}

const syncPaymentIfNeeded = async () => {
  if (!orderId.value || !order.value || !canPay(order.value.status) || syncingPayment.value) return
  const paymentNo = readPaymentNo()
  if (!paymentNo) return
  syncingPayment.value = true
  const res = await syncPayment(orderId.value, { payment_no: paymentNo })
  syncingPayment.value = false
  if (res?.code === 0 && res?.data?.synced) {
    clearPaymentNo()
    await load(false)
  }
}

const pay = () => {
  uni.showModal({
    title: '订单支付',
    content: '本次仅支持支付宝沙箱支付，买家账号 mcbfws5876@sandbox.com，支付密码 111111。确认继续吗？',
    success: async (m) => {
      if (!m.confirm) return
      paying.value = true
      const res = await payOrder(orderId.value)
      paying.value = false
      if (res?.code === 0) {
        const payForm = res?.data?.payForm || res?.data?.pay_form
        savePaymentNo(res?.data?.paymentNo || res?.data?.payment_no)
        submitAlipayForm(payForm)
        return
      }
      uni.showToast({ title: res?.msg || '发起支付失败', icon: 'none' })
    },
  })
}

const cancel = () => {
  uni.showModal({
    title: '取消订单',
    content: '确认取消该订单吗？',
    success: async (m) => {
      if (!m.confirm) return
      const res = await cancelOrder(orderId.value, { idempotency_key: `p-${orderId.value}-cancel-${Date.now()}` })
      if (res.code === 0) {
        uni.showToast({ title: '已取消', icon: 'success' })
        load()
      }
    },
  })
}

const goTrack = () => {
  if (!orderId.value) return
  uni.setStorageSync('passenger-track-order-id', String(orderId.value))
  uni.switchTab({ url: '/pages/tracking/tracking' })
}

const goChat = () => {
  if (!orderId.value) return
  uni.navigateTo({
    url: `/pages/orderChat/orderChat?orderId=${encodeURIComponent(orderId.value)}&mobile=${encodeURIComponent(driverPhone.value || '')}`,
  })
}

const callDriver = () => {
  if (!driverPhone.value) {
    uni.showToast({ title: '暂无司机电话', icon: 'none' })
    return
  }
  uni.makePhoneCall({ phoneNumber: String(driverPhone.value) })
}

onLoad((options) => {
  orderId.value = String(options.orderId || options.id || '')
})
onShow(async () => {
  await load()
  await syncPaymentIfNeeded()
})
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx 24rpx 40rpx; background: #f4f7fb; }
.status-card { border-radius: 24rpx; padding: 32rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); background: #fff; }
.status-card.success { background: linear-gradient(135deg, #07c160, #34d399); }
.status-card.primary { background: linear-gradient(135deg, #1677ff, #4a9bff); }
.status-card.warn { background: linear-gradient(135deg, #f5a623, #fbbf24); }
.status-card.muted { background: linear-gradient(135deg, #98a2b3, #cbd2dc); }
.status-text { display: block; font-size: 38rpx; font-weight: 700; color: #fff; }
.status-tip { display: block; margin-top: 8rpx; font-size: 26rpx; color: rgba(255,255,255,0.85); }
.route-card { background: #fff; border-radius: 24rpx; padding: 32rpx; margin-top: 20rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.route-row { display: flex; align-items: center; gap: 16rpx; font-size: 30rpx; font-weight: 600; color: #1f2937; }
.route-line { margin-left: 7rpx; height: 56rpx; border-left: 2rpx dashed #cbd2dc; }
.dot { width: 18rpx; height: 18rpx; border-radius: 50%; }
.dot.start { background: #1677ff; }
.dot.end { background: #ee0a24; }
.info-card { background: #fff; border-radius: 24rpx; padding: 16rpx 32rpx; margin-top: 20rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.info-row { display: flex; justify-content: space-between; align-items: center; padding: 24rpx 0; }
.info-row + .info-row { border-top: 1rpx solid #f0f2f5; }
.info-label { font-size: 26rpx; color: #8a93a6; }
.info-value { font-size: 28rpx; font-weight: 600; color: #1f2937; }
.info-value.price { color: #ee0a24; }
.action-row { display: flex; gap: 20rpx; margin-top: 32rpx; }
.skeleton { display: flex; flex-direction: column; gap: 20rpx; }
.skeleton-item { height: 200rpx; border-radius: 24rpx; background: linear-gradient(90deg, #eef1f6 25%, #e6eaf2 37%, #eef1f6 63%); background-size: 400% 100%; animation: shimmer 1.4s infinite; }
.skeleton-item.big { height: 320rpx; }
@keyframes shimmer { 0% { background-position: 100% 0; } 100% { background-position: 0 0; } }
</style>
