<template>
  <div class="workorder01-review">
    <RidePageHeader title="认证审核" subtitle="审核司机身份认证资料，通过或驳回申请">
      <el-button icon="Download" @click="exportRows">导出</el-button>
    </RidePageHeader>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="Search" @click="search">查询</el-button>
        <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 140px" @change="search">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已驳回" :value="2" />
            <el-option label="补充材料" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" clearable placeholder="姓名/证件号" style="width: 220px" @keyup.enter="search" />
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" row-key="id" style="width: 100%">
        <el-table-column label="用户ID" prop="userId" width="100" />
        <el-table-column label="姓名" prop="realName" min-width="120" />
        <el-table-column label="证件号" prop="certNumber" min-width="180" />
        <el-table-column label="驾驶证号" prop="driverLicenseNo" min-width="140" />
        <el-table-column label="准驾车型" prop="licenseType" width="110" />
        <el-table-column label="所属城市" prop="city" width="120" />
        <el-table-column label="证件类型" prop="certType" width="120" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" prop="createdAt" width="180" />
        <el-table-column label="操作" fixed="right" width="220">
          <template #default="{ row }">
            <el-button link type="primary" icon="View" @click="openDetail(row)">详情</el-button>
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

    <el-drawer v-model="detailVisible" title="认证详情" size="520px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户ID">{{ detailRow.userId }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detailRow.realName }}</el-descriptions-item>
        <el-descriptions-item label="证件号">{{ detailRow.certNumber }}</el-descriptions-item>
        <el-descriptions-item label="驾驶证号">{{ detailRow.driverLicenseNo }}</el-descriptions-item>
        <el-descriptions-item label="准驾车型">{{ detailRow.licenseType }}</el-descriptions-item>
        <el-descriptions-item label="所属城市">{{ detailRow.city }}</el-descriptions-item>
        <el-descriptions-item label="证件类型">{{ detailRow.certType }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detailRow.status) }}</el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ detailRow.createdAt }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>

    <el-dialog v-model="rejectVisible" title="驳回审核" width="460px">
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
import { approveAudit, getAuditDetail, getAuditList, rejectAudit } from '@/api/rideHailing/workorder01'
import RidePageHeader from '@/components/RidePageHeader/index.vue'

defineOptions({ name: 'Workorder01Review' })

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const detailVisible = ref(false)
const detailRow = ref({})
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
  id: row.id ?? row.ID,
  userId: row.userId ?? row.user_id,
  realName: row.realName ?? row.real_name,
  certNumber: row.certNumber ?? row.cert_number,
  driverLicenseNo: row.driverLicenseNo ?? row.driver_license_no,
  licenseType: row.licenseType ?? row.license_type,
  city: row.city ?? '',
  certType: row.certType ?? row.cert_type,
  createdAt: row.createdAt ?? row.created_at,
})

const getTableData = async () => {
  loading.value = true
  try {
    const res = await getAuditList({
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

const openDetail = async (row) => {
  detailRow.value = row
  detailVisible.value = true
  const res = await getAuditDetail(row.id)
  if (res.code === 0 && res.data) {
    detailRow.value = normalize(res.data)
  }
}

const approve = async (row) => {
  const res = await approveAudit(row.id)
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
  const res = await rejectAudit(rejectForm.row.id, { rejectReason: rejectForm.rejectReason })
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
    3: '补充材料',
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
