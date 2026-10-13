<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  total: number
  pageSize: number
  currentPage: number
}>()

const emit = defineEmits<{
  'update:currentPage': [page: number]
}>()

const totalPages = computed(() =>
  props.pageSize > 0 ? Math.max(1, Math.ceil(props.total / props.pageSize)) : 1,
)

const showPager = computed(() => props.total > props.pageSize)

function onPageChange(page: number) {
  emit('update:currentPage', page)
}
</script>

<template>
  <nav v-if="total > 0" class="paper-list-pagination" aria-label="列表分页">
    <p class="paper-list-pagination-meta">
      <span class="paper-list-pagination-total">
        共 <strong>{{ total }}</strong> 条
      </span>
      <span v-if="showPager" class="paper-list-pagination-sep">·</span>
      <span v-if="showPager" class="paper-list-pagination-page">
        第 {{ currentPage }} / {{ totalPages }} 页
      </span>
    </p>
    <el-pagination
      v-if="showPager"
      class="paper-list-pagination-control"
      :current-page="currentPage"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      background
      :hide-on-single-page="false"
      @current-change="onPageChange"
    />
  </nav>
</template>

<style scoped>
.paper-list-pagination {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 16px;
  align-items: center;
  justify-content: flex-end;
  margin-top: 16px;
  padding: 0;
}

.paper-list-pagination-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 10px;
  align-items: center;
  margin: 0;
  font-size: 13px;
  line-height: 1.4;
  color: #64748b;
}

.paper-list-pagination-total strong {
  font-weight: 600;
  color: #334155;
  font-variant-numeric: tabular-nums;
}

.paper-list-pagination-sep {
  color: #cbd5e1;
}

.paper-list-pagination-page {
  font-variant-numeric: tabular-nums;
  color: #475569;
}

.paper-list-pagination-control {
  margin: 0;
}

.paper-list-pagination-control :deep(.btn-prev),
.paper-list-pagination-control :deep(.btn-next),
.paper-list-pagination-control :deep(.el-pager li) {
  min-width: 36px;
  height: 36px;
  font-size: 13px;
  font-weight: 500;
  line-height: 36px;
  border-radius: 10px;
}

.paper-list-pagination-control :deep(.el-pager li) {
  color: #475569;
  background: #fff;
  border: 1px solid #e2e8f0;
}

.paper-list-pagination-control :deep(.el-pager li:hover) {
  color: #4338ca;
  border-color: #c7d2fe;
}

.paper-list-pagination-control :deep(.el-pager li.is-active) {
  color: #fff;
  background: linear-gradient(135deg, #7c3aed 0%, #6366f1 100%);
  border-color: transparent;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.35);
}

.paper-list-pagination-control :deep(.btn-prev),
.paper-list-pagination-control :deep(.btn-next) {
  color: #475569;
  background: #fff;
  border: 1px solid #e2e8f0;
}

.paper-list-pagination-control :deep(.btn-prev:hover:not(:disabled)),
.paper-list-pagination-control :deep(.btn-next:hover:not(:disabled)) {
  color: #4338ca;
  border-color: #c7d2fe;
}

.paper-list-pagination-control :deep(.btn-prev:disabled),
.paper-list-pagination-control :deep(.btn-next:disabled) {
  opacity: 0.45;
}

@media (max-width: 560px) {
  .paper-list-pagination {
    justify-content: flex-end;
  }
}
</style>
