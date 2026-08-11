<template>
  <view class="page">
    <view class="panel top">
      <text class="title">司机工作台</text>
      <text class="sub">订单、车辆、认证、定位统一管理</text>
    </view>

    <view class="menu-card">
      <view class="menu-item" @click="navigate('/pages/certification/certification')"><text>司机认证</text><u-icon name="arrow-right" /></view>
      <view class="menu-item" @click="navigate('/pages/vehicle/vehicle')"><text>我的车辆</text><u-icon name="arrow-right" /></view>
      <view class="menu-item" @click="navigate('/pages/publishTrip/publishTrip')"><text>发布行程</text><u-icon name="arrow-right" /></view>
      <view class="menu-item" @click="navigate('/pages/myTrips/myTrips')"><text>我的行程</text><u-icon name="arrow-right" /></view>
      <view class="menu-item" @click="navigate('/pages/trackReplay/trackReplay')"><text>轨迹回放</text><u-icon name="arrow-right" /></view>
    </view>

    <view class="menu-card">
      <view class="menu-item" @click="navigate('/pages/aiAlerts/aiAlerts')"><text>AI 预警</text><u-icon name="arrow-right" /></view>
      <view class="menu-item" @click="navigate('/pages/incomeLedger/incomeLedger')"><text>收入明细</text><u-icon name="arrow-right" /></view>
      <view class="menu-item danger" @click="logout"><text>退出登录</text><u-icon name="arrow-right" /></view>
    </view>
  </view>
</template>

<script setup>
import { useUserStore } from '@/store/user'
import { navigate } from '@/utils/nav'

const userStore = useUserStore()
const logout = () => {
  uni.showModal({ title: '退出登录', content: '确认退出吗？', success: (m) => { if (m.confirm) { userStore.clearSession(); uni.reLaunch({ url: '/pages/login/login' }) } } })
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.panel.top { background: linear-gradient(135deg, #1677ff, #4a9bff); color: #fff; border-radius: 24rpx; padding: 32rpx; }
.title { display: block; font-size: 38rpx; font-weight: 700; }
.sub { display: block; margin-top: 8rpx; font-size: 24rpx; opacity: .85; }
.menu-card { background: #fff; border-radius: 24rpx; padding: 0 24rpx; margin-top: 20rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.menu-item { display: flex; justify-content: space-between; align-items: center; padding: 30rpx 0; font-size: 30rpx; color: #1f2937; }
.menu-item + .menu-item { border-top: 1rpx solid #f0f2f5; }
.menu-item.danger { color: #f5222d; }
</style>
