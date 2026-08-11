<template>
  <div class="ride-passengers">
    <RidePageHeader title="乘客管理" subtitle="乘客资料、注册信息与基础运营数据">
      <el-button icon="Download" @click="exportRows">导出</el-button>
    </RidePageHeader>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="Search" @click="search">查询</el-button>
        <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" clearable placeholder="姓名/编号/手机号" style="width: 200px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 130px" @change="search">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" row-key="id" style="width: 100%">
        <el-table-column label="用户编号" prop="personNo" min-width="150" />
        <el-table-column label="姓名" prop="name" width="110" />
        <el-table-column label="手机号" prop="phoneMasked" width="140" />
        <el-table-column label="身份证" prop="idCardMasked" width="180" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'enabled' ? 'success' : 'danger'" size="small">{{ row.status === 'enabled' ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="View" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 'enabled'" link type="danger" @click="toggleStatus(row, 'disabled')">禁用</el-button>
            <el-button v-else link type="success" @click="toggleStatus(row, 'enabled')">启用</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-drawer v-model="detailVisible" title="乘客详情" size="480px">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="用户编号">{{ detail.personNo }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ detail.phoneMasked }}</el-descriptions-item>
        <el-descriptions-item label="身份证">{{ detail.idCardMasked }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status === 'enabled' ? '启用' : '禁用' }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listPassengerRecords, getPassengerRecord, batchPassengerStatus } from '@/api/rideHailing/passengers'
import { exportPersons } from '@/api/rideHailing/workorder05'
import RidePageHeader from '@/components/RidePageHeader/index.vue'

defineOptions({ name: 'RideHailingPassengers' })

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const detailVisible = ref(false)
const detail = ref(null)
const searchForm = reactive({ keyword: '', status: '' })

const getTableData = async () => {
  loading.value = true
  try {
    const res = await listPassengerRecords({ page: page.value, pageSize: pageSize.value, ...searchForm })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

const search = () => { page.value = 1; getTableData() }
const resetSearch = () => { Object.assign(searchForm, { keyword: '', status: '' }); search() }
const handleCurrentChange = (val) => { page.value = val; getTableData() }
const handleSizeChange = (val) => { pageSize.value = val; page.value = 1; getTableData() }

const openDetail = async (row) => {
  const res = await getPassengerRecord(row.id)
  if (res.code === 0) { detail.value = res.data; detailVisible.value = true }
}

const toggleStatus = (row, status) => {
  ElMessageBox.confirm(`确认${status === 'enabled' ? '启用' : '禁用'}乘客 ${row.name}？`, '提示', { type: 'warning' }).then(async () => {
    const res = await batchPassengerStatus({ ids: [row.id], status })
    if (res.code === 0) { ElMessage.success('操作成功'); getTableData() }
  }).catch(() => {})
}

const exportRows = async () => {
  const res = await exportPersons()
  if (res.code === 0) ElMessage.success(`导出任务已创建：${res.data.taskId}`)
}

getTableData()
</script>

<style scoped>
.search-form { margin-bottom: 8px; }
</style>
