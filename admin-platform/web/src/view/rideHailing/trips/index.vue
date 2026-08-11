<template>
  <div class="trip-page">
    <RidePageHeader title="行程审核" subtitle="统一管理待审核、已通过和已驳回行程">
      <el-button type="primary" @click="load">刷新</el-button>
    </RidePageHeader>
    <div class="gva-table-box">
      <el-form :inline="true" :model="query" class="filter">
        <el-form-item label="状态"><el-select v-model="query.status" clearable placeholder="全部" @change="load"><el-option label="待审核" :value="10" /><el-option label="已通过" :value="20" /><el-option label="已驳回" :value="30" /></el-select></el-form-item>
        <el-form-item label="关键词"><el-input v-model="query.keyword" clearable @keyup.enter="load" /></el-form-item>
        <el-button type="primary" @click="load">查询</el-button>
      </el-form>
      <el-table v-loading="loading" :data="rows" row-key="id">
        <el-table-column prop="id" label="行程ID" width="150" />
        <el-table-column label="路线" min-width="240"><template #default="{ row }">{{ row.originName }} → {{ row.destName }}</template></el-table-column>
        <el-table-column prop="publisherId" label="发布人" width="100" />
        <el-table-column prop="departureTime" label="出发时间" width="180" />
        <el-table-column prop="shareCost" label="价格" width="100"><template #default="{ row }">¥{{ row.shareCost ?? '-' }}</template></el-table-column>
        <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="tagType(row.status)">{{ statusText(row.status) }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="220" fixed="right"><template #default="{ row }"><el-button v-if="row.status === 10" link type="success" @click="review(row, true)">通过</el-button><el-button v-if="row.status === 10" link type="danger" @click="review(row, false)">驳回</el-button><el-button link @click="archive(row)">归档</el-button></template></el-table-column>
      </el-table>
      <div class="gva-pagination"><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="load" /></div>
    </div>
    <el-dialog v-model="rejectVisible" title="驳回行程" width="460px"><el-input v-model="reason" type="textarea" :rows="4" placeholder="请输入5-200字驳回原因" /><template #footer><el-button @click="rejectVisible = false">取消</el-button><el-button type="danger" @click="submitReject">确认驳回</el-button></template></el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RidePageHeader from '@/components/RidePageHeader/index.vue'
import { deactivateTrip, listTrips, reviewTrip } from '@/api/rideHailing/trips'

defineOptions({ name: 'RideHailingTrips' })
const loading = ref(false); const rows = ref([]); const total = ref(0); const page = ref(1); const pageSize = ref(20); const rejectVisible = ref(false); const reason = ref(''); const selected = ref(null)
const query = reactive({ status: undefined, keyword: '' })
const load = async () => { loading.value = true; try { const res = await listTrips({ page: page.value, pageSize: pageSize.value, ...query }); rows.value = res.data?.list || []; total.value = res.data?.total || 0 } finally { loading.value = false } }
const review = async (row, approved) => { if (approved) { const res = await reviewTrip(row.id, { approved: true }); if (res?.code !== 0) return; ElMessage.success('审核通过'); await load(); return } selected.value = row; reason.value = ''; rejectVisible.value = true }
const submitReject = async () => { if (reason.value.trim().length < 5) return ElMessage.warning('驳回原因至少5个字'); const res = await reviewTrip(selected.value.id, { approved: false, reason: reason.value }); if (res?.code !== 0) return; rejectVisible.value = false; ElMessage.success('已驳回'); await load() }
const archive = async (row) => { await ElMessageBox.confirm('确认归档该行程吗？', '提示'); await deactivateTrip(row.id, { reason: '管理端归档' }); ElMessage.success('已归档'); await load() }
const statusText = (value) => ({ 10: '待审核', 20: '已通过', 30: '已驳回' }[value] || '未知')
const tagType = (value) => ({ 10: 'warning', 20: 'success', 30: 'danger' }[value] || 'info')
load()
</script>

<style scoped>.filter { margin-bottom: 12px }.gva-pagination { margin-top: 16px; display: flex; justify-content: flex-end }</style>
