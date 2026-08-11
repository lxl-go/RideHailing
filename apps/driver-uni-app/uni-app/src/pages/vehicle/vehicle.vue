<template>
  <view class="page">
    <view class="head">
      <view>
        <text class="title">我的车辆</text>
        <text class="subtitle">维护可接单车辆信息</text>
      </view>
      <u-button size="mini" type="primary" :plain="true" @click="startAdd">添加车辆</u-button>
    </view>

    <u-empty v-if="!loading && list.length === 0" text="还没有车辆，添加后等待管理员审核" />

    <view v-else class="card-list">
      <view v-for="v in list" :key="vehicleKey(v)" class="vehicle-card">
        <view class="vehicle-head">
          <view>
            <text class="vehicle-no">{{ displayPlate(v) }}</text>
            <text class="vehicle-desc">{{ displayVehicle(v) }}</text>
          </view>
          <u-tag :text="statusText(v)" :type="statusType(v)" size="mini" />
        </view>
        <view class="vehicle-meta">
          <text>座位 {{ v.seats || 4 }}</text>
          <text>{{ v.color || '颜色未填' }}</text>
        </view>
        <view class="vehicle-actions">
          <u-button v-if="canEdit(v)" size="mini" :plain="true" @click="startEdit(v)">编辑</u-button>
          <u-button v-if="canDelete(v)" size="mini" type="error" :plain="true" @click="removeVehicle(v)">删除</u-button>
        </view>
      </view>
    </view>

    <u-popup v-model:show="showForm" mode="bottom" border-radius="24" :closeable="true">
      <view class="popup">
        <text class="popup-title">{{ editingId ? '编辑车辆' : '添加车辆' }}</text>
        <u-form class="vehicle-form" :model="form" label-position="left" label-width="144rpx">
          <u-form-item label="车牌号">
            <u-input v-model="form.plateNo" border placeholder="如 京A12345" @input="clearFormError" />
          </u-form-item>
          <u-form-item label="品牌">
            <u-input v-model="form.brand" border placeholder="如 比亚迪" @input="clearFormError" />
          </u-form-item>
          <u-form-item label="车型">
            <u-input v-model="form.vehicleType" border placeholder="如 秦 PLUS" @input="clearFormError" />
          </u-form-item>
          <u-form-item label="座位数">
            <u-input v-model.number="form.seats" type="number" border placeholder="4" @input="clearFormError" />
          </u-form-item>
          <u-form-item label="颜色">
            <u-input v-model="form.color" border placeholder="如 白色" @input="clearFormError" />
          </u-form-item>
        </u-form>
        <view v-if="formError" class="form-error">{{ formError }}</view>
        <u-button type="primary" :loading="submitting" :disabled="submitting" @click="submitForm">保存</u-button>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { listVehicles, saveVehicle, updateVehicle, deleteVehicle } from '@/api/vehicle'
import { listMessages, ackMessage } from '@/api/message'
import { normalizeVehicleForView } from '@/utils/vehicleViewModel.mjs'

const list = ref([])
const loading = ref(false)
const showForm = ref(false)
const submitting = ref(false)
const formError = ref('')
const editingId = ref('')
const form = reactive({ plateNo: '', brand: '', model: '', vehicleType: '', seats: 4, color: '' })

const normalizeVehicle = normalizeVehicleForView

const vehicleKey = (item) => `${item.source || 'vehicle'}-${item.auditId || item.id}`
const displayPlate = (item) => normalizeVehicle(item).plateNo || '未填写车牌'
const displayVehicle = (item) => {
  const vehicle = normalizeVehicle(item)
  return [vehicle.brand, vehicle.vehicleType || vehicle.model].filter(Boolean).join(' · ') || '车型未填'
}
const vehicleStatus = (item = {}) => normalizeVehicle(item).status
const statusText = (item) => {
  const status = vehicleStatus(item)
  if (status === 0) return '审核中'
  if (status === 2) return '已驳回'
  if (status === 3) return '司机已删除'
  return '使用中'
}
const statusType = (item) => {
  const status = vehicleStatus(item)
  if (status === 0) return 'warning'
  if (status === 2) return 'error'
  if (status === 3) return 'info'
  return 'success'
}
const canEdit = (item) => normalizeVehicle(item).canEdit === true || (normalizeVehicle(item).source === 'vehicle' && vehicleStatus(item) === 1)
const canDelete = (item) => normalizeVehicle(item).canDelete === true || (normalizeVehicle(item).source === 'vehicle' && vehicleStatus(item) === 1)

const resetForm = (vehicle = {}) => {
  const normalized = normalizeVehicle(vehicle)
  formError.value = ''
  Object.assign(form, {
    plateNo: normalized.plateNo,
    brand: normalized.brand,
    model: normalized.model,
    vehicleType: normalized.vehicleType,
    seats: normalized.seats,
    color: normalized.color,
  })
}

const load = async () => {
  loading.value = true
  const res = await listVehicles()
  loading.value = false
  list.value = res.code === 0 ? (res.data?.items || res.data || []).map(normalizeVehicle) : []
  loadVehicleMessages()
}

const startAdd = () => {
  editingId.value = ''
  resetForm()
  showForm.value = true
}

const startEdit = (vehicle) => {
  if (!canEdit(vehicle)) {
    uni.showToast({ title: '车辆正在审核中，请等待管理员审核', icon: 'none' })
    return
  }
  editingId.value = normalizeVehicle(vehicle).id
  resetForm(vehicle)
  showForm.value = true
}

const buildPayload = () => ({
  plateNo: form.plateNo.replace(/\s+/g, '').toUpperCase(),
  brand: form.brand.trim(),
  model: form.model.trim(),
  vehicleType: form.vehicleType.trim(),
  seats: Number(form.seats) || 4,
  color: form.color.trim(),
})

const platePattern = /^[\u4e00-\u9fa5][A-Z][A-Z0-9挂学警港澳]{5,6}$/
const isTextLegal = (value = '', maxLen, required = false) => {
  const text = value.trim()
  if (required && !text) return false
  if ([...text].length > maxLen) return false
  return !/[\r\n\t]/.test(text)
}

const validateVehiclePayload = (payload) => {
  if (!payload.plateNo) return '请填写车牌号'
  if (!platePattern.test(payload.plateNo)) return '车牌号格式不正确，请填写省份简称+字母+号码，如 京A12345'
  if (!isTextLegal(payload.brand, 64, true)) return '请填写合法品牌'
  if (!isTextLegal(payload.vehicleType, 32)) return '车型不能超过32个字'
  if (!isTextLegal(payload.color, 32)) return '颜色不能超过32个字'
  if (payload.seats < 1 || payload.seats > 9) return '座位数请输入1-9'
  return ''
}

const clearFormError = () => {
  formError.value = ''
}

const submitForm = async () => {
  const payload = buildPayload()
  const validationMessage = validateVehiclePayload(payload)
  if (validationMessage) {
    formError.value = validationMessage
    uni.showToast({ title: validationMessage, icon: 'none' })
    return
  }
  submitting.value = true
  try {
    const res = editingId.value ? await updateVehicle(editingId.value, payload) : await saveVehicle(payload)
    if (res.code === 0) {
      uni.showToast({ title: '已提交审核', icon: 'success' })
      showForm.value = false
      load()
      return
    }
    formError.value = res?.msg || '保存失败，请稍后重试'
  } catch {
    formError.value = '保存失败，请检查网络或稍后重试'
    uni.showToast({ title: formError.value, icon: 'none' })
  } finally {
    submitting.value = false
  }
}

const removeVehicle = (vehicle) => {
  if (!canDelete(vehicle)) {
    uni.showToast({ title: '车辆审核完成前不能删除', icon: 'none' })
    return
  }
  const normalized = normalizeVehicle(vehicle)
  uni.showModal({
    title: '删除车辆',
    content: `确认删除 ${displayPlate(normalized)} 吗？`,
    success: async ({ confirm }) => {
      if (!confirm) return
      const res = await deleteVehicle(normalized.id)
      if (res.code === 0) {
        uni.showToast({ title: '已删除', icon: 'success' })
        load()
      }
    },
  })
}

const parsePayload = (message) => {
  try {
    return JSON.parse(message.payload || '{}')
  } catch {
    return {}
  }
}

const loadVehicleMessages = async () => {
  const res = await listMessages()
  if (res.code !== 0) return
  const message = (res.data?.items || []).find((item) => item.topic === 'vehicle.audit' && parsePayload(item).result === 'rejected')
  if (!message) return
  const payload = parsePayload(message)
  uni.showModal({
    title: message.title || '车辆认证已驳回',
    content: payload.rejectReason || '车辆资料未通过审核，请修改后重新提交',
    showCancel: false,
    confirmText: '知晓',
    success: async () => {
      await ackMessage(message.id)
    },
  })
}

onShow(load)
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20rpx; }
.title { display: block; font-size: 36rpx; font-weight: 700; color: #1f2937; }
.subtitle { display: block; margin-top: 8rpx; font-size: 24rpx; color: #8a93a6; }
.card-list { display: flex; flex-direction: column; gap: 20rpx; }
.vehicle-card { background: #fff; border-radius: 16rpx; padding: 28rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.vehicle-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 20rpx; }
.vehicle-no { display: block; font-size: 32rpx; font-weight: 700; color: #1f2937; }
.vehicle-desc { display: block; margin-top: 8rpx; font-size: 25rpx; color: #64748b; }
.vehicle-meta { display: flex; gap: 24rpx; margin-top: 16rpx; font-size: 26rpx; color: #8a93a6; }
.vehicle-actions { display: flex; justify-content: flex-end; gap: 16rpx; margin-top: 24rpx; min-height: 52rpx; }
.popup { padding: 40rpx 32rpx; }
.popup-title { display: block; font-size: 34rpx; font-weight: 700; margin-bottom: 24rpx; color: #1f2937; }
.vehicle-form :deep(.u-form-item__body__left) { flex: 0 0 144rpx; }
.vehicle-form :deep(.u-form-item__body__left__content__label) { white-space: nowrap; word-break: keep-all; }
.vehicle-form :deep(.u-form-item__body__right) { min-width: 0; }
.form-error { margin: 4rpx 0 20rpx; padding: 18rpx 22rpx; border-radius: 8rpx; background: #fff1f2; color: #be123c; font-size: 24rpx; line-height: 1.5; }
</style>
