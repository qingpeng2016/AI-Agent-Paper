import type { ApiEnvelope } from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

export type UserInvoiceItem = {
  id: number
  order_no: string
  title: string
  amount: number | string
  status: 'pending' | 'issued' | 'failed' | string
  created_at: string
}

export type UserInvoiceListPage = {
  items: UserInvoiceItem[]
  total: number
  page: number
  page_size: number
}

export type InvoiceConfigItem = {
  id: number
  profile_type: 'enterprise' | 'personal' | string
  title: string
  tax_no?: string
  bank_name?: string
  bank_account?: string
  address?: string
  phone?: string
  is_default: boolean
  created_at: string
  updated_at: string
}

export type EnterpriseInvoiceLookupItem = {
  title: string
  tax_no: string
  bank_name?: string
  bank_account?: string
  address?: string
  phone?: string
}

export type EnterpriseInvoiceLookupResult = {
  match: EnterpriseInvoiceLookupItem
  candidates?: EnterpriseInvoiceLookupItem[]
}

export type SaveInvoiceConfigBody = {
  profile_type?: 'enterprise' | 'personal'
  title: string
  tax_no?: string
  bank_name?: string
  bank_account?: string
  address?: string
  phone?: string
  is_default?: boolean
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createInvoiceApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async list(params?: { page?: number; page_size?: number }): Promise<UserInvoiceListPage> {
      const search = new URLSearchParams()
      if (params?.page != null) search.set('page', String(params.page))
      if (params?.page_size != null) search.set('page_size', String(params.page_size))
      const qs = search.toString()
      const path = qs ? `/api/v1/users/invoices?${qs}` : '/api/v1/users/invoices'
      const res = await http.get<ApiEnvelope<UserInvoiceListPage>>(path)
      return unwrap(res)
    },
    async listConfigs(): Promise<InvoiceConfigItem[]> {
      const res = await http.get<ApiEnvelope<InvoiceConfigItem[]>>('/api/v1/users/invoice-configs')
      return unwrap(res)
    },
    async lookupEnterprise(keyword: string): Promise<EnterpriseInvoiceLookupResult> {
      const q = encodeURIComponent(keyword.trim())
      const res = await http.get<ApiEnvelope<EnterpriseInvoiceLookupResult>>(
        `/api/v1/users/invoice-configs/enterprise-lookup?keyword=${q}`,
      )
      return unwrap(res)
    },
    async createConfig(body: SaveInvoiceConfigBody): Promise<InvoiceConfigItem> {
      const res = await http.post<ApiEnvelope<InvoiceConfigItem>>(
        '/api/v1/users/invoice-configs',
        body,
      )
      return unwrap(res)
    },
    async updateConfig(id: number, body: SaveInvoiceConfigBody): Promise<InvoiceConfigItem> {
      const res = await http.put<ApiEnvelope<InvoiceConfigItem>>(
        `/api/v1/users/invoice-configs/${id}`,
        body,
      )
      return unwrap(res)
    },
    async setDefaultConfig(id: number, body: SaveInvoiceConfigBody): Promise<InvoiceConfigItem> {
      try {
        await http.post<ApiEnvelope<null>>(
          `/api/v1/users/invoice-configs/${id}/set-default`,
          {},
        )
        const list = await this.listConfigs()
        const hit = list.find((i) => i.id === id)
        if (hit) return hit
      } catch (err) {
        const msg = err instanceof Error ? err.message : ''
        const useFallback =
          /404|not found|Invalid JSON/i.test(msg) || msg.includes('无法')
        if (!useFallback) throw err
      }
      return this.updateConfig(id, { ...body, is_default: true })
    },
  }
}
