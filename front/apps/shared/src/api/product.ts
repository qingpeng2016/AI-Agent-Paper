import type { ApiEnvelope } from '../types/user'
import type { ProductCatalogResponse } from '../types/product'
import type { ProductDetailResponse } from '../types/product-detail'
import type { NavMenuResponse } from '../types/nav'
import { createHttpClient, type HttpClientOptions } from './http'

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createProductApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async list(productsCategoryId?: number): Promise<ProductCatalogResponse> {
      const qs =
        productsCategoryId && productsCategoryId > 0
          ? `?products_category_id=${productsCategoryId}`
          : ''
      const res = await http.get<ApiEnvelope<ProductCatalogResponse>>(
        `/api/v1/products${qs}`,
      )
      const data = unwrap(res)
      return { categories: data.categories ?? [] }
    },
    async navMenu(): Promise<NavMenuResponse> {
      const res = await http.get<ApiEnvelope<NavMenuResponse>>('/api/v1/products/nav-menu')
      return unwrap(res) ?? { mega_menu: [], brand_menus: [] }
    },
    async detailBySlug(slug: string): Promise<ProductDetailResponse> {
      const res = await http.get<ApiEnvelope<ProductDetailResponse>>(
        `/api/v1/products/slug/${encodeURIComponent(slug)}`,
      )
      return unwrap(res)
    },
  }
}
