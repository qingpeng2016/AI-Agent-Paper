type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export type PaperExperimentPlanItem = {
  id: string
  version: number
  status: string
  title?: string
  summary?: string
  content_medium?: string
  format: string
  literature_review_id?: string
  experiment_data_uri?: string
  created_at: string
}

export async function uploadExperimentPlanData(
  manuscriptId: string,
  experimentPlanId: string,
  file: File,
): Promise<PaperExperimentPlanItem> {
  const msNum = Number(manuscriptId)
  const planNum = Number(experimentPlanId)
  if (!Number.isFinite(msNum) || msNum <= 0 || !Number.isFinite(planNum) || planNum <= 0) {
    throw new Error('invalid ids')
  }
  const body = new FormData()
  body.append('manuscript_id', String(msNum))
  body.append('experiment_plan_id', String(planNum))
  body.append('file', file)
  const res = await fetch('/api/v1/paper/experiment-plans/upload-experiment-data', {
    method: 'POST',
    credentials: 'include',
    body,
  })
  if (!res.ok) {
    throw new Error(`upload experiment-data http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperExperimentPlanItem>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'upload experiment-data failed')
  }
  return envelope.data
}

export async function deleteExperimentPlanData(
  manuscriptId: string,
  experimentPlanId: string,
): Promise<PaperExperimentPlanItem> {
  const msNum = Number(manuscriptId)
  const planNum = Number(experimentPlanId)
  if (!Number.isFinite(msNum) || msNum <= 0 || !Number.isFinite(planNum) || planNum <= 0) {
    throw new Error('invalid ids')
  }
  const res = await fetch('/api/v1/paper/experiment-plans/delete-experiment-data', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      manuscript_id: msNum,
      experiment_plan_id: planNum,
    }),
  })
  if (!res.ok) {
    throw new Error(`delete experiment-data http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperExperimentPlanItem>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'delete experiment-data failed')
  }
  return envelope.data
}

export type PaperExperimentPlanListResponse = {
  manuscript_id: number
  items: PaperExperimentPlanItem[]
}

export async function fetchExperimentPlanDetail(
  manuscriptId: string,
  experimentPlanId: string,
): Promise<PaperExperimentPlanItem> {
  const msNum = Number(manuscriptId)
  const planNum = Number(experimentPlanId)
  if (!Number.isFinite(msNum) || msNum <= 0 || !Number.isFinite(planNum) || planNum <= 0) {
    throw new Error('invalid ids')
  }
  const q = new URLSearchParams({
    manuscript_id: String(msNum),
    experiment_plan_id: String(planNum),
  })
  const res = await fetch(`/api/v1/paper/experiment-plans/detail?${q}`, { credentials: 'include' })
  if (!res.ok) {
    throw new Error(`experiment-plan detail http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperExperimentPlanItem>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'experiment-plan detail failed')
  }
  return envelope.data
}

export async function fetchExperimentPlans(
  manuscriptId: string,
): Promise<PaperExperimentPlanListResponse> {
  const idNum = Number(manuscriptId)
  if (!Number.isFinite(idNum) || idNum <= 0) {
    return { manuscript_id: 0, items: [] }
  }
  const res = await fetch(
    `/api/v1/paper/experiment-plans?manuscript_id=${encodeURIComponent(idNum)}`,
    { credentials: 'include' },
  )
  if (!res.ok) {
    throw new Error(`experiment-plans http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperExperimentPlanListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'experiment-plans failed')
  }
  return envelope.data
}

export function formatExperimentPlanStatus(status: string): string {
  const map: Record<string, string> = {
    completed: '已完成',
    draft: '草稿',
    deleted: '已删除',
  }
  return map[status] ?? status
}

export function experimentPlanRowTitle(item: PaperExperimentPlanItem): string {
  const t = item.title?.trim()
  if (t) return t
  return `实验方案 v${item.version}`
}
