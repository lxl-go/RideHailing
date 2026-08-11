<template>
  <div class="workorder02-shuttle">
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="Plus" @click="openCreate">新增班线</el-button>
        <el-button icon="Search" @click="search">查询</el-button>
        <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
        <el-button icon="Upload" :disabled="selectedRows.length === 0" @click="batchPublish">批量发布</el-button>
        <el-button icon="Download" @click="exportRows">导出</el-button>
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 140px" @change="search">
            <el-option label="草稿" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已停运" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" clearable placeholder="线路/站点/编号" style="width: 220px" @keyup.enter="search" />
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" row-key="id" style="width: 100%" @selection-change="selectedRows = $event">
        <el-table-column type="selection" width="48" />
        <el-table-column label="线路编号" prop="lineCode" width="110" />
        <el-table-column label="线路名称" prop="lineName" min-width="150" />
        <el-table-column label="站点时序" prop="route" min-width="260" show-overflow-tooltip />
        <el-table-column label="发车/到达" width="150">
          <template #default="{ row }">{{ row.departTime }} / {{ row.arriveTime }}</template>
        </el-table-column>
        <el-table-column label="车型" prop="vehicleType" width="120" />
        <el-table-column label="座位" width="120">
          <template #default="{ row }">{{ row.remainSeats }} / {{ row.totalSeats }}</template>
        </el-table-column>
        <el-table-column label="排序" prop="sortNo" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="240">
          <template #default="{ row }">
            <el-button link type="primary" icon="View" @click="openDetail(row)">详情</el-button>
            <el-button link type="primary" icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button link type="success" icon="Upload" :disabled="row.status === 1" @click="publishOne(row)">发布</el-button>
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

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新增班线' : '编辑班线'" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="96px">
        <el-form-item label="线路编号" prop="lineCode">
          <el-input v-model="form.lineCode" maxlength="20" />
        </el-form-item>
        <el-form-item label="线路名称" prop="lineName">
          <el-input v-model="form.lineName" maxlength="40" />
        </el-form-item>
        <el-form-item label="站点时序" prop="route">
          <el-input v-model="form.route" type="textarea" :rows="3" placeholder="示例：静安寺 -> 中山公园 -> 虹桥站" />
        </el-form-item>
        <el-form-item label="发车时间" prop="departTime">
          <el-time-picker v-model="form.departTime" value-format="HH:mm" format="HH:mm" placeholder="发车时间" />
        </el-form-item>
        <el-form-item label="到达时间" prop="arriveTime">
          <el-time-picker v-model="form.arriveTime" value-format="HH:mm" format="HH:mm" placeholder="到达时间" />
        </el-form-item>
        <el-form-item label="车型" prop="vehicleType">
          <el-input v-model="form.vehicleType" maxlength="20" />
        </el-form-item>
        <el-form-item label="总座位" prop="totalSeats">
          <el-input-number v-model="form.totalSeats" :min="1" :max="60" />
        </el-form-item>
        <el-form-item label="余票" prop="remainSeats">
          <el-input-number v-model="form.remainSeats" :min="0" :max="form.totalSeats || 60" />
        </el-form-item>
        <el-form-item label="排序" prop="sortNo">
          <el-input-number v-model="form.sortNo" :min="1" :max="999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" title="班线详情" size="520px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="线路编号">{{ detailRow.lineCode }}</el-descriptions-item>
        <el-descriptions-item label="线路名称">{{ detailRow.lineName }}</el-descriptions-item>
        <el-descriptions-item label="站点时序">{{ detailRow.route }}</el-descriptions-item>
        <el-descriptions-item label="发车时间">{{ detailRow.departTime }}</el-descriptions-item>
        <el-descriptions-item label="预计到达">{{ detailRow.arriveTime }}</el-descriptions-item>
        <el-descriptions-item label="余票">{{ detailRow.remainSeats }} / {{ detailRow.totalSeats }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detailRow.status) }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createShuttleLine,
  exportShuttleLines,
  getShuttleLine,
  listShuttleLines,
  publishShuttleLines,
  updateShuttleLine,
} from '@/api/rideHailing/workorder02'

defineOptions({ name: 'Workorder02Shuttle' })

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const selectedRows = ref([])
const formVisible = ref(false)
const detailVisible = ref(false)
const formMode = ref('create')
const formRef = ref()
const detailRow = ref({})
const searchForm = reactive({
  keyword: '',
  status: undefined,
})
const form = reactive({
  id: undefined,
  lineCode: '',
  lineName: '',
  route: '',
  departTime: '',
  arriveTime: '',
  vehicleType: '',
  totalSeats: 20,
  remainSeats: 20,
  sortNo: 1,
})
const rules = {
  lineCode: [{ required: true, message: '请输入线路编号', trigger: 'blur' }],
  lineName: [{ required: true, message: '请输入线路名称', trigger: 'blur' }],
  route: [{ required: true, message: '请输入站点时序', trigger: 'blur' }],
  departTime: [{ required: true, message: '请选择发车时间', trigger: 'change' }],
  arriveTime: [{ required: true, message: '请选择到达时间', trigger: 'change' }],
  vehicleType: [{ required: true, message: '请输入车型', trigger: 'blur' }],
}

const assignForm = (row = {}) => {
  Object.assign(form, {
    id: row.id,
    lineCode: row.lineCode || '',
    lineName: row.lineName || '',
    route: row.route || '',
    departTime: row.departTime || '',
    arriveTime: row.arriveTime || '',
    vehicleType: row.vehicleType || '',
    totalSeats: row.totalSeats || 20,
    remainSeats: row.remainSeats ?? row.totalSeats ?? 20,
    sortNo: row.sortNo || 1,
  })
}

const getTableData = async () => {
  loading.value = true
  try {
    const res = await listShuttleLines({
      page: page.value,
      pageSize: pageSize.value,
      keyword: searchForm.keyword,
      status: searchForm.status,
    })
    if (res.code === 0) {
      tableData.value = res.data.list
      total.value = res.data.total
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
  searchForm.keyword = ''
  searchForm.status = undefined
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

const openCreate = () => {
  formMode.value = 'create'
  assignForm()
  formVisible.value = true
}

const openEdit = (row) => {
  formMode.value = 'edit'
  assignForm(row)
  formVisible.value = true
}

const openDetail = async (row) => {
  detailRow.value = row
  detailVisible.value = true
  const res = await getShuttleLine(row.id)
  if (res.code === 0 && res.data) {
    detailRow.value = res.data
  }
}

const submitForm = async () => {
  await formRef.value?.validate()
  if (form.remainSeats > form.totalSeats) {
    ElMessage.warning('余票不能大于总座位')
    return
  }
  const res = formMode.value === 'create' ? await createShuttleLine(form) : await updateShuttleLine(form.id, form)
  if (res.code === 0) {
    ElMessage.success('保存成功')
    formVisible.value = false
    getTableData()
  }
}

const publishOne = async (row) => {
  const res = await publishShuttleLines([row.id])
  if (res.code === 0) {
    ElMessage.success('发布成功')
    getTableData()
  }
}

const batchPublish = async () => {
  await ElMessageBox.confirm(`确认发布 ${selectedRows.value.length} 条班线？`, '批量发布', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning',
  })
  const res = await publishShuttleLines(selectedRows.value.map((row) => row.id))
  if (res.code === 0) {
    ElMessage.success('批量发布成功')
    getTableData()
  }
}

const exportRows = async () => {
  const res = await exportShuttleLines()
  if (res.code === 0) {
    ElMessage.success(`导出任务已创建：${res.data.taskId}`)
  }
}

const statusText = (status) => {
  const map = {
    0: '草稿',
    1: '已发布',
    2: '已停运',
  }
  return map[status] || '-'
}

const statusTag = (status) => {
  const map = {
    0: 'info',
    1: 'success',
    2: 'danger',
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
