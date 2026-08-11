<template>
  <div class="ride-orders">
    <RidePageHeader title="订单管理" subtitle="查询真实顺风车订单，查看订单详情与状态流转">
      <el-button v-if="can('export')" icon="Download" @click="exportRows">导出</el-button>
    </RidePageHeader>

    <div class="gva-table-box">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="订单号">
          <el-input v-model="searchForm.orderNo" clearable placeholder="订单号" style="width: 200px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 140px" @change="search">
            <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button v-if="can('search')" type="primary" icon="Search" @click="search">查询</el-button>
          <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" row-key="orderNo" style="width: 100%">
        <el-table-column label="订单号" prop="orderNo" min-width="170" />
        <el-table-column label="乘客ID" prop="passengerId" min-width="150" />
        <el-table-column label="司机ID" prop="driverId" min-width="150" />
        <el-table-column label="路线" prop="routeName" min-width="180" />
        <el-table-column label="金额" prop="payAmount" width="110">
          <template #default="{ row }">¥{{ row.payAmount ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" prop="createdAt" min-width="170" />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button v-if="can('detail')" link type="primary" icon="View" @click="openDetail(row)">详情</el-button>
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

    <el-drawer v-model="detailVisible" title="订单详情" size="520px">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="订单号">{{ detail.orderNo }}</el-descriptions-item>
        <el-descriptions-item label="乘客ID">{{ detail.passengerId }}</el-descriptions-item>
        <el-descriptions-item label="司机ID">{{ detail.driverId }}</el-descriptions-item>
        <el-descriptions-item label="路线">{{ detail.routeName }}</el-descriptions-item>
        <el-descriptions-item label="支付金额">¥{{ detail.payAmount }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
      </el-descriptions>
      <el-timeline class="history-list">
        <el-timeline-item v-for="item in history" :key="item.id" :timestamp="formatTime(item.createdAt)">
          {{ statusText(item.fromStatus) }} -> {{ statusText(item.toStatus) }} / {{ item.reason }}
        </el-timeline-item>
      </el-timeline>
    </el-drawer>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { listOrders, getOrderDetail, exportOrders } from '@/api/rideHailing/workorder04'
import RidePageHeader from '@/components/RidePageHeader/index.vue'

defineOptions({ name: 'RideHailingOrders' })

const btnAuth = useBtnAuth()
const can = (key) => Object.keys(btnAuth || {}).length === 0 || Boolean(btnAuth[key])
const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const detailVisible = ref(false)
const detail = ref(null)
const history = ref([])
const searchForm = reactive({ orderNo: '', serviceType: 'carpool', status: '' })
const statusOptions = [
  { label: '待支付', value: 'pending' },
  { label: '已支付', value: 'paid' },
  { label: '接驾中', value: 'picking' },
  { label: '送达中', value: 'delivering' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' },
  { label: '退款中', value: 'refunding' },
  { label: '已退款', value: 'refunded' },
]

const getTableData = async () => {
  loading.value = true
  try {
    const res = await listOrders({ page: page.value, pageSize: pageSize.value, ...searchForm })
    if (res.code === 0) {
      tableData.value = res.data?.list || []
      total.value = res.data?.total || 0
    }
  } finally {
    loading.value = false
  }
}

const search = () => { page.value = 1; getTableData() }
const resetSearch = () => { Object.assign(searchForm, { orderNo: '', serviceType: 'carpool', status: '' }); search() }
const handleCurrentChange = (val) => { page.value = val; getTableData() }
const handleSizeChange = (val) => { pageSize.value = val; page.value = 1; getTableData() }

const openDetail = async (row) => {
  const res = await getOrderDetail(row.orderNo)
  if (res.code === 0) {
    detail.value = res.data.order || res.data
    history.value = res.data.history || []
    detailVisible.value = true
  }
}

const exportRows = async () => {
  const res = await exportOrders({ pageSize: pageSize.value })
  if (res.code === 0) ElMessage.success(`导出任务已创建：${res.data.taskId}`)
}

const statusText = (value) => (statusOptions.find(item => item.value === value)?.label || value || '-')
const statusType = (value) => ({ paid: 'success', picking: 'warning', delivering: 'warning', completed: 'info', cancelled: 'danger', refunded: 'success', refunding: 'warning' }[value] || '')
const formatTime = (value) => value ? new Date(value).toLocaleString() : ''

getTableData()
</script>

<style scoped>
.search-form {
  margin-bottom: 8px;
}

.history-list {
  margin-top: 24px;
}
</style>
