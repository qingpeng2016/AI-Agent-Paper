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
    is_default?: boolean
  }>
  audit_levels: Array<{
    code: string
    label: string
    citation_strength: number
    claim_strength: number
    kill_argument_strength: number
    audit_rounds: number
    is_default?: boolean
  }>
}

type ApiEnvelope<T> = { code: number; message?: string; data: T }

export type PaperSelectOption = { value: string; label: string }

const FORM_OPTIONS_TIMEOUT_MS = 8_000

export type TopicDiscoveryRunRequest = {
  manuscript_title?: string
  discipline_code: string
  keywords: string[]
  description: string
  direction?: string
  venue: string
  source_codes: string[]
  intensity: string
  audit_level: string
  human_checkpoint: boolean
  /** start | continue | run_all */
  action: string
}

export type TopicDiscoveryLiteratureLink = {
  title: string
  url: string
  external_key?: string
  source_code?: string
}

export type TopicDiscoveryStepDTO = {
  stage_code: string
  status: string
  summary_text?: string
  result?: unknown
  meta?: unknown
}

export function parseRetrieveLiteratureLinks(
  run?: TopicDiscoveryRunResponse | null,
): TopicDiscoveryLiteratureLink[] {
  if (!run) return []
  const step = run.steps.find((s) => s.stage_code === 'retrieve')
  if (!step) return []
  const meta = (step.meta ?? {}) as Record<string, unknown>
  const fromMeta = meta.literature_links as TopicDiscoveryLiteratureLink[] | undefined
  if (Array.isArray(fromMeta) && fromMeta.length > 0) {
    return fromMeta.filter((l) => l.url?.trim())
  }
  const result = step.result as Record<string, unknown> | undefined
  const hits = (result?.literature_hits as Array<Record<string, unknown>>) ?? []
  return hits
    .map((h) => {
      const nested = (h.meta as Record<string, unknown> | undefined) ?? {}
      const url = String(h.url ?? nested.url ?? '').trim()
      return {
        title: String(h.title ?? h.external_key ?? '文献'),
        url,
        external_key: h.external_key != null ? String(h.external_key) : undefined,
        source_code: h.source_code != null ? String(h.source_code) : undefined,
      }
    })
    .filter((h) => h.url)
}

export type TopicDiscoveryRunResponse = {
  manuscript_id: number
  run_version: number
  run_status: string
  human_checkpoint: boolean
  pause_after_stage?: string
  steps: TopicDiscoveryStepDTO[]
}

const RUN_TIMEOUT_MS = 600_000

export async function postTopicDiscoveryRun(
  body: TopicDiscoveryRunRequest,
): Promise<TopicDiscoveryRunResponse> {
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), RUN_TIMEOUT_MS)
  let res: Response
  try {
    res = await fetch('/api/v1/paper/topic-discovery/run', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: controller.signal,
    })
  } finally {
    window.clearTimeout(timer)
  }
  if (!res.ok) {
    throw new Error(`topic-discovery run http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<TopicDiscoveryRunResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'topic-discovery run failed')
  }
  return envelope.data
}

export async function fetchCurrentTopicDiscoveryRun(): Promise<TopicDiscoveryRunResponse | null> {
  const res = await fetch('/api/v1/paper/topic-discovery/run/current', {
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

export async function commitTopicDiscoveryManuscript(title?: string): Promise<TopicDiscoveryRunResponse> {
  const res = await fetch('/api/v1/paper/topic-discovery/commit-manuscript', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ manuscript_title: title ?? '' }),
  })
  if (!res.ok) {
    throw new Error(`topic-discovery commit http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<TopicDiscoveryRunResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'topic-discovery commit failed')
  }
  return envelope.data
}

export async function cancelTopicDiscoveryRun(): Promise<void> {
  const res = await fetch('/api/v1/paper/topic-discovery/run/cancel', {
    method: 'POST',
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error(`topic-discovery cancel http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<unknown>
  if (envelope.code !== 200) {
    throw new Error(envelope.message ?? 'topic-discovery cancel failed')
  }
}

function pickDefaultCode(items: Array<{ code: string; is_default?: boolean }>): string {
  return items.find((x) => x.is_default)?.code ?? items[0]?.code ?? ''
}

export async function fetchTopicDiscoveryFormOptions(): Promise<{
  disciplineSelectOptions: PaperSelectOption[]
  intensityOptions: PaperSelectOption[]
  auditOptions: PaperSelectOption[]
  defaultIntensityCode: string
  defaultAuditLevelCode: string
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
    defaultIntensityCode: pickDefaultCode(d.execution_intensities),
    defaultAuditLevelCode: pickDefaultCode(d.audit_levels),
  }
}
