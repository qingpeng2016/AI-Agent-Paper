type ApiEnvelope<T> = {
  code: number
  message?: string
  data?: T
}

export const LITERATURE_REVIEW_STATUS_GENERATING_EXPERIMENT_PLAN = 'gen_exp_plan' as const
export const LITERATURE_REVIEW_STATUS_EXPERIMENT_PLAN_FAILED = 'gen_exp_fail' as const

export type LiteratureReviewFrom = 'upload' | 'generated'

export type PaperLiteratureReviewItem = {
  id: string
  version: number
  status: string
  from?: LiteratureReviewFrom | string
  experiment_plan_id?: string
  title?: string
  content_medium?: string
  pdf_uri?: string
  word_uri?: string
  ppt_uri?: string
  format: string
  references_medium?: string
  created_at: string
}

export type LiteratureReviewBibliographyEntry = {
  cite_key: string
  title: string
  authors: string
  year: string
  venue?: string
  citation: string
  abstract?: string
}

const LITERATURE_REVIEW_DEMO_CITE_KEYS = [
  'ref_depression_network_1',
  'ref_depression_network_2',
  'ref_depression_network_3',
  'ref_depression_network_4',
  'ref_depression_network_5',
] as const

/** 参考文献 tab：优先 references_medium；否则有综述正文时展示 demo 占位条目 */
export function literatureReviewBibliographyEntries(
  item: Pick<PaperLiteratureReviewItem, 'version' | 'content_medium' | 'references_medium'>,
): LiteratureReviewBibliographyEntry[] {
  const custom = item.references_medium?.trim()
  if (custom) {
    return custom
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((citation, idx) => {
        const cite_key = `ref_${idx + 1}`
        const title =
          citation.length > 72 ? `${citation.slice(0, 72)}…` : citation
        return {
          cite_key,
          title,
          authors: '—',
          year: '—',
          citation,
          abstract: citation,
        }
      })
  }
  if (!item.content_medium?.trim()) return []
  const count = Math.min(LITERATURE_REVIEW_DEMO_CITE_KEYS.length, 3 + (item.version % 3))
  const demoTitles = [
    '默认模式网络在抑郁症中的功能连接异常',
    '小世界性与模块化指标在抑郁脑网络中的 meta 分析',
    '结构 DTI 脑网络与 HAMD 评分的相关研究',
    '动态功能连接揭示抑郁状态的网络重组',
    '机器学习图神经网络用于抑郁分类的可解释性',
  ]
  return LITERATURE_REVIEW_DEMO_CITE_KEYS.slice(0, count).map((cite_key, i) => {
    const year = String(2018 + i)
    const title = demoTitles[i] ?? `抑郁脑网络研究 ${i + 1}`
    const authors = i % 2 === 0 ? 'Zhang, Li et al.' : 'Wang, Chen et al.'
    const venue = 'NeuroImage (Demo)'
    const citation = `${authors} (${year}). ${title}. ${venue}.`
    return {
      cite_key,
      title,
      authors,
      year,
      venue,
      citation,
      abstract: `占位摘要：该文献与综述 v${item.version} 相关，讨论抑郁障碍背景下大规模脑网络指标（cite key：${cite_key}）。接入 Corpus 后将展示完整摘要与 PDF 链接。`,
    }
  })
}

export const LITERATURE_REVIEW_LIST_PAGE_SIZE = 20

export type PaperLiteratureReviewListResponse = {
  manuscript_id: number
  items: PaperLiteratureReviewItem[]
  total?: number
  page?: number
  page_size?: number
}

export async function fetchLiteratureReviewDetail(
  manuscriptId: string,
  literatureReviewId: string,
): Promise<PaperLiteratureReviewItem> {
  const msNum = Number(manuscriptId)
  const reviewNum = Number(literatureReviewId)
  if (!Number.isFinite(msNum) || msNum <= 0 || !Number.isFinite(reviewNum) || reviewNum <= 0) {
    throw new Error('invalid ids')
  }
  const q = new URLSearchParams({
    manuscript_id: String(msNum),
    literature_review_id: String(reviewNum),
  })
  const res = await fetch(`/api/v1/paper/literature-reviews/detail?${q}`, { credentials: 'include' })
  if (!res.ok) {
    throw new Error(`literature-review detail http ${res.status}`)
  }
  const envelope = (await res.json()) as ApiEnvelope<PaperLiteratureReviewItem>
  if (envelope.code !== 200 || !envelope.data) {
    throw new Error(envelope.message ?? 'literature-review detail failed')
  }
  return envelope.data
}

export async function fetchLiteratureReviews(
  manuscriptId: string,
  page = 1,
  pageSize = LITERATURE_REVIEW_LIST_PAGE_SIZE,
): Promise<PaperLiteratureReviewListResponse> {
  const idNum = Number(manuscriptId)
  if (!Number.isFinite(idNum) || idNum <= 0) {
    return { manuscript_id: 0, items: [], total: 0, page: 1, page_size: pageSize }
  }
  const q = new URLSearchParams({
    manuscript_id: String(idNum),
    page: String(page > 0 ? page : 1),
    page_size: String(pageSize > 0 ? pageSize : LITERATURE_REVIEW_LIST_PAGE_SIZE),
  })
  const res = await fetch(`/api/v1/paper/literature-reviews?${q}`, { credentials: 'include' })
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

export function literatureReviewHasAttachmentUri(uri?: string): boolean {
  return Boolean(uri?.trim())
}

/** 文献综述 tab：PDF、Word 路径均存在时才显示「下载」 */
export function literatureReviewCanDownloadDocuments(
  item: Pick<PaperLiteratureReviewItem, 'pdf_uri' | 'word_uri'>,
): boolean {
  return (
    literatureReviewHasAttachmentUri(item.pdf_uri) &&
    literatureReviewHasAttachmentUri(item.word_uri)
  )
}

export function literatureReviewCanDownloadPpt(
  item: Pick<PaperLiteratureReviewItem, 'ppt_uri'>,
): boolean {
  return literatureReviewHasAttachmentUri(item.ppt_uri)
}

export function formatLiteratureReviewFrom(from?: string): string {
  if (from === 'upload') return '上传'
  return '生成'
}

export type LiteratureReviewFromTone = 'upload' | 'generated' | 'default'

export function literatureReviewFromTone(from?: string): LiteratureReviewFromTone {
  if (from === 'upload') return 'upload'
  if (from === 'generated') return 'generated'
  return 'default'
}

export function literatureReviewStorageBasename(uri?: string): string {
  const u = uri?.trim()
  if (!u) return ''
  const parts = u.split('/')
  return parts[parts.length - 1] ?? u
}

export function literatureReviewPdfFileName(item: PaperLiteratureReviewItem): string {
  return (
    literatureReviewStorageBasename(item.pdf_uri) || `文献综述_v${item.version}.pdf`
  )
}

export function literatureReviewWordFileName(item: PaperLiteratureReviewItem): string {
  return (
    literatureReviewStorageBasename(item.word_uri) || `文献综述_v${item.version}.docx`
  )
}

export function literatureReviewPptFileName(item: PaperLiteratureReviewItem): string {
  return (
    literatureReviewStorageBasename(item.ppt_uri) || `文献综述汇报_v${item.version}.pptx`
  )
}

export type LiteratureReviewAttachmentKind = 'pdf' | 'docx' | 'ppt'

export function literatureReviewAttachmentFileName(
  item: PaperLiteratureReviewItem,
  kind: LiteratureReviewAttachmentKind,
): string {
  if (kind === 'pdf') return literatureReviewPdfFileName(item)
  if (kind === 'docx') return literatureReviewWordFileName(item)
  return literatureReviewPptFileName(item)
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

export type LiteratureReviewStatusTone =
  | 'success'
  | 'draft'
  | 'progress'
  | 'danger'
  | 'muted'
  | 'default'

/** UI 状态色：completed / draft / gen_exp_plan / … */
export function literatureReviewStatusTone(status: string): LiteratureReviewStatusTone {
  if (status === 'completed' || status === 'superseded') return 'success'
  if (status === 'draft') return 'draft'
  if (status === LITERATURE_REVIEW_STATUS_GENERATING_EXPERIMENT_PLAN) return 'progress'
  if (status === LITERATURE_REVIEW_STATUS_EXPERIMENT_PLAN_FAILED) return 'danger'
  if (status === 'deleted') return 'muted'
  return 'default'
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
