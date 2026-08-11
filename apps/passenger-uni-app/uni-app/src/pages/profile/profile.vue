<template>
  <view class="page">
    <view class="profile-head">
      <u-avatar :src="profile?.avatar" size="120" />
      <view class="head-info">
        <text class="name">{{ profile?.name || profile?.nickname || '乘客用户' }}</text>
        <text class="mobile">{{ profile?.mobile || '未绑定手机' }}</text>
      </view>
    </view>

    <view class="stat-card">
      <view class="stat-item"><text class="stat-num">{{ stats.completed }}</text><text class="stat-label">完成行程</text></view>
      <view class="stat-divider" />
      <view class="stat-item"><text class="stat-num">{{ stats.coupons }}</text><text class="stat-label">优惠券</text></view>
      <view class="stat-divider" />
      <view class="stat-item"><text class="stat-num">{{ stats.points }}</text><text class="stat-label">积分</text></view>
    </view>

    <view class="menu-card">
      <view class="menu-item" @click="navigate('/pages/orders/orders')">
        <u-icon name="order" size="22" color="#1677ff" />
        <text class="menu-text">我的订单</text>
        <u-icon name="arrow-right" color="#c0c4cc" />
      </view>
      <view class="menu-item" @click="navigate('/pages/coupons/coupons')">
        <u-icon name="coupon" size="22" color="#f5222d" />
        <text class="menu-text">优惠券</text>
        <u-icon name="arrow-right" color="#c0c4cc" />
      </view>
      <view class="menu-item" @click="navigate('/pages/floodReport/floodReport')">
        <u-icon name="warning" size="22" color="#f5a623" />
        <text class="menu-text">积水上报</text>
        <u-icon name="arrow-right" color="#c0c4cc" />
      </view>
    </view>

    <view class="menu-card">
      <view class="menu-item" @click="logout">
        <u-icon name="close" size="22" color="#98a2b3" />
        <text class="menu-text">退出登录</text>
        <u-icon name="arrow-right" color="#c0c4cc" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/store/user'
import { listCoupons } from '@/api/profile'
import { navigate } from '@/utils/nav'

const userStore = useUserStore()
const profile = ref(null)
const stats = reactive({ completed: 0, coupons: 0, points: 0 })

const loadProfile = async () => {
  const res = await userStore.loadProfile()
  if (res.code === 0) profile.value = res.data
}

const loadCoupons = async () => {
  const res = await listCoupons()
  if (res.code === 0) stats.coupons = res.data?.total ?? (res.data?.length || 0)
}

const logout = () => {
  uni.showModal({
    title: '退出登录', content: '确认退出当前账号吗？',
    success: (m) => {
      if (m.confirm) {
        userStore.clearSession()
        uni.reLaunch({ url: '/pages/login/login' })
      }
    }
  })
}

onShow(() => { loadProfile(); loadCoupons() })
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.profile-head { display: flex; align-items: center; gap: 24rpx; background: #fff; border-radius: 24rpx; padding: 32rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.head-info { flex: 1; }
.name { display: block; font-size: 36rpx; font-weight: 700; color: #1f2937; }
.mobile { display: block; margin-top: 8rpx; font-size: 26rpx; color: #8a93a6; }
.stat-card { display: flex; align-items: center; background: #fff; border-radius: 24rpx; padding: 32rpx 0; margin-top: 20rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.stat-item { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 8rpx; }
.stat-num { font-size: 40rpx; font-weight: 700; color: #1f2937; }
.stat-label { font-size: 24rpx; color: #8a93a6; }
.stat-divider { width: 1rpx; height: 64rpx; background: #eef1f6; }
.menu-card { background: #fff; border-radius: 24rpx; margin-top: 20rpx; padding: 0 24rpx; box-shadow: 0 8rpx 24rpx rgba(16,24,40,0.06); }
.menu-item { display: flex; align-items: center; gap: 16rpx; padding: 32rpx 0; }
.menu-item + .menu-item { border-top: 1rpx solid #f0f2f5; }
.menu-text { flex: 1; font-size: 30rpx; color: #1f2937; }
</style>
