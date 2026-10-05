import { computed, ref } from 'vue'
import type { UserNotificationItem } from '@ai-token-mall/shared'
import { notificationApi } from '@/api'

/** 顶栏消息弹窗最多展示条数 */
export const BELL_NOTIFICATION_LIMIT = 20

const messages = ref<UserNotificationItem[]>([])
const unreadCount = ref(0)
const loading = ref(false)

export function useSiteMessages() {
  const sortedMessages = computed(() => [...messages.value])

  const recentForBell = computed(() => sortedMessages.value.slice(0, BELL_NOTIFICATION_LIMIT))

  async function refreshList(opts?: { unreadOnly?: boolean; pageSize?: number }) {
    loading.value = true
    try {
      const data = await notificationApi.list({
        page: 1,
        page_size: opts?.pageSize ?? 50,
        unread_only: opts?.unreadOnly,
      })
      messages.value = data.items
    } finally {
      loading.value = false
    }
  }

  async function refreshUnreadCount() {
    unreadCount.value = await notificationApi.unreadCount()
  }

  async function refreshAll(opts?: { unreadOnly?: boolean; pageSize?: number }) {
    await Promise.all([refreshList(opts), refreshUnreadCount()])
  }

  function isUnread(m: UserNotificationItem): boolean {
    return !m.read
  }

  async function markRead(id: number) {
    await notificationApi.markRead(id)
    const row = messages.value.find((m) => m.id === id)
    if (row) {
      row.read = true
    }
    if (unreadCount.value > 0) {
      unreadCount.value -= 1
    } else {
      await refreshUnreadCount()
    }
  }

  async function markAllRead() {
    await notificationApi.markAllRead()
    for (const m of messages.value) {
      m.read = true
    }
    unreadCount.value = 0
  }

  return {
    messages,
    sortedMessages,
    recentForBell,
    unreadCount,
    loading,
    refreshList,
    refreshUnreadCount,
    refreshAll,
    markRead,
    markAllRead,
    isUnread,
  }
}
