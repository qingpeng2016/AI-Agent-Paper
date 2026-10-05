<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { UserNotificationItem } from '@ai-token-mall/shared'
import type { MemberTab } from '@/mocks/member'
import { siteMessageCategoryLabel, type SiteMessageCategory } from '@/mocks/siteMessages'
import { useSiteMessages } from '@/composables/useSiteMessages'
import {
  beginRouteNavigationLoading,
  endRouteNavigationLoading,
} from '@/composables/useRouteNavigationLoading'

type FilterKey = 'all' | 'unread'

const readFilter = ref<FilterKey>('all')
const categoryFilter = ref<SiteMessageCategory | 'all'>('all')
const router = useRouter()

const { sortedMessages, unreadCount, markRead, markAllRead, isUnread, refreshAll, loading } =
  useSiteMessages()

const filtered = computed(() => {
  let rows = sortedMessages.value
  if (categoryFilter.value !== 'all') {
    rows = rows.filter((m) => m.category === categoryFilter.value)
  }
  return rows
})

const categoryOptions: { id: SiteMessageCategory | 'all'; label: string }[] = [
  { id: 'all', label: '全部分类' },
  { id: 'order', label: '订单' },
  { id: 'subscription', label: '套餐' },
  { id: 'billing', label: '发票/资金' },
  { id: 'system', label: '系统' },
]

function categoryLabel(cat: string): string {
  if (cat in siteMessageCategoryLabel) {
    return siteMessageCategoryLabel[cat as SiteMessageCategory]
  }
  return cat
}

async function load() {
  beginRouteNavigationLoading()
  try {
    await refreshAll({ unreadOnly: readFilter.value === 'unread', pageSize: 50 })
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载消息失败')
  } finally {
    endRouteNavigationLoading()
  }
}

async function onMarkAllRead() {
  beginRouteNavigationLoading()
  try {
    await markAllRead()
    ElMessage.success('已全部标为已读')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  } finally {
    endRouteNavigationLoading()
  }
}

async function onOpen(m: UserNotificationItem) {
  try {
    if (isUnread(m)) {
      await markRead(m.id)
    }
  } catch {
    /* ignore */
  }
  const tab = m.link_tab as MemberTab | undefined
  if (tab) {
    void router.push({ path: '/member', query: { tab } })
  }
}

watch(readFilter, () => {
  void load()
})

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="msg-panel">
    <div class="msg-toolbar">
      <div class="msg-tabs" role="tablist" aria-label="消息筛选">
        <button
          type="button"
          class="msg-tab"
          :class="{ 'msg-tab--active': readFilter === 'all' }"
          @click="readFilter = 'all'"
        >
          全部
        </button>
        <button
          type="button"
          class="msg-tab"
          :class="{ 'msg-tab--active': readFilter === 'unread' }"
          @click="readFilter = 'unread'"
        >
          未读
          <span v-if="unreadCount" class="msg-tab-badge">{{ unreadCount }}</span>
        </button>
      </div>
      <div class="msg-toolbar-right">
        <select v-model="categoryFilter" class="msg-select" aria-label="分类">
          <option v-for="opt in categoryOptions" :key="opt.id" :value="opt.id">
            {{ opt.label }}
          </option>
        </select>
        <button
          type="button"
          class="atm-btn-ghost btn-xs"
          :disabled="unreadCount === 0 || loading"
          @click="onMarkAllRead"
        >
          全部已读
        </button>
      </div>
    </div>

    <ul v-if="filtered.length" class="msg-list">
      <li v-for="m in filtered" :key="m.id">
        <article
          class="msg-card"
          :class="{ 'msg-card--unread': isUnread(m) }"
          role="button"
          tabindex="0"
          @click="onOpen(m)"
          @keydown.enter="onOpen(m)"
        >
          <div class="msg-card-head">
            <span v-if="isUnread(m)" class="msg-unread-dot" aria-label="未读" />
            <span class="msg-tag">{{ categoryLabel(m.category) }}</span>
            <time class="msg-time">{{ m.created_at }}</time>
          </div>
          <h3 class="msg-title">{{ m.title }}</h3>
          <p class="msg-body">{{ m.body }}</p>
          <p v-if="m.link_tab" class="msg-link-hint">点击查看相关页面 →</p>
        </article>
      </li>
    </ul>

    <div v-else-if="!loading" class="msg-empty">
      <p>{{ readFilter === 'unread' ? '暂无未读消息' : '暂无消息' }}</p>
    </div>
  </div>
</template>

<style scoped>
.msg-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.msg-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.msg-tabs {
  display: inline-flex;
  gap: 4px;
  padding: 4px;
  background: #f1f5f9;
  border-radius: 12px;
}

.msg-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  color: var(--atm-text-muted);
  background: transparent;
  border: none;
  border-radius: 9px;
  cursor: pointer;
}

.msg-tab--active {
  color: var(--atm-primary-dark, #6d28d9);
  background: #fff;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.12);
}

.msg-tab-badge {
  min-width: 18px;
  padding: 0 6px;
  font-size: 11px;
  line-height: 18px;
  color: #fff;
  text-align: center;
  background: var(--atm-primary, #7c3aed);
  border-radius: 999px;
}

.msg-toolbar-right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.msg-select {
  padding: 8px 12px;
  font-size: 13px;
  font-family: inherit;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.msg-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.msg-card {
  padding: 16px 18px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  cursor: pointer;
  transition:
    border-color 0.15s,
    box-shadow 0.15s;
}

.msg-card:hover {
  border-color: #c4b5fd;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.08);
}

.msg-card--unread {
  background: linear-gradient(135deg, #faf5ff 0%, #fff 55%);
  border-color: #ddd6fe;
}

.msg-card-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.msg-unread-dot {
  width: 8px;
  height: 8px;
  background: var(--atm-primary, #7c3aed);
  border-radius: 50%;
}

.msg-tag {
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 700;
  color: var(--atm-primary-dark, #6d28d9);
  background: #ede9fe;
  border-radius: 999px;
}

.msg-time {
  margin-left: auto;
  font-size: 12px;
  color: var(--atm-text-muted);
}

.msg-title {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 700;
  color: var(--atm-text);
}

.msg-body {
  margin: 0;
  font-size: 13px;
  line-height: 1.55;
  color: var(--atm-text-muted);
}

.msg-link-hint {
  margin: 10px 0 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary);
}

.msg-empty {
  padding: 48px 16px;
  text-align: center;
  font-size: 14px;
  color: var(--atm-text-muted);
  background: #f8fafc;
  border-radius: 14px;
}
</style>
