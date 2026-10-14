<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchPaperManuscripts,
  formatManuscriptDraftStatus,
  formatManuscriptSubmissionDisplay,
  formatSubmissionDeadline,
  submissionDeadlineIsUrgent,
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

function draftDisplay(item: PaperManuscriptApiItem) {
  return formatManuscriptDraftStatus(item.draft_status, item.draft_version)
}

function submissionDisplay(item: PaperManuscriptApiItem) {
  return formatManuscriptSubmissionDisplay(item)
}

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
              <div class="ms-draft-chip" :class="`ms-draft-chip--${draftDisplay(item).tone}`">
                <span class="ms-draft-chip__status">{{ draftDisplay(item).statusLabel }}</span>
                <span v-if="draftDisplay(item).versionLabel" class="ms-draft-chip__ver">
                  {{ draftDisplay(item).versionLabel }}
                </span>
              </div>
            </td>
            <td class="ms-col-submission">
              <div
                class="ms-submission-card"
                :class="`ms-submission-card--${submissionDisplay(item).tone}`"
              >
                <p class="ms-submission-card__venue">{{ submissionDisplay(item).primary }}</p>
                <div
                  v-if="submissionDisplay(item).secondary || item.submission_deadline"
                  class="ms-submission-card__meta"
                >
                  <span v-if="submissionDisplay(item).secondary" class="ms-submission-card__status">
                    {{ submissionDisplay(item).secondary }}
                  </span>
                  <span
                    v-if="item.submission_deadline"
                    class="ms-submission-card__deadline"
                    :class="{ 'ms-submission-card__deadline--urgent': submissionDeadlineIsUrgent(item) }"
                  >
                    <span class="ms-submission-card__deadline-label">截止</span>
                    {{ formatSubmissionDeadline(item.submission_deadline) }}
                  </span>
                </div>
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

/* 定稿：状态 + 版本分开展示 */
.ms-draft-chip {
  display: inline-flex;
  gap: 0;
  align-items: stretch;
  overflow: hidden;
  font-size: 12px;
  line-height: 1;
  border: 1px solid transparent;
  border-radius: 8px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}

.ms-draft-chip__status {
  padding: 6px 10px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.ms-draft-chip__ver {
  padding: 6px 9px;
  font-family: ui-monospace, 'SF Mono', Menlo, monospace;
  font-size: 11px;
  font-weight: 600;
  border-left: 1px solid rgba(15, 23, 42, 0.08);
}

.ms-draft-chip--default {
  color: #475569;
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
  border-color: #e2e8f0;
}

.ms-draft-chip--default .ms-draft-chip__ver {
  color: #64748b;
  background: rgba(255, 255, 255, 0.55);
}

.ms-draft-chip--ready {
  color: #1d4ed8;
  background: linear-gradient(180deg, #eff6ff 0%, #dbeafe 100%);
  border-color: #bfdbfe;
}

.ms-draft-chip--ready .ms-draft-chip__ver {
  color: #2563eb;
  background: rgba(255, 255, 255, 0.5);
}

.ms-draft-chip--frozen {
  color: #047857;
  background: linear-gradient(180deg, #ecfdf5 0%, #d1fae5 100%);
  border-color: #a7f3d0;
}

.ms-draft-chip--frozen .ms-draft-chip__ver {
  color: #059669;
  background: rgba(255, 255, 255, 0.45);
}

/* 投递：venue 一行 + 状态/截止一行 */
.ms-submission-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-width: 100%;
  padding: 8px 10px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.ms-submission-card__venue {
  margin: 0;
  overflow: hidden;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.35;
  color: #1e293b;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ms-submission-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.ms-submission-card__status {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1.3;
  border-radius: 6px;
}

.ms-submission-card__deadline {
  display: inline-flex;
  gap: 4px;
  align-items: center;
  padding: 3px 8px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.3;
  color: #64748b;
  font-variant-numeric: tabular-nums;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
}

.ms-submission-card__deadline-label {
  font-weight: 600;
  color: #94a3b8;
}

.ms-submission-card__deadline--urgent {
  color: #b45309;
  background: #fffbeb;
  border-color: #fcd34d;
}

.ms-submission-card__deadline--urgent .ms-submission-card__deadline-label {
  color: #d97706;
}

.ms-submission-card--muted {
  background: #fafafa;
  border-color: #e5e7eb;
}

.ms-submission-card--muted .ms-submission-card__venue {
  font-weight: 500;
  color: #64748b;
}

.ms-submission-card--muted .ms-submission-card__status {
  color: #64748b;
  background: #f1f5f9;
}

.ms-submission-card--progress {
  background: linear-gradient(135deg, #f5f3ff 0%, #eef2ff 100%);
  border-color: #c7d2fe;
}

.ms-submission-card--progress .ms-submission-card__status {
  color: #4338ca;
  background: rgba(255, 255, 255, 0.7);
}

.ms-submission-card--warn {
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  border-color: #fcd34d;
}

.ms-submission-card--warn .ms-submission-card__status {
  color: #b45309;
  background: rgba(255, 255, 255, 0.75);
}

.ms-submission-card--success {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  border-color: #6ee7b7;
}

.ms-submission-card--success .ms-submission-card__status {
  color: #047857;
  background: rgba(255, 255, 255, 0.65);
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
