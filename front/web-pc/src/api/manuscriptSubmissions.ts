import type { ManuscriptSubmissionStatus } from '@/api/manuscripts'

type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export type PaperManuscriptSubmissionItem = {
  id: string
  venue_code: string
  venue_label: string
  track: string
  round: number
  status: ManuscriptSubmissionStatus
  draft_version: number
  anonymous: boolean
  portal_url?: string
  external_id?: string
  title?: string
  abstract?: string
  main_pdf_filename?: string
  decision?: string
  comments_summary?: string
  revision_deadline?: string
  submitted_at?: string
  created_at: string
  updated_at: string
}

export type PaperManuscriptSubmissionChecklistItem = {
  id: string
  ok: boolean
  label?: string
}

export type PaperManuscriptSubmissionEvent = {
  id: string
  event_type: string
  message: string
  created_at: string
}

export type PaperManuscriptSubmissionDetail = PaperManuscriptSubmissionItem & {
  checklist?: PaperManuscriptSubmissionChecklistItem[]
  events?: PaperManuscriptSubmissionEvent[]
}

export type PaperManuscriptSubmissionListResponse = {
  manuscript_id: number
  items: PaperManuscriptSubmissionItem[]
  total?: number
  page?: number
  page_size?: number
}

export const MANUSCRIPT_SUBMISSION_LIST_PAGE_SIZE = 50

function parseManuscriptId(manuscriptId: string): number {
  const n = Number(manuscriptId)
  if (!Number.isFinite(n) || n <= 0) throw new Error('invalid manuscript_id')
  return n
}

export async function fetchManuscriptSubmissions(
  manuscriptId: string,
  page = 1,
  pageSize = MANUSCRIPT_SUBMISSION_LIST_PAGE_SIZE,
): Promise<PaperManuscriptSubmissionListResponse> {
  const ms = parseManuscriptId(manuscriptId)
  const q = new URLSearchParams({
    manuscript_id: String(ms),
    page: String(page),
    page_size: String(pageSize),
  })
  const res = await fetch(`/api/v1/paper/manuscript-submissions?${q}`, { credentials: 'include' })
  if (!res.ok) throw new Error(`manuscript-submissions list http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptSubmissionListResponse>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript-submissions list failed')
  }
  return envelope.data
}

export async function fetchManuscriptSubmissionDetail(
  manuscriptId: string,
  submissionId: string,
): Promise<PaperManuscriptSubmissionDetail> {
  const ms = parseManuscriptId(manuscriptId)
  const sub = Number(submissionId)
  if (!Number.isFinite(sub) || sub <= 0) throw new Error('invalid submission_id')
  const q = new URLSearchParams({
    manuscript_id: String(ms),
    submission_id: String(sub),
  })
  const res = await fetch(`/api/v1/paper/manuscript-submissions/detail?${q}`, { credentials: 'include' })
  if (!res.ok) throw new Error(`manuscript-submissions detail http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptSubmissionDetail>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript-submissions detail failed')
  }
  return envelope.data
}

export type CreateManuscriptSubmissionPayload = {
  venue_code: string
  venue_label: string
  track: 'main' | 'short'
  draft_version: number
  anonymous: boolean
  portal_url?: string
  external_id?: string
  title?: string
  abstract?: string
  marked_submitted: boolean
}

export async function createManuscriptSubmission(
  manuscriptId: string,
  payload: CreateManuscriptSubmissionPayload,
): Promise<PaperManuscriptSubmissionDetail> {
  const ms = parseManuscriptId(manuscriptId)
  const res = await fetch('/api/v1/paper/manuscript-submissions', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      manuscript_id: ms,
      ...payload,
    }),
  })
  if (!res.ok) throw new Error(`manuscript-submissions create http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptSubmissionDetail>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript-submissions create failed')
  }
  return envelope.data
}

export async function uploadManuscriptSubmissionMainPdf(
  manuscriptId: string,
  submissionId: string,
  file: File,
): Promise<PaperManuscriptSubmissionDetail> {
  const ms = parseManuscriptId(manuscriptId)
  const sub = Number(submissionId)
  if (!Number.isFinite(sub) || sub <= 0) throw new Error('invalid submission_id')
  const body = new FormData()
  body.append('manuscript_id', String(ms))
  body.append('submission_id', String(sub))
  body.append('file', file)
  const res = await fetch('/api/v1/paper/manuscript-submissions/upload-main-pdf', {
    method: 'POST',
    credentials: 'include',
    body,
  })
  if (!res.ok) throw new Error(`manuscript-submissions upload http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptSubmissionDetail>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript-submissions upload failed')
  }
  return envelope.data
}

export async function updateManuscriptSubmissionRevision(
  manuscriptId: string,
  submissionId: string,
  decision: string,
  comments: string,
): Promise<PaperManuscriptSubmissionDetail> {
  const ms = parseManuscriptId(manuscriptId)
  const sub = Number(submissionId)
  if (!Number.isFinite(sub) || sub <= 0) throw new Error('invalid submission_id')
  const res = await fetch('/api/v1/paper/manuscript-submissions/revision', {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      manuscript_id: ms,
      submission_id: sub,
      decision: decision.trim(),
      comments: comments.trim(),
    }),
  })
  if (!res.ok) throw new Error(`manuscript-submissions revision http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptSubmissionDetail>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript-submissions revision failed')
  }
  return envelope.data
}

export async function markManuscriptSubmissionResubmitted(
  manuscriptId: string,
  submissionId: string,
): Promise<PaperManuscriptSubmissionDetail> {
  const ms = parseManuscriptId(manuscriptId)
  const sub = Number(submissionId)
  if (!Number.isFinite(sub) || sub <= 0) throw new Error('invalid submission_id')
  const res = await fetch('/api/v1/paper/manuscript-submissions/mark-resubmitted', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      manuscript_id: ms,
      submission_id: sub,
    }),
  })
  if (!res.ok) throw new Error(`manuscript-submissions resubmit http ${res.status}`)
  const envelope = (await res.json()) as ApiEnvelope<PaperManuscriptSubmissionDetail>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'manuscript-submissions resubmit failed')
  }
  return envelope.data
}

export function submissionTrackLabel(track: string): string {
  return track === 'short' ? 'Short' : 'Main'
}

export function submissionSnapshotLabel(draftVersion: number): string {
  return draftVersion >= 1 ? `v${draftVersion}` : '—'
}
