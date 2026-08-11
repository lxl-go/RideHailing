<template>
  <div class="gva-table-box">
    <div class="gva-btn-list">
      <el-button type="primary" icon="Refresh" @click="$emit('refresh')">刷新</el-button>
      <el-button icon="Download" @click="$emit('export')">导出</el-button>
      <el-button type="warning" icon="Operation" :disabled="!selectedRows.length" @click="$emit('batch')">批量操作</el-button>
    </div>

    <el-table v-loading="loading" :data="tableData" row-key="id" style="width: 100%" @selection-change="$emit('selectionChange', $event)">
      <el-table-column type="selection" width="50" />
      <el-table-column label="编号" width="120">
        <template #default="{ row }">{{ row.id || row.ID || row.orderNo }}</template>
      </el-table-column>
      <el-table-column :label="mainColumnLabel" min-width="220">
        <template #default="{ row }">
          <div class="main-cell">
            <span>{{ row.title || row.realName || row.plateNumber || row.name || row.orderNo }}</span>
            <el-tag v-if="row.source" size="small" type="info">{{ row.source }}</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="业务归属" min-width="130">
        <template #default="{ row }">{{ row.owner || row.userId || row.driverId || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.status || '-' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="180">
        <template #default="{ row }">{{ row.updatedAt || row.CreatedAt || row.createdAt || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" icon="View" @click="$emit('detail', row)">详情</el-button>
          <el-button link type="success" icon="Check" @click="$emit('action', row, 'approve')">通过</el-button>
          <el-button link type="danger" icon="Close" @click="$emit('action', row, 'reject')">驳回</el-button>
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
        @current-change="$emit('pageChange', $event)"
        @size-change="$emit('sizeChange', $event)"
      />
    </div>
  </div>
</template>

<script setup>
defineProps({
  loading: Boolean,
  tableData: { type: Array, default: () => [] },
  total: { type: Number, default: 0 },
  page: { type: Number, default: 1 },
  pageSize: { type: Number, default: 20 },
  selectedRows: { type: Array, default: () => [] },
  mainColumnLabel: { type: String, default: '业务信息' },
})

defineEmits(['refresh', 'export', 'batch', 'selectionChange', 'detail', 'action', 'pageChange', 'sizeChange'])
</script>

<style scoped>
.main-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
