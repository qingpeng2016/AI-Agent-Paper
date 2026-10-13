<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchUserLiteratureCategories,
  PAPER_LIST_PAGE_SIZE,
  type PaperUserLiteratureCategoryItem,
} from '@/api/userLiterature'
import PaperListPagination from './PaperListPagination.vue'

const props = defineProps<{
  active: boolean
}>()

const items = ref<PaperUserLiteratureCategoryItem[]>([])
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
    const data = await fetchUserLiteratureCategories({
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

function onAddCategory() {
  ElMessage.info('新建分类功能即将接入')
}

function onEditCategory(item: PaperUserLiteratureCategoryItem) {
  ElMessage.info(`编辑分类「${item.name}」功能即将接入`)
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
  <div class="ul-cat-pane">
    <div class="ul-cat-toolbar">
      <p class="ul-cat-lead">管理文献分类；排序值越小，列表中越靠前。</p>
      <button type="button" class="ul-cat-btn-primary" @click="onAddCategory">新建分类</button>
    </div>
    <p v-if="loading" class="ul-meta">正在加载分类…</p>
    <p v-else-if="!total" class="ul-empty">暂无分类（status=active）。可先执行分类种子 SQL。</p>
    <template v-else>
      <table class="ul-table">
        <colgroup>
          <col class="ul-col-cat-name" />
          <col class="ul-col-cat-sort" />
          <col class="ul-col-cat-created" />
          <col class="ul-col-cat-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>分类名称</th>
            <th>排序</th>
            <th>创建时间</th>
            <th class="ul-col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td class="ul-col-title">{{ item.name }}</td>
            <td class="ul-col-cat-sort">{{ item.sort }}</td>
            <td class="ul-col-created">{{ formatCreatedAt(item.created_at) }}</td>
            <td class="ul-col-actions">
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
</template>

<style scoped>
.ul-cat-pane {
  padding-top: 4px;
}

.ul-cat-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.ul-cat-lead {
  flex: 1;
  min-width: 200px;
  margin: 0;
  font-size: 13px;
  line-height: 1.55;
  color: #64748b;
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
  width: 44%;
}

.ul-table col.ul-col-cat-sort {
  width: 12%;
}

.ul-table col.ul-col-cat-created {
  width: 26%;
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

.ul-col-cat-sort {
  font-variant-numeric: tabular-nums;
  color: #475569;
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
</style>
