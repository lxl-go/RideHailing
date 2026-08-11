<template>
  <div class="audit-detail">
    <el-descriptions :column="1" border>
      <el-descriptions-item label="Audit ID">{{ detail.id || detail.ID || '-' }}</el-descriptions-item>
      <el-descriptions-item label="User ID">{{ detail.userId || detail.user_id || '-' }}</el-descriptions-item>
      <el-descriptions-item label="Real name">{{ detail.realName || detail.real_name || '-' }}</el-descriptions-item>
      <el-descriptions-item label="Cert type">{{ certTypeText(detail.certType ?? detail.cert_type) }}</el-descriptions-item>
      <el-descriptions-item label="Cert number">{{ maskCert(detail.certNumber || detail.cert_number) }}</el-descriptions-item>
      <el-descriptions-item label="Status">{{ statusText(detail.status) }}</el-descriptions-item>
      <el-descriptions-item label="Submit count">{{ detail.submitCount || detail.submit_count || 0 }}</el-descriptions-item>
      <el-descriptions-item label="Reject reason">{{ detail.rejectReason || detail.reject_reason || '-' }}</el-descriptions-item>
      <el-descriptions-item label="Reviewed at">{{ detail.reviewedAt || detail.reviewed_at || '-' }}</el-descriptions-item>
      <el-descriptions-item label="Review hours">{{ detail.reviewDurationHours || detail.review_duration_hours || 0 }}</el-descriptions-item>
    </el-descriptions>

    <div class="image-grid">
      <div v-for="image in images" :key="image.label" class="image-box">
        <span>{{ image.label }}</span>
        <el-image
          v-if="image.url"
          :src="image.url"
          :preview-src-list="previewImages"
          fit="cover"
        />
        <el-empty v-else description="No image" :image-size="48" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

defineOptions({
  name: 'Workorder01ReviewDetail'
})

const props = defineProps({
  detail: {
    type: Object,
    default: () => ({})
  }
})

const images = computed(() => [
  { label: 'Front image', url: props.detail.frontImageUrl || props.detail.front_image_url },
  { label: 'Back image', url: props.detail.backImageUrl || props.detail.back_image_url },
  { label: 'Handheld image', url: props.detail.handheldImageUrl || props.detail.handheld_image_url }
])

const previewImages = computed(() => images.value.map((item) => item.url).filter(Boolean))

const statusText = (status) => {
  const map = {
    0: 'Pending',
    1: 'Approved',
    2: 'Rejected',
    3: 'Supplement required'
  }
  return map[status] || '-'
}

const certTypeText = (certType) => {
  const map = {
    1: 'ID card',
    2: 'Driver license',
    3: 'Vehicle license'
  }
  return map[certType] || '-'
}

const maskCert = (value) => {
  if (!value) return '-'
  const text = String(value)
  if (text.length <= 8) return text
  return `${text.slice(0, 4)}********${text.slice(-4)}`
}
</script>

<style scoped>
.audit-detail {
  display: grid;
  gap: 16px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 12px;
}

.image-box {
  display: grid;
  gap: 8px;
}

.image-box span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.image-box :deep(.el-image) {
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 6px;
  background: var(--el-fill-color-light);
}
</style>
