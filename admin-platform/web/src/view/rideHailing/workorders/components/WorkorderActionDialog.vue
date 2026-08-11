<template>
  <el-dialog v-model="visible" width="460px" :title="actionTitle">
    <el-form :model="actionForm" label-width="92px">
      <el-form-item label="处理结果">
        <el-tag :type="actionForm.action === 'approve' ? 'success' : 'danger'">
          {{ actionForm.action === 'approve' ? '通过' : '驳回' }}
        </el-tag>
      </el-form-item>
      <el-form-item v-if="actionForm.action === 'reject'" label="驳回原因" required>
        <el-input v-model="actionForm.rejectReason" type="textarea" :rows="4" maxlength="200" show-word-limit />
      </el-form-item>
      <el-form-item label="说明">
        <span class="dialog-tip">当前为占位交互，后端接口完成后可直接接入提交与审计链路。</span>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="$emit('submit', actionForm)">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, reactive, watch } from 'vue'

const props = defineProps({
  modelValue: Boolean,
  action: { type: String, default: 'approve' },
  actionRow: { type: Object, default: null },
})

const emit = defineEmits(['update:modelValue', 'submit'])

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const actionForm = reactive({
  action: 'approve',
  rejectReason: '',
  row: null,
})

watch(
  () => props.modelValue,
  (value) => {
    if (value) {
      actionForm.action = props.action
      actionForm.row = props.actionRow
      actionForm.rejectReason = ''
    }
  }
)

const actionTitle = computed(() => {
  return actionForm.action === 'approve' ? '确认通过' : '确认驳回'
})
</script>

<style scoped>
.dialog-tip {
  color: #606266;
  line-height: 1.6;
}
</style>
