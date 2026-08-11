<template>
  <view class="login-page" :class="{ 'is-dark': isDark }">
    <view class="topbar">
      <view class="brand-lockup">
        <view class="brand-mark">
          <u-icon name="car" size="22" color="#0f172a" />
        </view>
        <text class="brand-name">顺路车主</text>
      </view>
      <button class="theme-toggle" type="button" @click="toggleTheme">
        <u-icon :name="isDark ? 'sun' : 'moon'" size="18" :color="isDark ? '#f8fafc' : '#0f172a'" />
      </button>
    </view>

    <view class="hero-panel">
      <view class="shift-copy">
        <text class="eyebrow">Driver console</text>
        <text class="title">准备上线，接住下一单</text>
        <text class="subtitle">登录后进入司机工作台，查看待接订单、定位上报和今日收入。</text>
      </view>

      <view class="status-board">
        <view class="status-row">
          <text class="status-label">今日热区</text>
          <text class="status-value">静安 · 虹桥 · 张江</text>
        </view>
        <view class="meter">
          <view class="meter-fill" />
        </view>
        <view class="quick-stats">
          <view class="stat-item">
            <text class="stat-number">18</text>
            <text class="stat-label">待接</text>
          </view>
          <view class="stat-item">
            <text class="stat-number">4.9</text>
            <text class="stat-label">服务分</text>
          </view>
          <view class="stat-item">
            <text class="stat-number">¥320</text>
            <text class="stat-label">预估</text>
          </view>
        </view>
      </view>
    </view>

    <view class="login-card">
      <view class="card-heading">
        <text class="card-title">手机号验证码登录</text>
        <text class="card-subtitle">用于确认司机身份与车辆服务状态</text>
      </view>

      <u-form :model="form" label-position="left" label-width="132rpx">
        <u-form-item label="手机号">
          <u-input
            v-model="form.mobile"
            type="number"
            maxlength="11"
            placeholder="请输入司机绑定手机号"
            border="surround"
            clearable
            prefixIcon="phone"
          />
        </u-form-item>

        <u-form-item label="验证码">
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
          <text>{{ loggingIn ? '登录中...' : '进入司机工作台' }}</text>
          <u-icon name="arrow-right" size="18" color="#101828" />
        </button>
      </u-form>

      <view class="assurance-strip">
        <view class="assurance-item">
          <u-icon name="checkmark-circle" size="16" color="#0f9f6e" />
          <text>实名司机</text>
        </view>
        <view class="assurance-item">
          <u-icon name="map" size="16" color="#0f9f6e" />
          <text>定位保护</text>
        </view>
        <view class="assurance-item">
          <u-icon name="rmb-circle" size="16" color="#0f9f6e" />
          <text>收入清晰</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { reactive, ref, onBeforeUnmount } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { loginDriver, sendDriverLoginCode } from '@/api/auth'
import { useUserStore } from '@/store/user'
import { buildAuthenticatedLoginRedirect } from '@/utils/authNavigation.mjs'

const userStore = useUserStore()
const form = reactive({ mobile: '', code: '' })
const sending = ref(false)
const loggingIn = ref(false)
const countdown = ref(0)
const isDark = ref(uni.getStorageSync('driverLoginTheme') === 'dark')
let timer = null

const validMobile = () => /^1\d{10}$/.test(form.mobile)

const toggleTheme = () => {
  isDark.value = !isDark.value
  uni.setStorageSync('driverLoginTheme', isDark.value ? 'dark' : 'light')
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
    tokenKey: 'driverAccessToken',
    homeUrl: '/pages/home/home',
  })
  if (!redirect.shouldRedirect) return
  console.info('[driver-nav] skip login because session exists')
  uni.reLaunch({ url: redirect.url })
}

const sendCode = async () => {
  if (!validMobile()) return uni.showToast({ title: '请输入正确的手机号', icon: 'none' })
  sending.value = true
  const res = await sendDriverLoginCode(form.mobile)
  sending.value = false
  if (res.code === 0) {
    uni.showToast({ title: '验证码已发送', icon: 'success' })
    startCountdown()
  }
}

const login = async () => {
  if (!validMobile() || !form.code) return uni.showToast({ title: '请填写手机号和验证码', icon: 'none' })
  loggingIn.value = true
  const res = await loginDriver(form.mobile, form.code)
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
  padding: 56rpx 32rpx 40rpx;
  box-sizing: border-box;
  overflow-x: hidden;
  color: #101828;
  background:
    linear-gradient(145deg, rgba(255, 183, 77, 0.22), transparent 28%),
    linear-gradient(195deg, rgba(20, 184, 166, 0.18), transparent 36%),
    #f5f7f4;
}

.login-page.is-dark {
  color: #f8fafc;
  background:
    linear-gradient(145deg, rgba(245, 158, 11, 0.18), transparent 30%),
    linear-gradient(215deg, rgba(20, 184, 166, 0.18), transparent 38%),
    #111827;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 38rpx;
}

.brand-lockup {
  display: flex;
  align-items: center;
  gap: 16rpx;
  min-width: 0;
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
  background: #f59e0b;
  padding: 0;
}

.theme-toggle {
  background: rgba(255, 255, 255, 0.72);
  flex: 0 0 64rpx;
  overflow: hidden;
  font-size: 0;
  line-height: 1;
  box-shadow: 0 10rpx 30rpx rgba(15, 23, 42, 0.08);
}

.theme-toggle::after {
  border: 0;
}

.is-dark .theme-toggle {
  background: rgba(255, 255, 255, 0.12);
}

.brand-name {
  font-size: 28rpx;
  font-weight: 700;
  color: currentColor;
}

.hero-panel {
  position: relative;
  display: grid;
  gap: 28rpx;
  margin-bottom: 34rpx;
}

.eyebrow {
  display: block;
  margin-bottom: 14rpx;
  color: #0f766e;
  font-size: 22rpx;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.is-dark .eyebrow {
  color: #5eead4;
}

.title {
  display: block;
  max-width: 640rpx;
  font-size: 52rpx;
  font-weight: 900;
  line-height: 1.08;
  color: currentColor;
}

.subtitle {
  display: block;
  max-width: 620rpx;
  margin-top: 18rpx;
  font-size: 27rpx;
  line-height: 1.65;
  color: #5f6b7a;
}

.is-dark .subtitle {
  color: #cbd5e1;
}

.status-board,
.login-card {
  width: 100%;
  box-sizing: border-box;
  border: 1rpx solid rgba(15, 23, 42, 0.08);
  border-radius: 16rpx;
  background: rgba(255, 255, 255, 0.86);
  box-shadow: 0 24rpx 70rpx rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(18rpx);
}

.is-dark .status-board,
.is-dark .login-card {
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(17, 24, 39, 0.86);
  box-shadow: 0 24rpx 70rpx rgba(0, 0, 0, 0.26);
}

.status-board {
  padding: 28rpx;
}

.status-row,
.quick-stats {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18rpx;
}

.status-label,
.stat-label,
.card-subtitle {
  font-size: 23rpx;
  color: #667085;
}

.is-dark .status-label,
.is-dark .stat-label,
.is-dark .card-subtitle {
  color: #94a3b8;
}

.status-value {
  font-size: 25rpx;
  font-weight: 800;
  color: currentColor;
}

.meter {
  height: 12rpx;
  margin: 24rpx 0;
  overflow: hidden;
  border-radius: 999rpx;
  background: rgba(15, 23, 42, 0.1);
}

.meter-fill {
  width: 68%;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #f59e0b, #14b8a6);
}

.stat-item {
  min-width: 0;
}

.stat-number {
  display: block;
  font-size: 31rpx;
  font-weight: 900;
  color: currentColor;
}

.login-card {
  padding: 32rpx;
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
  line-height: 1.5;
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
  font-weight: 800;
}

.code-button {
  height: 80rpx;
  color: #0f766e;
  font-size: 24rpx;
  background: rgba(20, 184, 166, 0.12);
}

.code-button[disabled] {
  color: #98a2b3;
}

.login-button {
  width: 100%;
  height: 92rpx;
  gap: 12rpx;
  margin-top: 14rpx;
  color: #101828;
  font-size: 30rpx;
  background: linear-gradient(135deg, #f59e0b, #facc15);
  box-shadow: 0 18rpx 40rpx rgba(245, 158, 11, 0.3);
}

.login-button[disabled] {
  opacity: 0.72;
}

.assurance-strip {
  display: flex;
  justify-content: space-between;
  gap: 12rpx;
  margin-top: 26rpx;
}

.assurance-item {
  display: flex;
  align-items: center;
  gap: 8rpx;
  min-width: 0;
  font-size: 22rpx;
  color: #475467;
}

.is-dark .assurance-item {
  color: #cbd5e1;
}

:deep(.u-form-item__body__left__content__label) {
  color: #344054 !important;
  white-space: nowrap;
  font-size: 27rpx !important;
  font-weight: 700;
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
