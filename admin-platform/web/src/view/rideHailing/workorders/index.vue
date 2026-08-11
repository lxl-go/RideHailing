<template>
  <div class="ride-workbench">
    <div class="hero">
      <div>
        <h2>网约车运营工作台</h2>
        <p>集中查看订单、司机、乘客、审核、财务、营销、派单和治理模块。</p>
      </div>
      <el-button type="primary" icon="Refresh" @click="reload">刷新</el-button>
    </div>

    <el-row :gutter="16">
      <el-col v-for="item in cards" :key="item.id" :span="8">
        <el-card class="work-card" shadow="never">
          <div class="card-head">
            <div>
              <strong>{{ item.title }}</strong>
              <p>{{ item.summary }}</p>
            </div>
            <el-tag :type="item.statusType">{{ item.status }}</el-tag>
          </div>
          <div class="card-footer">
            <el-button v-if="item.link" link type="primary" @click="go(item.link)">打开页面</el-button>
            <span v-else class="placeholder">待补齐</span>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

defineOptions({ name: 'RideHailingWorkbench' })

const router = useRouter()

const cards = [
  { id: 'dashboard', title: '运营概览', summary: '查看今日订单、活跃司机、乘客和营收。', status: '核心', statusType: 'success', link: '/ride-hailing/dashboard' },
  { id: 'orders', title: '订单管理', summary: '查询、筛选、查看订单详情和状态。', status: '核心', statusType: 'success', link: '/ride-hailing/orders' },
  { id: 'drivers', title: '司机管理', summary: '司机档案、状态、认证和运营信息。', status: '核心', statusType: 'success', link: '/ride-hailing/drivers' },
  { id: 'passengers', title: '乘客管理', summary: '乘客资料、注册信息与基础运营数据。', status: '核心', statusType: 'success', link: '/ride-hailing/passengers' },
  { id: 'audit', title: '认证审核', summary: '审核司机身份与认证资料。', status: '运营', statusType: 'warning', link: '/ride-hailing/audit' },
  { id: 'vehicle', title: '车辆审核', summary: '审核车辆资料和合规状态。', status: '运营', statusType: 'warning', link: '/ride-hailing/audit/vehicle' },
  { id: 'finance', title: '财务管理', summary: '交易流水、退款、异常和汇总。', status: '财务', statusType: 'warning', link: '/ride-hailing/finance' },
  { id: 'analytics', title: '数据分析', summary: '趋势、转化、复购和导出分析。', status: '数据', statusType: 'warning', link: '/ride-hailing/analytics' },
  { id: 'marketing', title: '营销管理', summary: '优惠券、活动与用户触达。', status: '增长', statusType: 'warning', link: '/ride-hailing/marketing' },
  { id: 'dispatch', title: '派单中心', summary: '规则配置、派单、复核与轨迹回放。', status: '调度', statusType: 'warning', link: '/ride-hailing/dispatch' },
  { id: 'ai', title: 'AI 助手', summary: '出行问答、雨天路线、积水上报。', status: '智能', statusType: 'warning', link: '/ride-hailing/ai' },
  { id: 'governance', title: '平台治理', summary: '路由、权限、数据源和定时任务治理。', status: '系统', statusType: 'info', link: '/ride-hailing/governance' }
]

const go = (path) => router.push(path)
const reload = () => window.location.reload()
</script>

<style scoped>
.ride-workbench {
  padding: 8px 0 24px;
}

.hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.hero h2 {
  margin: 0 0 6px;
  font-size: 24px;
}

.hero p {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.work-card {
  margin-bottom: 16px;
}

.card-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.card-head p {
  margin: 8px 0 0;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.placeholder {
  color: var(--el-text-color-placeholder);
}
</style>