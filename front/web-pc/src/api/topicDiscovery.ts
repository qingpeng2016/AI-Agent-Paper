import { TOPIC_DISCOVERY_FLOW_STEPS } from '@paper/types'

export type TopicDiscoveryFormOptionsResponse = {
  disciplines: Array<{
    code: string
    label: string
    name_en?: string
    sort: number
  }>
  execution_intensities: Array<{
    code: string
    label: string
    exec_rounds: number
    max_papers: number
    max_ideas: number
    is_default?: boolean
  }>
  audit_levels: Array<{
    code: string
    label: string
    llm_description_citation_check_code: number
    llm_description_claim_check_code: number
    llm_description_kill_check_code: number
    is_default?: boolean
  }>
  literature_sources?: Array<{
    code: string
    label: string
    sort: number
    default_selected: boolean
  }>
  venues?: Array<{
    code: string
    label: string
    venue_type: string
    sort: number
  }>
}

export type VenueOption = {
  code: string
  label: string
  venueType: string
  sort: number
}

export type LiteratureSourceOption = {
  code: string
  label: string
  sort: number
  defaultSelected: boolean
}

type ApiEnvelope<T> = { code: number; message?: string; data: T }

export type PaperSelectOption = { value: string; label: string }

/** 与 paper_exec_intensity_topic 种子一致；form-options 失败或 execution_intensities 为空时使用 */
export const EXEC_INTENSITY_FALLBACK_OPTIONS: PaperSelectOption[] = [
  { value: 'fast', label: '更快' },
  { value: 'balanced', label: '平衡' },
  { value: 'deep', label: '更深' },
]

export const DEFAULT_EXEC_INTENSITY_CODE = 'balanced'

const FORM_OPTIONS_TIMEOUT_MS = 8_000

export type TopicDiscoveryRunRequest = {
  manuscript_id: number
  manuscript_title?: string
  discipline_code: string
  keywords: string[]
  description: string
  direction?: string
  venue: string
  venue_codes?: string[]
  source_codes: string[]
  intensity: string
  audit_level: string
  human_checkpoint: boolean
  random_references?: boolean
  user_literature_ids?: string[]
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
  input_params?: Record<string, unknown>
  result?: unknown
  extra?: unknown
}

/** 从 retrieve 步的 input_params 回填选题表单（刷新 / 检查点继续） */
export function formatVenueTextFromCodes(
  codes: readonly string[],
  venues: readonly VenueOption[],
): string {
  if (!codes.length) return ''
  return codes
    .map((c) => venues.find((v) => v.code === c)?.label ?? c)
    .filter(Boolean)
    .join(' / ')
}

export function inferVenueCodesFromText(
  text: string,
  venues: readonly VenueOption[],
): string[] {
  const parts = text
    .split(/[/／|,，、]+/)
    .map((s) => s.trim())
    .filter(Boolean)
  const out: string[] = []
  for (const part of parts) {
    const lower = part.toLowerCase()
    const byCode = venues.find((v) => v.code === part || v.code.toLowerCase() === lower)
    if (byCode) {
      out.push(byCode.code)
      continue
    }
    const byLabel = venues.find(
      (v) => v.label === part || v.label.toLowerCase() === lower,
    )
    if (byLabel) out.push(byLabel.code)
  }
  return [...new Set(out)]
}

export function applyStoredTopicRunInputToForm(
  form: {
    disciplineCode: string
    keywords: string
    description: string
    venueCodes: string[]
    sourceCodes: string[]
    intensity: string
    auditLevel: string
    humanCheckpoint: boolean
    randomReferences: boolean
    userLiteratureIds: string[]
  },
  run?: TopicDiscoveryRunResponse | null,
  venues: readonly VenueOption[] = [],
): void {
  if (!run) return
  if (isTopicRunTerminalCompleted(run)) return
  const retrieve = run.steps.find((s) => s.stage_code === 'retrieve')
  const p = retrieve?.input_params
  if (!p || typeof p !== 'object') return

  let keywords = Array.isArray(p.keywords)
    ? (p.keywords as unknown[]).map((k) => String(k).trim()).filter(Boolean)
    : []
  let description = String(p.description ?? '').trim()
  const legacyDirection = String(p.direction ?? '').trim()

  if (keywords.length === 0 && legacyDirection) {
    const kwMatch = legacyDirection.match(/关键词[：:]\s*(.+)$/m)
    if (kwMatch) {
      keywords = kwMatch[1]
        .split(/[,，;\n、]+/)
        .map((s) => s.trim())
        .filter(Boolean)
    }
    const contentMatch = legacyDirection.match(/研究内容[：:]\s*([\s\S]*?)(?:\n关键词|$)/)
    if (contentMatch) {
      description = contentMatch[1].trim()
    } else if (!description) {
      description = legacyDirection
    }
  }

  if (keywords.length) form.keywords = keywords.join(', ')
  if (description) form.description = description

  const discipline = String(p.discipline_code ?? '').trim()
  if (discipline) form.disciplineCode = discipline
  if (Array.isArray(p.venue_codes) && p.venue_codes.length) {
    form.venueCodes = (p.venue_codes as unknown[]).map((c) => String(c).trim()).filter(Boolean)
  } else {
    const venue = String(p.venue ?? '').trim()
    if (venue && venues.length) {
      form.venueCodes = inferVenueCodesFromText(venue, venues)
    }
  }
  if (Array.isArray(p.source_codes) && p.source_codes.length) {
    form.sourceCodes = (p.source_codes as unknown[]).map((c) => String(c).trim()).filter(Boolean)
  }
  const intensity = String(p.intensity ?? '').trim()
  if (intensity) form.intensity = intensity
  const audit = String(p.audit_level ?? '').trim()
  if (audit) form.auditLevel = audit
  if (typeof p.human_checkpoint === 'boolean') {
    form.humanCheckpoint = p.human_checkpoint
  }
  if (typeof p.random_references === 'boolean') {
    form.randomReferences = p.random_references
  }
  if (Array.isArray(p.user_literature_ids)) {
    form.userLiteratureIds = (p.user_literature_ids as unknown[])
      .map((id) => String(id).trim())
      .filter(Boolean)
  }
}

export function parseRetrieveLiteratureLinks(
  run?: TopicDiscoveryRunResponse | null,
): TopicDiscoveryLiteratureLink[] {
  if (!run) return []
  const step = run.steps.find((s) => s.stage_code === 'retrieve')
  if (!step) return []
  const extra = (step.extra ?? {}) as Record<string, unknown>
  const fromExtra = extra.literature_links as TopicDiscoveryLiteratureLink[] | undefined
  if (Array.isArray(fromExtra) && fromExtra.length > 0) {
    return fromExtra.filter((l) => l.url?.trim())
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

/** 三步均 completed（或 run_status=completed）：终态，不再把 retrieve 输入回填到表单 */
export function isTopicRunTerminalCompleted(
  data: TopicDiscoveryRunResponse | null | undefined,
): boolean {
  if (!data) return false
  if (data.run_status === 'completed') return true
  return TOPIC_DISCOVERY_FLOW_STEPS.every((def) => {
    const st = data.steps.find((s) => s.stage_code === def.stageCode)?.status
    return st === 'completed'
  })
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

export async function fetchCurrentTopicDiscoveryRun(
  manuscriptId: string | number,
): Promise<TopicDiscoveryRunResponse | null> {
  const id = typeof manuscriptId === 'string' ? manuscriptId.trim() : String(manuscriptId)
  if (!id || !/^\d+$/.test(id)) {
    return null
  }
  const res = await fetch(
    `/api/v1/paper/topic-discovery/run/current?manuscript_id=${encodeURIComponent(id)}`,
    { credentials: 'include' },
  )
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

export async function cancelTopicDiscoveryRun(manuscriptId: string | number): Promise<void> {
  const idNum = Number(typeof manuscriptId === 'string' ? manuscriptId.trim() : manuscriptId)
  const res = await fetch('/api/v1/paper/topic-discovery/run/cancel', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ manuscript_id: idNum }),
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

export function defaultLiteratureSourceCodesFromOptions(
  sources: LiteratureSourceOption[],
): string[] {
  return sources.filter((s) => s.defaultSelected).map((s) => s.code)
}

export async function fetchTopicDiscoveryFormOptions(): Promise<{
  disciplineSelectOptions: PaperSelectOption[]
  venueOptions: VenueOption[]
  venueSelectOptions: PaperSelectOption[]
  intensityOptions: PaperSelectOption[]
  auditOptions: PaperSelectOption[]
  literatureSources: LiteratureSourceOption[]
  defaultLiteratureSourceCodes: string[]
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
  const literatureSources = (d.literature_sources ?? [])
    .slice()
    .sort((a, b) => a.sort - b.sort || a.code.localeCompare(b.code))
    .map((x) => ({
      code: x.code,
      label: x.label,
      sort: x.sort,
      defaultSelected: Boolean(x.default_selected),
    }))
  const venueOptions = (d.venues ?? [])
    .slice()
    .sort((a, b) => a.sort - b.sort || a.code.localeCompare(b.code))
    .map((x) => ({
      code: x.code,
      label: x.label,
      venueType: x.venue_type,
      sort: x.sort,
    }))
  return {
    disciplineSelectOptions: d.disciplines
      .slice()
      .sort((a, b) => a.sort - b.sort || a.code.localeCompare(b.code))
      .map((x) => ({ value: x.code, label: x.label })),
    venueOptions,
    venueSelectOptions: venueOptions.map((x) => ({ value: x.code, label: x.label })),
    intensityOptions: d.execution_intensities.map((x) => ({ value: x.code, label: x.label })),
    auditOptions: d.audit_levels.map((x) => ({ value: x.code, label: x.label })),
    literatureSources,
    defaultLiteratureSourceCodes: defaultLiteratureSourceCodesFromOptions(literatureSources),
    defaultIntensityCode: pickDefaultCode(d.execution_intensities),
    defaultAuditLevelCode: pickDefaultCode(d.audit_levels),
  }
}
