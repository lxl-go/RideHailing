<template>
  <div class="dispatch-page">
    <div class="gva-table-box">
      <div class="page-head">
        <div>
          <h2>派单中心与调度规则</h2>
          <p>订单优先级、司机评分、规则配置、派单审计和轨迹回放。</p>
        </div>
        <div class="head-actions">
          <el-button icon="Refresh" @click="refreshAll">刷新</el-button>
          <el-button v-if="can('export')" type="primary" icon="Download" @click="handleExport">导出</el-button>
        </div>
      </div>

      <el-tabs v-model="activeTab" @tab-change="loadActiveTab">
        <el-tab-pane label="订单中心" name="orders" />
        <el-tab-pane label="派单规则" name="configs" />
        <el-tab-pane label="派单审计" name="audits" />
        <el-tab-pane label="轨迹回放" name="tracks" />
      </el-tabs>

      <template v-if="activeTab === 'orders'">
        <el-form :inline="true" :model="orderSearch" class="search-form">
          <el-form-item label="订单号"><el-input v-model="orderSearch.orderNo" clearable style="width: 180px" @keyup.enter="searchOrders" /></el-form-item>
          <el-form-item label="状态">
            <el-select v-model="orderSearch.status" clearable style="width: 130px" @change="searchOrders">
              <el-option label="待支付" value="pending" />
              <el-option label="已支付" value="paid" />
              <el-option label="已完成" value="completed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
          <el-form-item label="车牌"><el-input v-model="orderSearch.plate" clearable style="width: 150px" @keyup.enter="searchOrders" /></el-form-item>
          <el-form-item label="手机号"><el-input v-model="orderSearch.phone" clearable style="width: 150px" @keyup.enter="searchOrders" /></el-form-item>
          <el-form-item>
            <el-button type="primary" icon="Search" @click="searchOrders">查询</el-button>
            <el-button icon="RefreshLeft" @click="resetOrders">重置</el-button>
          </el-form-item>
        </el-form>

        <el-table v-loading="loading" :data="orders" row-key="id" empty-text="暂无派单订单">
          <el-table-column prop="id" label="订单ID" min-width="160" />
          <el-table-column prop="orderNo" label="订单号" min-width="170" />
          <el-table-column prop="serviceType" label="类型" width="100" />
          <el-table-column prop="passengerName" label="乘客" width="120" />
          <el-table-column prop="driverName" label="司机" width="120" />
          <el-table-column prop="driverId" label="司机ID" min-width="150" />
          <el-table-column prop="vehicleNo" label="车辆" width="120" />
          <el-table-column prop="departTime" label="出发时间" min-width="170" />
          <el-table-column prop="status" label="状态" width="110">
            <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="aiRiskLevel" label="AI风险" width="110">
            <template #default="{ row }"><el-tag v-if="row.aiRiskLevel" type="warning">{{ row.aiRiskLevel }}</el-tag><span v-else>-</span></template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button v-if="can('score')" link type="primary" @click="openScore(row)">评分</el-button>
              <el-button v-if="can('reassign')" link type="primary" @click="openReassign(row)">改派</el-button>
              <el-button v-if="can('cancel')" link type="danger" @click="openCancel(row)">取消</el-button>
            </template>
          </el-table-column>
        </el-table>
      </template>

      <template v-else-if="activeTab === 'configs'">
        <div class="gva-btn-list">
          <el-button v-if="can('saveConfig')" type="primary" icon="Plus" @click="openConfig">新增规则</el-button>
        </div>
        <el-table v-loading="loading" :data="configs" row-key="id" empty-text="暂无派单规则">
          <el-table-column prop="configNo" label="规则编号" min-width="170" />
          <el-table-column prop="city" label="城市" width="120" />
          <el-table-column prop="fleetId" label="车队" width="120" />
          <el-table-column prop="hotZone" label="热区" width="120" />
          <el-table-column label="白天权重" min-width="180">
            <template #default="{ row }">距离 {{ row.dayDistanceWeight }} / 评分 {{ row.dayRatingWeight }} / 空闲 {{ row.dayIdleWeight }}</template>
          </el-table-column>
          <el-table-column label="夜间权重" min-width="180">
            <template #default="{ row }">距离 {{ row.nightDistanceWeight }} / 评分 {{ row.nightRatingWeight }} / 空闲 {{ row.nightIdleWeight }}</template>
          </el-table-column>
          <el-table-column label="状态" width="120">
            <template #default="{ row }"><el-tag :type="row.published ? 'success' : 'info'">{{ row.published ? '已发布' : '草稿' }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button v-if="can('publishConfig')" link type="primary" @click="publishConfig(row)">发布</el-button>
              <el-button v-if="can('rollbackConfig')" link @click="rollbackConfig(row)">回滚</el-button>
            </template>
          </el-table-column>
        </el-table>
      </template>

      <template v-else-if="activeTab === 'audits'">
        <el-table v-loading="loading" :data="audits" row-key="id" empty-text="暂无派单审计">
          <el-table-column prop="auditNo" label="审计编号" min-width="180" />
          <el-table-column prop="orderNo" label="订单号" min-width="160" />
          <el-table-column prop="action" label="动作" width="100" />
          <el-table-column prop="driverName" label="司机" width="120" />
          <el-table-column prop="driverId" label="司机ID" min-width="150" />
          <el-table-column prop="score" label="分数" width="100" />
          <el-table-column prop="decisionReason" label="原因" min-width="240" show-overflow-tooltip />
          <el-table-column prop="traceId" label="Trace" min-width="160" show-overflow-tooltip />
        </el-table>
      </template>

      <template v-else>
        <el-form :inline="true" :model="trackSearch" class="search-form">
          <el-form-item label="司机ID"><el-input v-model="trackSearch.driverId" clearable placeholder="字符串ID" style="width: 180px" @keyup.enter="loadTracks" /></el-form-item>
          <el-form-item><el-button type="primary" icon="Search" @click="loadTracks">查询</el-button></el-form-item>
        </el-form>
        <el-table v-loading="loading" :data="tracks" row-key="id" empty-text="暂无轨迹点">
          <el-table-column prop="driverId" label="司机ID" min-width="150" />
          <el-table-column prop="city" label="城市" width="100" />
          <el-table-column prop="fleetId" label="车队" width="120" />
          <el-table-column prop="hotZone" label="热区" width="120" />
          <el-table-column prop="lat" label="纬度" width="120" />
          <el-table-column prop="lng" label="经度" width="120" />
          <el-table-column prop="reportedAt" label="上报时间" min-width="170" />
        </el-table>
      </template>

      <div v-if="activeTab !== 'configs'" class="gva-pagination">
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

    <el-drawer v-model="actionVisible" :title="drawerTitle" size="520px">
      <el-form v-if="actionType === 'score'" :model="scoreForm" label-width="110px">
        <el-form-item label="订单ID"><el-input v-model="scoreForm.orderId" placeholder="字符串ID" /></el-form-item>
        <el-form-item label="城市"><el-input v-model="scoreForm.city" /></el-form-item>
        <el-form-item label="候选司机JSON"><el-input v-model="candidateJSON" type="textarea" :rows="8" /></el-form-item>
        <el-button type="primary" :loading="actionLoading" @click="submitScore">计算评分</el-button>
        <el-alert v-if="scoreResult.auditNo" class="result-box" type="success" :closable="false" :title="`推荐司机 ${scoreResult.selectedDriverId}，分数 ${scoreResult.score}`" />
        <pre v-if="scoreResult.scoreDetail" class="detail-json">{{ scoreResult.scoreDetail }}</pre>
      </el-form>

      <el-form v-else-if="actionType === 'reassign'" :model="reassignForm" label-width="110px">
        <el-form-item label="司机ID"><el-input v-model="reassignForm.driverId" placeholder="字符串ID" /></el-form-item>
        <el-form-item label="司机姓名"><el-input v-model="reassignForm.driverName" /></el-form-item>
        <el-form-item label="车牌"><el-input v-model="reassignForm.vehicleNo" /></el-form-item>
        <el-form-item label="原因"><el-input v-model="reassignForm.reason" /></el-form-item>
        <el-button type="primary" :loading="actionLoading" @click="submitReassign">确认改派</el-button>
      </el-form>

      <el-form v-else :model="cancelForm" label-width="110px">
        <el-form-item label="原因"><el-input v-model="cancelForm.reason" /></el-form-item>
        <el-button type="danger" :loading="actionLoading" @click="submitCancel">确认取消</el-button>
      </el-form>
    </el-drawer>

    <el-dialog v-model="configVisible" title="派单规则" width="720px">
      <el-form :model="configForm" label-width="120px">
        <el-row :gutter="12">
          <el-col :span="12"><el-form-item label="城市"><el-input v-model="configForm.city" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="车队"><el-input v-model="configForm.fleetId" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="热区"><el-input v-model="configForm.hotZone" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="启用"><el-switch v-model="configForm.enabled" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="白天距离"><el-input-number v-model="configForm.dayDistanceWeight" :min="0" :max="1" :step="0.05" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="白天评分"><el-input-number v-model="configForm.dayRatingWeight" :min="0" :max="1" :step="0.05" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="白天空闲"><el-input-number v-model="configForm.dayIdleWeight" :min="0" :max="1" :step="0.05" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="夜间距离"><el-input-number v-model="configForm.nightDistanceWeight" :min="0" :max="1" :step="0.05" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="夜间评分"><el-input-number v-model="configForm.nightRatingWeight" :min="0" :max="1" :step="0.05" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="夜间空闲"><el-input-number v-model="configForm.nightIdleWeight" :min="0" :max="1" :step="0.05" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="configVisible = false">取消</el-button>
        <el-button type="primary" @click="submitConfig">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  cancelDispatchOrder,
  exportDispatch,
  listDispatchAudits,
  listDispatchConfigs,
  listDispatchOrders,
  publishDispatchConfig,
  reassignDispatchOrder,
  replayDispatchTrack,
  rollbackDispatchConfig,
  saveDispatchConfig,
  scoreDrivers,
} from '@/api/rideHailing/workorder11'

defineOptions({ name: 'RideHailingWorkorder11Dispatch' })

const btnAuth = useBtnAuth()
const can = (key) => Object.keys(btnAuth || {}).length === 0 || Boolean(btnAuth[key])
const loading = ref(false)
const actionLoading = ref(false)
const activeTab = ref('orders')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const orders = ref([])
const configs = ref([])
const audits = ref([])
const tracks = ref([])
const actionVisible = ref(false)
const configVisible = ref(false)
const actionType = ref('score')
const currentOrder = ref({})
const candidateJSON = ref('')
const scoreResult = reactive({})
const orderSearch = reactive({ orderNo: '', status: 'paid', plate: '', phone: '' })
const trackSearch = reactive({ driverId: '' })
const scoreForm = reactive({ orderId: '', city: '上海', candidates: [], idempotencyKey: '' })
const reassignForm = reactive({ driverId: '', driverName: '', vehicleNo: '', reason: '运营人工改派', idempotencyKey: '' })
const cancelForm = reactive({ reason: '运营关闭订单', idempotencyKey: '' })
const configForm = reactive(defaultConfig())

const drawerTitle = computed(() => ({ score: '司机评分', reassign: '订单改派', cancel: '取消订单' }[actionType.value]))

function defaultConfig() {
  return { city: '上海', fleetId: '', hotZone: '', dayDistanceWeight: 0.65, dayRatingWeight: 0.25, dayIdleWeight: 0.1, nightDistanceWeight: 0.15, nightRatingWeight: 0.75, nightIdleWeight: 0.1, enabled: true }
}

const loadOrders = async () => {
  loading.value = true
  try {
    const res = await listDispatchOrders({ page: page.value, pageSize: pageSize.value, ...orderSearch })
    if (res.code === 0) {
      orders.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

const loadConfigs = async () => {
  loading.value = true
  try {
    const res = await listDispatchConfigs()
    if (res.code === 0) configs.value = res.data || []
  } finally {
    loading.value = false
  }
}

const loadAudits = async () => {
  loading.value = true
  try {
    const res = await listDispatchAudits({ page: page.value, pageSize: pageSize.value })
    if (res.code === 0) {
      audits.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

const loadTracks = async () => {
  loading.value = true
  try {
    const driverId = String(trackSearch.driverId || '').trim()
    const res = await replayDispatchTrack({ page: page.value, pageSize: pageSize.value, driverId: driverId || undefined })
    if (res.code === 0) {
      tracks.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

const loadActiveTab = () => {
  const loaders = { orders: loadOrders, configs: loadConfigs, audits: loadAudits, tracks: loadTracks }
  loaders[activeTab.value]()
}

const refreshAll = () => loadActiveTab()
const searchOrders = () => { page.value = 1; loadOrders() }
const resetOrders = () => { Object.assign(orderSearch, { orderNo: '', status: 'paid', plate: '', phone: '' }); searchOrders() }

const openScore = (row) => {
  currentOrder.value = row
  actionType.value = 'score'
  Object.assign(scoreResult, {})
  Object.assign(scoreForm, { orderId: String(row.id || ''), city: '上海', idempotencyKey: `score-${row.id}-${Date.now()}` })
  candidateJSON.value = JSON.stringify([{ driverId: String(row.driverId || '2001'), driverName: row.driverName || 'Driver A', city: '上海', online: true, distanceKm: 2, rating: 4.9, idleMinutes: 40 }], null, 2)
  actionVisible.value = true
}

const openReassign = (row) => {
  currentOrder.value = row
  actionType.value = 'reassign'
  Object.assign(reassignForm, { driverId: String(row.driverId || ''), driverName: row.driverName || '', vehicleNo: row.vehicleNo || '', reason: '运营人工改派', idempotencyKey: `reassign-${row.id}-${Date.now()}` })
  actionVisible.value = true
}

const openCancel = (row) => {
  currentOrder.value = row
  actionType.value = 'cancel'
  Object.assign(cancelForm, { reason: '运营关闭订单', idempotencyKey: `cancel-${row.id}-${Date.now()}` })
  actionVisible.value = true
}

const submitScore = async () => {
  actionLoading.value = true
  try {
    const candidates = JSON.parse(candidateJSON.value || '[]').map((item) => ({ ...item, driverId: String(item.driverId || '').trim() }))
    const res = await scoreDrivers({ ...scoreForm, orderId: String(scoreForm.orderId || '').trim(), candidates })
    if (res.code === 0) {
      Object.assign(scoreResult, res.data || {})
      ElMessage.success('评分完成')
      activeTab.value = 'audits'
      loadAudits()
    }
  } finally {
    actionLoading.value = false
  }
}

const submitReassign = async () => {
  actionLoading.value = true
  try {
    const res = await reassignDispatchOrder(currentOrder.value.id, { ...reassignForm, driverId: String(reassignForm.driverId || '').trim() })
    if (res.code === 0) {
      ElMessage.success('改派完成')
      actionVisible.value = false
      loadOrders()
    }
  } finally {
    actionLoading.value = false
  }
}

const submitCancel = async () => {
  actionLoading.value = true
  try {
    const res = await cancelDispatchOrder(currentOrder.value.id, cancelForm)
    if (res.code === 0) {
      ElMessage.success('订单已取消')
      actionVisible.value = false
      loadOrders()
    }
  } finally {
    actionLoading.value = false
  }
}

const openConfig = () => {
  Object.assign(configForm, defaultConfig())
  configVisible.value = true
}

const submitConfig = async () => {
  const res = await saveDispatchConfig(configForm)
  if (res.code === 0) {
    ElMessage.success('规则已保存')
    configVisible.value = false
    loadConfigs()
  }
}

const publishConfig = async (row) => {
  const res = await publishDispatchConfig(row.id)
  if (res.code === 0) {
    ElMessage.success('规则已发布')
    loadConfigs()
  }
}

const rollbackConfig = async (row) => {
  const res = await rollbackDispatchConfig(row.id)
  if (res.code === 0) {
    ElMessage.success('规则已回滚')
    loadConfigs()
  }
}

const handleExport = async () => {
  const res = await exportDispatch()
  if (res.code === 0) ElMessage.success(`导出任务已创建：${res.data?.taskId || '-'}`)
}

const handleCurrentChange = (val) => { page.value = val; loadActiveTab() }
const handleSizeChange = (val) => { pageSize.value = val; page.value = 1; loadActiveTab() }
const statusText = (status) => ({ pending: '待支付', paid: '已支付', completed: '已完成', cancelled: '已取消' }[status] || status)
const statusType = (status) => ({ paid: 'success', completed: 'info', cancelled: 'warning', pending: 'warning' }[status] || 'info')

refreshAll()
</script>

<style scoped>
.dispatch-page {
  padding: 8px 0 24px;
}

.page-head,
.head-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.page-head {
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

.search-form {
  margin-bottom: 8px;
}

.result-box {
  margin-top: 12px;
}

.detail-json {
  max-height: 220px;
  overflow: auto;
  margin-top: 12px;
  padding: 12px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  background: var(--el-fill-color-lighter);
  white-space: pre-wrap;
}
</style>
