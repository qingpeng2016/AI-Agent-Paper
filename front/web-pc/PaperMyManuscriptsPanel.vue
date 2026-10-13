<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchPaperManuscripts,
  PAPER_LIST_PAGE_SIZE,
  type PaperManuscriptApiItem,
} from '@/api/manuscripts'
import PaperListPagination from './PaperListPagination.vue'

const props = defineProps<{
  activeManuscriptId: string
}>()

const emit = defineEmits<{
  select: [id: string]
}>()

const items = ref<PaperManuscriptApiItem[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)

function formatCreatedAt(iso?: string): string {
  if (!iso?.trim()) return '—'
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

async function loadPage(targetPage = page.value) {
  loading.value = true
  try {
    const data = await fetchPaperManuscripts({
      page: targetPage,
      page_size: PAPER_LIST_PAGE_SIZE,
    })
    items.value = (data.items ?? []).filter((m) => m.status === 'active')
    total.value = data.total ?? items.value.length
    page.value = data.page ?? targetPage
    if (items.value.length === 0 && total.value > 0 && page.value > 1) {
      await loadPage(page.value - 1)
    }
  } catch (e) {
    items.value = []
    total.value = 0
    const msg = e instanceof Error ? e.message : '加载论文列表失败'
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  void loadPage(p)
}

function selectAsCurrent(id: string) {
  if (!id || id === props.activeManuscriptId) return
  emit('select', id)
}

watch(
  () => props.activeManuscriptId,
  () => {
    void loadPage(page.value)
  },
)

void loadPage(1)
</script>

<template>
  <section class="ms-panel">
    <p v-if="loading" class="ms-meta">正在加载论文列表…</p>
    <p v-else-if="!total" class="ms-empty">暂无进行中的论文（status=active）。</p>
    <template v-else>
      <table class="ms-table">
        <colgroup>
          <col class="ms-col-title" />
          <col class="ms-col-created" />
          <col class="ms-col-disc" />
          <col class="ms-col-status" />
          <col class="ms-col-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>标题</th>
            <th>创建时间</th>
            <th>学科</th>
            <th>状态</th>
            <th class="ms-col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td class="ms-col-title">{{ item.title }}</td>
            <td class="ms-col-created">{{ formatCreatedAt(item.created_at) }}</td>
            <td class="ms-col-disc">
              {{ item.discipline_label || item.discipline_code || '—' }}
            </td>
            <td class="ms-col-status">
              <span v-if="item.is_current" class="ms-pill ms-pill--current">当前</span>
              <span v-else class="ms-muted">—</span>
            </td>
            <td class="ms-col-actions">
              <button
                v-if="!item.is_current"
                type="button"
                class="ms-btn-primary"
                @click="selectAsCurrent(item.id)"
              >
                设为当前
              </button>
              <span v-else class="ms-muted">已选中</span>
            </td>
          </tr>
        </tbody>
      </table>
      <PaperListPagination
        :total="total"
        :page-size="PAPER_LIST_PAGE_SIZE"
        :current-page="page"
        @update:current-page="onPageChange"
      />
    </template>
  </section>
</template>

<style scoped>
.ms-panel {
  padding: 20px 22px 24px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.ms-meta {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.ms-empty {
  margin: 0;
  padding: 32px 16px;
  font-size: 14px;
  color: #64748b;
  text-align: center;
  background: #f8fafc;
  border-radius: 12px;
}

.ms-table {
  width: 100%;
  font-size: 14px;
  table-layout: fixed;
  border-collapse: collapse;
}

.ms-table col.ms-col-title {
  width: 40%;
}

.ms-table col.ms-col-created {
  width: 18%;
}

.ms-table col.ms-col-disc {
  width: 12%;
}

.ms-table col.ms-col-status {
  width: 12%;
}

.ms-table col.ms-col-actions {
  width: 18%;
}

.ms-table th,
.ms-table td {
  padding: 13px 12px;
  line-height: 1.5;
  text-align: left;
  vertical-align: middle;
  border-bottom: 1px solid #e2e8f0;
}

.ms-table th {
  font-weight: 700;
  color: #64748b;
}

.ms-col-title {
  overflow: hidden;
  font-weight: 400;
  color: #1e1b4b;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ms-col-created {
  font-weight: 400;
  color: #334155;
  font-variant-numeric: tabular-nums;
}

.ms-col-disc {
  overflow: hidden;
  color: #475569;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ms-col-actions {
  text-align: right;
  white-space: nowrap;
}

.ms-pill {
  display: inline-block;
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.35;
  border-radius: 999px;
}

.ms-pill--current {
  color: #047857;
  background: #d1fae5;
}

.ms-muted {
  font-size: 13px;
  color: #94a3b8;
}

.ms-btn-primary {
  padding: 5px 10px;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: linear-gradient(135deg, #7c3aed, #6366f1);
  border: none;
  border-radius: 7px;
}

</style>
