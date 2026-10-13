type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export type ManuscriptContentLanguage = 'zh' | 'en'

export const PAPER_LIST_PAGE_SIZE = 20

export type PaperManuscriptApiItem = {
  id: string
  title: string
  status: 'active' | 'archived'
  is_current: boolean
  content_language: ManuscriptContentLanguage
  discipline_code?: string
  discipline_label?: string
  created_at?: string
  draft_version?: number
}

export function formatManuscriptDraftVersion(draftVersion?: number): string {
  const v = draftVersion ?? 0
  if (!Number.isFinite(v) || v < 1) return '—'
  return `v${v}`
}

export type PaperManuscriptListResponse = {
  current_manuscript_id?: number
  items: PaperManuscriptApiItem[]
  total?: number
  page?: number
  page_size?: number
}

export async function fetchPaperManuscripts(options?: {
  page?: number
  page_size?: number
}): Promise<PaperManuscriptListResponse> {
  const q = new URLSearchParams()
  if (options?.page != null && options.page > 0) {
    q.set('page', String(options.page))
    q.set('page_size', String(options.page_size ?? PAPER_LIST_PAGE_SIZE))
  }
  const qs = q.toString()
  const url = qs ? `/api/v1/paper/manuscripts?${qs}` : '/api/v1/paper/manuscripts'
  const res = await fetch(url, { credentials: 'include' })
  if (!res.ok) {
    throw new Error(`manuscripts list http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscripts list failed')
  }
  return envelope.data
}

export async function createPaperManuscript(
  title: string,
  disciplineCode: string,
  contentLanguage: ManuscriptContentLanguage = 'en',
): Promise<PaperManuscriptListResponse> {
  const lang = contentLanguage === 'zh' ? 'zh' : 'en'
  const res = await fetch('/api/v1/paper/manuscripts', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title: title.trim(),
      discipline_code: disciplineCode,
      content_language: lang,
    }),
  })
  if (!res.ok) {
    throw new Error(`manuscripts create http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscripts create failed')
  }
  return envelope.data
}

export async function setCurrentPaperManuscript(manuscriptId: string): Promise<PaperManuscriptListResponse> {
  const idNum = Number(manuscriptId)
  if (!Number.isFinite(idNum) || idNum <= 0) {
    throw new Error('invalid manuscript_id')
  }
  const res = await fetch('/api/v1/paper/manuscripts/set-current', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ manuscript_id: idNum }),
  })
  if (!res.ok) {
    throw new Error(`manuscripts set-current http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscripts set-current failed')
  }
  return envelope.data
}
