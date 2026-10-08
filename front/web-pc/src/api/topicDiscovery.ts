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
