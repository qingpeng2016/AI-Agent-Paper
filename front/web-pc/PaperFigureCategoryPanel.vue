<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createFigureCategory,
  updateFigureCategory,
  fetchAllFigures,
  fetchCategoryFigureIds,
  fetchFigureCategories,
  PAPER_LIST_PAGE_SIZE,
  setCategoryFigureMembership,
  type PaperFigureCategoryItem,
  type PaperFigureItem,
} from '@/api/figures'
import PaperListPagination from './PaperListPagination.vue'
import PaperFigureCategoryByManuscriptPanel from './PaperFigureCategoryByManuscriptPanel.vue'

export type CategoryManageSubTabId = 'by-name' | 'by-manuscript'

const props = defineProps<{
  active: boolean
}>()

const categorySubTab = ref<CategoryManageSubTabId>('by-name')

const emit = defineEmits<{
  changed: []
}>()

const items = ref<PaperFigureCategoryItem[]>([])
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
    const data = await fetchFigureCategories({
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
    const msg = e instanceof Error ? e.message : '加载分类失败'
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  void loadPage(p)
}

const createOpen = ref(false)
const createName = ref('')
const createSubmitting = ref(false)

function openCreateDialog() {
  createName.value = ''
  createOpen.value = true
}

function closeCreateDialog() {
  if (createSubmitting.value) return
  createOpen.value = false
}

async function submitCreateCategory() {
  const name = createName.value.trim()
  if (!name) {
    ElMessage.warning('请填写分类名称')
    return
  }
  if (name.length > 128) {
    ElMessage.warning('分类名称不能超过 128 个字符')
    return
  }
  createSubmitting.value = true
  try {
    await createFigureCategory({ name })
    ElMessage.success('分类已创建')
    createOpen.value = false
    emit('changed')
    await loadPage(1)
  } catch (e) {
    const msg = e instanceof Error ? e.message : '创建分类失败'
    ElMessage.error(msg)
  } finally {
    createSubmitting.value = false
  }
}

function onAddCategory() {
  openCreateDialog()
}

const editOpen = ref(false)
const editCategory = ref<PaperFigureCategoryItem | null>(null)
const editName = ref('')
const editSubmitting = ref(false)

function openEditDialog(item: PaperFigureCategoryItem) {
  editCategory.value = item
  editName.value = item.name
  editOpen.value = true
}

function closeEditDialog() {
  if (editSubmitting.value) return
  editOpen.value = false
  editCategory.value = null
}

async function submitEditCategory() {
  if (!editCategory.value) return
  const name = editName.value.trim()
  if (!name) {
    ElMessage.warning('请填写分类名称')
    return
  }
  if (name.length > 128) {
    ElMessage.warning('分类名称不能超过 128 个字符')
    return
  }
  editSubmitting.value = true
  try {
    await updateFigureCategory({ id: editCategory.value.id, name })
    ElMessage.success('分类已更新')
    editOpen.value = false
    editCategory.value = null
    emit('changed')
    await loadPage(page.value)
  } catch (e) {
    const msg = e instanceof Error ? e.message : '更新分类失败'
    ElMessage.error(msg)
  } finally {
    editSubmitting.value = false
  }
}

function onEditCategory(item: PaperFigureCategoryItem) {
  openEditDialog(item)
}

const manageOpen = ref(false)
const manageCategory = ref<PaperFigureCategoryItem | null>(null)
const manageLoading = ref(false)
const manageSubmitting = ref(false)
const manageSearchDraft = ref('')
const manageSearchApplied = ref('')
const manageFigures = ref<PaperFigureItem[]>([])
const selectedFigureIds = ref<Set<string>>(new Set())

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

async function loadManageFigures() {
  manageFigures.value = await fetchAllFigures({
    q: manageSearchApplied.value || undefined,
  })
}

async function openManageCategory(item: PaperFigureCategoryItem) {
  manageCategory.value = item
  manageSearchDraft.value = ''
  manageSearchApplied.value = ''
  manageOpen.value = true
  manageLoading.value = true
  try {
    const membership = await fetchCategoryFigureIds(item.id)
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

function closeManageDialog() {
  if (manageSubmitting.value) return
  manageOpen.value = false
  manageCategory.value = null
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

async function submitManageCategory() {
  if (!manageCategory.value) return
  manageSubmitting.value = true
  try {
    await setCategoryFigureMembership({
      category_id: manageCategory.value.id,
      figure_ids: Array.from(selectedFigureIds.value),
    })
    ElMessage.success('已更新分类下的图表')
    manageOpen.value = false
    manageCategory.value = null
    emit('changed')
  } catch (e) {
    const msg = e instanceof Error ? e.message : '保存失败'
    ElMessage.error(msg)
  } finally {
    manageSubmitting.value = false
  }
}

function onManageCategory(item: PaperFigureCategoryItem) {
  void openManageCategory(item)
}

watch(
  () => [props.active, categorySubTab.value] as const,
  ([on, tab]) => {
    if (on && tab === 'by-name') void loadPage(page.value)
  },
  { immediate: true },
)
</script>

<template>
  <div class="ul-cat-pane">
    <div class="ul-cat-header">
      <div class="ul-cat-subtabs" role="tablist" aria-label="分类视图">
        <button
          type="button"
          role="tab"
          class="ul-cat-subtab"
          :class="{ 'ul-cat-subtab--active': categorySubTab === 'by-name' }"
          :aria-selected="categorySubTab === 'by-name'"
          @click="categorySubTab = 'by-name'"
        >
          按名称
        </button>
        <button
          type="button"
          role="tab"
          class="ul-cat-subtab"
          :class="{ 'ul-cat-subtab--active': categorySubTab === 'by-manuscript' }"
          :aria-selected="categorySubTab === 'by-manuscript'"
          @click="categorySubTab = 'by-manuscript'"
        >
          按论文
        </button>
      </div>
      <button type="button" class="ul-cat-btn-primary" @click="onAddCategory">新建分类</button>
    </div>

    <div v-show="categorySubTab === 'by-name'" class="ul-cat-subpane" role="tabpanel">
    <p v-if="loading" class="ul-meta">正在加载分类…</p>
    <p v-else-if="!total" class="ul-empty">暂无分类。可先执行分类种子 SQL。</p>
    <template v-else>
      <table class="ul-table">
        <colgroup>
          <col class="ul-col-cat-name" />
          <col class="ul-col-cat-created" />
          <col class="ul-col-cat-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>分类名称</th>
            <th>创建时间</th>
            <th class="ul-col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td class="ul-col-title">{{ item.name }}</td>
            <td class="ul-col-created">{{ formatCreatedAt(item.created_at) }}</td>
            <td class="ul-col-actions">
              <button type="button" class="ul-btn-link" @click="onManageCategory(item)">管理</button>
              <button type="button" class="ul-btn-link" @click="onEditCategory(item)">编辑</button>
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

    <PaperFigureCategoryByManuscriptPanel
      v-show="categorySubTab === 'by-manuscript'"
      :active="active && categorySubTab === 'by-manuscript'"
      @changed="emit('changed')"
    />

    <Teleport to="body">
      <div
        v-if="createOpen"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ul-cat-create-title"
        @click.self="closeCreateDialog"
        @keydown.escape="closeCreateDialog"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="ul-cat-create-title" class="paper-modal-title">新建分类</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="createSubmitting"
              @click="closeCreateDialog"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <label class="ul-cat-form-field">
              <span class="ul-cat-form-label">分类名称</span>
              <input
                v-model="createName"
                type="text"
                class="ul-cat-form-input"
                maxlength="128"
                :disabled="createSubmitting"
                @keydown.enter.prevent="submitCreateCategory"
              />
            </label>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="createSubmitting || !createName.trim()"
              @click="submitCreateCategory"
            >
              {{ createSubmitting ? '提交中…' : '创建' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="createSubmitting"
              @click="closeCreateDialog"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="editOpen && editCategory"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ul-cat-edit-title"
        @click.self="closeEditDialog"
        @keydown.escape="closeEditDialog"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="ul-cat-edit-title" class="paper-modal-title">编辑分类</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="editSubmitting"
              @click="closeEditDialog"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <label class="ul-cat-form-field">
              <span class="ul-cat-form-label">分类名称</span>
              <input
                v-model="editName"
                type="text"
                class="ul-cat-form-input"
                maxlength="128"
                :disabled="editSubmitting"
                @keydown.enter.prevent="submitEditCategory"
              />
            </label>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="editSubmitting || !editName.trim()"
              @click="submitEditCategory"
            >
              {{ editSubmitting ? '保存中…' : '保存' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="editSubmitting"
              @click="closeEditDialog"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="manageOpen && manageCategory"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ul-cat-manage-title"
        @click.self="closeManageDialog"
        @keydown.escape="closeManageDialog"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel ul-cat-manage-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="ul-cat-manage-title" class="paper-modal-title">
              管理分类「{{ manageCategory.name }}」
            </h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="manageSubmitting"
              @click="closeManageDialog"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body ul-cat-manage-body">
            <p class="ul-cat-manage-lead">勾选要归入该分类的图表（可多选）。取消勾选将从本分类移出。</p>
            <div class="ul-cat-manage-toolbar">
              <input
                v-model="manageSearchDraft"
                type="search"
                class="ul-cat-form-input ul-cat-manage-search"
                :disabled="manageLoading || manageSubmitting"
                @keydown.enter.prevent="applyManageSearch"
              />
              <button
                type="button"
                class="ul-cat-btn-secondary"
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
              @click="submitManageCategory"
            >
              {{ manageSubmitting ? '保存中…' : '保存' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="manageSubmitting"
              @click="closeManageDialog"
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
.ul-cat-pane {
  padding-top: 4px;
}

.ul-cat-header {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.ul-cat-subtabs {
  display: flex;
  gap: 4px;
  align-items: center;
}

.ul-cat-subtab {
  padding: 8px 14px;
  font-size: 14px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  background: transparent;
  border: none;
  border-radius: 8px;
}

.ul-cat-subtab:hover:not(.ul-cat-subtab--active) {
  color: #475569;
  background: #f1f5f9;
}

.ul-cat-subtab--active {
  color: #5b21b6;
  background: #ede9fe;
}

.ul-cat-subpane {
  padding-top: 0;
}

.ul-cat-btn-primary {
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: linear-gradient(135deg, #7c3aed, #6366f1);
  border: none;
  border-radius: 8px;
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

.ul-table col.ul-col-cat-name {
  width: 50%;
}

.ul-table col.ul-col-cat-created {
  width: 32%;
}

.ul-table col.ul-col-cat-actions {
  width: 18%;
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

.ul-cat-form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 14px;
}

.ul-cat-form-label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.ul-cat-form-input {
  width: 100%;
  padding: 9px 12px;
  font-size: 14px;
  color: #1e293b;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-sizing: border-box;
}

.ul-cat-form-input:focus {
  outline: none;
  border-color: #818cf8;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

.ul-cat-form-hint {
  margin: -4px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: #94a3b8;
}

.ul-cat-btn-secondary {
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 600;
  color: #475569;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
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
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.ul-cat-manage-search {
  flex: 1 1 200px;
  min-width: 0;
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
  font-weight: 600;
  color: #475569;
  cursor: pointer;
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
  min-width: 0;
  font-size: 14px;
  line-height: 1.45;
  color: #1e1b4b;
  word-break: break-word;
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

.ul-col-actions .ul-btn-link + .ul-btn-link {
  margin-left: 4px;
}
</style>
