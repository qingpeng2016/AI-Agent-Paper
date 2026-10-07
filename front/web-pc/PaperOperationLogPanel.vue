<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { DEMO_OPERATION_LOG_SEED } from './demoOperationLogs'
import {
  OPERATION_LOG_STORAGE_KEY,
  getOperationLogs,
  setOperationLogs,
  type PaperOperationLogEntry,
  type PaperOperationLogStatus,
} from './types'

const props = withDefaults(
  defineProps<{
    manuscriptId: string
    manuscriptTitle: string
    /** 嵌入个人中心：无外层卡片 */
    embedded?: boolean
    /** 仅展示日志列表（无统计、筛选） */
    listOnly?: boolean
  }>(),
  { embedded: false, listOnly: false },
)

const refreshTick = ref(0)

function ensureDemoSeed() {
  try {
    const raw = localStorage.getItem(OPERATION_LOG_STORAGE_KEY)
    if (raw) {
      const data = JSON.parse(raw) as { entries?: PaperOperationLogEntry[] }
      if (data.entries && data.entries.length > 0) return
    }
    setOperationLogs([...DEMO_OPERATION_LOG_SEED])
  } catch {
    setOperationLogs([...DEMO_OPERATION_LOG_SEED])
  }
}

onMounted(() => {
  ensureDemoSeed()
  refreshTick.value += 1
})

watch(
  () => props.manuscriptId,
  () => {
    refreshTick.value += 1
  },
)

const entries = computed(() => {
  void refreshTick.value
  return getOperationLogs()
})

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString('zh-CN', { hour12: false })
  } catch {
    return iso
  }
}

function formatTokens(n: number) {
  return n.toLocaleString('zh-CN')
}

function statusLabel(status: PaperOperationLogStatus) {
  if (status === 'success') return '成功'
  if (status === 'failed') return '失败'
  return '已取消'
}

function reload() {
  refreshTick.value += 1
}

defineExpose({ reload })
</script>

<template>
  <section class="oplog-panel" :class="{ 'oplog-panel--embedded': embedded }">
    <p v-if="!embedded && !listOnly" class="oplog-lead">
      展示账号下与工作流相关的 <strong>Token 消耗与运行记录</strong>（演示数据存于浏览器本地；接入后将同步 Token
      商城计费流水）。
    </p>

    <div v-if="entries.length === 0" class="oplog-empty">
      暂无操作日志。在各模块点击「运行」后，记录将出现在此处。
    </div>

    <div v-else class="oplog-table-wrap">
      <table class="oplog-table">
        <thead>
          <tr>
            <th>时间</th>
            <th>模块</th>
            <th>操作</th>
            <th>论文</th>
            <th class="oplog-num">消耗 tokens</th>
            <th>状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in entries" :key="row.id">
            <td class="oplog-time">{{ formatDate(row.occurredAt) }}</td>
            <td>{{ row.moduleLabel }}</td>
            <td class="oplog-action">
              {{ row.action }}
              <span v-if="row.note" class="oplog-note">{{ row.note }}</span>
            </td>
            <td>{{ row.manuscriptTitle ?? '—' }}</td>
            <td class="oplog-num oplog-total">{{ formatTokens(row.tokensTotal) }}</td>
            <td>
              <span class="oplog-status" :class="`oplog-status--${row.status}`">{{
                statusLabel(row.status)
              }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.oplog-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.oplog-panel--embedded {
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 0;
  box-shadow: none;
}

.oplog-lead {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.55;
  color: #475569;
}

.oplog-empty {
  padding: 32px 16px;
  text-align: center;
  font-size: 14px;
  color: #64748b;
  background: #f8fafc;
  border-radius: 12px;
}

.oplog-table-wrap {
  overflow-x: auto;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.oplog-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.oplog-table th,
.oplog-table td {
  padding: 11px 14px;
  text-align: left;
  border-bottom: 1px solid #eef2f6;
  vertical-align: top;
}

.oplog-table th {
  background: #f8fafc;
  color: #64748b;
  font-weight: 600;
  font-size: 12px;
  white-space: nowrap;
}

.oplog-table tbody tr:hover {
  background: #fafbff;
}

.oplog-table tbody tr:last-child td {
  border-bottom: none;
}

.oplog-num {
  text-align: right;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.oplog-time {
  white-space: nowrap;
  color: #64748b;
  font-size: 12px;
}

.oplog-action {
  max-width: 320px;
  line-height: 1.45;
  color: #334155;
}

.oplog-note {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #94a3b8;
}

.oplog-total {
  font-weight: 600;
  color: #4f46e5;
}

.oplog-status {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.oplog-status--success {
  background: #dcfce7;
  color: #166534;
}

.oplog-status--failed {
  background: #fee2e2;
  color: #991b1b;
}

.oplog-status--cancelled {
  background: #f1f5f9;
  color: #64748b;
}
</style>
