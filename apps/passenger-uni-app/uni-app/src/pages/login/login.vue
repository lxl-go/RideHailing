<template>
  <view class="login-page" :class="{ 'is-dark': isDark }">
    <view class="topbar">
      <view class="brand-lockup">
        <view class="brand-mark">
          <u-icon name="map" size="21" color="#ffffff" />
        </view>
        <text class="brand-name">顺路出行</text>
      </view>
      <button class="theme-toggle" type="button" @click="toggleTheme">
        <u-icon :name="isDark ? 'sun' : 'moon'" size="18" :color="isDark ? '#f8fafc' : '#12201b'" />
      </button>
    </view>

    <view class="map-hero">
      <view class="route-line" />
      <view class="route-card pickup-card">
        <view class="route-dot pickup-dot" />
        <view>
          <text class="route-label">上车点</text>
          <text class="route-value">当前位置附近</text>
        </view>
      </view>
      <view class="route-card destination-card">
        <view class="route-dot destination-dot" />
        <view>
          <text class="route-label">目的地</text>
          <text class="route-value">登录后快速发布行程</text>
        </view>
      </view>
      <view class="car-chip">
        <u-icon name="car" size="18" color="#ffffff" />
        <text>附近 6 辆</text>
      </view>
    </view>

    <view class="hero-copy">
      <text class="eyebrow">Passenger ride</text>
      <text class="title">去哪里，现在就出发</text>
      <text class="subtitle">短信验证后发布需求、查看订单，并跟踪司机位置和行程动态。</text>
    </view>

    <view class="login-card">
      <view class="card-heading">
        <text class="card-title">手机号登录</text>
        <text class="card-subtitle">未注册手机号将自动创建乘客账号</text>
      </view>

      <u-form :model="form" ref="formRef" label-position="left" label-width="132rpx">
        <u-form-item label="手机号" prop="mobile">
          <u-input
            v-model="form.mobile"
            type="number"
            maxlength="11"
            placeholder="请输入手机号"
            border="surround"
            clearable
            prefixIcon="phone"
          />
        </u-form-item>

        <u-form-item label="验证码" prop="code">
          <view class="code-row">
            <u-input
              v-model="form.code"
              type="number"
              maxlength="6"
              placeholder="请输入验证码"
              border="surround"
              clearable
              prefixIcon="lock"
            />
            <button class="code-button" type="button" :disabled="sending || countdown > 0" @click="sendCode">
              {{ countdown > 0 ? `${countdown}s` : sending ? '发送中' : '获取验证码' }}
            </button>
          </view>
        </u-form-item>

        <button class="login-button" type="button" :disabled="loggingIn" @click="login">
          <text>{{ loggingIn ? '登录中...' : '一键进入行程' }}</text>
          <u-icon name="arrow-right" size="18" color="#ffffff" />
        </button>
      </u-form>

      <view class="safe-row">
        <view class="safe-item">
          <u-icon name="checkmark-circle" size="16" color="#16a34a" />
          <text>隐私保护</text>
        </view>
        <view class="safe-item">
          <u-icon name="coupon" size="16" color="#16a34a" />
          <text>优惠可用</text>
        </view>
        <view class="safe-item">
          <u-icon name="bell" size="16" color="#16a34a" />
          <text>行程提醒</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { reactive, ref, onBeforeUnmount } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { loginPassenger, sendPassengerLoginCode } from '@/api/auth'
import { useUserStore } from '@/store/user'
import { buildAuthenticatedLoginRedirect } from '@/utils/authNavigation.mjs'

const userStore = useUserStore()
const form = reactive({ mobile: '', code: '' })
const sending = ref(false)
const loggingIn = ref(false)
const countdown = ref(0)
const isDark = ref(uni.getStorageSync('passengerLoginTheme') === 'dark')
let timer = null

const validMobile = () => /^1\d{10}$/.test(form.mobile)

const toggleTheme = () => {
  isDark.value = !isDark.value
  uni.setStorageSync('passengerLoginTheme', isDark.value ? 'dark' : 'light')
}

const startCountdown = () => {
  countdown.value = 60
  clearInterval(timer)
  timer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) clearInterval(timer)
  }, 1000)
}

const redirectIfAuthenticated = () => {
  const redirect = buildAuthenticatedLoginRedirect({
    readStorage: (key) => uni.getStorageSync(key),
    tokenKey: 'passengerAccessToken',
    homeUrl: '/pages/home/home',
  })
  if (!redirect.shouldRedirect) return
  console.info('[passenger-nav] skip login because session exists')
  uni.reLaunch({ url: redirect.url })
}

const sendCode = async () => {
  if (!validMobile()) return uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
  sending.value = true
  const res = await sendPassengerLoginCode(form.mobile)
  sending.value = false
  if (res.code === 0) {
    uni.showToast({ title: '验证码已发送', icon: 'success' })
    startCountdown()
  }
}

const login = async () => {
  if (!validMobile() || !form.code) return uni.showToast({ title: '请填写手机号和验证码', icon: 'none' })
  loggingIn.value = true
  const res = await loginPassenger(form.mobile, form.code)
  loggingIn.value = false
  if (res.code === 0) {
    userStore.setSession(res.data)
    uni.reLaunch({ url: '/pages/home/home' })
  } else {
    uni.showToast({ title: res.msg || '登录失败，请检查验证码', icon: 'none' })
  }
}

onShow(redirectIfAuthenticated)
onBeforeUnmount(() => clearInterval(timer))
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  padding: 54rpx 32rpx 40rpx;
  box-sizing: border-box;
  overflow-x: hidden;
  color: #12201b;
  background:
    radial-gradient(circle at 20% 8%, rgba(52, 211, 153, 0.28), transparent 28%),
    linear-gradient(180deg, #eefdf4 0%, #f8fafc 56%, #eef6ff 100%);
}

.login-page.is-dark {
  color: #f8fafc;
  background:
    radial-gradient(circle at 20% 10%, rgba(34, 197, 94, 0.22), transparent 30%),
    linear-gradient(180deg, #0f1f1a 0%, #111827 62%, #0b1220 100%);
}

.topbar,
.brand-lockup,
.route-card,
.car-chip,
.safe-row,
.safe-item {
  display: flex;
  align-items: center;
}

.topbar {
  justify-content: space-between;
  margin-bottom: 26rpx;
}

.brand-lockup {
  gap: 14rpx;
}

.brand-mark,
.theme-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64rpx;
  height: 64rpx;
  border: 0;
  border-radius: 32rpx;
  padding: 0;
}

.brand-mark {
  background: linear-gradient(135deg, #16a34a, #0ea5e9);
}

.theme-toggle {
  background: rgba(255, 255, 255, 0.76);
  flex: 0 0 64rpx;
  overflow: hidden;
  font-size: 0;
  line-height: 1;
  box-shadow: 0 10rpx 26rpx rgba(2, 6, 23, 0.08);
}

.theme-toggle::after {
  border: 0;
}

.is-dark .theme-toggle {
  background: rgba(255, 255, 255, 0.12);
}

.brand-name {
  font-size: 28rpx;
  font-weight: 850;
  color: currentColor;
}

.map-hero {
  position: relative;
  height: 360rpx;
  margin-bottom: 30rpx;
  overflow: hidden;
  border: 1rpx solid rgba(15, 23, 42, 0.08);
  border-radius: 16rpx;
  background:
    linear-gradient(90deg, rgba(15, 23, 42, 0.06) 1rpx, transparent 1rpx),
    linear-gradient(0deg, rgba(15, 23, 42, 0.06) 1rpx, transparent 1rpx),
    linear-gradient(135deg, rgba(240, 253, 244, 0.95), rgba(239, 246, 255, 0.95));
  background-size: 56rpx 56rpx, 56rpx 56rpx, 100% 100%;
  box-shadow: 0 24rpx 70rpx rgba(22, 163, 74, 0.14);
}

.is-dark .map-hero {
  border-color: rgba(255, 255, 255, 0.1);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.07) 1rpx, transparent 1rpx),
    linear-gradient(0deg, rgba(255, 255, 255, 0.07) 1rpx, transparent 1rpx),
    linear-gradient(135deg, rgba(15, 31, 26, 0.92), rgba(15, 23, 42, 0.96));
  background-size: 56rpx 56rpx, 56rpx 56rpx, 100% 100%;
}

.route-line {
  position: absolute;
  left: 106rpx;
  top: 92rpx;
  width: 4rpx;
  height: 138rpx;
  border-radius: 999rpx;
  background: linear-gradient(180deg, #16a34a, #0ea5e9);
}

.route-card {
  position: absolute;
  left: 52rpx;
  right: 52rpx;
  gap: 18rpx;
  min-height: 96rpx;
  padding: 20rpx 22rpx;
  box-sizing: border-box;
  border-radius: 14rpx;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 14rpx 34rpx rgba(15, 23, 42, 0.1);
}

.pickup-card {
  top: 54rpx;
}

.destination-card {
  top: 188rpx;
}

.is-dark .route-card {
  background: rgba(15, 23, 42, 0.82);
}

.route-dot {
  width: 22rpx;
  height: 22rpx;
  flex: 0 0 22rpx;
  border-radius: 999rpx;
}

.pickup-dot {
  background: #16a34a;
}

.destination-dot {
  background: #0ea5e9;
}

.route-label {
  display: block;
  font-size: 22rpx;
  color: #667085;
}

.route-value {
  display: block;
  margin-top: 4rpx;
  font-size: 28rpx;
  font-weight: 850;
  color: currentColor;
}

.is-dark .route-label {
  color: #94a3b8;
}

.car-chip {
  position: absolute;
  right: 44rpx;
  bottom: 34rpx;
  gap: 10rpx;
  padding: 12rpx 18rpx;
  border-radius: 999rpx;
  color: #ffffff;
  font-size: 23rpx;
  font-weight: 800;
  background: #12201b;
}

.hero-copy {
  margin-bottom: 28rpx;
}

.eyebrow {
  display: block;
  color: #16a34a;
  font-size: 22rpx;
  font-weight: 850;
  letter-spacing: 0;
  text-transform: uppercase;
}

.title {
  display: block;
  margin-top: 12rpx;
  font-size: 50rpx;
  font-weight: 950;
  line-height: 1.1;
  color: currentColor;
}

.subtitle {
  display: block;
  margin-top: 16rpx;
  font-size: 27rpx;
  line-height: 1.62;
  color: #5f6b7a;
}

.is-dark .subtitle {
  color: #cbd5e1;
}

.login-card {
  width: 100%;
  padding: 32rpx;
  box-sizing: border-box;
  border: 1rpx solid rgba(15, 23, 42, 0.08);
  border-radius: 16rpx;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 24rpx 70rpx rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(18rpx);
}

.is-dark .login-card {
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(15, 23, 42, 0.84);
}

.card-heading {
  margin-bottom: 24rpx;
}

.card-title {
  display: block;
  font-size: 34rpx;
  font-weight: 900;
  color: currentColor;
}

.card-subtitle {
  display: block;
  margin-top: 8rpx;
  color: #667085;
  font-size: 23rpx;
  line-height: 1.5;
}

.is-dark .card-subtitle {
  color: #94a3b8;
}

.code-row {
  display: grid;
  width: 100%;
  grid-template-columns: minmax(0, 1fr) 184rpx;
  gap: 14rpx;
  align-items: center;
}

.code-button,
.login-button {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
  border: 0;
  border-radius: 12rpx;
  padding: 0;
  font-weight: 850;
}

.code-button {
  height: 80rpx;
  color: #047857;
  font-size: 24rpx;
  background: rgba(22, 163, 74, 0.12);
}

.code-button[disabled] {
  color: #98a2b3;
}

.login-button {
  width: 100%;
  height: 92rpx;
  gap: 12rpx;
  margin-top: 14rpx;
  color: #ffffff;
  font-size: 30rpx;
  background: linear-gradient(135deg, #16a34a, #0ea5e9);
  box-shadow: 0 18rpx 44rpx rgba(14, 165, 233, 0.24);
}

.login-button[disabled] {
  opacity: 0.72;
}

.safe-row {
  justify-content: space-between;
  gap: 12rpx;
  margin-top: 26rpx;
}

.safe-item {
  gap: 8rpx;
  min-width: 0;
  color: #475467;
  font-size: 22rpx;
}

.is-dark .safe-item {
  color: #cbd5e1;
}

:deep(.u-form-item__body__left__content__label) {
  color: #344054 !important;
  white-space: nowrap;
  font-size: 27rpx !important;
  font-weight: 750;
}

:deep(.u-form-item__body__left) {
  flex: 0 0 132rpx !important;
  width: 132rpx !important;
  min-width: 132rpx !important;
}

.is-dark :deep(.u-form-item__body__left__content__label) {
  color: #dbeafe !important;
}

:deep(.u-input) {
  min-width: 0;
}
</style>
