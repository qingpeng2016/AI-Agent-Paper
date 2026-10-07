import type { ApiEnvelope } from '../types/user'
import type { TutorialArticle, TutorialListResponse } from '../types/tutorial'
import { createHttpClient, type HttpClientOptions } from './http'

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createTutorialApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async list(categoryCode?: string): Promise<TutorialListResponse> {
      const qs =
        categoryCode && categoryCode !== 'all'
          ? `?category_code=${encodeURIComponent(categoryCode)}`
          : ''
      const res = await http.get<ApiEnvelope<TutorialListResponse>>(
        `/api/v1/tutorials${qs}`,
      )
      const data = unwrap(res)
      return {
        title: data.title ?? '',
        categories: data.categories ?? [],
        articles: data.articles ?? [],
      }
    },
    async detail(slug: string): Promise<TutorialArticle> {
      const res = await http.get<ApiEnvelope<TutorialArticle>>(
        `/api/v1/tutorials/articles/${encodeURIComponent(slug)}`,
      )
      return unwrap(res)
    },
  }
}
