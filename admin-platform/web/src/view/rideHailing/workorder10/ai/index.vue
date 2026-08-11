<template>
  <div class="ai-page">
    <div class="gva-table-box">
      <div class="page-head">
        <div>
          <h2>AI智能出行助手</h2>
          <p>出行问答、暴雨路线规划、积水上报和AI调用审计。</p>
        </div>
        <div class="head-actions">
          <el-button icon="Refresh" @click="refreshAll">刷新</el-button>
          <el-button type="primary" icon="Download" @click="handleExport">导出</el-button>
        </div>
      </div>

      <div class="kpi-grid">
        <el-statistic title="调用总数" :value="summary.totalCalls || 0" />
        <el-statistic title="成功调用" :value="summary.successCalls || 0" />
        <el-statistic title="降级次数" :value="summary.fallbackCalls || 0" />
        <el-statistic title="平均耗时(ms)" :value="summary.avgLatencyMs || 0" />
      </div>

      <el-row :gutter="16" class="action-row">
        <el-col :span="12">
          <el-card shadow="never">
            <template #header>测试问答</template>
            <el-form :model="chatForm" label-width="90px">
              <el-form-item label="会话ID"><el-input v-model="chatForm.sessionId" /></el-form-item>
              <el-form-item label="角色">
                <el-select v-model="chatForm.userRole" style="width: 100%">
                  <el-option label="乘客" value="passenger" />
                  <el-option label="司机" value="driver" />
                  <el-option label="管理员" value="admin" />
                </el-select>
              </el-form-item>
              <el-form-item label="问题">
                <el-input v-model="chatForm.text" type="textarea" :rows="3" />
              </el-form-item>
              <el-button type="primary" icon="Promotion" :loading="chatLoading" @click="submitChat">发送</el-button>
              <el-alert v-if="chatResult.answer" class="result-box" :type="chatResult.fallback ? 'warning' : 'success'" :closable="false" :title="chatResult.answer" />
            </el-form>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never">
            <template #header>暴雨路线</template>
            <el-form :model="routeForm" label-width="90px">
              <el-row :gutter="10">
                <el-col :span="12"><el-form-item label="城市"><el-input v-model="routeForm.city" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="角色"><el-select v-model="routeForm.userRole" style="width: 100%"><el-option label="乘客" value="passenger" /><el-option label="司机" value="driver" /></el-select></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="起点"><el-input v-model="routeForm.origin" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="终点"><el-input v-model="routeForm.destination" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="天气"><el-input v-model="routeForm.weather" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="偏好"><el-input v-model="routeForm.preference" /></el-form-item></el-col>
              </el-row>
              <el-form-item label="避让"><el-input v-model="routeForm.avoid" /></el-form-item>
              <el-button type="primary" icon="Location" :loading="routeLoading" @click="submitRoute">规划</el-button>
              <el-alert v-if="routeResult.rawResult" class="result-box" :type="routeResult.fallback ? 'warning' : 'success'" :closable="false" :title="routeResult.rawResult" />
            </el-form>
          </el-card>
        </el-col>
      </el-row>

      <el-tabs v-model="activeTab" class="data-tabs" @tab-change="loadActiveTab">
        <el-tab-pane label="会话日志" name="conversation" />
        <el-tab-pane label="路线规划" name="route" />
        <el-tab-pane label="积水上报" name="flood" />
      </el-tabs>

      <el-table v-if="activeTab === 'conversation'" v-loading="loading" :data="conversationLogs" row-key="id" empty-text="暂无AI会话记录">
        <el-table-column prop="sessionId" label="会话ID" min-width="150" show-overflow-tooltip />
        <el-table-column prop="userRole" label="角色" width="100" />
        <el-table-column prop="question" label="问题" min-width="220" show-overflow-tooltip />
        <el-table-column prop="success" label="状态" width="110">
          <template #default="{ row }"><el-tag :type="row.success ? 'success' : 'warning'">{{ row.success ? '成功' : '降级' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="latencyMs" label="耗时(ms)" width="110" />
        <el-table-column prop="traceId" label="Trace" min-width="160" show-overflow-tooltip />
      </el-table>

      <el-table v-else-if="activeTab === 'route'" v-loading="loading" :data="routeLogs" row-key="id" empty-text="暂无路线规划记录">
        <el-table-column prop="routePlanNo" label="规划编号" min-width="180" />
        <el-table-column prop="city" label="城市" width="100" />
        <el-table-column prop="origin" label="起点" min-width="150" show-overflow-tooltip />
        <el-table-column prop="destination" label="终点" min-width="150" show-overflow-tooltip />
        <el-table-column prop="success" label="状态" width="110">
          <template #default="{ row }"><el-tag :type="row.success ? 'success' : 'warning'">{{ row.success ? '成功' : '降级' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="rawResult" label="结果" min-width="260" show-overflow-tooltip />
      </el-table>

      <el-table v-else v-loading="loading" :data="floodReports" row-key="id" empty-text="暂无积水上报记录">
        <el-table-column prop="reportNo" label="上报编号" min-width="180" />
        <el-table-column prop="city" label="城市" width="100" />
        <el-table-column prop="locationText" label="位置" min-width="220" show-overflow-tooltip />
        <el-table-column prop="depthCm" label="水深(cm)" width="110" />
        <el-table-column prop="confidence" label="置信度" width="100" />
        <el-table-column prop="auditStatus" label="审核" width="150">
          <template #default="{ row }"><el-tag :type="row.auditStatus === 'pending_manual_review' ? 'warning' : 'success'">{{ auditText(row.auditStatus) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="auditFlood(row, 'confirmed')">确认</el-button>
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
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  auditFloodReport,
  chatAI,
  exportAI,
  getAISummary,
  listAIConversationLogs,
  listAIFloodReports,
  listAIRoutePlanLogs,
  planRainRoute,
} from '@/api/rideHailing/workorder10'

defineOptions({ name: 'RideHailingWorkorder10AI' })

const loading = ref(false)
const chatLoading = ref(false)
const routeLoading = ref(false)
const activeTab = ref('conversation')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const summary = reactive({})
const conversationLogs = ref([])
const routeLogs = ref([])
const floodReports = ref([])
const chatResult = reactive({})
const routeResult = reactive({})
const chatForm = reactive({ sessionId: `admin-${Date.now()}`, userRole: 'admin', userId: 0, text: '请给出今天暴雨出行建议' })
const routeForm = reactive({ sessionId: `route-${Date.now()}`, city: '上海', origin: '静安寺', destination: '虹桥火车站', weather: '暴雨黄色预警', avoid: '积水路段、隧道', preference: '安全优先', userRole: 'passenger' })

const loadSummary = async () => {
  const res = await getAISummary()
  if (res.code === 0) Object.assign(summary, res.data || {})
}

const loadActiveTab = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value }
    if (activeTab.value === 'conversation') {
      const res = await listAIConversationLogs(params)
      if (res.code === 0) {
        conversationLogs.value = res.data.list || []
        total.value = res.data.total || 0
      }
    } else if (activeTab.value === 'route') {
      const res = await listAIRoutePlanLogs(params)
      if (res.code === 0) {
        routeLogs.value = res.data.list || []
        total.value = res.data.total || 0
      }
    } else {
      const res = await listAIFloodReports(params)
      if (res.code === 0) {
        floodReports.value = res.data.list || []
        total.value = res.data.total || 0
      }
    }
  } finally {
    loading.value = false
  }
}

const refreshAll = async () => {
  await Promise.all([loadSummary(), loadActiveTab()])
}

const submitChat = async () => {
  chatLoading.value = true
  try {
    const res = await chatAI(chatForm)
    if (res.code === 0) {
      Object.assign(chatResult, res.data || {})
      ElMessage.success(res.data?.fallback ? 'AI已降级返回' : 'AI问答已完成')
      refreshAll()
    }
  } finally {
    chatLoading.value = false
  }
}

const submitRoute = async () => {
  routeLoading.value = true
  try {
    const res = await planRainRoute(routeForm)
    if (res.code === 0) {
      Object.assign(routeResult, res.data || {})
      ElMessage.success(res.data?.fallback ? '路线规划已降级返回' : '路线规划已完成')
      activeTab.value = 'route'
      refreshAll()
    }
  } finally {
    routeLoading.value = false
  }
}

const auditFlood = async (row, status) => {
  const res = await auditFloodReport({ reportNo: row.reportNo, auditStatus: status, auditRemark: '管理端确认' })
  if (res.code === 0) {
    ElMessage.success('积水上报已审核')
    loadActiveTab()
    loadSummary()
  }
}

const handleExport = async () => {
  const res = await exportAI()
  if (res.code === 0) ElMessage.success(`导出任务已创建：${res.data?.taskId || '-'}`)
}

const handleCurrentChange = (val) => {
  page.value = val
  loadActiveTab()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  page.value = 1
  loadActiveTab()
}

const auditText = (status) => ({ pending_manual_review: '待人工复核', auto_confirmed: '自动确认', confirmed: '已确认' }[status] || status || '-')

refreshAll()
</script>

<style scoped>
.ai-page {
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
}

.page-head p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.kpi-grid :deep(.el-statistic) {
  min-height: 78px;
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
}

.action-row,
.data-tabs {
  margin-top: 16px;
}

.result-box {
  margin-top: 12px;
}

@media (max-width: 980px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
