import type { UserLiteratureFrom } from '@paper/types'

type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export const PAPER_LIST_PAGE_SIZE = 20

export type PaperUserLiteratureItem = {
  id: string
  from: UserLiteratureFrom | string
  file_ext: string
  file_name: string
  literature_title: string
  note?: string
  size_bytes: number
  storage_uri?: string
  category_ids?: string[]
  category_names?: string[]
  created_at?: string
}

export function formatUserLiteratureCategories(names?: string[]): string {
  if (!names?.length) return '—'
  const parts = names.map((n) => n.trim()).filter(Boolean)
  return parts.length ? parts.join('、') : '—'
}

export type PaperUserLiteratureListResponse = {
  items: PaperUserLiteratureItem[]
  total?: number
  page?: number
  page_size?: number
}

const FROM_LABELS: Record<string, string> = {
  user_upload: '本地上传',
  arxiv: 'arXiv',
  openalex: 'OpenAlex',
  semantic_scholar: 'Semantic Scholar',
  crossref: 'Crossref',
  pubmed: 'PubMed',
}

export function formatUserLiteratureFrom(from?: string): string {
  if (!from?.trim()) return '—'
  return FROM_LABELS[from] ?? from
}

export type UserLiteratureFromTone =
  | 'upload'
  | 'arxiv'
  | 'catalog'
  | 'default'

export function userLiteratureFromTone(from?: string): UserLiteratureFromTone {
  if (from === 'user_upload') return 'upload'
  if (from === 'arxiv') return 'arxiv'
  if (
    from === 'openalex' ||
    from === 'semantic_scholar' ||
    from === 'crossref' ||
    from === 'pubmed'
  ) {
    return 'catalog'
  }
  return 'default'
}

export function formatFileSizeBytes(n?: number): string {
  const v = n ?? 0
  if (v < 1024) return `${v} B`
  if (v < 1024 * 1024) return `${(v / 1024).toFixed(1)} KB`
  return `${(v / (1024 * 1024)).toFixed(2)} MB`
}

export type PaperUserLiteratureCategoryItem = {
  id: string
  name: string
  sort: number
  status: string
  created_at?: string
}

export type PaperUserLiteratureCategoryListResponse = {
  items: PaperUserLiteratureCategoryItem[]
  total?: number
  page?: number
  page_size?: number
}

export async function fetchUserLiteratureCategories(options: {
  page: number
  page_size?: number
}): Promise<PaperUserLiteratureCategoryListResponse> {
  const q = new URLSearchParams()
  q.set('page', String(options.page))
  q.set('page_size', String(options.page_size ?? PAPER_LIST_PAGE_SIZE))
  const res = await fetch(`/api/v1/paper/user-literature/categories?${q}`, {
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error(`user-literature categories http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperUserLiteratureCategoryListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'user-literature categories list failed')
  }
  return envelope.data
}

export async function fetchUserLiterature(options: {
  page: number
  page_size?: number
  q?: string
  /** 空=全部；none=未分类；否则分类 id */
  category_id?: string
}): Promise<PaperUserLiteratureListResponse> {
  const q = new URLSearchParams()
  q.set('page', String(options.page))
  q.set('page_size', String(options.page_size ?? PAPER_LIST_PAGE_SIZE))
  const keyword = options.q?.trim()
  if (keyword) q.set('q', keyword)
  const cat = options.category_id?.trim()
  if (cat) q.set('category_id', cat)
  const res = await fetch(`/api/v1/paper/user-literature?${q}`, { credentials: 'include' })
  if (!res.ok) {
    throw new Error(`user-literature list http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperUserLiteratureListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'user-literature list failed')
  }
  return envelope.data
}
