<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  REFERENCE_UPLOAD_ACCEPT,
  REFERENCE_UPLOAD_STORAGE_KEY,
  type UploadedReferenceItem,
  createUploadedReferenceFromFile,
} from './types'

const { manuscriptId, manuscriptTitle } = defineProps<{
  manuscriptId: string
  manuscriptTitle: string
}>()

const fileInputRef = ref<HTMLInputElement | null>(null)
const itemsByManuscript = ref<Record<string, UploadedReferenceItem[]>>({})
const searchQuery = ref('')

function loadFromStorage() {
  try {
    const raw = localStorage.getItem(REFERENCE_UPLOAD_STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as { byManuscript?: Record<string, UploadedReferenceItem[]> }
    if (data.byManuscript) itemsByManuscript.value = data.byManuscript
  } catch {
    /* ignore */
  }
}

function persist() {
  localStorage.setItem(
    REFERENCE_UPLOAD_STORAGE_KEY,
    JSON.stringify({ byManuscript: itemsByManuscript.value }),
  )
}

loadFromStorage()

const items = computed({
  get() {
    if (!manuscriptId) return []
    return itemsByManuscript.value[manuscriptId] ?? []
  },
  set(list: UploadedReferenceItem[]) {
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
  const added: UploadedReferenceItem[] = []
  for (const file of Array.from(files)) {
    added.push(createUploadedReferenceFromFile(file))
  }
  items.value = [...items.value, ...added]
  input.value = ''
  ElMessage.success(`已添加 ${added.length} 篇（演示：仅存元数据，接入后上传至服务端）`)
}

async function removeItem(id: string) {
  const target = items.value.find((it) => it.id === id)
  if (!target) return
  try {
    await ElMessageBox.confirm(`确定移除「${target.title}」？`, '删除文献', {
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
</script>

<template>
  <section class="ref-panel">
    <p class="ref-lead">
      为当前论文「<strong>{{ manuscriptTitle }}</strong>」上传 PDF、BibTeX 等文件，供写作与引用门禁对照（演示：本地仅存文件名与元数据）。
    </p>

    <div class="ref-toolbar">
      <input
        ref="fileInputRef"
        type="file"
        class="ref-file-input"
        multiple
        :accept="REFERENCE_UPLOAD_ACCEPT"
        @change="onFilesSelected"
      />
      <button type="button" class="ref-btn-primary" @click="openFilePicker">上传文献</button>
      <input
        v-model="searchQuery"
        type="search"
        class="ref-search"
        placeholder="搜索标题、文件名、备注…"
      />
      <span class="ref-count">共 {{ items.length }} 篇</span>
    </div>

    <div v-if="filteredItems.length === 0" class="ref-empty">
      <template v-if="items.length === 0">尚未上传文献，点击「上传文献」添加 PDF / .bib 等。</template>
      <template v-else>没有匹配的文献，请调整搜索词。</template>
    </div>

    <ul v-else class="ref-list">
      <li v-for="item in filteredItems" :key="item.id" class="ref-row">
        <div class="ref-row-main">
          <span class="ref-kind" :class="`ref-kind--${item.kind}`">{{ item.kind.toUpperCase() }}</span>
          <label class="ref-field">
            <span class="ref-field-label">标题</span>
            <input
              :value="item.title"
              type="text"
              class="ref-input"
              @change="updateTitle(item.id, ($event.target as HTMLInputElement).value)"
            />
          </label>
          <p class="ref-meta">
            <span>{{ item.fileName }}</span>
            <span>{{ formatSize(item.sizeBytes) }}</span>
            <span>{{ formatDate(item.uploadedAt) }}</span>
          </p>
          <label class="ref-field ref-field--note">
            <span class="ref-field-label">备注</span>
            <input
              :value="item.note ?? ''"
              type="text"
              class="ref-input ref-input--muted"
              placeholder="可选：DOI、用途说明…"
              @change="updateNote(item.id, ($event.target as HTMLInputElement).value)"
            />
          </label>
        </div>
        <button type="button" class="ref-btn-danger" @click="removeItem(item.id)">删除</button>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.ref-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.ref-lead {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.55;
  color: #475569;
}

.ref-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  margin-bottom: 20px;
}

.ref-file-input {
  display: none;
}

.ref-btn-primary {
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: linear-gradient(135deg, #6366f1 0%, #7c3aed 100%);
  border: none;
  border-radius: 10px;
}

.ref-btn-primary:hover {
  filter: brightness(1.05);
}

.ref-search {
  flex: 1;
  min-width: 180px;
  padding: 10px 12px;
  font-size: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.ref-count {
  font-size: 13px;
  color: #64748b;
}

.ref-empty {
  padding: 32px 16px;
  font-size: 14px;
  color: #94a3b8;
  text-align: center;
  background: #f8fafc;
  border-radius: 12px;
}

.ref-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.ref-row {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  padding: 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.ref-row-main {
  flex: 1;
  min-width: 0;
}

.ref-kind {
  display: inline-block;
  margin-bottom: 8px;
  padding: 2px 8px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.06em;
  border-radius: 4px;
}

.ref-kind--pdf {
  color: #b91c1c;
  background: #fef2f2;
}

.ref-kind--bib {
  color: #1d4ed8;
  background: #eff6ff;
}

.ref-kind--other {
  color: #64748b;
  background: #f1f5f9;
}

.ref-field {
  display: block;
  margin-bottom: 8px;
}

.ref-field--note {
  margin-bottom: 0;
}

.ref-field-label {
  display: block;
  margin-bottom: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.ref-input {
  width: 100%;
  padding: 8px 10px;
  font-size: 14px;
  color: #1e1b4b;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.ref-input--muted {
  font-size: 13px;
}

.ref-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin: 0 0 10px;
  font-size: 12px;
  color: #64748b;
}

.ref-btn-danger {
  flex-shrink: 0;
  padding: 8px 12px;
  font-size: 13px;
  color: #b91c1c;
  cursor: pointer;
  background: #fff;
  border: 1px solid #fecaca;
  border-radius: 8px;
}

.ref-btn-danger:hover {
  background: #fef2f2;
}
</style>
