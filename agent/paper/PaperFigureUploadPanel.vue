<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  FIGURE_UPLOAD_ACCEPT,
  FIGURE_UPLOAD_STORAGE_KEY,
  type UploadedFigureItem,
  createUploadedFigureFromFile,
} from './types'

const { manuscriptId, manuscriptTitle, embedded } = withDefaults(
  defineProps<{
    manuscriptId: string
    manuscriptTitle: string
    /** 嵌入图表管理卡片：无外层边框 */
    embedded?: boolean
  }>(),
  { embedded: false },
)

const fileInputRef = ref<HTMLInputElement | null>(null)
const itemsByManuscript = ref<Record<string, UploadedFigureItem[]>>({})
const searchQuery = ref('')

function loadFromStorage() {
  try {
    const raw = localStorage.getItem(FIGURE_UPLOAD_STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as { byManuscript?: Record<string, UploadedFigureItem[]> }
    if (data.byManuscript) itemsByManuscript.value = data.byManuscript
  } catch {
    /* ignore */
  }
}

function persist() {
  localStorage.setItem(
    FIGURE_UPLOAD_STORAGE_KEY,
    JSON.stringify({ byManuscript: itemsByManuscript.value }),
  )
}

loadFromStorage()

const items = computed({
  get() {
    if (!manuscriptId) return []
    return itemsByManuscript.value[manuscriptId] ?? []
  },
  set(list: UploadedFigureItem[]) {
    if (!manuscriptId) return
    itemsByManuscript.value = { ...itemsByManuscript.value, [manuscriptId]: list }
    persist()
  },
})

const filteredItems = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return items.value
  return items.value.filter(
    (it) =>
      it.title.toLowerCase().includes(q) ||
      it.fileName.toLowerCase().includes(q) ||
      (it.note ?? '').toLowerCase().includes(q),
  )
})

watch(
  () => manuscriptId,
  () => {
    searchQuery.value = ''
  },
)

function openFilePicker() {
  fileInputRef.value?.click()
}

function onFilesSelected(ev: Event) {
  const input = ev.target as HTMLInputElement
  const files = input.files
  if (!files?.length || !manuscriptId) return
  const added: UploadedFigureItem[] = []
  for (const file of Array.from(files)) {
    added.push(createUploadedFigureFromFile(file))
  }
  items.value = [...items.value, ...added]
  input.value = ''
  ElMessage.success(`已添加 ${added.length} 个图表（演示：仅存元数据，接入后上传至服务端）`)
}

async function removeItem(id: string) {
  const target = items.value.find((it) => it.id === id)
  if (!target) return
  try {
    await ElMessageBox.confirm(`确定移除「${target.title}」？`, '删除图表', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
    items.value = items.value.filter((it) => it.id !== id)
    ElMessage.success('已移除')
  } catch {
    /* cancelled */
  }
}

function updateTitle(id: string, title: string) {
  items.value = items.value.map((it) => (it.id === id ? { ...it, title } : it))
}

function updateNote(id: string, note: string) {
  items.value = items.value.map((it) => (it.id === id ? { ...it, note } : it))
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString('zh-CN', { hour12: false })
  } catch {
    return iso
  }
}

function kindLabel(kind: UploadedFigureItem['kind']) {
  if (kind === 'image') return 'IMG'
  if (kind === 'pdf') return 'PDF'
  if (kind === 'vector') return 'VEC'
  return 'FILE'
}
</script>

<template>
  <section class="fig-upload-panel" :class="{ 'fig-upload-panel--embedded': embedded }">
    <p class="fig-upload-lead">
      为当前论文「<strong>{{ manuscriptTitle }}</strong>」上传 figure / 表格截图等文件，供写作与论文审查引用（演示：本地仅存文件名与元数据）。
    </p>

    <div class="fig-upload-toolbar">
      <input
        ref="fileInputRef"
        type="file"
        class="fig-upload-file-input"
        multiple
        :accept="FIGURE_UPLOAD_ACCEPT"
        @change="onFilesSelected"
      />
      <button type="button" class="fig-upload-btn-primary" @click="openFilePicker">上传图表</button>
      <input
        v-model="searchQuery"
        type="search"
        class="fig-upload-search"
        placeholder="搜索标题、文件名、备注…"
      />
      <span class="fig-upload-count">共 {{ items.length }} 个</span>
    </div>

    <div v-if="filteredItems.length === 0" class="fig-upload-empty">
      <template v-if="items.length === 0">尚未上传图表，点击「上传图表」添加 PNG / PDF / SVG 等。</template>
      <template v-else>没有匹配的图表，请调整搜索词。</template>
    </div>

    <ul v-else class="fig-upload-list">
      <li v-for="item in filteredItems" :key="item.id" class="fig-upload-row">
        <div class="fig-upload-row-main">
          <span class="fig-upload-kind" :class="`fig-upload-kind--${item.kind}`">{{
            kindLabel(item.kind)
          }}</span>
          <label class="fig-upload-field">
            <span class="fig-upload-field-label">标题</span>
            <input
              :value="item.title"
              type="text"
              class="fig-upload-input"
              @change="updateTitle(item.id, ($event.target as HTMLInputElement).value)"
            />
          </label>
          <p class="fig-upload-meta">
            <span>{{ item.fileName }}</span>
            <span>{{ formatSize(item.sizeBytes) }}</span>
            <span>{{ formatDate(item.uploadedAt) }}</span>
          </p>
          <label class="fig-upload-field fig-upload-field--note">
            <span class="fig-upload-field-label">备注</span>
            <input
              :value="item.note ?? ''"
              type="text"
              class="fig-upload-input fig-upload-input--muted"
              placeholder="可选：Figure 编号、对应实验…"
              @change="updateNote(item.id, ($event.target as HTMLInputElement).value)"
            />
          </label>
        </div>
        <button type="button" class="fig-upload-btn-danger" @click="removeItem(item.id)">删除</button>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.fig-upload-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.fig-upload-panel--embedded {
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 0;
  box-shadow: none;
}

.fig-upload-lead {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.65;
  color: #475569;
}

.fig-upload-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  margin-bottom: 20px;
}

.fig-upload-file-input {
  display: none;
}

.fig-upload-btn-primary {
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  border: none;
  border-radius: 10px;
}

.fig-upload-search {
  flex: 1;
  min-width: 180px;
  padding: 10px 14px;
  font-size: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.fig-upload-count {
  font-size: 13px;
  color: #64748b;
}

.fig-upload-empty {
  padding: 32px 16px;
  font-size: 14px;
  color: #64748b;
  text-align: center;
  background: #f8fafc;
  border-radius: 12px;
}

.fig-upload-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 0;
  margin: 0;
  list-style: none;
}

.fig-upload-row {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  padding: 16px 18px;
  background: #fafbff;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
}

.fig-upload-row-main {
  flex: 1;
  min-width: 0;
}

.fig-upload-kind {
  display: inline-block;
  padding: 2px 8px;
  margin-bottom: 10px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  border-radius: 6px;
}

.fig-upload-kind--image {
  color: #0369a1;
  background: #e0f2fe;
}

.fig-upload-kind--pdf {
  color: #b45309;
  background: #fef3c7;
}

.fig-upload-kind--vector {
  color: #6d28d9;
  background: #ede9fe;
}

.fig-upload-kind--other {
  color: #475569;
  background: #f1f5f9;
}

.fig-upload-field {
  display: block;
  margin-bottom: 8px;
}

.fig-upload-field-label {
  display: block;
  margin-bottom: 4px;
  font-size: 12px;
  color: #64748b;
}

.fig-upload-input {
  width: 100%;
  padding: 8px 12px;
  font-size: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.fig-upload-input--muted {
  font-size: 13px;
}

.fig-upload-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin: 0 0 8px;
  font-size: 12px;
  color: #94a3b8;
}

.fig-upload-btn-danger {
  flex-shrink: 0;
  padding: 8px 14px;
  font-size: 13px;
  color: #dc2626;
  cursor: pointer;
  background: #fff;
  border: 1px solid #fecaca;
  border-radius: 8px;
}

.fig-upload-btn-danger:hover {
  background: #fef2f2;
}
</style>
