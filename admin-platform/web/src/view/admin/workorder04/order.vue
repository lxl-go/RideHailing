<template>
  <div class="workorder04-order">
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button v-if="can('search')" type="primary" icon="Search" @click="search">查询</el-button>
        <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
        <el-button v-if="can('refund')" icon="RefreshRight" :disabled="selectedRows.length === 0" @click="openBatchRefund">批量退款</el-button>
        <el-button v-if="can('export')" icon="Download" @click="exportRows">导出</el-button>
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="订单号">
          <el-input v-model="searchForm.orderNo" clearable placeholder="订单号" style="width: 220px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.serviceType" clearable placeholder="全部" style="width: 140px" @change="search">
            <el-option label="顺风车" value="carpool" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 150px" @change="search">
            <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </el-form>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="订单记录" name="orders" />
        <el-tab-pane label="退款记录" name="refunds" />
      </el-tabs>

      <el-table v-loading="loading" :data="tableData" row-key="orderNo" style="width: 100%" @selection-change="handleSelectionChange">
        <template v-if="activeTab === 'orders'">
          <el-table-column type="selection" width="48" />
          <el-table-column label="订单号" prop="orderNo" min-width="170" />
          <el-table-column label="类型" prop="serviceType" width="100">
            <template #default="{ row }">{{ serviceTypeText(row.serviceType) }}</template>
          </el-table-column>
          <el-table-column label="乘客ID" prop="passengerId" min-width="150" />
          <el-table-column label="司机ID" prop="driverId" min-width="150" />
          <el-table-column label="路线" prop="routeName" min-width="180" />
          <el-table-column label="金额" prop="payAmount" width="110" />
          <el-table-column label="状态" prop="status" width="120">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="190" fixed="right">
            <template #default="{ row }">
              <el-button v-if="can('detail')" link type="primary" @click="openDetail(row)">详情</el-button>
              <el-button v-if="can('refund')" link type="primary" @click="openRefund(row)">退款</el-button>
            </template>
          </el-table-column>
        </template>
        <template v-else>
          <el-table-column label="退款单号" prop="refundNo" min-width="170" />
          <el-table-column label="订单号" prop="orderNo" min-width="160" />
          <el-table-column label="退款金额" prop="refundAmount" width="120" />
          <el-table-column label="手续费" prop="cancelFee" width="110" />
          <el-table-column label="审核方式" prop="reviewType" width="110" />
          <el-table-column label="状态" prop="status" width="120">
            <template #default="{ row }">
              <el-tag :type="refundStatusType(row.status)">{{ refundStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button v-if="can('review')" link type="primary" :disabled="row.status !== 'pending'" @click="openReview(row)">复核</el-button>
            </template>
          </el-table-column>
        </template>
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

    <el-drawer v-model="detailVisible" title="订单详情" size="48%">
      <el-descriptions v-if="detail.order" :column="2" border>
        <el-descriptions-item label="订单号">{{ detail.order.orderNo }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ serviceTypeText(detail.order.serviceType) }}</el-descriptions-item>
        <el-descriptions-item label="乘客ID">{{ detail.order.passengerId }}</el-descriptions-item>
        <el-descriptions-item label="司机ID">{{ detail.order.driverId }}</el-descriptions-item>
        <el-descriptions-item label="路线">{{ detail.order.routeName }}</el-descriptions-item>
        <el-descriptions-item label="支付金额">{{ detail.order.payAmount }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.order.status) }}</el-descriptions-item>
      </el-descriptions>
      <el-timeline class="history-list">
        <el-timeline-item v-for="item in detail.history" :key="item.id" :timestamp="formatTime(item.createdAt)">
          {{ statusText(item.fromStatus) }} -> {{ statusText(item.toStatus) }} / {{ item.reason }}
        </el-timeline-item>
      </el-timeline>
    </el-drawer>

    <el-dialog v-model="refundVisible" title="申请退款" width="420px">
      <el-form :model="refundForm" label-width="90px">
        <el-form-item label="订单号"><el-input v-model="refundForm.orderNo" disabled /></el-form-item>
        <el-form-item label="原因"><el-input v-model="refundForm.reason" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRefund">提交</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="reviewVisible" title="人工复核" width="420px">
      <el-form :model="reviewForm" label-width="90px">
        <el-form-item label="退款单号"><el-input v-model="reviewForm.refundNo" disabled /></el-form-item>
        <el-form-item label="结论">
          <el-radio-group v-model="reviewForm.decision">
            <el-radio-button label="approved">通过</el-radio-button>
            <el-radio-button label="rejected">拒绝</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="reviewForm.remark" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewVisible = false">取消</el-button>
        <el-button type="primary" @click="submitReview">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { applyRefund, batchRefund, exportOrders, getOrderDetail, listOrders, listRefunds, reviewRefund } from '@/api/rideHailing/workorder04'

defineOptions({ name: 'Workorder04Order' })

const btnAuth = useBtnAuth()
const can = (key) => Object.keys(btnAuth || {}).length === 0 || Boolean(btnAuth[key])
const loading = ref(false)
const activeTab = ref('orders')
const tableData = ref([])
const selectedRows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const detailVisible = ref(false)
const refundVisible = ref(false)
const reviewVisible = ref(false)
const detail = reactive({ order: null, history: [] })
const searchForm = reactive({ orderNo: '', serviceType: '', status: '' })
const refundForm = reactive({ orderNo: '', reason: '用户申请退款' })
const reviewForm = reactive({ refundNo: '', decision: 'approved', reviewer: 'admin', remark: '' })

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
    const params = { page: page.value, pageSize: pageSize.value, ...searchForm }
    const res = activeTab.value === 'orders' ? await listOrders(params) : await listRefunds(params)
    if (res.code === 0) {
      tableData.value = res.data?.list || []
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
const resetSearch = () => { Object.assign(searchForm, { orderNo: '', serviceType: '', status: '' }); search() }
const handleTabChange = () => { selectedRows.value = []; search() }
const handleSelectionChange = (rows) => { selectedRows.value = rows }

const openDetail = async (row) => {
  const res = await getOrderDetail(row.orderNo)
  if (res.code === 0) {
    detail.order = res.data.order
    detail.history = res.data.history || []
    detailVisible.value = true
  }
}

const openRefund = (row) => {
  refundForm.orderNo = row.orderNo
  refundForm.reason = '用户申请退款'
  refundVisible.value = true
}

const submitRefund = async () => {
  const res = await applyRefund({
    orderNo: refundForm.orderNo,
    reason: refundForm.reason,
    idempotentKey: `admin-${refundForm.orderNo}-${Date.now()}`,
    operator: 'admin',
  })
  if (res.code === 0) {
    ElMessage.success('退款申请已提交')
    refundVisible.value = false
    getTableData()
  }
}

const openBatchRefund = async () => {
  const res = await batchRefund({
    orderNos: selectedRows.value.map(item => item.orderNo),
    reason: '批量退款',
    operator: 'admin',
    idempotentSeed: `batch-${Date.now()}`,
  })
  if (res.code === 0) {
    ElMessage.success(`批量处理完成：${res.data.items.length}条`)
    getTableData()
  }
}

const openReview = (row) => {
  reviewForm.refundNo = row.refundNo
  reviewForm.decision = 'approved'
  reviewForm.remark = ''
  reviewVisible.value = true
}

const submitReview = async () => {
  const res = await reviewRefund(reviewForm)
  if (res.code === 0) {
    ElMessage.success('复核已提交')
    reviewVisible.value = false
    getTableData()
  }
}

const exportRows = async () => {
  const res = await exportOrders({ pageSize: pageSize.value })
  if (res.code === 0) ElMessage.success(`导出任务已创建：${res.data.taskId}`)
}

const handleCurrentChange = (val) => { page.value = val; getTableData() }
const handleSizeChange = (val) => { pageSize.value = val; page.value = 1; getTableData() }
const serviceTypeText = (value) => ({ carpool: '顺风车', shuttle: '班车' }[value] || value || '-')
const statusText = (value) => (statusOptions.find(item => item.value === value)?.label || value || '-')
const statusType = (value) => ({ paid: 'success', picking: 'warning', delivering: 'warning', completed: 'info', cancelled: 'danger', refunded: 'success', refunding: 'warning' }[value] || '')
const refundStatusText = (value) => ({ pending: '待审核', approved: '已通过', rejected: '已拒绝', refunded: '已退款' }[value] || value)
const refundStatusType = (value) => ({ pending: 'warning', approved: 'success', rejected: 'danger', refunded: 'success' }[value] || '')
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
