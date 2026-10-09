type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export type PaperLiteratureReviewItem = {
  id: string
  version: number
  is_current: boolean
  status: string
  structure?: string
  title?: string
  summary?: string
  content_medium?: string
  format: string
  created_at: string
}

export type PaperLiteratureReviewListResponse = {
  manuscript_id: number
  items: PaperLiteratureReviewItem[]
}

export async function fetchLiteratureReviews(
  manuscriptId: string,
): Promise<PaperLiteratureReviewListResponse> {
  const idNum = Number(manuscriptId)
  if (!Number.isFinite(idNum) || idNum <= 0) {
    return { manuscript_id: 0, items: [] }
  }
  const res = await fetch(
    `/api/v1/paper/literature-reviews?manuscript_id=${encodeURIComponent(idNum)}`,
    { credentials: 'include' },
  )
  if (!res.ok) {
    throw new Error(`literature-reviews http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperLiteratureReviewListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'literature-reviews failed')
  }
  return envelope.data
}

export function formatLiteratureReviewTabLabel(item: PaperLiteratureReviewItem): string {
  const title = item.title?.trim()
  if (title) {
    return title.length > 28 ? `${title.slice(0, 28)}…` : title
  }
  let when = ''
  try {
    when = new Date(item.created_at).toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    when = item.created_at
  }
  return `v${item.version} · ${when}`
}
