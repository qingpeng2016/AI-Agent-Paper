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
  manuscript_ids?: string[]
  manuscript_titles?: string[]
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

export async function createUserLiteratureCategory(body: {
  name: string
}): Promise<PaperUserLiteratureCategoryItem> {
  const res = await fetch('/api/v1/paper/user-literature/categories', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`user-literature category create http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperUserLiteratureCategoryItem>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'user-literature category create failed')
  }
  return envelope.data
}

export async function updateUserLiteratureCategory(body: {
  id: string
  name: string
}): Promise<PaperUserLiteratureCategoryItem> {
  const res = await fetch('/api/v1/paper/user-literature/categories', {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`user-literature category update http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperUserLiteratureCategoryItem>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'user-literature category update failed')
  }
  return envelope.data
}

export type PaperUserLiteratureCategoryLiteratureIDsResponse = {
  category_id: string
  literature_ids: string[]
}

export async function fetchCategoryLiteratureIds(
  categoryId: string,
): Promise<PaperUserLiteratureCategoryLiteratureIDsResponse> {
  const q = new URLSearchParams()
  q.set('category_id', categoryId)
  const res = await fetch(
    `/api/v1/paper/user-literature/categories/literature-ids?${q}`,
    { credentials: 'include' },
  )
  if (!res.ok) {
    throw new Error(`category literature-ids http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperUserLiteratureCategoryLiteratureIDsResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'category literature-ids failed')
  }
  return envelope.data
}

export async function setCategoryLiteratureMembership(body: {
  category_id: string
  literature_ids: string[]
}): Promise<PaperUserLiteratureCategoryLiteratureIDsResponse> {
  const res = await fetch('/api/v1/paper/user-literature/categories/literature-membership', {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`category literature-membership http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperUserLiteratureCategoryLiteratureIDsResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'category literature-membership failed')
  }
  return envelope.data
}

export type PaperManuscriptWithLiteratureItem = {
  id: string
  title: string
  is_current: boolean
  literature_ids?: string[]
  literature_titles?: string[]
  created_at?: string
}

export type PaperManuscriptWithLiteratureListResponse = {
  items: PaperManuscriptWithLiteratureItem[]
  total?: number
  page?: number
  page_size?: number
}

export async function fetchManuscriptsWithLiterature(options: {
  page: number
  page_size?: number
}): Promise<PaperManuscriptWithLiteratureListResponse> {
  const q = new URLSearchParams()
  q.set('page', String(options.page))
  q.set('page_size', String(options.page_size ?? PAPER_LIST_PAGE_SIZE))
  const res = await fetch(`/api/v1/paper/user-literature/manuscripts-with-literature?${q}`, {
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error(`manuscripts-with-literature http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptWithLiteratureListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscripts-with-literature list failed')
  }
  return envelope.data
}

export type PaperManuscriptLiteratureIDsResponse = {
  manuscript_id: string
  literature_ids: string[]
}

export async function fetchManuscriptLiteratureIds(
  manuscriptId: string,
): Promise<PaperManuscriptLiteratureIDsResponse> {
  const q = new URLSearchParams()
  q.set('manuscript_id', manuscriptId)
  const res = await fetch(`/api/v1/paper/user-literature/manuscripts/literature-ids?${q}`, {
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error(`manuscript literature-ids http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptLiteratureIDsResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript literature-ids failed')
  }
  return envelope.data
}

export async function setManuscriptLiteratureMembership(body: {
  manuscript_id: string
  literature_ids: string[]
}): Promise<PaperManuscriptLiteratureIDsResponse> {
  const res = await fetch('/api/v1/paper/user-literature/manuscripts/literature-membership', {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`manuscript literature-membership http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptLiteratureIDsResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript literature-membership failed')
  }
  return envelope.data
}

const USER_LITERATURE_FETCH_PAGE_SIZE = 100

export async function fetchAllUserLiterature(options?: {
  q?: string
}): Promise<PaperUserLiteratureItem[]> {
  const all: PaperUserLiteratureItem[] = []
  let page = 1
  let total = 0
  const keyword = options?.q?.trim()
  do {
    const data = await fetchUserLiterature({
      page,
      page_size: USER_LITERATURE_FETCH_PAGE_SIZE,
      q: keyword || undefined,
    })
    const batch = data.items ?? []
    all.push(...batch)
    total = data.total ?? all.length
    if (batch.length === 0) break
    page += 1
  } while (all.length < total)
  return all
}

export async function fetchUserLiterature(options: {
  page: number
  page_size?: number
  q?: string
  /** 空=全部；none=未分类；否则分类 id */
  category_id?: string
  /** 空=全部；none=未关联论文；否则 paper_manuscript id */
  manuscript_id?: string
}): Promise<PaperUserLiteratureListResponse> {
  const q = new URLSearchParams()
  q.set('page', String(options.page))
  q.set('page_size', String(options.page_size ?? PAPER_LIST_PAGE_SIZE))
  const keyword = options.q?.trim()
  if (keyword) q.set('q', keyword)
  const cat = options.category_id?.trim()
  if (cat) q.set('category_id', cat)
  const ms = options.manuscript_id?.trim()
  if (ms) q.set('manuscript_id', ms)
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
