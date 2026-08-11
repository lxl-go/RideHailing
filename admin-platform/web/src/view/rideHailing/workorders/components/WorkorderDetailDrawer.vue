<template>
  <el-drawer v-model="visible" size="520px" :show-close="false">
    <template #header>
      <div class="drawer-header">
        <span>{{ title }}详情</span>
        <el-button @click="visible = false">关闭</el-button>
      </div>
    </template>
    <el-descriptions :column="1" border>
      <el-descriptions-item v-for="item in detailItems" :key="item.label" :label="item.label">
        {{ item.value }}
      </el-descriptions-item>
    </el-descriptions>
  </el-drawer>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: Boolean,
  title: { type: String, default: '' },
  detailRow: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['update:modelValue'])

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const detailItems = computed(() => {
  return Object.entries(props.detailRow || {}).map(([key, value]) => ({
    label: key,
    value: Array.isArray(value) ? value.join(', ') : value ?? '-',
  }))
})
</script>

<style scoped>
.drawer-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
