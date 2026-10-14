<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchPaperManuscripts,
  formatManuscriptDraftVersion,
  PAPER_LIST_PAGE_SIZE,
  type PaperManuscriptApiItem,
} from '@/api/manuscripts'
import PaperListPagination from './PaperListPagination.vue'
import PaperTableActionsHead from './PaperTableActionsHead.vue'

type SubmissionDialogTabId = 'submit' | 'revise'

const SUBMISSION_DIALOG_TABS: { id: SubmissionDialogTabId; label: string }[] = [
  { id: 'submit', label: '投稿' },
  { id: 'revise', label: '改稿' },
]

const submissionDialog = ref<{
  item: PaperManuscriptApiItem
  tab: SubmissionDialogTabId
} | null>(null)

const props = defineProps<{
  activeManuscriptId: string
  manuscriptSwitching?: boolean
}>()

const emit = defineEmits<{
  startWriting: [item: PaperManuscriptApiItem]
}>()

function onStartWritingClick(item: PaperManuscriptApiItem) {
  if (props.manuscriptSwitching) return
  emit('startWriting', item)
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

function openSubmissionDialog(item: PaperManuscriptApiItem) {
  submissionDialog.value = { item, tab: 'submit' }
}

function closeSubmissionDialog() {
  submissionDialog.value = null
}

function setSubmissionDialogTab(tab: SubmissionDialogTabId) {
  const ctx = submissionDialog.value
  if (!ctx) return
  submissionDialog.value = { ...ctx, tab }
}

function runSubmissionDialogPrimary() {
  const ctx = submissionDialog.value
  if (!ctx) return
  const label = ctx.tab === 'submit' ? '提交投稿' : '保存改稿计划'
  ElMessage.info(`${label}（接口接入后替换）`)
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
          <col class="ms-col-status" />
          <col class="ms-col-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>标题</th>
            <th>创建时间</th>
            <th>学科</th>
            <th>状态</th>
            <th class="ms-col-actions ul-col-actions">
              <PaperTableActionsHead tip="进入写作或投稿流程" />
            </th>
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
            <td class="ms-col-actions ul-col-actions">
              <button
                type="button"
                class="ul-btn-link"
                :disabled="manuscriptSwitching"
                @click="onStartWritingClick(item)"
              >
                写作
              </button>
              <button type="button" class="ul-btn-link" @click="openSubmissionDialog(item)">
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

  <Teleport to="body">
    <div
      v-if="submissionDialog"
      class="paper-modal-overlay ms-submission-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="ms-submission-title"
      @click.self="closeSubmissionDialog"
      @keydown.escape="closeSubmissionDialog"
    >
      <div
        class="paper-message-box paper-message-box--default paper-modal-panel ms-submission-panel"
        @click.stop
      >
        <header class="paper-modal-header ms-submission-header">
          <div class="ms-submission-heading">
            <h2 id="ms-submission-title" class="paper-modal-title">
              {{ submissionDialog.item.title }}
            </h2>
            <p class="ms-meta ms-submission-meta">
              版本号 {{ formatManuscriptDraftVersion(submissionDialog.item.draft_version) }}
              <span v-if="submissionDialog.item.discipline_label"> · </span>
              <span v-if="submissionDialog.item.discipline_label">{{
                submissionDialog.item.discipline_label
              }}</span>
            </p>
          </div>
          <button
            type="button"
            class="paper-modal-close"
            aria-label="关闭"
            @click="closeSubmissionDialog"
          >
            ×
          </button>
        </header>
        <nav class="ms-submission-tabs" role="tablist" aria-label="投稿与改稿">
          <button
            v-for="t in SUBMISSION_DIALOG_TABS"
            :key="t.id"
            type="button"
            role="tab"
            class="ms-submission-tab"
            :class="{ 'ms-submission-tab--active': submissionDialog.tab === t.id }"
            :aria-selected="submissionDialog.tab === t.id"
            @click="setSubmissionDialogTab(t.id)"
          >
            {{ t.label }}
          </button>
        </nav>
        <div class="paper-modal-body ms-submission-body">
          <template v-if="submissionDialog.tab === 'submit'">
            <p class="ms-lead">准备向目标会议/期刊提交当前稿（PDF、补充材料、Checklist 等）。</p>
            <label class="ms-field">
              <span class="ms-label">目标 venue</span>
              <input type="text" class="ms-input" placeholder="如 NeurIPS 2026、IEEE TMI" disabled />
            </label>
            <label class="ms-field">
              <span class="ms-label">投稿系统链接 / 备注</span>
              <textarea
                class="ms-textarea"
                rows="3"
                placeholder="OpenReview 链接、deadline、匿名说明等（接入后编辑）"
                disabled
              />
            </label>
            <p class="ms-meta ms-submission-hint">
              接入后将校验 PDF 页数、匿名化与必需附件；成功提交后更新投递状态。
            </p>
          </template>
          <template v-else>
            <p class="ms-lead">记录审稿意见、修回计划与再投/同会修回轮次。</p>
            <label class="ms-field">
              <span class="ms-label">审稿结论</span>
              <input
                type="text"
                class="ms-input"
                placeholder="Major Revision / Reject / Accept …"
                disabled
              />
            </label>
            <label class="ms-field">
              <span class="ms-label">意见摘要与改稿要点</span>
              <textarea
                class="ms-textarea"
                rows="5"
                placeholder="逐条 reviewer 意见、需补实验/改 claim/改图表（接入 Agent 后解析）"
                disabled
              />
            </label>
            <p class="ms-meta ms-submission-hint">
              改稿完成后可回到「智能论文全链路」更新综述/实验/写作，或在此标记修回已提交。
            </p>
          </template>
        </div>
        <footer class="paper-message-box__btns ms-submission-footer">
          <button type="button" class="paper-btn-primary" @click="runSubmissionDialogPrimary">
            {{ submissionDialog.tab === 'submit' ? '提交投稿' : '保存改稿计划' }}
          </button>
        </footer>
      </div>
    </div>
  </Teleport>
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

.ms-submission-heading .ms-meta {
  font-size: 13px;
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

.ms-submission-panel {
  display: flex;
  flex-direction: column;
  width: min(640px, calc(100vw - 32px));
  max-height: min(88vh, 720px);
}

.ms-submission-header {
  align-items: flex-start;
}

.ms-submission-heading {
  min-width: 0;
  padding-right: 12px;
}

.ms-submission-meta {
  margin: 6px 0 0;
  font-size: 13px;
}

.ms-submission-tabs {
  display: flex;
  flex-shrink: 0;
  gap: 6px;
  padding: 0 22px;
  margin-top: 16px;
  border-bottom: 1px solid #e2e8f0;
}

.ms-submission-tab {
  position: relative;
  padding: 12px 20px;
  margin-bottom: -1px;
  font-size: 15px;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  border-radius: 8px 8px 0 0;
}

.ms-submission-tab--active {
  font-weight: 600;
  color: #1e293b;
  border-bottom-color: #6366f1;
}

.ms-submission-body {
  flex: 1;
  overflow: auto;
  padding: 18px 22px 8px;
}

.ms-lead {
  margin: 0 0 16px;
  font-size: 14px;
  line-height: 1.55;
  color: #475569;
}

.ms-field {
  display: block;
  margin-bottom: 14px;
}

.ms-label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  font-weight: 600;
  color: #334155;
}

.ms-input,
.ms-textarea {
  box-sizing: border-box;
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  color: #334155;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.ms-textarea {
  resize: vertical;
  min-height: 88px;
}

.ms-submission-hint {
  margin: 12px 0 0;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 8px;
}

.ms-submission-footer {
  flex-shrink: 0;
}

</style>
