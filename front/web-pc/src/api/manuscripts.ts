type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export type ManuscriptLanguage = 'zh' | 'en'

export const PAPER_LIST_PAGE_SIZE = 20

export type ManuscriptDraftStatus = 'writing' | 'ready' | 'frozen'

export type ManuscriptSubmissionStatus =
  | 'not_submitted'
  | 'submitted'
  | 'under_review'
  | 'revision'
  | 'rebuttal'
  | 'resubmitted'
  | 'accepted'
  | 'rejected'

export type PaperManuscriptApiItem = {
  id: string
  title: string
  status: 'active' | 'archived'
  is_current: boolean
  language: ManuscriptLanguage
  discipline_code?: string
  discipline_label?: string
  created_at?: string
  draft_version?: number
  draft_status?: ManuscriptDraftStatus
  submission_status?: ManuscriptSubmissionStatus
  submission_label?: string
  submission_deadline?: string
}

const DRAFT_STATUS_LABEL: Record<ManuscriptDraftStatus, string> = {
  writing: '撰写中',
  ready: '可投递',
  frozen: '已定稿',
}

const SUBMISSION_STATUS_LABEL: Record<ManuscriptSubmissionStatus, string> = {
  not_submitted: '未投递',
  submitted: '已提交',
  under_review: '审稿中',
  revision: '待改稿',
  rebuttal: 'Rebuttal',
  resubmitted: '已修回',
  accepted: '录用',
  rejected: '拒稿',
}

export type ManuscriptDraftStatusTone = 'default' | 'ready' | 'frozen'

export function formatManuscriptDraftStatus(
  status?: string,
  draftVersion?: number,
): {
  label: string
  statusLabel: string
  versionLabel?: string
  tone: ManuscriptDraftStatusTone
} {
  const key = (status ?? 'writing') as ManuscriptDraftStatus
  const statusLabel = DRAFT_STATUS_LABEL[key] ?? DRAFT_STATUS_LABEL.writing
  const ver = formatManuscriptDraftVersion(draftVersion)
  const versionLabel = ver !== '—' ? ver : undefined
  const label = versionLabel ? `${statusLabel} · ${versionLabel}` : statusLabel
  const tone: ManuscriptDraftStatusTone =
    key === 'frozen' ? 'frozen' : key === 'ready' ? 'ready' : 'default'
  return { label, statusLabel, versionLabel, tone }
}

/** 修回 / rebuttal 且存在截止日时用于强调样式 */
export function submissionDeadlineIsUrgent(item: PaperManuscriptApiItem): boolean {
  const st = item.submission_status
  if (st !== 'revision' && st !== 'rebuttal') return false
  return Boolean(item.submission_deadline?.trim())
}

export type ManuscriptSubmissionStatusTone = 'default' | 'progress' | 'warn' | 'success' | 'muted'

export function formatManuscriptSubmissionDisplay(item: PaperManuscriptApiItem): {
  primary: string
  secondary?: string
  tone: ManuscriptSubmissionStatusTone
} {
  const st = (item.submission_status ?? 'not_submitted') as ManuscriptSubmissionStatus
  const tone: ManuscriptSubmissionStatusTone =
    st === 'accepted'
      ? 'success'
      : st === 'not_submitted' || st === 'rejected'
        ? 'muted'
        : st === 'revision' || st === 'rebuttal'
          ? 'warn'
          : 'progress'
  const label = item.submission_label?.trim()
  if (label) {
    return {
      primary: label,
      secondary: SUBMISSION_STATUS_LABEL[st] ?? st,
      tone,
    }
  }
  return {
    primary: SUBMISSION_STATUS_LABEL[st] ?? '未投递',
    secondary: undefined,
    tone,
  }
}

export function formatSubmissionDeadline(iso?: string): string {
  if (!iso?.trim()) return ''
  try {
    return new Date(iso).toLocaleString('zh-CN', { dateStyle: 'short', timeStyle: 'short' })
  } catch {
    return iso
  }
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
  language: ManuscriptLanguage = 'en',
): Promise<PaperManuscriptListResponse> {
  const lang = language === 'zh' ? 'zh' : 'en'
  const res = await fetch('/api/v1/paper/manuscripts', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title: title.trim(),
      discipline_code: disciplineCode,
      language: lang,
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
