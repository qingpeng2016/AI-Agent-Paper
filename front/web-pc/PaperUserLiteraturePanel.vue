<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { REFERENCE_UPLOAD_ACCEPT } from '@paper/types'
import {
  fetchUserLiterature,
  fetchUserLiteratureCategories,
  formatFileSizeBytes,
  formatUserLiteratureCategories,
  formatUserLiteratureFrom,
  PAPER_LIST_PAGE_SIZE,
  userLiteratureFromTone,
  type PaperUserLiteratureItem,
} from '@/api/userLiterature'
import PaperListPagination from './PaperListPagination.vue'
import PaperSelect from './PaperSelect.vue'
import PaperUserLiteratureCategoryPanel from './PaperUserLiteratureCategoryPanel.vue'

export type UserLiteratureTabId = 'list' | 'categories'

const activeTab = ref<UserLiteratureTabId>('list')

const items = ref<PaperUserLiteratureItem[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const detailOpen = ref(false)
const detailItem = ref<PaperUserLiteratureItem | null>(null)

const uploadOpen = ref(false)
const uploadFile = ref<File | null>(null)
const uploadInputRef = ref<HTMLInputElement | null>(null)
const uploadSubmitting = ref(false)
const uploadTitle = ref('')
const uploadNote = ref('')

const filterKeywordDraft = ref('')
const filterKeywordApplied = ref('')
const filterCategoryId = ref('')

const categoryFilterOptions = ref<{ value: string; label: string }[]>([
  { value: '', label: '全部分类' },
  { value: 'none', label: '未分类' },
])

const categorySelectOptions = computed(() => categoryFilterOptions.value)

async function loadCategoryFilterOptions() {
  try {
    const data = await fetchUserLiteratureCategories({ page: 1, page_size: 200 })
    const fromApi = (data.items ?? []).map((c) => ({
      value: c.id,
      label: c.name,
    }))
    categoryFilterOptions.value = [
      { value: '', label: '全部分类' },
      { value: 'none', label: '未分类' },
      ...fromApi,
    ]
  } catch {
    /* 保留默认「全部 / 未分类」 */
  }
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
    const data = await fetchUserLiterature({
      page: targetPage,
      page_size: PAPER_LIST_PAGE_SIZE,
      q: filterKeywordApplied.value || undefined,
      category_id: filterCategoryId.value || undefined,
    })
    items.value = data.items ?? []
    total.value = data.total ?? items.value.length
    page.value = data.page ?? targetPage
    if (items.value.length === 0 && total.value > 0 && page.value > 1) {
      await loadPage(page.value - 1)
    }
  } catch (e) {
    items.value = []
    total.value = 0
    const msg = e instanceof Error ? e.message : '加载文献列表失败'
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  void loadPage(p)
}

function applyFilters() {
  filterKeywordApplied.value = filterKeywordDraft.value.trim()
  page.value = 1
  void loadPage(1)
}

function onFilterKeywordKeydown(ev: KeyboardEvent) {
  if (ev.key === 'Enter') applyFilters()
}

function openDetail(item: PaperUserLiteratureItem) {
  detailItem.value = item
  detailOpen.value = true
}

function closeDetail() {
  detailOpen.value = false
  detailItem.value = null
}

function resetUploadInput() {
  uploadFile.value = null
  uploadTitle.value = ''
  uploadNote.value = ''
  if (uploadInputRef.value) uploadInputRef.value.value = ''
}

function openUploadDialog() {
  resetUploadInput()
  uploadOpen.value = true
}

function closeUploadDialog() {
  if (uploadSubmitting.value) return
  uploadOpen.value = false
  resetUploadInput()
}

function openUploadPicker() {
  if (uploadSubmitting.value) return
  uploadInputRef.value?.click()
}

function onUploadFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  uploadFile.value = file
  if (file && !uploadTitle.value.trim()) {
    uploadTitle.value = file.name.replace(/\.[^.]+$/, '')
  }
}

function onUploadDrop(ev: DragEvent) {
  if (uploadSubmitting.value) return
  const file = ev.dataTransfer?.files?.[0]
  if (file) {
    uploadFile.value = file
    if (!uploadTitle.value.trim()) {
      uploadTitle.value = file.name.replace(/\.[^.]+$/, '')
    }
  }
}

async function submitUpload() {
  const file = uploadFile.value
  if (!file) {
    ElMessage.warning('请选择要上传的文献文件')
    return
  }
  uploadSubmitting.value = true
  try {
    ElMessage.info(`上传文献「${file.name}」功能即将接入`)
    closeUploadDialog()
  } finally {
    uploadSubmitting.value = false
  }
}

defineExpose({ openUploadDialog })

void loadCategoryFilterOptions()
void loadPage(1)
</script>

<template>
  <section class="ul-panel">
    <div class="ul-tabs" role="tablist" aria-label="文献管理">
      <button
        type="button"
        role="tab"
        class="ul-tab"
        :class="{ 'ul-tab--active': activeTab === 'list' }"
        :aria-selected="activeTab === 'list'"
        @click="activeTab = 'list'"
      >
        文献列表
      </button>
      <button
        type="button"
        role="tab"
        class="ul-tab"
        :class="{ 'ul-tab--active': activeTab === 'categories' }"
        :aria-selected="activeTab === 'categories'"
        @click="activeTab = 'categories'"
      >
        分类管理
      </button>
    </div>

    <div v-show="activeTab === 'list'" class="ul-tab-pane" role="tabpanel">
    <div class="ul-filter">
      <div class="ul-filter-label-row">
        <label class="ul-filter-label ul-filter-label--search" for="ul-filter-keyword">搜索</label>
        <label class="ul-filter-label ul-filter-label--category" for="ul-filter-category">
          分类
        </label>
      </div>
      <div class="ul-filter-control-row">
        <div class="ul-filter-control-slot ul-filter-control-slot--search">
          <input
            id="ul-filter-keyword"
            v-model="filterKeywordDraft"
            type="search"
            class="ul-filter-input"
            autocomplete="off"
            @keydown="onFilterKeywordKeydown"
          />
        </div>
        <div id="ul-filter-category" class="ul-filter-control-slot ul-filter-control-slot--category">
          <PaperSelect v-model="filterCategoryId" :options="categorySelectOptions" />
        </div>
        <div class="ul-filter-control-slot ul-filter-control-slot--action">
          <button type="button" class="ul-filter-btn ul-filter-btn--primary" @click="applyFilters">
            查询
          </button>
        </div>
      </div>
    </div>

    <p v-if="loading" class="ul-meta">正在加载文献列表…</p>
    <p v-else-if="!total" class="ul-empty">
      {{
        filterKeywordApplied || filterCategoryId
          ? '无匹配的文献，请调整筛选条件。'
          : '暂无文献（status=active）。'
      }}
    </p>
    <template v-else>
      <table class="ul-table">
        <colgroup>
          <col class="ul-col-title" />
          <col class="ul-col-cat" />
          <col class="ul-col-created" />
          <col class="ul-col-from" />
          <col class="ul-col-ext" />
          <col class="ul-col-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>标题</th>
            <th>分类</th>
            <th>创建时间</th>
            <th>来源</th>
            <th>格式</th>
            <th class="ul-col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td class="ul-col-title" :title="item.literature_title">
              {{ item.literature_title }}
            </td>
            <td class="ul-col-cat">{{ formatUserLiteratureCategories(item.category_names) }}</td>
            <td class="ul-col-created">{{ formatCreatedAt(item.created_at) }}</td>
            <td class="ul-col-from">
              <span
                class="ul-from-tag"
                :class="`ul-from-tag--${userLiteratureFromTone(item.from)}`"
              >
                {{ formatUserLiteratureFrom(item.from) }}
              </span>
            </td>
            <td class="ul-col-ext">{{ item.file_ext || '—' }}</td>
            <td class="ul-col-actions">
              <button type="button" class="ul-btn-link" @click="openDetail(item)">查看</button>
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
    </div>

    <PaperUserLiteratureCategoryPanel
      v-show="activeTab === 'categories'"
      :active="activeTab === 'categories'"
    />

    <Teleport to="body">
      <div
        v-if="uploadOpen"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ul-upload-title"
        @click.self="closeUploadDialog"
        @keydown.escape="closeUploadDialog"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="ul-upload-title" class="paper-modal-title">上传文献</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="uploadSubmitting"
              @click="closeUploadDialog"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <p class="ul-upload-lead">
              支持 PDF、BibTeX、纯文本等；上传后将写入「文献管理」语料库（本地上传）。
            </p>
            <label class="ul-upload-field">
              <span class="ul-upload-field-label">文献标题</span>
              <input
                v-model="uploadTitle"
                type="text"
                class="ul-upload-input"
                placeholder="默认可取自文件名"
                maxlength="512"
                :disabled="uploadSubmitting"
              />
            </label>
            <label class="ul-upload-field">
              <span class="ul-upload-field-label">备注（可选）</span>
              <textarea
                v-model="uploadNote"
                class="ul-upload-textarea"
                rows="2"
                placeholder="研究笔记、阅读进度等"
                :disabled="uploadSubmitting"
              />
            </label>
            <div
              class="ul-upload-drop"
              :class="{
                'ul-upload-drop--has-file': !!uploadFile,
                'ul-upload-drop--disabled': uploadSubmitting,
              }"
              role="button"
              tabindex="0"
              @click="openUploadPicker"
              @keydown.enter.prevent="openUploadPicker"
              @keydown.space.prevent="openUploadPicker"
              @dragover.prevent
              @drop.prevent="onUploadDrop"
            >
              <input
                ref="uploadInputRef"
                type="file"
                class="ul-upload-file-input"
                :accept="REFERENCE_UPLOAD_ACCEPT"
                :disabled="uploadSubmitting"
                @change="onUploadFileChange"
              />
              <p v-if="!uploadFile" class="ul-upload-drop-hint">
                点击或拖拽文献文件（PDF / Bib / TXT / Markdown / JSON）
              </p>
              <p v-else class="ul-upload-drop-name">{{ uploadFile.name }}</p>
            </div>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="uploadSubmitting || !uploadFile"
              @click="submitUpload"
            >
              {{ uploadSubmitting ? '上传中…' : '上传' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="uploadSubmitting"
              @click="closeUploadDialog"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

    <div v-if="detailOpen && detailItem" class="ul-modal-backdrop" @click.self="closeDetail">
      <div class="ul-modal" role="dialog" aria-labelledby="ul-detail-title">
        <header class="ul-modal-head">
          <h3 id="ul-detail-title">文献详情</h3>
          <button type="button" class="ul-modal-close" aria-label="关闭" @click="closeDetail">
            ×
          </button>
        </header>
        <dl class="ul-detail">
          <dt>题名</dt>
          <dd>{{ detailItem.literature_title }}</dd>
          <dt>分类</dt>
          <dd>{{ formatUserLiteratureCategories(detailItem.category_names) }}</dd>
          <dt>文件名</dt>
          <dd>{{ detailItem.file_name }}</dd>
          <dt>来源</dt>
          <dd>{{ formatUserLiteratureFrom(detailItem.from) }}</dd>
          <dt>格式</dt>
          <dd>{{ detailItem.file_ext }}</dd>
          <dt>大小</dt>
          <dd>{{ formatFileSizeBytes(detailItem.size_bytes) }}</dd>
          <dt>上传时间</dt>
          <dd>{{ formatCreatedAt(detailItem.created_at) }}</dd>
          <dt v-if="detailItem.note">备注</dt>
          <dd v-if="detailItem.note">{{ detailItem.note }}</dd>
          <dt v-if="detailItem.storage_uri">存储路径</dt>
          <dd v-if="detailItem.storage_uri" class="ul-mono">{{ detailItem.storage_uri }}</dd>
        </dl>
      </div>
    </div>
  </section>
</template>

<style scoped>
.ul-panel {
  padding: 20px 22px 24px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.ul-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 0 2px;
  margin-bottom: 20px;
  border-bottom: 1px solid #e2e8f0;
}

.ul-tab {
  position: relative;
  flex: 0 1 auto;
  min-width: 112px;
  padding: 12px 20px;
  margin-bottom: -1px;
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  border-radius: 8px 8px 0 0;
  transition:
    color 0.15s,
    background 0.15s,
    border-color 0.15s;
}

.ul-tab:hover:not(.ul-tab--active) {
  color: #334155;
  background: #f8fafc;
}

.ul-tab--active {
  font-weight: 600;
  color: #1e293b;
  background: #fff;
  border-bottom-color: #6366f1;
  box-shadow: inset 0 -1px 0 #fff;
}

.ul-tab-pane {
  padding-top: 4px;
}

.ul-filter {
  --ul-filter-control-h: 38px;

  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 18px;
}

.ul-filter-label-row,
.ul-filter-control-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 14px;
  align-items: center;
}

.ul-filter-control-row {
  align-items: stretch;
}

.ul-filter-label {
  font-size: 14px;
  font-weight: 700;
  line-height: 1.5;
  color: #64748b;
}

.ul-filter-label--search {
  flex: 0 0 auto;
  width: 280px;
  max-width: 100%;
}

.ul-filter-label--category {
  flex: 0 0 auto;
  width: 168px;
}

.ul-filter-control-slot {
  display: flex;
  align-items: flex-end;
  height: var(--ul-filter-control-h);
  box-sizing: border-box;
}

.ul-filter-control-slot--search {
  flex: 0 0 auto;
  width: 280px;
  max-width: 100%;
}

.ul-filter-control-slot--category {
  flex: 0 0 auto;
  width: 168px;
}

.ul-filter-control-slot--category :deep(.paper-select) {
  width: 100%;
}

.ul-filter-control-slot--action {
  flex-shrink: 0;
}

.ul-filter-input {
  width: 100%;
  height: 100%;
  padding: 0 12px;
  margin: 0;
  font-size: 14px;
  line-height: 1.25;
  color: #1e293b;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-sizing: border-box;
  appearance: none;
  -webkit-appearance: none;
}

.ul-filter-input::-webkit-search-decoration,
.ul-filter-input::-webkit-search-cancel-button {
  -webkit-appearance: none;
}

.ul-filter-input:focus {
  outline: none;
  border-color: #818cf8;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

.ul-filter :deep(.paper-select-trigger) {
  width: 100%;
  height: 100%;
  min-height: 0;
  padding: 0 10px 0 12px;
  font-size: 13px;
  background: #fff;
  border-radius: 8px;
  box-sizing: border-box;
}

.ul-filter-control-slot--category :deep(.paper-select) {
  height: 100%;
}

.ul-filter :deep(.paper-select-option) {
  font-size: 13px;
}

.ul-filter-btn {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  height: calc(var(--ul-filter-control-h) - 2px);
  padding: 0 16px;
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1;
  color: #475569;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-sizing: border-box;
}

.ul-filter-btn--primary {
  color: #fff;
  background: linear-gradient(135deg, #7c3aed, #6366f1);
  border: 1px solid transparent;
}

.ul-filter-btn:hover {
  border-color: #cbd5e1;
}

.ul-meta,
.ul-empty {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.ul-empty {
  padding: 32px 16px;
  text-align: center;
  background: #f8fafc;
  border-radius: 12px;
}

.ul-table {
  width: 100%;
  font-size: 14px;
  table-layout: fixed;
  border-collapse: collapse;
}

.ul-table col.ul-col-title {
  width: 28%;
}

.ul-table col.ul-col-cat {
  width: 14%;
}

.ul-table col.ul-col-created {
  width: 15%;
}

.ul-table col.ul-col-from {
  width: 17%;
}

.ul-table col.ul-col-ext {
  width: 6%;
}

.ul-table col.ul-col-actions {
  width: 18%;
}

.ul-col-cat {
  overflow: hidden;
  color: #475569;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ul-table th,
.ul-table td {
  padding: 13px 12px;
  line-height: 1.5;
  text-align: left;
  vertical-align: middle;
  border-bottom: 1px solid #e2e8f0;
}

.ul-table th {
  font-weight: 700;
  color: #64748b;
}

.ul-col-title {
  overflow: hidden;
  font-weight: 400;
  color: #1e1b4b;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ul-col-created {
  font-weight: 400;
  color: #334155;
  font-variant-numeric: tabular-nums;
}

.ul-col-from {
  white-space: nowrap;
}

.ul-col-ext {
  color: #475569;
  text-transform: lowercase;
}

.ul-col-actions {
  text-align: right;
  white-space: nowrap;
}

.ul-from-tag {
  display: inline-block;
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.35;
  border-radius: 999px;
}

.ul-from-tag--upload {
  color: #1d4ed8;
  background: #dbeafe;
}

.ul-from-tag--arxiv {
  color: #b45309;
  background: #fef3c7;
}

.ul-from-tag--catalog {
  color: #047857;
  background: #d1fae5;
}

.ul-from-tag--default {
  color: #475569;
  background: #f1f5f9;
}

.ul-btn-link {
  padding: 5px 10px;
  font-size: 13px;
  font-weight: 600;
  color: #6366f1;
  cursor: pointer;
  background: transparent;
  border: none;
  border-radius: 7px;
}

.ul-btn-link:hover {
  background: #eef2ff;
}

.ul-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(15, 23, 42, 0.45);
}

.ul-modal {
  width: min(520px, 100%);
  max-height: min(80vh, 640px);
  overflow: auto;
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 24px 48px rgba(30, 27, 75, 0.18);
}

.ul-modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #e2e8f0;
}

.ul-modal-head h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #1e1b4b;
}

.ul-modal-close {
  font-size: 22px;
  line-height: 1;
  color: #64748b;
  cursor: pointer;
  background: none;
  border: none;
}

.ul-detail {
  display: grid;
  grid-template-columns: 88px 1fr;
  gap: 10px 12px;
  padding: 16px 20px 20px;
  margin: 0;
  font-size: 14px;
}

.ul-detail dt {
  margin: 0;
  font-weight: 600;
  color: #64748b;
}

.ul-detail dd {
  margin: 0;
  color: #1e293b;
  word-break: break-word;
}

.ul-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  color: #475569;
}

.ul-upload-lead {
  margin: 0 0 14px;
  font-size: 13px;
  line-height: 1.55;
  color: #64748b;
}

.ul-upload-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 14px;
}

.ul-upload-field-label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.ul-upload-input,
.ul-upload-textarea {
  width: 100%;
  padding: 9px 12px;
  font-size: 14px;
  color: #1e293b;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-sizing: border-box;
}

.ul-upload-input:focus,
.ul-upload-textarea:focus {
  outline: none;
  border-color: #818cf8;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}

.ul-upload-textarea {
  resize: vertical;
  min-height: 56px;
}

.ul-upload-file-input {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  pointer-events: none;
}

.ul-upload-drop {
  position: relative;
  padding: 22px 16px;
  text-align: center;
  cursor: pointer;
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  transition:
    border-color 0.15s ease,
    background 0.15s ease;
}

.ul-upload-drop:hover:not(.ul-upload-drop--disabled) {
  background: #f5f3ff;
  border-color: #818cf8;
}

.ul-upload-drop--has-file {
  background: #fff;
  border-color: #a5b4fc;
  border-style: solid;
}

.ul-upload-drop--disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.ul-upload-drop-hint {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #64748b;
}

.ul-upload-drop-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
  color: var(--atm-text, #1e1b4b);
  word-break: break-all;
}
</style>
