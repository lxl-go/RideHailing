<template>
  <div class="workorder05-person">
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="Plus" @click="openEdit()">新增</el-button>
        <el-button icon="CircleCheck" :disabled="selectedRows.length === 0" @click="batchStatus('enabled')">启用</el-button>
        <el-button icon="CircleClose" :disabled="selectedRows.length === 0" @click="batchStatus('disabled')">禁用</el-button>
        <el-button icon="Delete" :disabled="selectedRows.length === 0" @click="batchStatus('deleted')">删除</el-button>
        <el-button icon="Upload" @click="importVisible = true">导入</el-button>
        <el-button icon="Download" @click="exportRows">导出</el-button>
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" clearable placeholder="姓名/编号/手机号" style="width: 220px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 130px" @change="search">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="searchForm.roleCode" clearable placeholder="全部" style="width: 160px" @change="search">
            <el-option v-for="item in roleOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="search">查询</el-button>
          <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <el-tabs v-model="activeType" @tab-change="search">
        <el-tab-pane label="员工管理" name="staff" />
        <el-tab-pane label="司机管理" name="driver" />
        <el-tab-pane label="用户管理" name="passenger" />
      </el-tabs>

      <el-table v-loading="loading" :data="tableData" row-key="id" style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="48" />
        <el-table-column label="人员编号" prop="personNo" min-width="160" />
        <el-table-column label="姓名" prop="name" width="120" />
        <el-table-column label="手机号" prop="phoneMasked" width="130" />
        <el-table-column label="身份证" prop="idCardMasked" width="170" />
        <el-table-column label="角色" min-width="190">
          <template #default="{ row }">
            <el-tag v-for="role in row.roles" :key="role.roleCode" class="role-tag" type="info">{{ role.roleName }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="车辆" prop="vehicleNo" width="120" />
        <el-table-column label="评分" prop="rating" width="90" />
        <el-table-column label="状态" prop="status" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'enabled' ? 'success' : 'danger'">{{ row.status === 'enabled' ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="openRoles(row)">角色</el-button>
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

    <el-dialog v-model="editVisible" :title="editForm.id ? '编辑人员' : '新增人员'" width="720px">
      <el-form :model="editForm" label-width="110px">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="类型">
              <el-select v-model="editForm.personType" style="width: 100%">
                <el-option label="员工" value="staff" />
                <el-option label="司机" value="driver" />
                <el-option label="乘客" value="passenger" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="姓名"><el-input v-model="editForm.name" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号"><el-input v-model="editForm.phone" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="身份证"><el-input v-model="editForm.idCardNo" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱"><el-input v-model="editForm.email" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="注册日期"><el-date-picker v-model="editForm.registerDate" value-format="YYYY-MM-DD" type="date" style="width: 100%" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="驾驶证"><el-input v-model="editForm.driverLicenseNo" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="车牌号"><el-input v-model="editForm.vehicleNo" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="车型"><el-input v-model="editForm.vehicleType" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="角色">
              <el-select v-model="editForm.roles" multiple style="width: 100%">
                <el-option v-for="item in roleOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="roleVisible" title="角色配置" width="420px">
      <el-select v-model="roleForm.roles" multiple style="width: 100%">
        <el-option v-for="item in roleOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <template #footer>
        <el-button @click="roleVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRoles">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="数据导入" width="760px">
      <el-input v-model="importText" type="textarea" :rows="8" />
      <div class="import-actions">
        <el-button icon="View" @click="previewImport">预检</el-button>
        <el-button type="primary" icon="UploadFilled" @click="commitImport">提交导入</el-button>
      </div>
      <el-alert v-if="importResult" :title="`总数 ${importResult.total}，成功 ${importResult.successCount}，错误 ${importResult.errorCount}`" type="info" show-icon />
      <el-table v-if="importResult?.errors?.length" :data="importResult.errors" size="small" class="error-table">
        <el-table-column label="行号" prop="rowNo" width="80" />
        <el-table-column label="字段" prop="field" width="120" />
        <el-table-column label="错误" prop="message" />
      </el-table>
    </el-dialog>

    <el-drawer v-model="detailVisible" title="人员详情" size="42%">
      <el-descriptions v-if="detail.id" :column="2" border>
        <el-descriptions-item label="编号">{{ detail.personNo }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ typeText(detail.personType) }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ detail.phoneMasked }}</el-descriptions-item>
        <el-descriptions-item label="身份证">{{ detail.idCardMasked }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ detail.email }}</el-descriptions-item>
        <el-descriptions-item label="车辆">{{ detail.vehicleNo }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  assignPersonRoles,
  batchDeleteDrivers,
  batchPersonStatus,
  commitPersonImport,
  createPerson,
  exportPersons,
  getPerson,
  listPersons,
  previewPersonImport,
  updatePerson,
} from '@/api/rideHailing/workorder05'

defineOptions({ name: 'Workorder05Person' })

const roleOptions = [
  { label: '员工', value: 'staff' },
  { label: '乘客', value: 'passenger' },
  { label: '班车司机', value: 'shuttle_driver' },
  { label: '接送司机', value: 'pickup_driver' },
  { label: '顺风车司机', value: 'carpool_driver' },
  { label: '调度员', value: 'dispatcher' },
  { label: '验票员', value: 'ticket_checker' },
  { label: '联系人员', value: 'contact' },
]

const loading = ref(false)
const activeType = ref('staff')
const tableData = ref([])
const selectedRows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const editVisible = ref(false)
const roleVisible = ref(false)
const importVisible = ref(false)
const detailVisible = ref(false)
const importText = ref('[{"personType":"passenger","name":"Passenger Import","phone":"13900002001","idCardNo":"110101199501011234","registerDate":"2026-07-29","roles":["passenger"]}]')
const importResult = ref(null)
const detail = reactive({})
const searchForm = reactive({ keyword: '', status: '', roleCode: '' })
const editForm = reactive(defaultEditForm())
const roleForm = reactive({ personId: 0, roles: [] })

function defaultEditForm() {
  return {
    id: 0,
    personType: activeType.value || 'staff',
    name: '',
    phone: '',
    email: '',
    idCardNo: '',
    driverLicenseNo: '',
    vehicleNo: '',
    vehicleType: '',
    commonAddress: '',
    paymentPreference: '',
    rating: 5,
    status: 'enabled',
    registerDate: new Date().toISOString().slice(0, 10),
    roles: [],
  }
}

const getTableData = async () => {
  loading.value = true
  try {
    const res = await listPersons({ page: page.value, pageSize: pageSize.value, personType: activeType.value, ...searchForm })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
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
  Object.assign(searchForm, { keyword: '', status: '', roleCode: '' })
  search()
}

const handleSelectionChange = (rows) => {
  selectedRows.value = rows
}

const openEdit = (row) => {
  Object.assign(editForm, defaultEditForm(), row || {})
  editForm.personType = row?.personType || activeType.value
  editForm.roles = row?.roles?.map(item => item.roleCode) || []
  editForm.phone = ''
  editForm.idCardNo = ''
  editVisible.value = true
}

const submitEdit = async () => {
  const payload = { ...editForm }
  if (editForm.id) {
    const res = await updatePerson(editForm.id, payload)
    if (res.code === 0) ElMessage.success('已更新')
  } else {
    const res = await createPerson(payload)
    if (res.code === 0) ElMessage.success('已创建')
  }
  editVisible.value = false
  getTableData()
}

const openRoles = (row) => {
  roleForm.personId = row.id
  roleForm.roles = row.roles.map(item => item.roleCode)
  roleVisible.value = true
}

const submitRoles = async () => {
  const res = await assignPersonRoles(roleForm)
  if (res.code === 0) {
    ElMessage.success('角色已保存')
    roleVisible.value = false
    getTableData()
  }
}

const batchStatus = async (status) => {
  const ids = selectedRows.value.map(item => item.id)
  if (status === 'deleted' && activeType.value === 'driver') {
    const res = await batchDeleteDrivers({ ids, reason: 'batch delete drivers' })
    if (res.code === 0) {
      ElMessage.success('batch delete completed')
      getTableData()
    }
    return
  }
  const res = await batchPersonStatus({ ids: selectedRows.value.map(item => item.id), status, reason: status === 'disabled' ? '批量禁用' : '' })
  if (res.code === 0) {
    ElMessage.success('批量操作完成')
    getTableData()
  }
}

const previewImport = async () => {
  const rows = JSON.parse(importText.value)
  const res = await previewPersonImport({ sourceType: 'json', operator: 'admin', rows })
  if (res.code === 0) importResult.value = res.data
}

const commitImport = async () => {
  const rows = JSON.parse(importText.value)
  const res = await commitPersonImport({ sourceType: 'json', operator: 'admin', rows })
  importResult.value = res.data
  if (res.code === 0) {
    ElMessage.success('导入成功')
    importVisible.value = false
    getTableData()
  }
}

const openDetail = async (row) => {
  const res = await getPerson(row.id)
  if (res.code === 0) {
    Object.assign(detail, res.data)
    detailVisible.value = true
  }
}

const exportRows = async () => {
  const res = await exportPersons()
  if (res.code === 0) ElMessage.success(`导出任务已创建：${res.data.taskId}`)
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

const typeText = (type) => ({ staff: '员工', driver: '司机', passenger: '乘客' }[type] || type)

getTableData()
</script>

<style scoped>
.search-form {
  margin-bottom: 8px;
}

.role-tag {
  margin-right: 6px;
  margin-bottom: 4px;
}

.import-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin: 12px 0;
}

.error-table {
  margin-top: 12px;
}
</style>
