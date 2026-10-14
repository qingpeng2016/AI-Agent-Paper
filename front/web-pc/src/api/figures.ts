type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export const PAPER_LIST_PAGE_SIZE = 20

export type PaperFigureItem = {
  id: string
  origin: string
  file_kind: string
  file_name: string
  title: string
  note?: string
  size_bytes: number
  storage_uri?: string
  category_ids?: string[]
  category_names?: string[]
  manuscript_ids?: string[]
  manuscript_titles?: string[]
  created_at?: string
}

export function formatFigureCategories(names?: string[]): string {
  if (!names?.length) return '—'
  const parts = names.map((n) => n.trim()).filter(Boolean)
  return parts.length ? parts.join('、') : '—'
}

export type PaperFigureListResponse = {
  items: PaperFigureItem[]
  total?: number
  page?: number
  page_size?: number
}

const ORIGIN_LABELS: Record<string, string> = {
  upload: '本地上传',
  generated: '一键生成',
}

export function formatFigureOrigin(origin?: string): string {
  if (!origin?.trim()) return '—'
  return ORIGIN_LABELS[origin] ?? origin
}

/** 与文献管理 `userLiteratureFromTone` 共用 pill 样式类名 */
export type FigureOriginTone = 'upload' | 'arxiv' | 'catalog' | 'default'

export function figureOriginTone(origin?: string): FigureOriginTone {
  if (origin === 'upload') return 'upload'
  if (origin === 'generated') return 'catalog'
  return 'default'
}

export function formatFileSizeBytes(n?: number): string {
  const v = n ?? 0
  if (v < 1024) return `${v} B`
  if (v < 1024 * 1024) return `${(v / 1024).toFixed(1)} KB`
  return `${(v / (1024 * 1024)).toFixed(2)} MB`
}

export type PaperFigureCategoryItem = {
  id: string
  name: string
  status: string
  created_at?: string
}

export type PaperFigureCategoryListResponse = {
  items: PaperFigureCategoryItem[]
  total?: number
  page?: number
  page_size?: number
}

export async function fetchFigureCategories(options: {
  page: number
  page_size?: number
}): Promise<PaperFigureCategoryListResponse> {
  const q = new URLSearchParams()
  q.set('page', String(options.page))
  q.set('page_size', String(options.page_size ?? PAPER_LIST_PAGE_SIZE))
  const res = await fetch(`/api/v1/paper/figures/categories?${q}`, { credentials: 'include' })
  if (!res.ok) throw new Error(`figures categories http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperFigureCategoryListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'figures categories list failed')
  }
  return envelope.data
}

export async function createFigureCategory(body: {
  name: string
}): Promise<PaperFigureCategoryItem> {
  const res = await fetch('/api/v1/paper/figures/categories', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`figure category create http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperFigureCategoryItem>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'figure category create failed')
  }
  return envelope.data
}

export async function updateFigureCategory(body: {
  id: string
  name: string
}): Promise<PaperFigureCategoryItem> {
  const res = await fetch('/api/v1/paper/figures/categories', {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`figure category update http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperFigureCategoryItem>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'figure category update failed')
  }
  return envelope.data
}

export type PaperFigureCategoryFigureIDsResponse = {
  category_id: string
  figure_ids: string[]
}

export async function fetchCategoryFigureIds(
  categoryId: string,
): Promise<PaperFigureCategoryFigureIDsResponse> {
  const q = new URLSearchParams()
  q.set('category_id', categoryId)
  const res = await fetch(`/api/v1/paper/figures/categories/figure-ids?${q}`, {
    credentials: 'include',
  })
  if (!res.ok) throw new Error(`category figure-ids http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperFigureCategoryFigureIDsResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'category figure-ids failed')
  }
  return envelope.data
}

export async function setCategoryFigureMembership(body: {
  category_id: string
  figure_ids: string[]
}): Promise<PaperFigureCategoryFigureIDsResponse> {
  const res = await fetch('/api/v1/paper/figures/categories/figure-membership', {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`category figure-membership http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperFigureCategoryFigureIDsResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'category figure-membership failed')
  }
  return envelope.data
}

export type PaperManuscriptWithFiguresItem = {
  id: string
  title: string
  is_current: boolean
  figure_ids?: string[]
  figure_titles?: string[]
  created_at?: string
}

export type PaperManuscriptWithFiguresListResponse = {
  items: PaperManuscriptWithFiguresItem[]
  total?: number
  page?: number
  page_size?: number
}

export async function fetchManuscriptsWithFigures(options: {
  page: number
  page_size?: number
}): Promise<PaperManuscriptWithFiguresListResponse> {
  const q = new URLSearchParams()
  q.set('page', String(options.page))
  q.set('page_size', String(options.page_size ?? PAPER_LIST_PAGE_SIZE))
  const res = await fetch(`/api/v1/paper/figures/manuscripts-with-figures?${q}`, {
    credentials: 'include',
  })
  if (!res.ok) throw new Error(`manuscripts-with-figures http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptWithFiguresListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscripts-with-figures list failed')
  }
  return envelope.data
}

export type PaperManuscriptFigureIDsResponse = {
  manuscript_id: string
  figure_ids: string[]
}

export async function fetchManuscriptFigureIds(
  manuscriptId: string,
): Promise<PaperManuscriptFigureIDsResponse> {
  const q = new URLSearchParams()
  q.set('manuscript_id', manuscriptId)
  const res = await fetch(`/api/v1/paper/figures/manuscripts/figure-ids?${q}`, {
    credentials: 'include',
  })
  if (!res.ok) throw new Error(`manuscript figure-ids http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptFigureIDsResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript figure-ids failed')
  }
  return envelope.data
}

export async function setManuscriptFigureMembership(body: {
  manuscript_id: string
  figure_ids: string[]
}): Promise<PaperManuscriptFigureIDsResponse> {
  const res = await fetch('/api/v1/paper/figures/manuscripts/figure-membership', {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`manuscript figure-membership http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptFigureIDsResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript figure-membership failed')
  }
  return envelope.data
}

const FIGURE_FETCH_PAGE_SIZE = 100

export async function fetchAllFigures(options?: { q?: string }): Promise<PaperFigureItem[]> {
  const all: PaperFigureItem[] = []
  let page = 1
  let total = 0
  const keyword = options?.q?.trim()
  do {
    const data = await fetchFigures({
      page,
      page_size: FIGURE_FETCH_PAGE_SIZE,
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

export async function fetchFigures(options: {
  page: number
  page_size?: number
  q?: string
  category_id?: string
  manuscript_id?: string
}): Promise<PaperFigureListResponse> {
  const q = new URLSearchParams()
  q.set('page', String(options.page))
  q.set('page_size', String(options.page_size ?? PAPER_LIST_PAGE_SIZE))
  const keyword = options.q?.trim()
  if (keyword) q.set('q', keyword)
  const cat = options.category_id?.trim()
  if (cat) q.set('category_id', cat)
  const ms = options.manuscript_id?.trim()
  if (ms) q.set('manuscript_id', ms)
  const res = await fetch(`/api/v1/paper/figures?${q}`, { credentials: 'include' })
  if (!res.ok) throw new Error(`figures list http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperFigureListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'figures list failed')
  }
  return envelope.data
}
