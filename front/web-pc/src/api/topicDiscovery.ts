export type TopicDiscoveryFormOptionsResponse = {
  disciplines: Array<{
    code: string
    label: string
    name_en?: string
    sort: number
    literature_source_codes?: string[]
  }>
  execution_intensities: Array<{
    code: string
    label: string
    multiplier: number
    max_papers: number
    max_ideas: number
  }>
  audit_levels: Array<{
    code: string
    label: string
    citation_strength: number
    claim_strength: number
    kill_argument_strength: number
    audit_rounds: number
  }>
}

type ApiEnvelope<T> = { code: number; message?: string; data: T }

export type PaperSelectOption = { value: string; label: string }

const FORM_OPTIONS_TIMEOUT_MS = 8_000

export type TopicDiscoveryRunRequest = {
  manuscript_id?: number
  manuscript_title?: string
  discipline_code: string
  direction: string
  venue: string
  source_codes: string[]
  intensity: string
  audit_level: string
  human_checkpoint: boolean
  /** start | continue | run_all */
  action: string
}

export type TopicDiscoveryStepDTO = {
  stage_code: string
  status: string
  summary_text?: string
  result?: unknown
  meta?: unknown
}

export type TopicDiscoveryRunResponse = {
  manuscript_id: number
  run_version: number
  run_status: string
  human_checkpoint: boolean
  pause_after_stage?: string
  steps: TopicDiscoveryStepDTO[]
}

export async function postTopicDiscoveryRun(
  body: TopicDiscoveryRunRequest,
): Promise<TopicDiscoveryRunResponse> {
  const res = await fetch('/api/v1/paper/topic-discovery/run', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`topic-discovery run http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<TopicDiscoveryRunResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'topic-discovery run failed')
  }
  return envelope.data
}

export async function fetchCurrentTopicDiscoveryRun(
  manuscriptId: number,
): Promise<TopicDiscoveryRunResponse | null> {
  const q = new URLSearchParams({ manuscript_id: String(manuscriptId) })
  const res = await fetch(`/api/v1/paper/topic-discovery/run/current?${q}`, {
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error(`topic-discovery current http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<TopicDiscoveryRunResponse | null>
  if (envelope.code !== 200) {
    throw new Error(envelope.message ?? 'topic-discovery current failed')
  }
  return envelope.data ?? null
}

export async function fetchTopicDiscoveryFormOptions(): Promise<{
  disciplineSelectOptions: PaperSelectOption[]
  intensityOptions: PaperSelectOption[]
  auditOptions: PaperSelectOption[]
}> {
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), FORM_OPTIONS_TIMEOUT_MS)
  let res: Response
  try {
    res = await fetch('/api/v1/paper/topic-discovery/form-options', {
      signal: controller.signal,
    })
  } finally {
    window.clearTimeout(timer)
  }
  if (!res.ok) {
    throw new Error(`form-options http ${res.status}`)
  }
  const body = (await res.json()) as ApiEnvelope<TopicDiscoveryFormOptionsResponse>
  if (body.code !== 200 || !body.data) {
    throw new Error(body.message ?? 'form-options failed')
  }
  const d = body.data
  return {
    disciplineSelectOptions: d.disciplines.map((x) => ({ value: x.code, label: x.label })),
    intensityOptions: d.execution_intensities.map((x) => ({ value: x.code, label: x.label })),
    auditOptions: d.audit_levels.map((x) => ({ value: x.code, label: x.label })),
  }
}
