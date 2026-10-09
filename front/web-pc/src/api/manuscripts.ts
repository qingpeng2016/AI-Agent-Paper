type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export type PaperManuscriptApiItem = {
  id: string
  title: string
  status: 'active' | 'archived'
  is_current: boolean
  discipline_code?: string
  discipline_label?: string
}

export type PaperManuscriptListResponse = {
  current_manuscript_id?: number
  items: PaperManuscriptApiItem[]
}

export async function fetchPaperManuscripts(): Promise<PaperManuscriptListResponse> {
  const res = await fetch('/api/v1/paper/manuscripts', { credentials: 'include' })
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
): Promise<PaperManuscriptListResponse> {
  const res = await fetch('/api/v1/paper/manuscripts', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title: title.trim(), discipline_code: disciplineCode }),
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
