<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { UserNotificationItem } from '@ai-token-mall/shared'
import type { MemberTab } from '@/mocks/member'
import { siteMessageCategoryLabel, type SiteMessageCategory } from '@/mocks/siteMessages'
import { BELL_NOTIFICATION_LIMIT, useSiteMessages } from '@/composables/useSiteMessages'
import { isLoggedIn } from '@/composables/useSessionUser'

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const router = useRouter()
const route = useRoute()

const { recentForBell, unreadCount, markRead, markAllRead, isUnread, refreshAll, refreshUnreadCount } =
  useSiteMessages()

function toggle() {
  open.value = !open.value
}

function close() {
  open.value = false
}

function onDocClick(e: MouseEvent) {
  if (!open.value || !root.value) return
  if (!root.value.contains(e.target as Node)) close()
}

function categoryLabel(cat: string): string {
  if (cat in siteMessageCategoryLabel) {
    return siteMessageCategoryLabel[cat as SiteMessageCategory]
  }
  return cat
}

async function openMessage(m: UserNotificationItem) {
  if (isUnread(m)) {
    try {
      await markRead(m.id)
    } catch {
      /* ignore */
    }
  }
  close()
  const tab = m.link_tab as MemberTab | undefined
  if (tab) {
    void router.push({ path: '/member', query: { tab } })
  } else {
    void router.push({ path: '/member', query: { tab: 'messages' } })
  }
}

watch(open, (v) => {
  if (!v || !isLoggedIn()) return
  void (async () => {
    await refreshAll({ pageSize: BELL_NOTIFICATION_LIMIT })
    if (unreadCount.value > 0) {
      try {
        await markAllRead()
      } catch {
        /* ignore */
      }
    }
  })()
})

watch(
  () => route.fullPath,
  () => {
    if (isLoggedIn()) {
      void refreshUnreadCount()
    }
  },
)

onMounted(() => {
  document.addEventListener('click', onDocClick)
  if (isLoggedIn()) {
    void refreshUnreadCount()
  }
})

onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<template>
  <div ref="root" class="site-bell">
    <button
      type="button"
      class="site-bell-btn"
      aria-label="站内消息"
      :aria-expanded="open"
      @click.stop="toggle"
    >
      <svg class="site-bell-icon" viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M12 22a2.5 2.5 0 0 0 2.45-2h-4.9A2.5 2.5 0 0 0 12 22Zm7-6V11a7 7 0 0 0-5-6.71V4a2 2 0 1 0-4 0v.29A7 7 0 0 0 5 11v5l-2 2v1h18v-1l-2-2Z"
        />
      </svg>
      <span v-if="unreadCount > 0" class="site-bell-badge">{{
        unreadCount > 99 ? '99+' : unreadCount
      }}</span>
    </button>

    <div v-show="open" class="site-bell-panel" role="dialog" aria-label="最近20条消息">
      <header class="site-bell-head">
        <strong>站内消息</strong>
        <span v-if="unreadCount" class="site-bell-unread">{{ unreadCount }} 条未读</span>
      </header>
      <ul v-if="recentForBell.length" class="site-bell-list">
        <li v-for="m in recentForBell" :key="m.id">
          <button type="button" class="site-bell-item" @click="openMessage(m)">
            <span v-if="isUnread(m)" class="site-bell-dot" aria-hidden="true" />
            <span class="site-bell-item-main">
              <span class="site-bell-item-top">
                <span class="site-bell-tag">{{ categoryLabel(m.category) }}</span>
                <span class="site-bell-time">{{ m.created_at }}</span>
              </span>
              <span class="site-bell-title" :class="{ 'site-bell-title--unread': isUnread(m) }">{{
                m.title
              }}</span>
              <span class="site-bell-preview">{{ m.body }}</span>
            </span>
          </button>
        </li>
      </ul>
      <p v-else class="site-bell-empty">暂无消息</p>
    </div>
  </div>
</template>

<style scoped>
.site-bell {
  position: relative;
}

.site-bell-btn {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  padding: 0;
  color: var(--atm-text-muted);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  cursor: pointer;
  transition:
    color 0.15s,
    border-color 0.15s,
    background 0.15s;
}

.site-bell-btn:hover {
  color: var(--atm-primary);
  background: #f5f3ff;
  border-color: #ddd6fe;
}

.site-bell-icon {
  width: 20px;
  height: 20px;
}

.site-bell-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 18px;
  padding: 0 5px;
  font-size: 10px;
  font-weight: 700;
  line-height: 18px;
  color: #fff;
  text-align: center;
  background: #ef4444;
  border: 2px solid var(--atm-bg, #fff);
  border-radius: 999px;
}

.site-bell-panel {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  z-index: 220;
  display: flex;
  flex-direction: column;
  width: min(380px, calc(100vw - 24px));
  max-height: min(560px, 78vh);
  overflow: hidden;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow:
    0 4px 8px rgba(30, 27, 75, 0.06),
    0 24px 56px rgba(30, 27, 75, 0.16);
}

.site-bell-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 14px 16px 10px;
  font-size: 14px;
  color: var(--atm-text);
  border-bottom: 1px solid #f1f5f9;
}

.site-bell-unread {
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary);
}

.site-bell-list {
  flex: 1;
  margin: 0;
  padding: 6px 8px;
  overflow-y: auto;
  list-style: none;
}

.site-bell-item {
  display: flex;
  gap: 8px;
  width: 100%;
  padding: 10px 8px;
  text-align: left;
  background: transparent;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  font-family: inherit;
}

.site-bell-item:hover {
  background: #f8fafc;
}

.site-bell-dot {
  flex-shrink: 0;
  width: 8px;
  height: 8px;
  margin-top: 6px;
  background: var(--atm-primary, #7c3aed);
  border-radius: 50%;
}

.site-bell-item-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.site-bell-item-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.site-bell-tag {
  padding: 2px 8px;
  font-size: 10px;
  font-weight: 700;
  color: var(--atm-primary-dark, #6d28d9);
  background: #f5f3ff;
  border-radius: 999px;
}

.site-bell-time {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--atm-text-muted);
}

.site-bell-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-text);
}

.site-bell-title--unread {
  color: #1e1b4b;
}

.site-bell-preview {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-size: 12px;
  line-height: 1.45;
  color: var(--atm-text-muted);
}

.site-bell-empty {
  margin: 0;
  padding: 28px 16px;
  font-size: 13px;
  text-align: center;
  color: var(--atm-text-muted);
}

</style>
