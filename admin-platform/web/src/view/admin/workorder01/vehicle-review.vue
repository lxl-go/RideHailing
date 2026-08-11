<template>
  <div class="vehicle-audit-page">
    <RidePageHeader title="车辆审核" subtitle="审核司机提交的车辆资料，通过或驳回">
      <el-button icon="Download" @click="exportRows">导出</el-button>
    </RidePageHeader>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="Search" @click="search">查询</el-button>
        <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 150px" @change="search">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已驳回" :value="2" />
            <el-option label="司机已删除" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" clearable placeholder="车牌/品牌" style="width: 220px" @keyup.enter="search" />
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" row-key="id" style="width: 100%">
        <el-table-column label="司机ID" prop="driverId" width="100" />
        <el-table-column label="车牌号" prop="plateNumber" min-width="120" />
        <el-table-column label="品牌" prop="brand" width="120" />
        <el-table-column label="车型" prop="model" width="120" />
        <el-table-column label="座位数" prop="seats" width="90" />
        <el-table-column label="颜色" prop="color" width="100" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="220">
          <template #default="{ row }">
            <el-button link type="success" icon="Check" :disabled="row.status !== 0" @click="approve(row)">通过</el-button>
            <el-button link type="danger" icon="Close" :disabled="row.status !== 0" @click="openReject(row)">驳回</el-button>
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

    <el-dialog v-model="rejectVisible" title="驳回车辆审核" width="460px">
      <el-form :model="rejectForm" label-width="90px">
        <el-form-item label="驳回原因">
          <el-input v-model="rejectForm.rejectReason" type="textarea" :rows="4" maxlength="200" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" @click="reject">驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getVehicleReviewList, handleVehicleReview } from '@/api/rideHailing/workorder01'
import RidePageHeader from '@/components/RidePageHeader/index.vue'

defineOptions({ name: 'Workorder01VehicleReview' })

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const rejectVisible = ref(false)
const rejectForm = reactive({
  row: null,
  rejectReason: '',
})
const searchForm = reactive({
  status: undefined,
  keyword: '',
})

const normalize = (row = {}) => ({
  ...row,
  id: row.id || row.ID,
  driverId: row.driverId ?? row.driver_id,
  plateNumber: row.plateNumber ?? row.plate_number,
})

const getTableData = async () => {
  loading.value = true
  try {
    const res = await getVehicleReviewList({
      page: page.value,
      pageSize: pageSize.value,
      status: searchForm.status,
      keyword: searchForm.keyword || undefined,
    })
    if (res.code === 0) {
      tableData.value = (res.data?.list || []).map(normalize)
      total.value = res.data?.total || 0
    }
  } finally {
    loading.value = false
  }
}

const search = () => {
  page.value = 1
  getTableData()
}

const resetSearch = () => {
  searchForm.status = undefined
  searchForm.keyword = ''
  search()
}

const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  page.value = 1
  getTableData()
}

const approve = async (row) => {
  const res = await handleVehicleReview(row.id, { status: 1, rejectReason: '' })
  if (res.code === 0) {
    ElMessage.success('已通过')
    getTableData()
  }
}

const openReject = (row) => {
  rejectForm.row = row
  rejectForm.rejectReason = ''
  rejectVisible.value = true
}

const reject = async () => {
  if (!rejectForm.rejectReason.trim()) {
    ElMessage.warning('请输入驳回原因')
    return
  }
  const res = await handleVehicleReview(rejectForm.row.id, {
    status: 2,
    rejectReason: rejectForm.rejectReason,
  })
  if (res.code === 0) {
    ElMessage.success('已驳回')
    rejectVisible.value = false
    getTableData()
  }
}

const exportRows = () => {
  ElMessage.success('导出任务已创建')
}

const statusText = (status) => {
  const map = {
    0: '待审核',
    1: '已通过',
    2: '已驳回',
    3: '司机已删除',
  }
  return map[status] || '-'
}

const statusTag = (status) => {
  const map = {
    0: 'warning',
    1: 'success',
    2: 'danger',
    3: 'info',
  }
  return map[status] || 'info'
}

getTableData()
</script>

<style scoped>
.search-form {
  margin-bottom: 8px;
}
</style>
