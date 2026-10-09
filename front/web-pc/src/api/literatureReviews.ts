type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export const LITERATURE_REVIEW_STATUS_GENERATING_EXPERIMENT_PLAN = 'gen_exp_plan' as const
export const LITERATURE_REVIEW_STATUS_EXPERIMENT_PLAN_FAILED = 'gen_exp_fail' as const

export type PaperLiteratureReviewItem = {
  id: string
  version: number
  status: string
  experiment_plan_id?: string
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

export async function softDeleteLiteratureReview(
  manuscriptId: string,
  literatureReviewId: string,
): Promise<void> {
  const msNum = Number(manuscriptId)
  const reviewNum = Number(literatureReviewId)
  if (!Number.isFinite(msNum) || msNum <= 0 || !Number.isFinite(reviewNum) || reviewNum <= 0) {
    throw new Error('invalid ids')
  }
  const res = await fetch('/api/v1/paper/literature-reviews/soft-delete', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      manuscript_id: msNum,
      literature_review_id: reviewNum,
    }),
  })
  if (!res.ok) {
    throw new Error(`literature-reviews soft-delete http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<{ ok?: boolean }>
  if (envelope.code !== 200) {
    throw new Error(envelope.message ?? 'soft-delete failed')
  }
}

export async function generateExperimentPlanFromLiteratureReview(
  manuscriptId: string,
  literatureReviewId: string,
): Promise<void> {
  const msNum = Number(manuscriptId)
  const reviewNum = Number(literatureReviewId)
  if (!Number.isFinite(msNum) || msNum <= 0 || !Number.isFinite(reviewNum) || reviewNum <= 0) {
    throw new Error('invalid ids')
  }
  const res = await fetch('/api/v1/paper/literature-reviews/generate-experiment-plan', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      manuscript_id: msNum,
      literature_review_id: reviewNum,
    }),
  })
  if (!res.ok) {
    throw new Error(`generate-experiment-plan http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<{ ok?: boolean; status?: string }>
  if (envelope.code !== 200) {
    throw new Error(envelope.message ?? 'generate-experiment-plan failed')
  }
}

export function formatLiteratureReviewStatus(status: string): string {
  if (status === 'superseded') return '已完成'
  const map: Record<string, string> = {
    completed: '已完成',
    draft: '草稿',
    deleted: '已删除',
    [LITERATURE_REVIEW_STATUS_GENERATING_EXPERIMENT_PLAN]: '生成实验方案中',
    [LITERATURE_REVIEW_STATUS_EXPERIMENT_PLAN_FAILED]: '生成实验方案失败',
  }
  return map[status] ?? status
}

export function isLiteratureReviewGeneratingExperimentPlan(status: string): boolean {
  return status === LITERATURE_REVIEW_STATUS_GENERATING_EXPERIMENT_PLAN
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
