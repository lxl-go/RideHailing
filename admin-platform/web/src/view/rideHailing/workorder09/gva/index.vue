<template>
  <div class="gva-page">
    <div class="gva-table-box">
      <div class="hero">
        <div>
          <h2>WO-09 GVA 框架治理</h2>
          <p>动态路由、权限审计、多数据源和定时任务参数治理。</p>
        </div>
        <div class="actions">
          <el-button icon="Refresh" @click="loadData">刷新</el-button>
          <el-button type="primary" icon="Download" @click="handleExport">导出</el-button>
        </div>
      </div>

      <el-row :gutter="16">
        <el-col v-for="item in kpis" :key="item.label" :span="6">
          <el-card shadow="never" class="kpi-card">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
            <p>{{ item.note }}</p>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" class="section-row">
        <el-col :span="12">
          <el-card shadow="never">
            <template #header>动态路由治理</template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="菜单数">{{ route.totalMenus || 0 }}</el-descriptions-item>
              <el-descriptions-item label="隐藏路由">{{ route.hiddenMenus || 0 }}</el-descriptions-item>
              <el-descriptions-item label="默认菜单">{{ route.defaultMenus || 0 }}</el-descriptions-item>
              <el-descriptions-item label="版本">{{ route.routeVersion || '-' }}</el-descriptions-item>
              <el-descriptions-item label="白名单">
                <el-tag :type="route.whitelistStatus === 'PASS' ? 'success' : 'warning'">{{ route.whitelistStatus || '-' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="重复路由">{{ route.duplicateNames || 0 }}</el-descriptions-item>
            </el-descriptions>
            <el-alert v-if="routeWarnings.length" class="mt12" type="warning" :closable="false" :title="routeWarnings.join('；')" />
            <el-empty v-else description="路由白名单暂无告警" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never">
            <template #header>多数据源健康</template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="DB类型">{{ datasource.dbType || '-' }}</el-descriptions-item>
              <el-descriptions-item label="当前库">{{ datasource.activeDbName || '-' }}</el-descriptions-item>
              <el-descriptions-item label="健康状态">
                <el-tag :type="datasource.healthy ? 'success' : 'warning'">{{ datasource.healthy ? '正常' : '异常' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="打开连接">{{ datasource.pool?.openConnections || 0 }}</el-descriptions-item>
              <el-descriptions-item label="使用中">{{ datasource.pool?.inUse || 0 }}</el-descriptions-item>
              <el-descriptions-item label="空闲">{{ datasource.pool?.idle || 0 }}</el-descriptions-item>
            </el-descriptions>
            <el-alert v-if="datasource.warning" class="mt12" type="warning" :closable="false" :title="datasource.warning" />
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" class="section-row">
        <el-col :span="12">
          <el-card shadow="never">
            <template #header>权限审计</template>
            <el-descriptions :column="3" border class="mb12">
              <el-descriptions-item label="审计事件">{{ audit.dataAccessLogs || 0 }}</el-descriptions-item>
              <el-descriptions-item label="越权写">{{ audit.blockedWrites || 0 }}</el-descriptions-item>
              <el-descriptions-item label="无身份">{{ audit.noIdentityEvents || 0 }}</el-descriptions-item>
            </el-descriptions>
            <el-table :data="audit.recentDataAccessLogs || []" height="260" empty-text="暂无记录">
              <el-table-column prop="eventType" label="事件" width="130" />
              <el-table-column prop="targetTable" label="表" min-width="140" />
              <el-table-column prop="path" label="路径" min-width="180" show-overflow-tooltip />
              <el-table-column prop="requestId" label="请求ID" min-width="120" show-overflow-tooltip />
            </el-table>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never">
            <template #header>定时任务治理</template>
            <el-descriptions :column="3" border class="mb12">
              <el-descriptions-item label="任务数">{{ timedTask.totalTasks || 0 }}</el-descriptions-item>
              <el-descriptions-item label="启用">{{ timedTask.enabledTasks || 0 }}</el-descriptions-item>
              <el-descriptions-item label="停用">{{ timedTask.disabledTasks || 0 }}</el-descriptions-item>
            </el-descriptions>
            <el-table :data="timedTask.invalidTasks || []" height="260" empty-text="暂无记录">
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="name" label="任务" min-width="140" />
              <el-table-column prop="reason" label="问题" min-width="220" show-overflow-tooltip />
            </el-table>
          </el-card>
        </el-col>
      </el-row>

      <el-card shadow="never" class="section-row">
        <template #header>操作日志链路质量</template>
        <el-descriptions :column="3" border class="mb12">
          <el-descriptions-item label="操作日志">{{ audit.operationRecords || 0 }}</el-descriptions-item>
          <el-descriptions-item label="缺链路">{{ audit.missingTraceRecords || 0 }}</el-descriptions-item>
          <el-descriptions-item label="治理结论">
            <el-tag :type="(audit.missingTraceRecords || 0) === 0 ? 'success' : 'warning'">{{ (audit.missingTraceRecords || 0) === 0 ? '达标' : '需补齐' }}</el-tag>
          </el-descriptions-item>
        </el-descriptions>
        <el-table :data="audit.recentOperationRecords || []" empty-text="暂无记录">
          <el-table-column prop="method" label="方法" width="90" />
          <el-table-column prop="path" label="路径" min-width="220" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="90" />
          <el-table-column prop="latency_ms" label="耗时(ms)" width="110" />
          <el-table-column prop="request_id" label="请求ID" min-width="160" show-overflow-tooltip />
          <el-table-column prop="trace_id" label="Trace" min-width="160" show-overflow-tooltip />
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { exportGvaGovernance, getGvaGovernanceSummary } from '@/api/rideHailing/workorder09'

defineOptions({ name: 'RideHailingWorkorder09Gva' })

const state = reactive({
  route: {},
  audit: {},
  datasource: {},
  timedTask: {},
})

const route = computed(() => state.route)
const audit = computed(() => state.audit)
const datasource = computed(() => state.datasource)
const timedTask = computed(() => state.timedTask)
const routeWarnings = computed(() => route.value.warnings || [])

const kpis = computed(() => [
  { label: '菜单总数', value: route.value.totalMenus || 0, note: `重复路由 ${route.value.duplicateNames || 0}` },
  { label: '审计事件', value: audit.value.dataAccessLogs || 0, note: `越权写 ${audit.value.blockedWrites || 0}` },
  { label: '操作日志', value: audit.value.operationRecords || 0, note: `缺链路 ${audit.value.missingTraceRecords || 0}` },
  { label: '定时任务', value: timedTask.value.totalTasks || 0, note: `启用 ${timedTask.value.enabledTasks || 0}` },
])

const loadData = async () => {
  const res = await getGvaGovernanceSummary()
  if (res.code === 0) {
    const data = res.data || {}
    state.route = data.route || {}
    state.audit = data.audit || {}
    state.datasource = data.datasource || {}
    state.timedTask = data.timedTask || {}
  }
}

const handleExport = async () => {
  const res = await exportGvaGovernance()
  if (res.code === 0) ElMessage.success(`导出任务已创建：${res.data?.taskId || '-'}`)
}

onMounted(loadData)
</script>

<style scoped>
.gva-page {
  padding: 8px 0 24px;
}

.hero,
.actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.hero {
  margin-bottom: 16px;
}

.hero h2 {
  margin: 0 0 6px;
  font-size: 24px;
}

.hero p,
.kpi-card p,
.kpi-card span {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.kpi-card strong {
  display: block;
  margin: 8px 0;
  font-size: 26px;
}

.section-row {
  margin-top: 16px;
}

.mt12 {
  margin-top: 12px;
}

.mb12 {
  margin-bottom: 12px;
}
</style>
