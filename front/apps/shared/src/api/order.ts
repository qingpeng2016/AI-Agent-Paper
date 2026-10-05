import type { ApiEnvelope } from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

export type OrderType = 'purchase' | 'renewal' | 'upgrade' | 'quota_addon'

export type CreateOrderBody = {
  product_id: number
  /** 默认 purchase；会员中心续费传 renewal */
  order_type?: OrderType
  /** 续费/升档/加购额度必填，新购传 0 或不传 */
  user_subscription_id?: number
  quantity: number
  channel: 'alipay' | 'wechat' | 'paypal'
  enterprise_invoice?: boolean
  /** 用户优惠券 user_coupons.id，0 或不传表示不使用 */
  user_coupon_id?: number
}

export type UserOrderListPage = {
  items: UserOrderItem[]
  total: number
  page: number
  page_size: number
}

export type UserOrderItem = {
  id: number
  order_no: string
  order_type: OrderType
  product_name: string
  quantity: number
  total_amount: number | string
  status: 'pending_payment' | 'completed' | 'failed' | 'cancelled'
  enterprise_invoice: boolean
  out_trade_no: string
  created_at: string
}

export type CreateOrderResult = {
  order_no: string
  out_trade_no: string
  order_id: number
  order_type: OrderType
  user_subscription_id: number
  channel: string
  status: string
  total_amount: number | string
  currency: string
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createOrderApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async list(params?: { page?: number; page_size?: number }): Promise<UserOrderListPage> {
      const search = new URLSearchParams()
      if (params?.page != null) search.set('page', String(params.page))
      if (params?.page_size != null) search.set('page_size', String(params.page_size))
      const qs = search.toString()
      const path = qs ? `/api/v1/mock/orders?${qs}` : '/api/v1/mock/orders'
      const res = await http.get<ApiEnvelope<UserOrderListPage>>(path)
      return unwrap(res)
    },
    async create(body: CreateOrderBody): Promise<CreateOrderResult> {
      const res = await http.post<ApiEnvelope<CreateOrderResult>>('/api/v1/mock/orders', body)
      return unwrap(res)
    },
    /** 创建订单并模拟支付成功（推荐；避免只创建 pending 订单） */
    async checkout(body: CreateOrderBody): Promise<CreateOrderResult> {
      const res = await http.post<ApiEnvelope<CreateOrderResult>>(
        '/api/v1/mock/orders/checkout',
        body,
      )
      return unwrap(res)
    },
    async mockNotify(channel: string, outTradeNo: string): Promise<void> {
      await http.post<ApiEnvelope<unknown>>(`/api/v1/mock/payments/notify/${channel}`, {
        out_trade_no: outTradeNo,
        trade_status: 'TRADE_SUCCESS',
      })
    },
  }
}
