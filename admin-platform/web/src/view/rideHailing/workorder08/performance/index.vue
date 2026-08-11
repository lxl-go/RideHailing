<template>
  <div class="performance-page">
    <div class="gva-table-box">
      <div class="summary-grid">
        <el-statistic title="场景数量" :value="summary.totalScenarios || 0" />
        <el-statistic title="报告数量" :value="summary.totalReports || 0" />
        <el-statistic title="通过报告" :value="summary.passReports || 0" />
        <div class="pass-rate">
          <span class="label">通过率</span>
          <el-progress :percentage="summary.passRate || 0" :status="passRateStatus" />
        </div>
      </div>

      <div class="runtime-strip">
        <span>Go {{ runtime.goVersion || '-' }}</span>
        <span>CPU {{ runtime.numCpu || 0 }}</span>
        <span>Goroutine {{ runtime.numGoroutine || 0 }}</span>
        <span>Heap {{ runtime.heapAllocMb || 0 }}MB / {{ runtime.heapSysMb || 0 }}MB</span>
        <span>GC {{ runtime.gcCycles || 0 }}</span>
      </div>

      <div class="gva-btn-list">
        <el-button type="primary" icon="Plus" @click="openReport">录入报告</el-button>
        <el-button icon="Refresh" @click="refreshAll">刷新指标</el-button>
        <el-button icon="Download" @click="handleExport">导出</el-button>
      </div>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="场景">
          <el-select v-model="searchForm.scenario" clearable placeholder="全部" style="width: 190px" @change="search">
            <el-option v-for="item in scenarios" :key="item.scenario" :label="item.name" :value="item.scenario" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务">
          <el-input v-model="searchForm.targetService" clearable placeholder="driver-api/message-svc" style="width: 190px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="结论">
          <el-select v-model="searchForm.verdict" clearable placeholder="全部" style="width: 120px" @change="search">
            <el-option label="PASS" value="PASS" />
            <el-option label="WARN" value="WARN" />
            <el-option label="FAIL" value="FAIL" />
          </el-select>
        </el-form-item>
        <el-form-item label="工具">
          <el-select v-model="searchForm.tool" clearable placeholder="全部" style="width: 120px" @change="search">
            <el-option label="k6" value="k6" />
            <el-option label="JMeter" value="jmeter" />
            <el-option label="wrk" value="wrk" />
            <el-option label="pprof" value="pprof" />
            <el-option label="trace" value="trace" />
            <el-option label="manual" value="manual" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="search">查询</el-button>
          <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="压测报告" name="reports" />
        <el-tab-pane label="验收目标" name="scenarios" />
      </el-tabs>

      <el-table v-if="activeTab === 'reports'" v-loading="loading" :data="reports" row-key="id" style="width: 100%">
        <el-table-column label="报告编号" prop="reportNo" min-width="170" />
        <el-table-column label="场景" prop="scenario" min-width="160">
          <template #default="{ row }">{{ scenarioName(row.scenario) }}</template>
        </el-table-column>
        <el-table-column label="服务" prop="targetService" width="140" />
        <el-table-column label="工具" prop="tool" width="90" />
        <el-table-column label="QPS" prop="qps" width="100" />
        <el-table-column label="P99(ms)" prop="p99Ms" width="100" />
        <el-table-column label="错误率" width="100">
          <template #default="{ row }">{{ percentText(row.errorRate) }}</template>
        </el-table-column>
        <el-table-column label="协程" width="100">
          <template #default="{ row }">{{ row.goroutinesBefore }} → {{ row.goroutinesAfter }}</template>
        </el-table-column>
        <el-table-column label="结论" prop="verdict" width="100">
          <template #default="{ row }">
            <el-tag :type="verdictType(row.verdict)">{{ row.verdict }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="产物路径" prop="artifactPath" min-width="240" show-overflow-tooltip />
      </el-table>

      <el-table v-else :data="scenarios" row-key="id" style="width: 100%">
        <el-table-column label="场景" prop="name" min-width="220" />
        <el-table-column label="范围" prop="scope" width="120" />
        <el-table-column label="目标QPS" prop="targetQps" width="120" />
        <el-table-column label="目标P99(ms)" prop="targetP99Ms" width="130" />
        <el-table-column label="最大错误率" width="120">
          <template #default="{ row }">{{ percentText(row.maxErrorRate) }}</template>
        </el-table-column>
        <el-table-column label="协程偏差" width="120">
          <template #default="{ row }">{{ row.maxGoroutineDeltaPercent }}%</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="activeTab === 'reports'" class="gva-pagination">
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

    <el-dialog v-model="reportVisible" title="录入 WO-08 压测报告" width="760px">
      <el-form :model="reportForm" label-width="110px">
        <el-row :gutter="12">
          <el-col :span="12"><el-form-item label="报告编号"><el-input v-model="reportForm.reportNo" placeholder="留空自动生成" /></el-form-item></el-col>
          <el-col :span="12">
            <el-form-item label="场景">
              <el-select v-model="reportForm.scenario" style="width: 100%">
                <el-option v-for="item in scenarios" :key="item.scenario" :label="item.name" :value="item.scenario" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12"><el-form-item label="目标服务"><el-input v-model="reportForm.targetService" /></el-form-item></el-col>
          <el-col :span="12">
            <el-form-item label="工具">
              <el-select v-model="reportForm.tool" style="width: 100%">
                <el-option label="k6" value="k6" />
                <el-option label="JMeter" value="jmeter" />
                <el-option label="wrk" value="wrk" />
                <el-option label="pprof" value="pprof" />
                <el-option label="trace" value="trace" />
                <el-option label="manual" value="manual" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8"><el-form-item label="QPS"><el-input-number v-model="reportForm.qps" :min="0" :precision="2" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="P50(ms)"><el-input-number v-model="reportForm.p50Ms" :min="0" :precision="2" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="P90(ms)"><el-input-number v-model="reportForm.p90Ms" :min="0" :precision="2" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="P99(ms)"><el-input-number v-model="reportForm.p99Ms" :min="0" :precision="2" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="错误率"><el-input-number v-model="reportForm.errorRate" :min="0" :max="1" :step="0.0001" :precision="6" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8">
            <el-form-item label="结论">
              <el-select v-model="reportForm.verdict" style="width: 100%">
                <el-option label="PASS" value="PASS" />
                <el-option label="WARN" value="WARN" />
                <el-option label="FAIL" value="FAIL" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8"><el-form-item label="协程前"><el-input-number v-model="reportForm.goroutinesBefore" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="协程后"><el-input-number v-model="reportForm.goroutinesAfter" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="堆后(MB)"><el-input-number v-model="reportForm.heapAfterMb" :min="0" :precision="2" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="产物路径"><el-input v-model="reportForm.artifactPath" placeholder="docs/performance/reports/wo08-driver-location.sample.json" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="结论备注"><el-input v-model="reportForm.notes" type="textarea" :rows="3" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="reportVisible = false">取消</el-button>
        <el-button type="primary" @click="submitReport">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createPerformanceReport,
  exportPerformanceReports,
  getPerformanceSummary,
  getRuntimeSnapshot,
  listPerformanceReports,
  listPerformanceScenarios,
} from '@/api/rideHailing/workorder08'

defineOptions({ name: 'RideHailingWorkorder08Performance' })

const loading = ref(false)
const activeTab = ref('reports')
const reportVisible = ref(false)
const reports = ref([])
const scenarios = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const summary = reactive({})
const runtime = reactive({})
const searchForm = reactive({ scenario: '', targetService: '', verdict: '', tool: '' })
const reportForm = reactive(defaultReportForm())

const passRateStatus = computed(() => {
  if ((summary.passRate || 0) >= 90) return 'success'
  if ((summary.passRate || 0) >= 60) return 'warning'
  return 'exception'
})

function defaultReportForm() {
  return {
    reportNo: '',
    scenario: 'admin_http',
    targetService: 'admin-api',
    tool: 'k6',
    qps: 1000,
    p50Ms: 60,
    p90Ms: 120,
    p99Ms: 190,
    errorRate: 0,
    goroutinesBefore: 0,
    goroutinesAfter: 0,
    heapBeforeMb: 0,
    heapAfterMb: 0,
    verdict: 'PASS',
    artifactPath: 'docs/performance/reports/wo08-admin-http.sample.json',
    notes: '',
  }
}

const loadSummary = async () => {
  const res = await getPerformanceSummary()
  if (res.code === 0) {
    Object.assign(summary, res.data || {})
    Object.assign(runtime, res.data?.runtime || {})
  }
}

const loadRuntime = async () => {
  const res = await getRuntimeSnapshot()
  if (res.code === 0) Object.assign(runtime, res.data || {})
}

const loadScenarios = async () => {
  const res = await listPerformanceScenarios()
  if (res.code === 0) scenarios.value = res.data || []
}

const getTableData = async () => {
  loading.value = true
  try {
    const res = await listPerformanceReports({ page: page.value, pageSize: pageSize.value, ...searchForm })
    if (res.code === 0) {
      reports.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } finally {
    loading.value = false
  }
}

const refreshAll = async () => {
  await Promise.all([loadSummary(), loadRuntime(), loadScenarios(), getTableData()])
}

const search = () => {
  page.value = 1
  getTableData()
}

const resetSearch = () => {
  Object.assign(searchForm, { scenario: '', targetService: '', verdict: '', tool: '' })
  search()
}

const openReport = () => {
  Object.assign(reportForm, defaultReportForm())
  reportVisible.value = true
}

const submitReport = async () => {
  const res = await createPerformanceReport(reportForm)
  if (res.code === 0) {
    ElMessage.success('压测报告已录入')
    reportVisible.value = false
    await loadSummary()
    search()
  }
}

const handleExport = async () => {
  const res = await exportPerformanceReports()
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

const scenarioName = (value) => scenarios.value.find((item) => item.scenario === value)?.name || value
const percentText = (value) => `${(((value || 0) * 100).toFixed(3)).replace(/\.?0+$/, '')}%`
const verdictType = (value) => ({ PASS: 'success', WARN: 'warning', FAIL: 'danger' }[value] || 'info')

refreshAll()
</script>

<style scoped>
.performance-page {
  padding: 8px 0 24px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.summary-grid :deep(.el-statistic),
.pass-rate {
  min-height: 78px;
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
}

.pass-rate .label {
  display: block;
  margin-bottom: 12px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.runtime-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 16px;
}

.runtime-strip span {
  padding: 7px 10px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-lighter);
  font-size: 13px;
}

.search-form {
  margin-bottom: 8px;
}

@media (max-width: 980px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 540px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
