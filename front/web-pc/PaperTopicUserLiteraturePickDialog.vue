<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { fetchPaperManuscripts } from '@/api/manuscripts'
import {
  fetchUserLiterature,
  fetchUserLiteratureCategories,
  type PaperUserLiteratureItem,
} from '@/api/userLiterature'
import PaperSelect from './PaperSelect.vue'

const props = defineProps<{
  open: boolean
  /** 打开弹窗时的已选文献 id */
  initialSelectedIds: string[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  close: []
  confirm: [ids: string[]]
}>()

const loading = ref(false)
const submitting = ref(false)
const filterCategoryId = ref('')
const filterManuscriptId = ref('')
const listItems = ref<PaperUserLiteratureItem[]>([])
const selectedIds = ref<Set<string>>(new Set())

const categoryOptions = ref<{ value: string; label: string }[]>([{ value: '', label: '全部' }])
const manuscriptOptions = ref<{ value: string; label: string }[]>([{ value: '', label: '全部' }])

function cloneSet(from: Set<string>): Set<string> {
  return new Set(from)
}

function isSelected(id: string): boolean {
  return selectedIds.value.has(id)
}

function toggleId(id: string) {
  const next = cloneSet(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

const allVisibleSelected = computed(() => {
  const rows = listItems.value
  if (!rows.length) return false
  return rows.every((r) => selectedIds.value.has(r.id))
})

const someVisibleSelected = computed(() => {
  const rows = listItems.value
  if (!rows.length) return false
  const n = rows.filter((r) => selectedIds.value.has(r.id)).length
  return n > 0 && n < rows.length
})

function toggleSelectAllVisible() {
  const rows = listItems.value
  if (!rows.length) return
  const next = cloneSet(selectedIds.value)
  if (allVisibleSelected.value) {
    for (const r of rows) next.delete(r.id)
  } else {
    for (const r of rows) next.add(r.id)
  }
  selectedIds.value = next
}

async function loadFilterOptions() {
  try {
    const [catData, msData] = await Promise.all([
      fetchUserLiteratureCategories({ page: 1, page_size: 200 }),
      fetchPaperManuscripts({ page: 1, page_size: 200 }),
    ])
    categoryOptions.value = [
      { value: '', label: '全部' },
      ...(catData.items ?? []).map((c) => ({ value: c.id, label: c.name })),
    ]
    manuscriptOptions.value = [
      { value: '', label: '全部' },
      ...(msData.items ?? []).map((m) => ({
        value: m.id,
        label: m.is_current ? `${m.title}（当前）` : m.title,
      })),
    ]
  } catch {
    categoryOptions.value = [{ value: '', label: '全部' }]
    manuscriptOptions.value = [{ value: '', label: '全部' }]
  }
}

async function loadList() {
  loading.value = true
  try {
    const data = await fetchUserLiterature({
      page: 1,
      page_size: 200,
      category_id: filterCategoryId.value || undefined,
      manuscript_id: filterManuscriptId.value || undefined,
    })
    listItems.value = data.items ?? []
  } catch {
    listItems.value = []
  } finally {
    loading.value = false
  }
}

async function onFiltersChange() {
  await loadList()
}

function onClose() {
  if (submitting.value) return
  emit('close')
}

function onConfirm() {
  submitting.value = true
  try {
    emit('confirm', Array.from(selectedIds.value))
    emit('close')
  } finally {
    submitting.value = false
  }
}

watch(
  () => props.open,
  async (visible) => {
    if (!visible) return
    filterCategoryId.value = ''
    filterManuscriptId.value = ''
    selectedIds.value = new Set(props.initialSelectedIds ?? [])
    submitting.value = false
    await loadFilterOptions()
    await loadList()
  },
)

watch([filterCategoryId, filterManuscriptId], () => {
  if (!props.open) return
  void onFiltersChange()
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="paper-modal-overlay wf-lit-generate-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="topic-ul-pick-title"
      @click.self="onClose"
      @keydown.escape="onClose"
    >
      <div
        class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel topic-ul-pick-panel"
        @click.stop
      >
        <header class="paper-modal-header">
          <h2 id="topic-ul-pick-title" class="paper-modal-title">选择我的文献</h2>
          <button
            type="button"
            class="paper-modal-close"
            aria-label="关闭"
            :disabled="submitting || disabled"
            @click="onClose"
          >
            ×
          </button>
        </header>
        <div class="paper-modal-body topic-ul-pick-body">
          <p class="topic-ul-pick-lead">
            可按名称、按论文筛选列表，并单独勾选要纳入本次选题发现的文献。
          </p>
          <div class="topic-ul-pick-filters">
            <label class="topic-ul-pick-filter">
              <span class="topic-ul-pick-filter-label">按名称</span>
              <PaperSelect
                v-model="filterCategoryId"
                :options="categoryOptions"
                :disabled="loading || submitting || disabled"
              />
            </label>
            <label class="topic-ul-pick-filter">
              <span class="topic-ul-pick-filter-label">按论文</span>
              <PaperSelect
                v-model="filterManuscriptId"
                :options="manuscriptOptions"
                :disabled="loading || submitting || disabled"
              />
            </label>
          </div>
          <p v-if="loading" class="topic-ul-pick-meta">正在加载文献…</p>
          <p v-else-if="!listItems.length" class="topic-ul-pick-empty">当前筛选下暂无文献</p>
          <div v-else class="topic-ul-pick-list-wrap">
            <label class="topic-ul-pick-select-all">
              <input
                type="checkbox"
                :checked="allVisibleSelected"
                :indeterminate="someVisibleSelected"
                :disabled="submitting || disabled"
                @change="toggleSelectAllVisible"
              />
              <span>全选当前列表（共 {{ listItems.length }} 篇）</span>
            </label>
            <ul class="topic-ul-pick-list" role="listbox" aria-multiselectable="true">
              <li v-for="lit in listItems" :key="lit.id" class="topic-ul-pick-item">
                <label class="topic-ul-pick-item-label">
                  <input
                    type="checkbox"
                    :checked="isSelected(lit.id)"
                    :disabled="submitting || disabled"
                    @change="toggleId(lit.id)"
                  />
                  <span class="topic-ul-pick-item-title" :title="lit.literature_title">
                    {{ lit.literature_title }}
                  </span>
                </label>
              </li>
            </ul>
          </div>
          <p class="topic-ul-pick-count">已选 {{ selectedIds.size }} 篇（含未在当前列表中的已选项）</p>
        </div>
        <footer class="paper-message-box__btns">
          <button
            type="button"
            class="paper-btn-primary wf-lit-delete-dialog-btn"
            :disabled="submitting || disabled"
            @click="onConfirm"
          >
            确定
          </button>
          <button
            type="button"
            class="paper-btn-primary wf-lit-delete-dialog-btn"
            :disabled="submitting"
            @click="onClose"
          >
            取消
          </button>
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.topic-ul-pick-panel {
  width: min(680px, calc(100vw - 32px));
  max-width: 680px;
}

.topic-ul-pick-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.topic-ul-pick-lead {
  margin: 0;
  font-size: 13px;
  line-height: 1.55;
  color: #64748b;
}

.topic-ul-pick-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 16px;
}

.topic-ul-pick-filter {
  display: flex;
  flex: 1 1 200px;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.topic-ul-pick-filter-label {
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
}

.topic-ul-pick-filter :deep(.paper-select) {
  width: 100%;
}

.topic-ul-pick-filter :deep(.paper-select-trigger) {
  min-height: 38px;
  font-size: 13px;
}

.topic-ul-pick-meta {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

.topic-ul-pick-empty {
  margin: 0;
  padding: 24px 12px;
  font-size: 14px;
  color: #94a3b8;
  text-align: center;
  background: #f8fafc;
  border-radius: 8px;
}

.topic-ul-pick-list-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.topic-ul-pick-select-all {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 13px;
  color: #475569;
}

.topic-ul-pick-list {
  max-height: min(360px, 48vh);
  padding: 0;
  margin: 0;
  overflow: auto;
  list-style: none;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.topic-ul-pick-item {
  border-bottom: 1px solid #f1f5f9;
}

.topic-ul-pick-item:last-child {
  border-bottom: none;
}

.topic-ul-pick-item-label {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 10px 12px;
  cursor: pointer;
}

.topic-ul-pick-item-label:hover {
  background: #f8fafc;
}

.topic-ul-pick-item-title {
  flex: 1;
  font-size: 14px;
  line-height: 1.45;
  color: #1e1b4b;
}

.topic-ul-pick-count {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}
</style>
