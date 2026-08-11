<template>
  <div class="marketing-page">
    <div class="gva-table-box">
      <div class="page-head">
        <div>
          <h2>营销中心</h2>
          <p>管理优惠券模板、发放核销记录和拉新活动。</p>
        </div>
        <div class="summary-grid">
          <el-statistic title="模板数量" :value="templateTotal" />
          <el-statistic title="已发券数" :value="couponTotal" />
          <el-statistic title="推荐奖励" :value="referral.totalRewards" />
          <el-statistic title="已发奖励" :value="referral.issuedRewards" />
        </div>
      </div>

      <div class="gva-btn-list">
        <el-button v-if="can('createTemplate')" type="primary" icon="Plus" @click="openTemplate">新增模板</el-button>
        <el-button v-if="can('issueCoupon')" icon="Ticket" @click="openIssue">手动发券</el-button>
        <el-button v-if="can('redeemCoupon')" icon="CircleCheck" @click="openRedeem">核销</el-button>
        <el-button v-if="can('export')" icon="Download" @click="handleExport">导出</el-button>
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" clearable placeholder="名称/编号" style="width: 210px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item v-if="activeTab === 'userCoupons'" label="用户ID">
          <el-input v-model="searchForm.userId" clearable placeholder="字符串ID" style="width: 180px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.couponType" clearable placeholder="全部" style="width: 130px" @change="search">
            <el-option label="现金券" value="cash" />
            <el-option label="折扣券" value="discount" />
            <el-option label="免乘券" value="free_ride" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" clearable placeholder="全部" style="width: 130px" @change="search">
            <el-option v-if="activeTab === 'templates'" label="启用" value="enabled" />
            <el-option v-if="activeTab === 'templates'" label="停用" value="disabled" />
            <el-option v-if="activeTab === 'templates'" label="草稿" value="draft" />
            <el-option v-if="activeTab === 'userCoupons'" label="未使用" value="unused" />
            <el-option v-if="activeTab === 'userCoupons'" label="已使用" value="used" />
            <el-option v-if="activeTab === 'userCoupons'" label="已过期" value="expired" />
            <el-option v-if="activeTab === 'userCoupons'" label="已退款" value="refunded" />
            <el-option v-if="activeTab === 'campaigns'" label="运行中" value="running" />
            <el-option v-if="activeTab === 'campaigns'" label="已暂停" value="paused" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="search">查询</el-button>
          <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <el-tabs v-model="activeTab" @tab-change="search">
        <el-tab-pane label="优惠券模板" name="templates" />
        <el-tab-pane label="发放/核销记录" name="userCoupons" />
        <el-tab-pane label="营销活动" name="campaigns" />
      </el-tabs>

      <el-table v-loading="loading" :data="tableData" row-key="id" style="width: 100%">
        <template v-if="activeTab === 'templates'">
          <el-table-column label="模板编号" prop="couponNo" min-width="170" />
          <el-table-column label="名称" prop="name" min-width="180" />
          <el-table-column label="类型" prop="couponType" width="110">
            <template #default="{ row }">{{ couponTypeText(row.couponType) }}</template>
          </el-table-column>
          <el-table-column label="优惠值" width="110">
            <template #default="{ row }">{{ couponValueText(row) }}</template>
          </el-table-column>
          <el-table-column label="门槛" prop="thresholdAmount" width="100" />
          <el-table-column label="范围" min-width="150">
            <template #default="{ row }">{{ scopeText(row) }}</template>
          </el-table-column>
          <el-table-column label="库存" width="120">
            <template #default="{ row }">{{ row.issuedCount }}/{{ row.totalStock || '不限' }}</template>
          </el-table-column>
          <el-table-column label="状态" prop="status" width="100">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button v-if="can('issueCoupon')" link type="primary" @click="openIssue(row)">发券</el-button>
              <el-button v-if="can('deleteTemplate')" link type="danger" :disabled="row.issuedCount > 0" @click="removeTemplate(row)">删除</el-button>
            </template>
          </el-table-column>
        </template>

        <template v-else-if="activeTab === 'userCoupons'">
          <el-table-column label="券码" prop="couponCode" min-width="190" />
          <el-table-column label="模板编号" prop="couponNo" min-width="170" />
          <el-table-column label="用户ID" prop="userId" min-width="150" />
          <el-table-column label="来源" prop="source" width="110" />
          <el-table-column label="状态" prop="status" width="110">
            <template #default="{ row }">
              <el-tag :type="userCouponStatusType(row.status)">{{ userCouponStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="订单号" prop="orderNo" min-width="160" />
          <el-table-column label="优惠金额" prop="discountAmount" width="120" />
        </template>

        <template v-else>
          <el-table-column label="活动编号" prop="campaignNo" min-width="180" />
          <el-table-column label="活动名称" prop="name" min-width="180" />
          <el-table-column label="渠道" prop="channel" width="120" />
          <el-table-column label="券模板" prop="couponNo" min-width="170" />
          <el-table-column label="状态" prop="status" width="110">
            <template #default="{ row }">
              <el-tag :type="row.status === 'running' ? 'success' : 'warning'">{{ campaignStatusText(row.status) }}</el-tag>
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

    <el-dialog v-model="templateVisible" title="新增优惠券模板" width="720px">
      <el-form :model="templateForm" label-width="110px">
        <el-row :gutter="12">
          <el-col :span="12"><el-form-item label="名称"><el-input v-model="templateForm.name" /></el-form-item></el-col>
          <el-col :span="12">
            <el-form-item label="类型">
              <el-select v-model="templateForm.couponType" style="width: 100%">
                <el-option label="现金券" value="cash" />
                <el-option label="折扣券" value="discount" />
                <el-option label="免乘券" value="free_ride" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12"><el-form-item label="面值"><el-input-number v-model="templateForm.faceValue" :min="0" :precision="2" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="折扣率"><el-input-number v-model="templateForm.discountRate" :min="0" :max="0.99" :step="0.01" :precision="2" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="使用门槛"><el-input-number v-model="templateForm.thresholdAmount" :min="0" :precision="2" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="库存"><el-input-number v-model="templateForm.totalStock" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="城市"><el-input v-model="templateForm.cityScope" /></el-form-item></el-col>
          <el-col :span="12">
            <el-form-item label="业务">
              <el-select v-model="templateForm.serviceScope" style="width: 100%">
                <el-option label="全部" value="all" />
                <el-option label="顺风车" value="carpool" />
                <el-option label="班车" value="shuttle" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12"><el-form-item label="开始日期"><el-date-picker v-model="templateForm.validFrom" value-format="YYYY-MM-DD" type="date" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="结束日期"><el-date-picker v-model="templateForm.validTo" value-format="YYYY-MM-DD" type="date" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="可叠加"><el-switch v-model="templateForm.stackable" /></el-form-item></el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="templateForm.status" style="width: 100%">
                <el-option label="启用" value="enabled" />
                <el-option label="草稿" value="draft" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="templateVisible = false">取消</el-button>
        <el-button type="primary" @click="submitTemplate">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="issueVisible" title="手动发券" width="460px">
      <el-form :model="issueForm" label-width="90px">
        <el-form-item label="模板编号"><el-input v-model="issueForm.couponNo" /></el-form-item>
        <el-form-item label="用户ID"><el-input v-model="issueForm.userId" placeholder="字符串ID，避免大整数精度丢失" /></el-form-item>
        <el-form-item label="来源">
          <el-select v-model="issueForm.source" style="width: 100%">
            <el-option label="手动发放" value="manual" />
            <el-option label="活动领取" value="campaign" />
            <el-option label="推荐奖励" value="referral" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="issueVisible = false">取消</el-button>
        <el-button type="primary" @click="submitIssue">发放</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="redeemVisible" title="核销优惠券" width="460px">
      <el-form :model="redeemForm" label-width="90px">
        <el-form-item label="券码"><el-input v-model="redeemForm.couponCode" /></el-form-item>
        <el-form-item label="订单号"><el-input v-model="redeemForm.orderNo" /></el-form-item>
        <el-form-item label="订单金额"><el-input-number v-model="redeemForm.orderAmount" :min="0" :precision="2" style="width: 100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="redeemVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRedeem">核销</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  createCouponTemplate,
  deleteCouponTemplate,
  exportMarketing,
  getReferralSummary,
  issueCoupon,
  listCampaigns,
  listCouponTemplates,
  listUserCoupons,
  redeemCoupon,
} from '@/api/rideHailing/workorder07'

defineOptions({ name: 'RideHailingMarketing' })

const today = new Date()
const formatDate = (date) => date.toISOString().slice(0, 10)
const btnAuth = useBtnAuth()
const can = (key) => Object.keys(btnAuth || {}).length === 0 || Boolean(btnAuth[key])

const loading = ref(false)
const activeTab = ref('templates')
const tableData = ref([])
const total = ref(0)
const templateTotal = ref(0)
const couponTotal = ref(0)
const page = ref(1)
const pageSize = ref(20)
const templateVisible = ref(false)
const issueVisible = ref(false)
const redeemVisible = ref(false)
const referral = reactive({ totalRewards: 0, issuedRewards: 0, pendingRewards: 0 })
const searchForm = reactive({ keyword: '', userId: '', couponType: '', status: '' })
const templateForm = reactive(defaultTemplateForm())
const issueForm = reactive({ couponNo: '', userId: '', userType: 'passenger', source: 'manual', operator: 'admin' })
const redeemForm = reactive({ couponCode: '', orderNo: '', orderAmount: 0 })

function defaultTemplateForm() {
  return {
    name: '',
    couponType: 'cash',
    faceValue: 20,
    discountRate: 0.8,
    thresholdAmount: 50,
    validFrom: formatDate(today),
    validTo: formatDate(new Date(today.getFullYear(), today.getMonth() + 1, today.getDate())),
    cityScope: 'Beijing',
    serviceScope: 'all',
    timeScope: '',
    stackable: false,
    totalStock: 100,
    status: 'enabled',
  }
}

const loadSummary = async () => {
  const [templateRes, couponRes, referralRes] = await Promise.all([
    listCouponTemplates({ page: 1, pageSize: 1 }),
    listUserCoupons({ page: 1, pageSize: 1 }),
    getReferralSummary(),
  ])
  templateTotal.value = templateRes.data?.total || 0
  couponTotal.value = couponRes.data?.total || 0
  Object.assign(referral, referralRes.data || {})
}

const getTableData = async () => {
  loading.value = true
  try {
    const baseParams = { page: page.value, pageSize: pageSize.value }
    let res
    if (activeTab.value === 'templates') {
      res = await listCouponTemplates({ ...baseParams, keyword: searchForm.keyword, couponType: searchForm.couponType, status: searchForm.status })
    } else if (activeTab.value === 'userCoupons') {
      res = await listUserCoupons({ ...baseParams, userId: searchForm.userId, status: searchForm.status })
    } else {
      res = await listCampaigns({ ...baseParams, keyword: searchForm.keyword, status: searchForm.status })
    }
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
  Object.assign(searchForm, { keyword: '', userId: '', couponType: '', status: '' })
  search()
}

const openTemplate = () => {
  Object.assign(templateForm, defaultTemplateForm())
  templateVisible.value = true
}

const submitTemplate = async () => {
  const res = await createCouponTemplate(templateForm)
  if (res.code === 0) {
    ElMessage.success('优惠券模板已创建')
    templateVisible.value = false
    await loadSummary()
    search()
  }
}

const openIssue = (row) => {
  issueForm.couponNo = row?.couponNo || issueForm.couponNo
  issueVisible.value = true
}

const submitIssue = async () => {
  const res = await issueCoupon({ ...issueForm, userId: String(issueForm.userId || '').trim() })
  if (res.code === 0) {
    ElMessage.success(`已发放券码：${res.data.couponCode}`)
    issueVisible.value = false
    activeTab.value = 'userCoupons'
    await loadSummary()
    search()
  }
}

const openRedeem = () => {
  redeemVisible.value = true
}

const submitRedeem = async () => {
  const res = await redeemCoupon(redeemForm)
  if (res.code === 0) {
    ElMessage.success(`已核销，优惠 ${res.data.discountAmount}`)
    redeemVisible.value = false
    activeTab.value = 'userCoupons'
    search()
  }
}

const removeTemplate = async (row) => {
  await ElMessageBox.confirm('删除前会校验是否已有发放记录，确认继续？', '删除模板', { type: 'warning' })
  const res = await deleteCouponTemplate(row.couponNo)
  if (res.code === 0) {
    ElMessage.success('模板已删除')
    await loadSummary()
    search()
  }
}

const handleExport = async () => {
  const res = await exportMarketing()
  if (res.code === 0) ElMessage.success(`导出任务已创建：${res.data?.taskId || '-'}`)
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

const couponTypeText = (value) => ({ cash: '现金券', discount: '折扣券', free_ride: '免乘券' }[value] || value)
const couponValueText = (row) => row.couponType === 'discount' ? `${(Number(row.discountRate || 0) * 10).toFixed(1)}折` : `¥${row.faceValue || 0}`
const scopeText = (row) => `${row.cityScope || '全部城市'} / ${row.serviceScope || 'all'}`
const statusText = (value) => ({ enabled: '启用', disabled: '停用', draft: '草稿' }[value] || value)
const statusType = (value) => ({ enabled: 'success', disabled: 'danger', draft: 'info' }[value] || '')
const campaignStatusText = (value) => ({ running: '运行中', paused: '已暂停', ended: '已结束' }[value] || value)
const userCouponStatusText = (value) => ({ unused: '未使用', used: '已使用', expired: '已过期', refunded: '已退款' }[value] || value)
const userCouponStatusType = (value) => ({ unused: 'success', used: 'info', expired: 'warning', refunded: 'danger' }[value] || '')

loadSummary()
getTableData()
</script>

<style scoped>
.marketing-page {
  padding: 8px 0 24px;
}

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.page-head h2 {
  margin: 0 0 6px;
  font-size: 24px;
  font-weight: 600;
}

.page-head p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  min-width: 560px;
}

.summary-grid :deep(.el-statistic) {
  min-height: 78px;
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
}

.search-form {
  margin-bottom: 8px;
}

@media (max-width: 1100px) {
  .page-head {
    display: block;
  }

  .summary-grid {
    min-width: 0;
    margin-top: 16px;
  }
}

@media (max-width: 900px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 520px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
