<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { PaperManuscriptItem } from './types'

const props = defineProps<{
  manuscripts: PaperManuscriptItem[]
  activeManuscriptId: string
}>()

const emit = defineEmits<{
  select: [id: string]
  updateManuscripts: [list: PaperManuscriptItem[]]
}>()

const tab = ref<'active' | 'archived'>('active')
const searchQuery = ref('')

const filteredList = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return props.manuscripts.filter((m) => {
    if (m.status !== tab.value) return false
    if (!q) return true
    const hay = `${m.title} ${m.venueHint}`.toLowerCase()
    return hay.includes(q)
  })
})

function setList(next: PaperManuscriptItem[]) {
  emit('updateManuscripts', next)
}

async function editManuscript(item: PaperManuscriptItem) {
  try {
    const { value } = await ElMessageBox.prompt('标题与目标 venue 可在下方分别编辑。', '编辑论文', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValue: item.title,
      inputPlaceholder: '论文标题',
    })
    const title = value?.trim()
    if (!title) return
    const { value: venue } = await ElMessageBox.prompt('可选', '目标会议/期刊', {
      confirmButtonText: '保存',
      cancelButtonText: '跳过',
      inputValue: item.venueHint,
      inputPlaceholder: '如 NeurIPS 2026',
    })
    const next = props.manuscripts.map((m) =>
      m.id === item.id
        ? { ...m, title, venueHint: venue?.trim() ?? m.venueHint }
        : m,
    )
    setList(next)
    ElMessage.success('已保存（演示）')
  } catch {
    /* cancelled */
  }
}

async function archiveManuscript(item: PaperManuscriptItem) {
  try {
    await ElMessageBox.confirm(`归档后「${item.title}」不会出现在侧栏当前论文下拉中，可随时恢复。`, '归档论文', {
      type: 'warning',
      confirmButtonText: '归档',
      cancelButtonText: '取消',
    })
    const next = props.manuscripts.map((m) =>
      m.id === item.id ? { ...m, status: 'archived' as const } : m,
    )
    setList(next)
    if (props.activeManuscriptId === item.id) {
      const fallback = next.find((m) => m.status === 'active')
      if (fallback) emit('select', fallback.id)
    }
    ElMessage.success('已归档')
  } catch {
    /* cancelled */
  }
}

function restoreManuscript(item: PaperManuscriptItem) {
  const next = props.manuscripts.map((m) =>
    m.id === item.id ? { ...m, status: 'active' as const } : m,
  )
  setList(next)
  ElMessage.success('已恢复为进行中')
}

function selectAsCurrent(id: string) {
  emit('select', id)
  ElMessage.success('已切换当前论文')
}
</script>

<template>
  <section class="ms-panel">
    <p class="ms-lead">
      管理本篇工作台下的论文项目（对应 <code>paper_manuscript</code>）。切换「当前论文」后，各模块运行与产出按篇隔离。
    </p>

    <div class="ms-toolbar">
      <input v-model="searchQuery" type="search" class="ms-search" placeholder="搜索标题、venue…" />
    </div>

    <div class="ms-tabs" role="tablist">
      <button
        type="button"
        class="ms-tab"
        :class="{ 'ms-tab--active': tab === 'active' }"
        @click="tab = 'active'"
      >
        进行中
      </button>
      <button
        type="button"
        class="ms-tab"
        :class="{ 'ms-tab--active': tab === 'archived' }"
        @click="tab = 'archived'"
      >
        已归档
      </button>
    </div>

    <div v-if="filteredList.length === 0" class="ms-empty">
      {{ tab === 'active' ? '暂无进行中的论文。' : '暂无归档论文。' }}
    </div>

    <ul v-else class="ms-list">
      <li v-for="item in filteredList" :key="item.id" class="ms-row">
        <div class="ms-row-main">
          <div class="ms-row-head">
            <strong class="ms-title">{{ item.title }}</strong>
            <span
              v-if="item.id === activeManuscriptId && item.status === 'active'"
              class="ms-badge ms-badge--current"
              >当前</span
            >
          </div>
          <p v-if="item.venueHint" class="ms-venue">{{ item.venueHint }}</p>
          <p class="ms-id">ID · {{ item.id }}</p>
        </div>
        <div class="ms-row-actions">
          <button
            v-if="item.status === 'active' && item.id !== activeManuscriptId"
            type="button"
            class="ms-btn-ghost"
            @click="selectAsCurrent(item.id)"
          >
            设为当前
          </button>
          <button type="button" class="ms-btn-ghost" @click="editManuscript(item)">编辑</button>
          <button
            v-if="item.status === 'active'"
            type="button"
            class="ms-btn-ghost ms-btn-ghost--warn"
            @click="archiveManuscript(item)"
          >
            归档
          </button>
          <button v-else type="button" class="ms-btn-ghost" @click="restoreManuscript(item)">恢复</button>
        </div>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.ms-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.ms-lead {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.55;
  color: #475569;
}

.ms-lead code {
  font-size: 12px;
  color: #6366f1;
}

.ms-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.ms-btn-primary {
  padding: 9px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #7c3aed, #6366f1);
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.ms-search {
  flex: 1;
  min-width: 200px;
  padding: 9px 12px;
  font-size: 14px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
}

.ms-tabs {
  display: inline-flex;
  gap: 4px;
  padding: 4px;
  margin-bottom: 16px;
  background: #f1f5f9;
  border-radius: 10px;
}

.ms-tab {
  padding: 7px 16px;
  font-size: 13px;
  font-weight: 500;
  color: #64748b;
  background: transparent;
  border: none;
  border-radius: 7px;
  cursor: pointer;
}

.ms-tab--active {
  color: #1e293b;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.06);
}

.ms-empty {
  padding: 32px 16px;
  text-align: center;
  font-size: 14px;
  color: #64748b;
  background: #f8fafc;
  border-radius: 12px;
}

.ms-list {
  margin: 0;
  padding: 0;
  list-style: none;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
  overflow: hidden;
}

.ms-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: flex-start;
  justify-content: space-between;
  padding: 16px 18px;
  border-bottom: 1px solid #f1f5f9;
}

.ms-row:last-child {
  border-bottom: none;
}

.ms-title {
  font-size: 15px;
  color: #0f172a;
}

.ms-row-head {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 4px;
}

.ms-badge {
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 6px;
}

.ms-badge--current {
  color: #4338ca;
  background: #eef2ff;
}

.ms-venue {
  margin: 0 0 4px;
  font-size: 13px;
  color: #64748b;
}

.ms-id {
  margin: 0;
  font-size: 11px;
  color: #94a3b8;
}

.ms-row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ms-btn-ghost {
  padding: 6px 12px;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  background: #fff;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  cursor: pointer;
}

.ms-btn-ghost:hover {
  border-color: #94a3b8;
}

.ms-btn-ghost--warn {
  color: #b45309;
  border-color: #fcd34d;
}
</style>
