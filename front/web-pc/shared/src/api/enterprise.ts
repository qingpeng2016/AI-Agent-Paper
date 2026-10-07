import type { ApiEnvelope } from '../types/user'
import type {
  EnterpriseProduct,
  EnterpriseProductListResponse,
} from '../types/enterprise'
import { createHttpClient, type HttpClientOptions } from './http'

export type { EnterpriseProduct, EnterpriseProductListResponse }

export type SubmitEnterpriseInquiryRequest = {
  company_name: string
  contact_name: string
  phone: string
  email?: string
}

export type SubmitEnterpriseInquiryResponse = {
  id: number
  message?: string
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createEnterpriseApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async getProduct(code: string): Promise<EnterpriseProduct> {
      const res = await http.get<ApiEnvelope<EnterpriseProduct>>(
        `/api/v1/enterprise/products/${encodeURIComponent(code)}`,
      )
      return unwrap(res)
    },
    async listProducts(): Promise<EnterpriseProduct[]> {
      const res = await http.get<ApiEnvelope<EnterpriseProductListResponse>>(
        '/api/v1/enterprise/products',
      )
      const data = unwrap(res)
      return data.products ?? []
    },
    async submitInquiry(
      body: SubmitEnterpriseInquiryRequest,
    ): Promise<SubmitEnterpriseInquiryResponse> {
      const res = await http.post<ApiEnvelope<SubmitEnterpriseInquiryResponse>>(
        '/api/v1/enterprise/inquiries',
        body,
      )
      return unwrap(res)
    },
  }
}
