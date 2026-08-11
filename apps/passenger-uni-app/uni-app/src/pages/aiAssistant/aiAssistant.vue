<template>
  <view class="page">
    <view class="hero">
      <text class="title">出行助手</text>
      <text class="sub">可查询路线建议、雨天绕行和积水上报</text>
    </view>

    <view class="card">
      <view class="card-title">快速提问</view>
      <view class="quick-tags">
        <view class="tag" v-for="item in quickQuestions" :key="item" @click="fill(item)">{{ item }}</view>
      </view>
      <view class="chat-input">
        <u-input v-model="input" placeholder="输入你的出行问题" border />
        <u-button type="primary" :loading="loading" @click="send">发送</u-button>
      </view>
    </view>

    <view class="card">
      <view class="card-title">路线与天气</view>
      <u-form :model="route" label-position="top">
        <u-form-item label="出发地"><u-input v-model="route.origin" placeholder="如 静安寺" border /></u-form-item>
        <u-form-item label="目的地"><u-input v-model="route.destination" placeholder="如 虹桥火车站" border /></u-form-item>
        <u-form-item label="天气情况"><u-input v-model="route.weather" placeholder="如 暴雨黄色预警" border /></u-form-item>
      </u-form>
      <u-button type="primary" :plain="true" :loading="routeLoading" @click="plan">生成建议</u-button>
    </view>

    <view class="card">
      <view class="card-title">对话记录</view>
      <view class="chat-box">
        <view v-for="(msg, i) in messages" :key="i" class="msg" :class="msg.role">
          <text class="msg-text">{{ msg.content }}</text>
        </view>
      </view>
      <u-alert v-if="errorText" type="warning" :description="errorText" />
    </view>
  </view>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { chatAI, chatWithRoute, planRainRoute } from '@/api/ai'

const quickQuestions = ['今天从静安到虹桥怎么走更稳妥', '暴雨天建议几点出发', '帮我看看有没有积水风险']
const input = ref('')
const loading = ref(false)
const routeLoading = ref(false)
const errorText = ref('')
const messages = ref([{ role: 'ai', content: '你好，我可以帮你做路线建议和雨天出行提醒。' }])
const route = reactive({ city: '上海', origin: '', destination: '', weather: '', avoid: '积水路段', preference: '优先安全', userRole: 'passenger' })

const fill = (text) => { input.value = text }

const pushAI = (content) => messages.value.push({ role: 'ai', content })

const send = async () => {
  const text = input.value.trim()
  if (!text || loading.value) return
  messages.value.push({ role: 'user', content: text })
  input.value = ''
  loading.value = true
  errorText.value = ''
  const sessionId = `p-${Date.now()}`
  const hasRoute = route.origin && route.destination
  const res = hasRoute
    ? await chatWithRoute({ chat: { sessionId, text, userRole: 'passenger' }, route: { ...route, sessionId } })
    : await chatAI({ sessionId, text, userRole: 'passenger' })
  loading.value = false
  if (res.code === 0) pushAI(res.data?.answer || res.data?.rawResult || '已收到你的问题。')
  else errorText.value = res.msg || '暂时无法响应'
}

const plan = async () => {
  if (!route.origin || !route.destination) return uni.showToast({ title: '请先填写起终点', icon: 'none' })
  routeLoading.value = true
  const res = await planRainRoute({ ...route, sessionId: `r-${Date.now()}` })
  routeLoading.value = false
  if (res.code === 0) pushAI(res.data?.answer || res.data?.summary || '已生成路线建议')
  else errorText.value = res.msg || '路线建议生成失败'
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.hero { margin-bottom: 20rpx; }
.title { display: block; font-size: 36rpx; font-weight: 700; color: #1f2937; }
.sub { display: block; margin-top: 8rpx; font-size: 24rpx; color: #8a93a6; }
.card { background: #fff; border-radius: 24rpx; padding: 24rpx; margin-bottom: 20rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.card-title { font-size: 30rpx; font-weight: 700; color: #1f2937; margin-bottom: 16rpx; }
.quick-tags { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 20rpx; }
.tag { padding: 14rpx 20rpx; border-radius: 999rpx; background: #f2f6ff; color: #1677ff; font-size: 24rpx; }
.chat-input { display: flex; gap: 16rpx; align-items: center; }
.chat-box { display: flex; flex-direction: column; gap: 14rpx; }
.msg { display: flex; }
.msg.user { justify-content: flex-end; }
.msg.ai { justify-content: flex-start; }
.msg-text { display: inline-block; max-width: 80%; padding: 16rpx 20rpx; border-radius: 18rpx; font-size: 26rpx; line-height: 1.6; }
.msg.user .msg-text { background: #1677ff; color: #fff; }
.msg.ai .msg-text { background: #f2f6ff; color: #1f2937; }
</style>
