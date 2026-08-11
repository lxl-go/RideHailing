<template>
  <view class="page">
    <view class="top-bar">
      <view>
        <text class="title">联系乘客</text>
        <text class="subtitle">订单 {{ orderId || '-' }}</text>
        <text class="connection">{{ connectionText }}</text>
      </view>
      <u-button size="mini" :plain="true" :disabled="!peerMobile" @click="callPeer">电话</u-button>
    </view>

    <scroll-view class="message-list" scroll-y :scroll-into-view="lastMessageId">
      <view v-if="messages.length === 0" class="empty-chat">
        <text class="empty-title">暂无聊天消息</text>
        <text class="empty-tip">可直接发送消息和乘客沟通行程细节</text>
      </view>
      <view
        v-for="(item, index) in messages"
        :id="`msg-${index}`"
        :key="item.client_msg_id || `${item.sent_at || index}-${index}`"
        class="message-row"
        :class="{ mine: isMine(item) }"
      >
        <view class="bubble">
          <text class="content">{{ item.content }}</text>
          <text class="time">{{ formatTime(item.sent_at) }}</text>
        </view>
      </view>
    </scroll-view>

    <view class="composer">
      <u-input v-model="draft" type="text" placeholder="输入消息" border />
      <u-button type="primary" :disabled="!draft.trim()" @click="sendMessage">发送</u-button>
    </view>
  </view>
</template>

<script setup>
import { computed, nextTick, onUnmounted, ref } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { buildSocketUrl, socketHeaders } from '@/utils/request'

const role = 'driver'
const orderId = ref('')
const peerMobile = ref('')
const draft = ref('')
const connected = ref(false)
const connectionStatus = ref('idle')
const messages = ref([])
const userId = ref('')
let socketTask = null
const seenMessageIds = new Set()
const pendingMessages = []
let connectTimer = null

const lastMessageId = computed(() => (messages.value.length ? `msg-${messages.value.length - 1}` : ''))
const connectionText = computed(() => {
  if (!orderId.value) return '未选择订单'
  if (connectionStatus.value === 'connected') return '已连接'
  if (connectionStatus.value === 'failed') return '连接失败，消息已暂存'
  if (connectionStatus.value === 'closed') return '已断开，消息已暂存'
  return '连接中'
})

const formatTime = (value) => {
  if (!value) return ''
  return String(value).replace('T', ' ').slice(5, 16)
}

const normalizeMessage = (raw) => {
  const message = raw && typeof raw === 'object' ? raw : { content: String(raw || '') }
  return {
    type: message.type || 'chat',
    content: String(message.content || message.message || ''),
    sender_role: message.sender_role || message.senderRole || '',
    sender_id: String(message.sender_id || message.senderId || ''),
    client_msg_id: message.client_msg_id || message.clientMsgId || '',
    sent_at: message.sent_at || message.sentAt || message.created_at || new Date().toISOString(),
  }
}

const appendMessage = (message) => {
  const normalized = normalizeMessage(message)
  if (!normalized.content) return
  const dedupeKey =
    normalized.client_msg_id ||
    `${normalized.sender_role}-${normalized.sender_id}-${normalized.sent_at}-${normalized.content}`
  if (seenMessageIds.has(dedupeKey)) return
  seenMessageIds.add(dedupeKey)
  messages.value.push(normalized)
  nextTick(() => {})
}

const isMine = (message) => {
  const messageRole = message?.sender_role || message?.senderRole
  return messageRole === role
}

const readAccessToken = () => {
  try {
    return String(uni.getStorageSync('driverAccessToken') || '')
  } catch {
    return ''
  }
}

const createChatSocket = (url) => {
  if (typeof window !== 'undefined' && typeof window.WebSocket === 'function') {
    const ws = new window.WebSocket(url)
    return {
      onOpen: (handler) => {
        ws.onopen = handler
      },
      onMessage: (handler) => {
        ws.onmessage = handler
      },
      onClose: (handler) => {
        ws.onclose = handler
      },
      onError: (handler) => {
        ws.onerror = handler
      },
      send: ({ data, fail }) => {
        if (ws.readyState !== window.WebSocket.OPEN) {
          fail?.()
          return
        }
        ws.send(data)
      },
      close: () => ws.close(),
    }
  }
  const task = uni.connectSocket({ url, header: socketHeaders() })
  return {
    onOpen: (handler) => {
      if (typeof task.onOpen === 'function') task.onOpen(handler)
      else uni.onSocketOpen(handler)
    },
    onMessage: (handler) => {
      if (typeof task.onMessage === 'function') task.onMessage(handler)
      else uni.onSocketMessage(handler)
    },
    onClose: (handler) => {
      if (typeof task.onClose === 'function') task.onClose(handler)
      else uni.onSocketClose(handler)
    },
    onError: (handler) => {
      if (typeof task.onError === 'function') task.onError(handler)
      else uni.onSocketError(handler)
    },
    send: (options) => {
      if (typeof task.send === 'function') task.send(options)
      else uni.sendSocketMessage(options)
    },
    close: (options) => {
      if (typeof task.close === 'function') task.close(options)
      else uni.closeSocket(options)
    },
  }
}

const connect = () => {
  if (!orderId.value || socketTask) return
  connectionStatus.value = 'connecting'
  userId.value = userId.value || String(uni.getStorageSync('driverUserId') || '')
  const accessToken = readAccessToken()
  const query = `role=${role}&user_id=${encodeURIComponent(userId.value)}&access_token=${encodeURIComponent(accessToken)}`
  const url = buildSocketUrl(`/api/v1/driver/orders/${encodeURIComponent(orderId.value)}/chat/ws?${query}`)
  socketTask = createChatSocket(url)
  clearTimeout(connectTimer)
  connectTimer = setTimeout(() => {
    if (!connected.value) connectionStatus.value = 'failed'
  }, 8000)
  socketTask.onOpen(() => {
    clearTimeout(connectTimer)
    connected.value = true
    connectionStatus.value = 'connected'
    flushPendingMessages()
  })
  socketTask.onMessage((event) => {
    try {
      appendMessage(JSON.parse(event.data))
    } catch {
      appendMessage({ sender_role: 'system', content: String(event.data || ''), sent_at: new Date().toISOString() })
    }
  })
  socketTask.onClose(() => {
    clearTimeout(connectTimer)
    connected.value = false
    connectionStatus.value = 'closed'
    socketTask = null
  })
  socketTask.onError(() => {
    clearTimeout(connectTimer)
    connected.value = false
    connectionStatus.value = 'failed'
    socketTask = null
    uni.showToast({ title: '聊天连接失败', icon: 'none' })
  })
}

const deliverMessage = (payload) => {
  if (!socketTask || !connected.value) {
    pendingMessages.push(payload)
    connect()
    return
  }
  socketTask.send({
    data: JSON.stringify(payload),
    fail: () => {
      pendingMessages.unshift(payload)
      connected.value = false
      connectionStatus.value = 'failed'
      socketTask = null
      uni.showToast({ title: '消息已暂存，连接恢复后发送', icon: 'none' })
    },
  })
}

const flushPendingMessages = () => {
  while (pendingMessages.length && socketTask && connected.value) {
    const payload = pendingMessages.shift()
    socketTask.send({
      data: JSON.stringify(payload),
      fail: () => {
        pendingMessages.unshift(payload)
        connected.value = false
        connectionStatus.value = 'failed'
      },
    })
    if (!connected.value) break
  }
}

const sendMessage = () => {
  const content = draft.value.trim()
  if (!content || !orderId.value) return
  const clientMsgId = `d-${orderId.value}-${Date.now()}`
  const payload = {
    type: 'chat',
    content,
    client_msg_id: clientMsgId,
    sent_at: new Date().toISOString(),
  }
  appendMessage({ ...payload, sender_role: role, sender_id: userId.value })
  deliverMessage(payload)
  draft.value = ''
}

const callPeer = () => {
  if (!peerMobile.value) {
    uni.showToast({ title: '暂无乘客电话', icon: 'none' })
    return
  }
  uni.makePhoneCall({ phoneNumber: String(peerMobile.value) })
}

const closeSocket = () => {
  clearTimeout(connectTimer)
  if (socketTask) {
    socketTask.close({})
    socketTask = null
  }
}

onLoad((query) => {
  orderId.value = String(query?.orderId || '')
  peerMobile.value = decodeURIComponent(String(query?.mobile || ''))
  userId.value = String(query?.userId || query?.user_id || uni.getStorageSync('driverUserId') || '')
  connect()
})

onUnload(closeSocket)
onUnmounted(closeSocket)
</script>

<style scoped>
.page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f4f7fb;
}
.top-bar {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx;
  background: #fff;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.05);
}
.title {
  display: block;
  font-size: 34rpx;
  font-weight: 700;
  color: #1f2937;
}
.subtitle {
  display: block;
  margin-top: 6rpx;
  font-size: 24rpx;
  color: #8a93a6;
}
.connection {
  display: inline-block;
  margin-top: 8rpx;
  padding: 4rpx 12rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
  color: #128c4a;
  background: #e9f8ef;
}
.message-list {
  flex: 1;
  min-height: 0;
  padding: 24rpx;
  box-sizing: border-box;
}
.empty-chat {
  min-height: 360rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #8a93a6;
}
.empty-title {
  font-size: 30rpx;
  color: #344054;
}
.empty-tip {
  margin-top: 12rpx;
  font-size: 24rpx;
}
.message-row {
  display: flex;
  margin-bottom: 18rpx;
}
.message-row.mine {
  justify-content: flex-end;
}
.bubble {
  max-width: 520rpx;
  padding: 18rpx 22rpx;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 6rpx 20rpx rgba(16, 24, 40, 0.05);
}
.mine .bubble {
  background: #16a34a;
}
.content {
  display: block;
  font-size: 28rpx;
  line-height: 1.5;
  color: #1f2937;
  word-break: break-word;
}
.mine .content,
.mine .time {
  color: #fff;
}
.time {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: #98a2b3;
}
.composer {
  flex-shrink: 0;
  display: flex;
  gap: 16rpx;
  align-items: center;
  padding: 18rpx 24rpx calc(18rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -6rpx 20rpx rgba(16, 24, 40, 0.06);
}
</style>
