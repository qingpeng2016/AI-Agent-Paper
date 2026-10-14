<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchAllFigures,
  fetchManuscriptFigureIds,
  fetchManuscriptsWithFigures,
  formatFigureCategories,
  PAPER_LIST_PAGE_SIZE,
  setManuscriptFigureMembership,
  type PaperManuscriptWithFiguresItem,
  type PaperFigureItem,
} from '@/api/figures'
import PaperListPagination from './PaperListPagination.vue'

const props = defineProps<{
  active: boolean
}>()

const emit = defineEmits<{
  changed: []
}>()

const items = ref<PaperManuscriptWithFiguresItem[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)

const manageOpen = ref(false)
const manageManuscript = ref<PaperManuscriptWithFiguresItem | null>(null)
const manageLoading = ref(false)
const manageSubmitting = ref(false)
const manageSearchDraft = ref('')
const manageSearchApplied = ref('')
const manageFigures = ref<PaperFigureItem[]>([])
const selectedFigureIds = ref<Set<string>>(new Set())

function formatCreatedAt(iso?: string): string {
  if (!iso?.trim()) return '—'
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

function cloneSelectedSet(from: Set<string>): Set<string> {
  return new Set(from)
}

function isFigureSelected(id: string): boolean {
  return selectedFigureIds.value.has(id)
}

function toggleFigureSelection(id: string) {
  const next = cloneSelectedSet(selectedFigureIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedFigureIds.value = next
}

const manageAllVisibleSelected = computed(() => {
  const rows = manageFigures.value
  if (!rows.length) return false
  return rows.every((r) => selectedFigureIds.value.has(r.id))
})

const manageSomeVisibleSelected = computed(() => {
  const rows = manageFigures.value
  if (!rows.length) return false
  const n = rows.filter((r) => selectedFigureIds.value.has(r.id)).length
  return n > 0 && n < rows.length
})

function toggleSelectAllVisible() {
  const rows = manageFigures.value
  if (!rows.length) return
  const next = cloneSelectedSet(selectedFigureIds.value)
  if (manageAllVisibleSelected.value) {
    for (const r of rows) next.delete(r.id)
  } else {
    for (const r of rows) next.add(r.id)
  }
  selectedFigureIds.value = next
}

async function loadPage(targetPage = page.value) {
  loading.value = true
  try {
    const data = await fetchManuscriptsWithFigures({
      page: targetPage,
      page_size: PAPER_LIST_PAGE_SIZE,
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

async function loadManageFigures() {
  manageFigures.value = await fetchAllFigures({
    q: manageSearchApplied.value || undefined,
  })
}

async function openManage(item: PaperManuscriptWithFiguresItem) {
  manageManuscript.value = item
  manageSearchDraft.value = ''
  manageSearchApplied.value = ''
  manageOpen.value = true
  manageLoading.value = true
  try {
    const membership = await fetchManuscriptFigureIds(item.id)
    selectedFigureIds.value = new Set(membership.figure_ids ?? [])
    await loadManageFigures()
  } catch (e) {
    manageOpen.value = false
    const msg = e instanceof Error ? e.message : '加载图表失败'
    ElMessage.error(msg)
  } finally {
    manageLoading.value = false
  }
}

function closeManage() {
  if (manageSubmitting.value) return
  manageOpen.value = false
  manageManuscript.value = null
}

async function applyManageSearch() {
  manageSearchApplied.value = manageSearchDraft.value.trim()
  manageLoading.value = true
  try {
    await loadManageFigures()
  } catch (e) {
    const msg = e instanceof Error ? e.message : '搜索失败'
    ElMessage.error(msg)
  } finally {
    manageLoading.value = false
  }
}

async function submitManage() {
  if (!manageManuscript.value) return
  manageSubmitting.value = true
  try {
    await setManuscriptFigureMembership({
      manuscript_id: manageManuscript.value.id,
      figure_ids: Array.from(selectedFigureIds.value),
    })
    ElMessage.success('已更新论文关联图表')
    manageOpen.value = false
    manageManuscript.value = null
    emit('changed')
    await loadPage(page.value)
  } catch (e) {
    const msg = e instanceof Error ? e.message : '保存失败'
    ElMessage.error(msg)
  } finally {
    manageSubmitting.value = false
  }
}

watch(
  () => props.active,
  (on) => {
    if (on) void loadPage(page.value)
  },
  { immediate: true },
)
</script>

<template>
  <div class="ul-ms-lit-pane">
    <p v-if="loading" class="ul-meta">正在加载论文…</p>
    <p v-else-if="!total" class="ul-empty">暂无论文。</p>
    <template v-else>
      <table class="ul-table">
        <colgroup>
          <col class="ul-col-ms-title" />
          <col class="ul-col-ms-lit" />
          <col class="ul-col-ms-created" />
          <col class="ul-col-ms-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>论文标题</th>
            <th>关联图表</th>
            <th>创建时间</th>
            <th class="ul-col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td class="ul-col-title">
              <span class="ul-ms-title-text" :title="item.title">{{ item.title }}</span>
              <span v-if="item.is_current" class="ul-ms-current-tag">当前</span>
            </td>
            <td class="ul-col-lit" :title="formatFigureCategories(item.figure_titles)">
              {{ formatFigureCategories(item.figure_titles) }}
            </td>
            <td class="ul-col-created">{{ formatCreatedAt(item.created_at) }}</td>
            <td class="ul-col-actions">
              <button type="button" class="ul-btn-link" @click="openManage(item)">管理</button>
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

    <Teleport to="body">
      <div
        v-if="manageOpen && manageManuscript"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ul-ms-lit-manage-title"
        @click.self="closeManage"
        @keydown.escape="closeManage"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel ul-cat-manage-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="ul-ms-lit-manage-title" class="paper-modal-title">
              论文图表 — {{ manageManuscript.title }}
            </h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="manageSubmitting"
              @click="closeManage"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body ul-cat-manage-body">
            <p class="ul-cat-manage-lead">
              勾选要关联到该论文的图表（可多选）。同一张图表可关联多篇论文。
            </p>
            <div class="ul-cat-manage-toolbar">
              <input
                v-model="manageSearchDraft"
                type="search"
                class="ul-cat-form-input ul-cat-manage-search"
                placeholder="搜索图表标题"
                :disabled="manageLoading || manageSubmitting"
                @keydown.enter.prevent="applyManageSearch"
              />
              <button
                type="button"
                class="ul-ms-btn-secondary"
                :disabled="manageLoading || manageSubmitting"
                @click="applyManageSearch"
              >
                查询
              </button>
            </div>
            <p v-if="manageLoading" class="ul-meta">正在加载图表…</p>
            <p v-else-if="!manageFigures.length" class="ul-cat-manage-empty">暂无图表</p>
            <div v-else class="ul-cat-manage-list-wrap">
              <label class="ul-cat-manage-select-all">
                <input
                  type="checkbox"
                  :checked="manageAllVisibleSelected"
                  :indeterminate="manageSomeVisibleSelected"
                  :disabled="manageSubmitting"
                  @change="toggleSelectAllVisible"
                />
                <span>全选当前列表（共 {{ manageFigures.length }} 张）</span>
              </label>
              <ul class="ul-cat-manage-list" role="listbox" aria-multiselectable="true">
                <li v-for="lit in manageFigures" :key="lit.id" class="ul-cat-manage-item">
                  <label class="ul-cat-manage-item-label">
                    <input
                      type="checkbox"
                      :checked="isFigureSelected(lit.id)"
                      :disabled="manageSubmitting"
                      @change="toggleFigureSelection(lit.id)"
                    />
                    <span class="ul-cat-manage-item-title" :title="lit.title">
                      {{ lit.title }}
                    </span>
                  </label>
                </li>
              </ul>
            </div>
            <p class="ul-cat-manage-count">已选 {{ selectedFigureIds.size }} 张</p>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="manageSubmitting || manageLoading"
              @click="submitManage"
            >
              {{ manageSubmitting ? '保存中…' : '保存' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="manageSubmitting"
              @click="closeManage"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.ul-ms-lit-pane {
  padding-top: 4px;
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

.ul-table col.ul-col-ms-title {
  width: 40%;
}

.ul-table col.ul-col-ms-lit {
  width: 30%;
}

.ul-table col.ul-col-ms-created {
  width: 18%;
}

.ul-table col.ul-col-ms-actions {
  width: 12%;
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
  color: #1e1b4b;
}

.ul-ms-title-text {
  display: inline;
  vertical-align: middle;
}

.ul-ms-current-tag {
  display: inline-block;
  margin-left: 8px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 600;
  color: #6d28d9;
  vertical-align: middle;
  background: #ede9fe;
  border-radius: 999px;
}

.ul-col-lit {
  overflow: hidden;
  color: #475569;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ul-col-created {
  font-variant-numeric: tabular-nums;
  color: #334155;
}

.ul-col-actions {
  text-align: right;
  white-space: nowrap;
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

.ul-cat-manage-panel {
  width: min(640px, calc(100vw - 32px));
  max-width: 640px;
}

.ul-cat-manage-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.ul-cat-manage-lead {
  margin: 0;
  font-size: 13px;
  line-height: 1.55;
  color: #64748b;
}

.ul-cat-manage-toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
}

.ul-cat-form-input {
  box-sizing: border-box;
  width: 100%;
  padding: 8px 12px;
  font-size: 14px;
  color: #1e1b4b;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.ul-cat-manage-search {
  flex: 1;
}

.ul-ms-btn-secondary {
  flex-shrink: 0;
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 600;
  color: #475569;
  cursor: pointer;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.ul-ms-btn-secondary:hover:not(:disabled) {
  background: #f1f5f9;
}

.ul-ms-btn-secondary:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.ul-cat-manage-list-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ul-cat-manage-select-all {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 13px;
  color: #475569;
}

.ul-cat-manage-list {
  max-height: min(360px, 50vh);
  padding: 0;
  margin: 0;
  overflow: auto;
  list-style: none;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.ul-cat-manage-item {
  border-bottom: 1px solid #f1f5f9;
}

.ul-cat-manage-item:last-child {
  border-bottom: none;
}

.ul-cat-manage-item-label {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 10px 12px;
  cursor: pointer;
}

.ul-cat-manage-item-label:hover {
  background: #f8fafc;
}

.ul-cat-manage-item-title {
  flex: 1;
  font-size: 14px;
  line-height: 1.45;
  color: #1e1b4b;
}

.ul-cat-manage-empty {
  margin: 0;
  padding: 24px 12px;
  font-size: 14px;
  color: #94a3b8;
  text-align: center;
}

.ul-cat-manage-count {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}
</style>
