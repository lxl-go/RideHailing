<template>
  <view class="page">
    <view class="panel">
      <text class="title">乘客工作台</text>
      <text class="subtitle">行程搜索、订单、评价和优惠券入口</text>
    </view>
  </view>
</template>

<script setup>
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 32rpx;
  background: #0f172a;
  color: #f8fafc;
}

.panel {
  border: 1px solid rgba(148, 163, 184, 0.35);
  border-radius: 8px;
  padding: 32rpx;
  background: rgba(15, 23, 42, 0.84);
}

.title {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
}

.subtitle {
  display: block;
  margin-top: 16rpx;
  color: #cbd5e1;
  font-size: 26rpx;
}
</style>
