import type { ApiEnvelope } from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

export type UserSubscriptionItem = {
  id: number
  product_id: number
  product_name: string
  status: string
  limit_tokens: number
  used_tokens: number
  period_end: string
  expires_at: string
  sku_product_name: string
  products_category_name: string
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createSubscriptionApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async list(): Promise<UserSubscriptionItem[]> {
      const res = await http.get<ApiEnvelope<UserSubscriptionItem[]>>('/api/v1/mock/subscriptions')
      return unwrap(res)
    },
  }
}
