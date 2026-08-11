<template>
  <view class="page">
    <view class="panel top">
      <text class="title">司机认证</text>
      <text class="sub">完成认证后才可以发布行程并接单</text>
    </view>

    <view v-if="status === 'approved'" class="status-card success">
      <u-icon name="checkmark-circle" size="48" color="#07c160" />
      <text class="status-text">认证已通过</text>
      <text class="status-tip">你可以开始发布行程和接单</text>
      <u-button type="primary" size="mini" @click="goHome">返回工作台</u-button>
    </view>

    <view v-else-if="status === 'reviewing'" class="status-card warn">
      <u-icon name="info-circle" size="48" color="#f5a623" />
      <text class="status-text">认证审核中</text>
      <text class="status-tip">平台正在审核你的资料，请耐心等待</text>
    </view>

    <template v-else>
      <view v-if="status === 'rejected'" class="status-card danger">
        <u-icon name="close-circle" size="48" color="#f56c6c" />
        <text class="status-text">认证未通过</text>
        <text class="status-tip">{{ rejectedReason || '请核对认证资料后重新提交' }}</text>
      </view>

      <view class="card">
        <u-form class="cert-form" :model="form" label-position="left" label-width="170rpx">
          <u-form-item label="真实姓名">
            <u-input v-model="form.real_name" border placeholder="与身份证一致" />
          </u-form-item>
          <u-form-item label="身份证号">
            <u-input v-model="form.id_card" border placeholder="18位身份证号" />
          </u-form-item>
          <u-form-item label="驾驶证号">
            <u-input v-model="form.license_no" border placeholder="驾驶证档案编号" />
          </u-form-item>
          <u-form-item label="准驾车型">
            <u-input v-model="form.license_type" border placeholder="如 C1 / B2" />
          </u-form-item>
          <u-form-item label="所属城市">
            <u-input v-model="form.city" border disabled disabledColor="#f5f7fb" placeholder="实名认证通过后自动填充" />
          </u-form-item>
        </u-form>
      </view>
      <u-button type="primary" :loading="submitting" @click="submit">
        {{ status === 'rejected' ? '重新提交认证' : '提交认证' }}
      </u-button>
    </template>
  </view>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { submitCertification, getCertificationStatus } from '@/api/driver'
import { listMessages, ackMessage } from '@/api/message'

const form = reactive({ real_name: '', id_card: '', license_no: '', license_type: 'C1', city: '' })
const submitting = ref(false)
const status = ref('')
const rejectedReason = ref('')

const fillForm = (data = {}) => {
  form.real_name = data.real_name ?? data.realName ?? form.real_name
  form.id_card = data.id_card_no ?? data.idCardNo ?? form.id_card
  form.license_no = data.license_no ?? data.licenseNo ?? form.license_no
  form.license_type = data.license_type ?? data.licenseType ?? form.license_type
  form.city = data.city ?? form.city
}

const loadStatus = async () => {
  const res = await getCertificationStatus()
  if (res.code === 0) {
    const data = res.data || {}
    status.value = data.status || ''
    rejectedReason.value = data.reject_reason || data.rejectReason || ''
    if (status.value === 'rejected') fillForm(data)
  }
  loadCertificationMessages()
}

const submit = async () => {
  if (!form.real_name || !form.id_card || !form.license_no || !form.license_type) {
    return uni.showToast({ title: '请填写完整资料', icon: 'none' })
  }
  submitting.value = true
  const res = await submitCertification(form)
  submitting.value = false
  if (res.code === 0) {
    form.city = res.data?.city || form.city
    rejectedReason.value = ''
    uni.showToast({ title: '已提交，等待审核', icon: 'success' })
    status.value = 'reviewing'
  }
}

const parsePayload = (message) => {
  try {
    return JSON.parse(message.payload || '{}')
  } catch {
    return {}
  }
}

const loadCertificationMessages = async () => {
  const res = await listMessages()
  if (res.code !== 0) return
  const message = (res.data?.items || []).find((item) => item.topic === 'certification.audit' && parsePayload(item).result === 'rejected')
  if (!message) return
  const payload = parsePayload(message)
  uni.showModal({
    title: message.title || '司机认证已驳回',
    content: payload.rejectReason || '认证资料未通过审核，请修改后重新提交',
    showCancel: false,
    confirmText: '知晓',
    success: async () => {
      await ackMessage(message.id)
    },
  })
}

const goHome = () => uni.switchTab({ url: '/pages/home/home' })

onShow(loadStatus)
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.panel.top { background: linear-gradient(135deg, #1677ff, #4a9bff); color: #fff; border-radius: 24rpx; padding: 32rpx; }
.title { display: block; font-size: 38rpx; font-weight: 700; }
.sub { display: block; margin-top: 8rpx; font-size: 24rpx; opacity: .85; }
.status-card { display: flex; flex-direction: column; align-items: center; gap: 12rpx; background: #fff; border-radius: 24rpx; padding: 48rpx 24rpx; margin-top: 20rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.status-card.danger { border: 1rpx solid rgba(245,108,108,0.2); background: #fffafa; }
.status-text { font-size: 34rpx; font-weight: 700; color: #1f2937; }
.status-tip { font-size: 26rpx; color: #8a93a6; text-align: center; line-height: 1.5; }
.card { background: #fff; border-radius: 24rpx; padding: 24rpx; margin: 20rpx 0; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.cert-form :deep(.u-form-item__body) { min-height: 88rpx; }
.cert-form :deep(.u-form-item__body__left) { flex: 0 0 170rpx; }
.cert-form :deep(.u-form-item__body__left__content__label) { white-space: nowrap; word-break: keep-all; }
.cert-form :deep(.u-form-item__body__right) { min-width: 0; }
</style>
