<template>
  <view class="page">
    <view class="top">
      <text class="title">我的优惠券</text>
      <text class="sub">出行前领取，可在下单时自动抵扣</text>
    </view>
    <u-empty v-if="!loading && list.length === 0" text="暂无优惠券" />
    <view v-else class="coupon-list">
      <view v-for="item in list" :key="item.id" class="coupon-card">
        <view class="coupon-left"><text class="money">¥{{ item.amount || item.discount || 0 }}</text><text class="name">{{ item.name || item.title || '出行券' }}</text></view>
        <view class="coupon-right"><text class="desc">{{ item.desc || item.description || '适用于顺风车订单' }}</text></view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { listCoupons } from '@/api/profile'

const list = ref([])
const loading = ref(false)

const load = async () => {
  loading.value = true
  const res = await listCoupons()
  loading.value = false
  if (res.code === 0) list.value = res.data?.items || res.data || []
}

onShow(load)
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.top { margin-bottom: 20rpx; }
.title { display: block; font-size: 36rpx; font-weight: 700; color: #1f2937; }
.sub { display: block; margin-top: 8rpx; font-size: 24rpx; color: #8a93a6; }
.coupon-list { display: flex; flex-direction: column; gap: 20rpx; }
.coupon-card { background: linear-gradient(135deg, #fff, #f8fbff); border-radius: 24rpx; padding: 24rpx; display: flex; justify-content: space-between; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.money { display: block; font-size: 44rpx; font-weight: 700; color: #ee0a24; }
.name { display: block; margin-top: 8rpx; font-size: 28rpx; font-weight: 600; color: #1f2937; }
.desc { font-size: 24rpx; color: #8a93a6; }
</style>
