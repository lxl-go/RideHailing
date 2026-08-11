<template>
  <view class="page">
    <view class="panel">
      <text class="title">司机工作台</text>
      <text class="subtitle">认证、车辆、行程发布和派单入口</text>
    </view>
  </view>
</template>

<script setup>
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 32rpx;
  background: #111827;
  color: #f9fafb;
}

.panel {
  border: 1px solid rgba(34, 197, 94, 0.35);
  border-radius: 8px;
  padding: 32rpx;
  background: rgba(17, 24, 39, 0.84);
}

.title {
  display: block;
  font-size: 40rpx;
  font-weight: 700;
}

.subtitle {
  display: block;
  margin-top: 16rpx;
  color: #d1d5db;
  font-size: 26rpx;
}
</style>
