<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchPaperManuscripts,
  formatManuscriptDraftStatus,
  formatManuscriptSubmissionDisplay,
  formatSubmissionDeadline,
  PAPER_LIST_PAGE_SIZE,
  type PaperManuscriptApiItem,
} from '@/api/manuscripts'
import PaperListPagination from './PaperListPagination.vue'
import PaperTableActionsHead from './PaperTableActionsHead.vue'

const props = defineProps<{
  activeManuscriptId: string
  manuscriptSwitching?: boolean
}>()

const emit = defineEmits<{
  startWriting: [item: PaperManuscriptApiItem]
  startSubmission: [item: PaperManuscriptApiItem]
}>()

function onStartWritingClick(item: PaperManuscriptApiItem) {
  if (props.manuscriptSwitching) return
  emit('startWriting', item)
}

function onStartSubmissionClick(item: PaperManuscriptApiItem) {
  if (props.manuscriptSwitching) return
  emit('startSubmission', item)
}

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
    <p v-else-if="!total" class="ms-empty">暂无进行中的论文。</p>
    <template v-else>
      <table class="ms-table">
        <colgroup>
          <col class="ms-col-title" />
          <col class="ms-col-created" />
          <col class="ms-col-disc" />
          <col class="ms-col-draft" />
          <col class="ms-col-submission" />
          <col class="ms-col-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>标题</th>
            <th>创建时间</th>
            <th>学科</th>
            <th>定稿</th>
            <th>投递</th>
            <th class="ms-col-actions ul-col-actions">
              <PaperTableActionsHead tip="进入写作或论文投稿" />
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td class="ms-col-title">
              <span v-if="item.is_current" class="ms-pill ms-pill--current">当前</span>
              <span class="ms-col-title-text">{{ item.title }}</span>
            </td>
            <td class="ms-col-created">{{ formatCreatedAt(item.created_at) }}</td>
            <td class="ms-col-disc">
              {{ item.discipline_label || item.discipline_code || '—' }}
            </td>
            <td class="ms-col-draft">
              <span
                class="ms-pill"
                :class="`ms-pill--draft-${formatManuscriptDraftStatus(item.draft_status, item.draft_version).tone}`"
              >
                {{ formatManuscriptDraftStatus(item.draft_status, item.draft_version).label }}
              </span>
            </td>
            <td class="ms-col-submission">
              <div class="ms-submission-cell">
                <span
                  class="ms-pill"
                  :class="`ms-pill--sub-${formatManuscriptSubmissionDisplay(item).tone}`"
                >
                  {{ formatManuscriptSubmissionDisplay(item).primary }}
                </span>
                <span
                  v-if="formatManuscriptSubmissionDisplay(item).secondary"
                  class="ms-submission-sub"
                >
                  {{ formatManuscriptSubmissionDisplay(item).secondary }}
                </span>
                <span v-if="item.submission_deadline" class="ms-submission-deadline">
                  截止 {{ formatSubmissionDeadline(item.submission_deadline) }}
                </span>
              </div>
            </td>
            <td class="ms-col-actions ul-col-actions">
              <button
                type="button"
                class="ul-btn-link"
                :disabled="manuscriptSwitching"
                @click="onStartWritingClick(item)"
              >
                写作
              </button>
              <button
                type="button"
                class="ul-btn-link"
                :disabled="manuscriptSwitching"
                @click="onStartSubmissionClick(item)"
              >
                投稿
              </button>
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
  width: 26%;
}

.ms-table col.ms-col-created {
  width: 14%;
}

.ms-table col.ms-col-disc {
  width: 10%;
}

.ms-table col.ms-col-draft {
  width: 14%;
}

.ms-table col.ms-col-submission {
  width: 22%;
}

.ms-table col.ms-col-actions {
  width: 14%;
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

.ms-col-title-text {
  vertical-align: middle;
}

.ms-col-title .ms-pill--current {
  margin-right: 8px;
  vertical-align: middle;
}

.ms-col-draft,
.ms-col-submission {
  vertical-align: middle;
}

.ms-submission-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-start;
}

.ms-submission-sub,
.ms-submission-deadline {
  font-size: 12px;
  line-height: 1.4;
  color: #64748b;
}

.ms-pill--draft-default {
  color: #475569;
  background: #f1f5f9;
}

.ms-pill--draft-ready {
  color: #1d4ed8;
  background: #dbeafe;
}

.ms-pill--draft-frozen {
  color: #047857;
  background: #d1fae5;
}

.ms-pill--sub-muted {
  color: #64748b;
  background: #f1f5f9;
}

.ms-pill--sub-progress {
  color: #4338ca;
  background: #e0e7ff;
}

.ms-pill--sub-warn {
  color: #b45309;
  background: #fef3c7;
}

.ms-pill--sub-success {
  color: #047857;
  background: #d1fae5;
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
</style>
