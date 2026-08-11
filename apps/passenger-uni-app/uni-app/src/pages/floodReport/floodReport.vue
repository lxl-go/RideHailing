<template>
  <view class="page">
    <view class="panel">
      <text class="title">积水上报</text>
      <text class="sub">如果路段积水、封路或绕行困难，可以快速上报给平台</text>
    </view>

    <view class="card">
      <u-form :model="form" label-position="top">
        <u-form-item label="地点"><u-input v-model="form.location" placeholder="积水路段或详细地址" border /></u-form-item>
        <u-form-item label="情况描述"><u-input v-model="form.description" placeholder="请描述积水、封闭或拥堵情况" border /></u-form-item>
        <u-form-item label="严重程度"><u-input v-model="form.level" placeholder="轻微 / 中等 / 严重" border /></u-form-item>
      </u-form>
    </view>

    <u-button type="primary" :loading="submitting" @click="submit">提交上报</u-button>
  </view>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { submitFloodReport } from '@/api/ai'

const form = reactive({ location: '', description: '', level: '中等' })
const submitting = ref(false)

const submit = async () => {
  if (!form.location || !form.description) return uni.showToast({ title: '请填写完整信息', icon: 'none' })
  submitting.value = true
  const res = await submitFloodReport(form)
  submitting.value = false
  if (res.code === 0) uni.showToast({ title: '已上报', icon: 'success' })
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.panel { margin-bottom: 20rpx; }
.title { display: block; font-size: 36rpx; font-weight: 700; color: #1f2937; }
.sub { display: block; margin-top: 8rpx; font-size: 24rpx; color: #8a93a6; }
.card { background: #fff; border-radius: 24rpx; padding: 24rpx; margin-bottom: 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
</style>
