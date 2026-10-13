<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchUserLiterature,
  formatFileSizeBytes,
  formatUserLiteratureFrom,
  PAPER_LIST_PAGE_SIZE,
  userLiteratureFromTone,
  type PaperUserLiteratureItem,
} from '@/api/userLiterature'
import PaperListPagination from './PaperListPagination.vue'

const items = ref<PaperUserLiteratureItem[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const detailOpen = ref(false)
const detailItem = ref<PaperUserLiteratureItem | null>(null)

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

function openDetail(item: PaperUserLiteratureItem) {
  detailItem.value = item
  detailOpen.value = true
}

function closeDetail() {
  detailOpen.value = false
  detailItem.value = null
}

void loadPage(1)
</script>

<template>
  <section class="ul-panel">
    <p v-if="loading" class="ul-meta">正在加载文献列表…</p>
    <p v-else-if="!total" class="ul-empty">暂无文献（status=active）。</p>
    <template v-else>
      <table class="ul-table">
        <colgroup>
          <col class="ul-col-title" />
          <col class="ul-col-created" />
          <col class="ul-col-from" />
          <col class="ul-col-ext" />
          <col class="ul-col-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>标题</th>
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
          <dt v-if="detailItem.storage_url">存储路径</dt>
          <dd v-if="detailItem.storage_url" class="ul-mono">{{ detailItem.storage_url }}</dd>
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
  width: 40%;
}

.ul-table col.ul-col-created {
  width: 18%;
}

.ul-table col.ul-col-from,
.ul-table col.ul-col-ext {
  width: 12%;
}

.ul-table col.ul-col-actions {
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
</style>
