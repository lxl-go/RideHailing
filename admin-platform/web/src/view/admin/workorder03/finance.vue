<template>
  <div class="finance-page">
    <RidePageHeader title="财务管理" subtitle="交易流水、退款记录、异常交易和司机收入">
      <el-button v-if="can('export')" icon="Download" @click="exportRows">导出</el-button>
    </RidePageHeader>

    <div class="gva-table-box">
      <div class="summary-grid">
        <el-statistic title="交易笔数" :value="summary.transactionCount" />
        <el-statistic title="交易金额" :value="summary.totalAmount" prefix="¥" :precision="2" />
        <el-statistic title="退款金额" :value="summary.refundAmount" prefix="¥" :precision="2" />
        <el-statistic title="异常交易" :value="summary.abnormalCount" />
        <el-statistic title="司机今日收入" :value="summary.driverIncomeDay" prefix="¥" :precision="2" />
        <el-statistic title="司机周收入" :value="summary.driverIncomeWeek" prefix="¥" :precision="2" />
        <el-statistic title="司机月收入" :value="summary.driverIncomeMonth" prefix="¥" :precision="2" />
        <el-statistic title="司机年收入" :value="summary.driverIncomeYear" prefix="¥" :precision="2" />
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="订单号">
          <el-input v-model="searchForm.orderNo" clearable placeholder="订单号" style="width: 220px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 150px" @change="search">
            <el-option label="成功" value="success" />
            <el-option label="待支付" value="pending" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button v-if="can('search')" type="primary" icon="Search" @click="search">查询</el-button>
          <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
          <el-button v-if="can('abnormal')" icon="Warning" @click="loadAbnormal">异常交易</el-button>
        </el-form-item>
      </el-form>

      <el-tabs v-model="activeTab" @tab-change="getTableData">
        <el-tab-pane label="交易流水" name="transactions" />
        <el-tab-pane label="退款记录" name="refunds" />
      </el-tabs>

      <el-table v-loading="loading" :data="tableData" style="width: 100%">
        <template v-if="activeTab === 'transactions'">
          <el-table-column label="订单号" prop="orderNo" min-width="160" />
          <el-table-column label="司机ID" prop="driverId" width="150" />
          <el-table-column label="乘客ID" prop="passengerId" width="150" />
          <el-table-column label="金额" prop="amount" width="120" />
          <el-table-column label="支付方式" prop="paymentMethod" width="120" />
          <el-table-column label="状态" prop="status" width="130" />
          <el-table-column label="异常" width="100">
            <template #default="{ row }">
              <el-tag :type="row.abnormal ? 'danger' : 'success'">{{ row.abnormal ? '异常' : '正常' }}</el-tag>
            </template>
          </el-table-column>
        </template>
        <template v-else>
          <el-table-column label="订单号" prop="orderNo" min-width="160" />
          <el-table-column label="退款单号" prop="refundNo" min-width="160" />
          <el-table-column label="退款金额" prop="refundAmount" width="130" />
          <el-table-column label="退款状态" prop="status" width="130" />
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
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  exportFinance,
  getFinanceSummary,
  listAbnormalTransactions,
  listFinanceRefunds,
  listFinanceTransactions,
} from '@/api/rideHailing/workorder03'
import RidePageHeader from '@/components/RidePageHeader/index.vue'

defineOptions({ name: 'Workorder03Finance' })

const btnAuth = useBtnAuth()
const can = (key) => Object.keys(btnAuth || {}).length === 0 || Boolean(btnAuth[key])
const loading = ref(false)
const activeTab = ref('transactions')
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const summary = reactive({
  transactionCount: 0,
  totalAmount: 0,
  refundAmount: 0,
  abnormalCount: 0,
  driverIncomeDay: 0,
  driverIncomeWeek: 0,
  driverIncomeMonth: 0,
  driverIncomeYear: 0,
})
const searchForm = reactive({ orderNo: '', status: '' })

const loadSummary = async () => {
  const res = await getFinanceSummary()
  if (res.code === 0 && res.data) Object.assign(summary, res.data)
}

const getTableData = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value, ...searchForm }
    const res = activeTab.value === 'transactions' ? await listFinanceTransactions(params) : await listFinanceRefunds(params)
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

const resetSearch = () => {
  Object.assign(searchForm, { orderNo: '', status: '' })
  search()
}

const loadAbnormal = async () => {
  const res = await listAbnormalTransactions()
  if (res.code === 0) {
    activeTab.value = 'transactions'
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  }
}

const exportRows = async () => {
  const res = await exportFinance()
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

loadSummary()
getTableData()
</script>

<style scoped>
.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.summary-grid :deep(.el-statistic) {
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
}

.search-form {
  margin-bottom: 8px;
}
</style>
