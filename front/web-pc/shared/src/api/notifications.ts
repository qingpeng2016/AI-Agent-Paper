import type { ApiEnvelope } from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

export type UserNotificationItem = {
  id: number
  category: string
  title: string
  body: string
  template_code: string
  read: boolean
  created_at: string
  link_tab?: string
}

export type UserNotificationListPage = {
  items: UserNotificationItem[]
  total: number
  page: number
  page_size: number
}

export type UserNotificationUnreadCount = {
  count: number
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createNotificationApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async list(params?: {
      page?: number
      page_size?: number
      unread_only?: boolean
    }): Promise<UserNotificationListPage> {
      const search = new URLSearchParams()
      if (params?.page != null) search.set('page', String(params.page))
      if (params?.page_size != null) search.set('page_size', String(params.page_size))
      if (params?.unread_only) search.set('unread_only', 'true')
      const qs = search.toString()
      const path = qs ? `/api/v1/users/notifications?${qs}` : '/api/v1/users/notifications'
      const res = await http.get<ApiEnvelope<UserNotificationListPage>>(path)
      return unwrap(res)
    },
    async unreadCount(): Promise<number> {
      const res = await http.get<ApiEnvelope<UserNotificationUnreadCount>>(
        '/api/v1/users/notifications/unread-count',
      )
      return unwrap(res).count
    },
    async markRead(id: number): Promise<void> {
      await http.post<ApiEnvelope<null>>(`/api/v1/users/notifications/${id}/read`, {})
    },
    async markAllRead(): Promise<void> {
      await http.post<ApiEnvelope<null>>('/api/v1/users/notifications/read-all', {})
    },
  }
}
