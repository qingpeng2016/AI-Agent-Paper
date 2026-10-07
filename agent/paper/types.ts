export type PaperModuleId =
  | 'topic-discovery'
  | 'literature-review'
  | 'experiment-planning'
  | 'auto-review'
  | 'paper-writing'
  | 'figure-generation'
  | 'manuscript-analysis'
  | 'reference-library'
  | 'my-manuscripts'
  | 'personal-center'
  /** 仅用于操作日志条目，侧栏无独立入口（默认配置在个人中心 Tab） */
  | 'environment'

export type PaperModuleMeta = {
  id: PaperModuleId
  label: string
  description: string
}

export const PAPER_MODULES: PaperModuleMeta[] = [
  {
    id: 'topic-discovery',
    label: '选题发现',
    description: '多源检索与校验入库，脑暴 idea 与新颖性检查；完成后在顶栏继续生成文献综述',
  },
  {
    id: 'literature-review',
    label: '文献综述',
    description: '整理选题产出为综述；完成后顶栏继续生成实验计划（实验规划模块）',
  },
  {
    id: 'experiment-planning',
    label: '实验规划',
    description: '设计假设、基线、指标与可执行实验步骤',
  },
  {
    id: 'auto-review',
    label: '结果审查',
    description: '写作前：审查实验方案与上传数据，模拟审稿人挑 plan / 证据硬伤',
  },
  {
    id: 'paper-writing',
    label: '论文写作',
    description: '按目标 venue 风格分节撰写与润色稿件',
  },
  {
    id: 'figure-generation',
    label: '图表管理',
    description: '上传自有 figure 资产，或基于实验数据一键生成可复现统计图',
  },
  {
    id: 'manuscript-analysis',
    label: '论文审查',
    description: '写作完成后：对全文做投稿前全面审查（贡献、实验、引用、venue fit、kill argument）',
  },
  {
    id: 'reference-library',
    label: '上传文献',
    description: '上传并管理本篇论文的 PDF、BibTeX 等文献文件',
  },
  {
    id: 'my-manuscripts',
    label: '我的论文',
    description: '切换、编辑与归档工作台下的论文项目（paper_manuscript）',
  },
  {
    id: 'personal-center',
    label: '个人中心',
    description: '我的信息、Token 操作日志与默认配置',
  },
  {
    id: 'environment',
    label: '默认配置',
    description: '默认科研偏好（入口在个人中心 · 默认配置 Tab）',
  },
]

export type PaperModuleGroupId = 'topic-experiment' | 'writing' | 'resources'

export type PaperModuleGroup = {
  id: PaperModuleGroupId
  label: string
  moduleIds: PaperModuleId[]
}

/** 侧栏二级菜单：一级 = 阶段，二级 = 具体模块 */
export const PAPER_MODULE_GROUPS: PaperModuleGroup[] = [
  {
    id: 'topic-experiment',
    label: '选题与实验',
    moduleIds: [
      'topic-discovery',
      'literature-review',
      'experiment-planning',
      'auto-review',
    ],
  },
  {
    id: 'writing',
    label: '撰写与审查',
    moduleIds: ['paper-writing', 'manuscript-analysis', 'figure-generation'],
  },
  {
    id: 'resources',
    label: '资料与设置',
    moduleIds: ['my-manuscripts', 'reference-library', 'personal-center'],
  },
]

const PAPER_MODULE_MAP = new Map(PAPER_MODULES.map((m) => [m.id, m]))

export function getPaperModuleMeta(id: PaperModuleId): PaperModuleMeta {
  const meta = PAPER_MODULE_MAP.get(id)
  if (!meta) throw new Error(`unknown paper module: ${id}`)
  return meta
}

export function findPaperModuleGroupId(moduleId: PaperModuleId): PaperModuleGroupId | undefined {
  return PAPER_MODULE_GROUPS.find((g) => g.moduleIds.includes(moduleId))?.id
}

export type UploadedReferenceKind = 'pdf' | 'bib' | 'other'

/** 对齐 paper_user_literature（本地 id 暂为字符串，接 API 后为 BIGINT） */
export type UploadedReferenceItem = {
  id: string
  fileName: string
  title: string
  uploadedAt: string
  sizeBytes: number
  kind: UploadedReferenceKind
  note?: string
}

export const REFERENCE_UPLOAD_STORAGE_KEY = 'atm:paper:reference-uploads:v1'

export const REFERENCE_UPLOAD_ACCEPT = '.pdf,.bib,.txt,.md,.json'

/** 选题发现：用户在本篇「上传文献」中上传的文件，作为模型分析语料（非第三方 API 检索） */
export const USER_LIBRARY_SOURCE_CODE = 'user_library' as const

export const TOPIC_USER_LIBRARY_SOURCE = {
  code: USER_LIBRARY_SOURCE_CODE,
  label: '上传文献',
} as const

export function getUserReferenceUploads(manuscriptId: string): UploadedReferenceItem[] {
  if (!manuscriptId) return []
  try {
    const raw = localStorage.getItem(REFERENCE_UPLOAD_STORAGE_KEY)
    if (!raw) return []
    const data = JSON.parse(raw) as { byManuscript?: Record<string, UploadedReferenceItem[]> }
    return data.byManuscript?.[manuscriptId] ?? []
  } catch {
    return []
  }
}

export function getLiteratureSourceLabel(code: string): string {
  if (code === USER_LIBRARY_SOURCE_CODE) return TOPIC_USER_LIBRARY_SOURCE.label
  return LITERATURE_SOURCE_OPTIONS.find((o) => o.code === code)?.label ?? code
}

export function inferReferenceKind(fileName: string): UploadedReferenceKind {
  const lower = fileName.toLowerCase()
  if (lower.endsWith('.pdf')) return 'pdf'
  if (lower.endsWith('.bib')) return 'bib'
  return 'other'
}

export function createUploadedReferenceFromFile(file: File): UploadedReferenceItem {
  const base = file.name.replace(/\.[^.]+$/, '')
  return {
    id: `ref-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    fileName: file.name,
    title: base || file.name,
    uploadedAt: new Date().toISOString(),
    sizeBytes: file.size,
    kind: inferReferenceKind(file.name),
  }
}

export type UploadedFigureKind = 'image' | 'pdf' | 'vector' | 'other'

/** 对齐 paper_output_figure（origin=upload） */
export type UploadedFigureItem = {
  id: string
  fileName: string
  title: string
  uploadedAt: string
  sizeBytes: number
  kind: UploadedFigureKind
  note?: string
}

export const FIGURE_UPLOAD_STORAGE_KEY = 'atm:paper:figure-uploads:v1'

export const FIGURE_UPLOAD_ACCEPT = '.png,.jpg,.jpeg,.webp,.svg,.pdf,.eps,.tif,.tiff'

export function getUserFigureUploads(manuscriptId: string): UploadedFigureItem[] {
  if (!manuscriptId) return []
  try {
    const raw = localStorage.getItem(FIGURE_UPLOAD_STORAGE_KEY)
    if (!raw) return []
    const data = JSON.parse(raw) as { byManuscript?: Record<string, UploadedFigureItem[]> }
    return data.byManuscript?.[manuscriptId] ?? []
  } catch {
    return []
  }
}

export function inferFigureKind(fileName: string): UploadedFigureKind {
  const lower = fileName.toLowerCase()
  if (/\.(png|jpe?g|webp|gif|bmp|tiff?)$/.test(lower)) return 'image'
  if (lower.endsWith('.pdf')) return 'pdf'
  if (lower.endsWith('.svg') || lower.endsWith('.eps')) return 'vector'
  return 'other'
}

export function createUploadedFigureFromFile(file: File): UploadedFigureItem {
  const base = file.name.replace(/\.[^.]+$/, '')
  return {
    id: `fig-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    fileName: file.name,
    title: base || file.name,
    uploadedAt: new Date().toISOString(),
    sizeBytes: file.size,
    kind: inferFigureKind(file.name),
  }
}

export type PaperOperationLogStatus = 'success' | 'failed' | 'cancelled'

export type PaperOperationLogEntry = {
  id: string
  occurredAt: string
  moduleId: PaperModuleId | 'system'
  moduleLabel: string
  action: string
  tokensPrompt: number
  tokensCompletion: number
  tokensTotal: number
  manuscriptId?: string
  manuscriptTitle?: string
  status: PaperOperationLogStatus
  note?: string
}

export const OPERATION_LOG_STORAGE_KEY = 'atm:paper:operation-logs:v1'

export function getOperationLogs(): PaperOperationLogEntry[] {
  try {
    const raw = localStorage.getItem(OPERATION_LOG_STORAGE_KEY)
    if (!raw) return []
    const data = JSON.parse(raw) as { entries?: PaperOperationLogEntry[] }
    const entries = data.entries ?? []
    return [...entries].sort(
      (a, b) => new Date(b.occurredAt).getTime() - new Date(a.occurredAt).getTime(),
    )
  } catch {
    return []
  }
}

export function setOperationLogs(entries: PaperOperationLogEntry[]) {
  localStorage.setItem(OPERATION_LOG_STORAGE_KEY, JSON.stringify({ entries }))
}

export function appendOperationLog(
  entry: Omit<PaperOperationLogEntry, 'id' | 'occurredAt' | 'tokensTotal'> & {
    id?: string
    occurredAt?: string
    tokensTotal?: number
  },
) {
  const tokensTotal =
    entry.tokensTotal ?? Math.max(0, entry.tokensPrompt + entry.tokensCompletion)
  const row: PaperOperationLogEntry = {
    id: entry.id ?? `op-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    occurredAt: entry.occurredAt ?? new Date().toISOString(),
    moduleId: entry.moduleId,
    moduleLabel: entry.moduleLabel,
    action: entry.action,
    tokensPrompt: entry.tokensPrompt,
    tokensCompletion: entry.tokensCompletion,
    tokensTotal,
    manuscriptId: entry.manuscriptId,
    manuscriptTitle: entry.manuscriptTitle,
    status: entry.status,
    note: entry.note,
  }
  const prev = getOperationLogs()
  setOperationLogs([row, ...prev])
  return row
}

/** 工作台「当前论文」（对应 paper_manuscript） */
export type PaperManuscriptItem = {
  id: string
  title: string
  venueHint: string
  status: 'active' | 'archived'
}

export const DEMO_PAPER_MANUSCRIPTS: PaperManuscriptItem[] = [
  {
    id: 'ms-demo-1',
    title: '稀疏注意力机制',
    venueHint: 'NeurIPS 2026',
    status: 'active',
  },
  {
    id: 'ms-demo-2',
    title: '多模态 RAG 审计',
    venueHint: 'ACL 2026',
    status: 'active',
  },
]

export type ExecutionIntensity = 'fast' | 'balanced' | 'deep'
export type AuditLevel = 'standard' | 'polished' | 'strict'

export type TopicDiscoveryForm = {
  /** 对齐 paper_discipline，影响检索策略与 Prompt 模板 */
  disciplineCode: string
  direction: string
  venue: string
  /** 多源检索（写入 paper_manuscript_literature_hit） */
  sourceCodes: string[]
  intensity: ExecutionIntensity
  auditLevel: AuditLevel
  humanCheckpoint: boolean
}

/** 选题发现模块内阶段（产品/Agent 约定 stage_code） */
export type TopicCheckpointKey = 'ideas_ready'

export type TopicFlowStepDef = {
  stageCode: string
  label: string
  checkpointKey?: TopicCheckpointKey
}

export const TOPIC_DISCOVERY_FLOW_STEPS: TopicFlowStepDef[] = [
  { stageCode: 'retrieve', label: '多源文献检索与校验入库' },
  { stageCode: 'generate_ideas', label: '脑暴候选选题' },
  { stageCode: 'novelty', label: '新颖性检查', checkpointKey: 'ideas_ready' },
  { stageCode: 'audit', label: '选题断言初 audit' },
]

export type TopicFlowStepStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'waiting_human'
  | 'skipped'
  | 'failed'

export type LiteratureReviewStructure = 'thematic' | 'chronological' | 'method'

/** 文献综述模块消费的选题发现 run 快照（只读输入） */
export type TopicDiscoveryArtifactSnapshot = {
  runStatus: 'idle' | 'running' | 'checkpoint' | 'completed' | 'failed'
  /** retrieve 已完成，语料已入库 */
  corpusReady: boolean
  /** 选题发现已完成（至新颖性 + audit），可生成文献综述 */
  runCompleted: boolean
  disciplineLabel: string
  direction: string
  venue: string
  sourceLabels: string[]
  literatureHitCount: number
  verifiedHitCount: number
  candidateIdeas: string[]
  noveltyLines: string[]
  experimentPlanLines: string[]
}

export const EMPTY_TOPIC_DISCOVERY_ARTIFACT: TopicDiscoveryArtifactSnapshot = {
  runStatus: 'idle',
  corpusReady: false,
  runCompleted: false,
  disciplineLabel: '—',
  direction: '（尚未运行选题发现）',
  venue: '—',
  sourceLabels: [],
  literatureHitCount: 0,
  verifiedHitCount: 0,
  candidateIdeas: [],
  noveltyLines: [],
  experimentPlanLines: [],
}

export type LiteratureReviewForm = {
  /** 如何把选题产出 + 入库文献排成一篇综述 */
  structure: LiteratureReviewStructure
  includeFieldSurvey: boolean
  includeIdeaAndNovelty: boolean
  includeExperimentContext: boolean
  includeGap: boolean
  intensity: ExecutionIntensity
  auditLevel: AuditLevel
}

export type ExperimentPlanningForm = {
  ideaSummary: string
  venue: string
  baselinesText: string
  resources: string
  intensity: ExecutionIntensity
}

export type AutoReviewForm = {
  reviewFocus: 'plan' | 'data' | 'both'
  dataFileLabel: string
  strictKill: boolean
  auditLevel: AuditLevel
}

export type PaperWritingForm = {
  venue: string
  sections: string[]
  useReviewArtifact: boolean
  useLitReview: boolean
  tone: 'concise' | 'standard'
}

export type FigureGenerationForm = {
  datasetLabel: string
  chartTypes: string[]
  includeErrorBars: boolean
  captionLang: 'zh' | 'en'
}

export type ManuscriptAnalysisForm = {
  depth: AuditLevel
  includeFigures: boolean
  includeBib: boolean
}

export const DEFAULT_LITERATURE_REVIEW: LiteratureReviewForm = {
  structure: 'thematic',
  includeFieldSurvey: true,
  includeIdeaAndNovelty: true,
  includeExperimentContext: true,
  includeGap: true,
  intensity: 'balanced',
  auditLevel: 'polished',
}

export const DEFAULT_EXPERIMENT_PLANNING: ExperimentPlanningForm = {
  ideaSummary: '训练-free 动态稀疏路由，固定 KV 预算 B',
  venue: 'NeurIPS/ICLR/ICML',
  baselinesText: 'Full Attention, StreamingLLM, H2O, SparseAttn',
  resources: '8×A100 80GB · 约 4 周',
  intensity: 'balanced',
}

export const DEFAULT_AUTO_REVIEW: AutoReviewForm = {
  reviewFocus: 'both',
  dataFileLabel: 'results_main.csv（已上传 · 演示）',
  strictKill: true,
  auditLevel: 'strict',
}

export const DEFAULT_PAPER_WRITING: PaperWritingForm = {
  venue: 'NeurIPS/ICLR/ICML',
  sections: ['abstract', 'intro', 'related', 'method', 'experiments'],
  useReviewArtifact: true,
  useLitReview: true,
  tone: 'standard',
}

export const DEFAULT_FIGURE_GENERATION: FigureGenerationForm = {
  datasetLabel: 'results_main.csv',
  chartTypes: ['bar', 'line'],
  includeErrorBars: true,
  captionLang: 'en',
}

export const DEFAULT_MANUSCRIPT_ANALYSIS: ManuscriptAnalysisForm = {
  depth: 'strict',
  includeFigures: true,
  includeBib: true,
}

export const PAPER_WRITING_SECTION_OPTIONS = [
  { value: 'abstract', label: 'Abstract' },
  { value: 'intro', label: 'Introduction' },
  { value: 'related', label: 'Related Work' },
  { value: 'method', label: 'Method' },
  { value: 'experiments', label: 'Experiments' },
  { value: 'conclusion', label: 'Conclusion' },
] as const

export const FIGURE_CHART_OPTIONS = [
  { value: 'bar', label: '柱状图（主结果）' },
  { value: 'line', label: '折线（训练/缩放）' },
  { value: 'scatter', label: '散点（吞吐-准确率）' },
  { value: 'ablation', label: '消融分组图' },
] as const

export const DEFAULT_TOPIC_DISCOVERY: TopicDiscoveryForm = {
  disciplineCode: 'cs_ai',
  direction: '',
  venue: 'NeurIPS/ICLR/ICML',
  sourceCodes: ['arxiv', 'openalex', 'semantic_scholar'],
  intensity: 'balanced',
  auditLevel: 'polished',
  humanCheckpoint: true,
}

export type EnvironmentPreferenceForm = {
  disciplineCode: string
  defaultVenueText: string
  literatureSourceCodes: string[]
  intensity: ExecutionIntensity
  auditLevel: AuditLevel
  humanCheckpoint: boolean
  referenceGateEnabled: boolean
}

export const LITERATURE_SOURCE_OPTIONS = [
  { code: 'arxiv', label: 'arXiv' },
  { code: 'openalex', label: 'OpenAlex' },
  { code: 'semantic_scholar', label: 'Semantic Scholar' },
  { code: 'crossref', label: 'Crossref' },
  { code: 'pubmed', label: 'PubMed' },
] as const

export const DISCIPLINE_OPTIONS = [
  { code: 'cs_ai', label: '计算机 / 人工智能' },
  { code: 'general', label: '跨学科通用' },
] as const

export const ENV_PREFERENCE_STORAGE_KEY = 'atm:paper:environment:v3'

export const DEFAULT_ENV_PREFERENCE: EnvironmentPreferenceForm = {
  disciplineCode: 'cs_ai',
  defaultVenueText: 'NeurIPS/ICLR/ICML',
  literatureSourceCodes: ['arxiv', 'openalex', 'semantic_scholar'],
  intensity: 'balanced',
  auditLevel: 'polished',
  humanCheckpoint: true,
  referenceGateEnabled: true,
}

