<script setup lang="ts">
import { computed, onUnmounted, reactive, ref, toRefs, watch } from 'vue'
import { ElMessage, ElTooltip } from 'element-plus'
import {
  fetchLiteratureReviewDetail,
  fetchLiteratureReviews,
  LITERATURE_REVIEW_LIST_PAGE_SIZE,
  formatLiteratureReviewFrom,
  literatureReviewBibliographyEntries,
  literatureReviewCanDownloadDocuments,
  literatureReviewCanDownloadPpt,
  literatureReviewFromTone,
  literatureReviewHasAttachmentUri,
  formatLiteratureReviewStatus,
  literatureReviewAttachmentFileName,
  literatureReviewPdfFileName,
  literatureReviewPptFileName,
  literatureReviewStatusTone,
  literatureReviewWordFileName,
  isLiteratureReviewGeneratingExperimentPlan,
  literatureReviewNeedsGenerateRun,
  softDeleteLiteratureReview,
  type LitReviewAttachmentKind,
  type LiteratureReviewBibliographyEntry,
  type PaperLiteratureReviewItem,
} from '@/api/literatureReviews'
import {
  experimentPlanRowTitle,
  fetchExperimentPlanDetail,
  fetchExperimentPlans,
  EXPERIMENT_PLAN_LIST_PAGE_SIZE,
  formatExperimentPlanStatus,
  experimentPlanStatusTone,
  uploadExperimentPlanData,
  type PaperExperimentPlanItem,
} from '@/api/experimentPlans'
import { DEMO_FIGURES, DEMO_MANUSCRIPT } from './demoModuleOutputs'
import PaperFigureUploadPanel from './PaperFigureUploadPanel.vue'
import PaperListPagination from './PaperListPagination.vue'
import PaperSelect from './PaperSelect.vue'
import {
  DEFAULT_FIGURE_GENERATION,
  FIGURE_CHART_OPTIONS,
  LITERATURE_SOURCE_OPTIONS,
  type FigureGenerationForm,
  type PaperModuleId,
} from './types'
import PaperTopicUserLiteraturePickDialog from './PaperTopicUserLiteraturePickDialog.vue'

export type FigureManagementTabId = 'upload' | 'generate'

const figureTab = defineModel<FigureManagementTabId>('figureTab', { default: 'upload' })

const emit = defineEmits<{
  navigateModule: [PaperModuleId, openExperimentPlanId?: string]
  consumedPendingExperimentPlan: []
}>()

const props = withDefaults(
  defineProps<{
    moduleId: PaperModuleId
    manuscriptId?: string
    manuscriptTitle?: string
    pendingOpenExperimentPlanId?: string | null
  }>(),
  {
    manuscriptId: '',
    manuscriptTitle: '未命名',
    pendingOpenExperimentPlanId: null,
  },
)
const { moduleId, manuscriptId, manuscriptTitle } = toRefs(props)

const figureForm = reactive<FigureGenerationForm>({
  ...DEFAULT_FIGURE_GENERATION,
  chartTypes: [...DEFAULT_FIGURE_GENERATION.chartTypes],
})

const resultVisible = ref<Partial<Record<PaperModuleId, boolean>>>({})

const litReviewItems = ref<PaperLiteratureReviewItem[]>([])
const litReviewListTotal = ref(0)
const litReviewListPage = ref(1)
const litReviewsLoading = ref(false)
const litReviewViewItem = ref<PaperLiteratureReviewItem | null>(null)
const litReviewViewDownloadKind = ref<'pdf' | 'docx' | null>(null)
const litReviewAttachmentPreview = ref<{
  item: PaperLiteratureReviewItem
  kind: LitReviewAttachmentKind
} | null>(null)
const litReviewDeletePending = ref<PaperLiteratureReviewItem | null>(null)
const litReviewDeleteSubmitting = ref(false)
const litReviewUploadExpPlanOpen = ref(false)
const litReviewUploadExpPlanReviewId = ref('')
const litReviewUploadExpPlanFile = ref<File | null>(null)
const litReviewUploadExpPlanInputRef = ref<HTMLInputElement | null>(null)
const litReviewUploadExpPlanSubmitting = ref(false)
const litReviewGeneratePending = ref<PaperLiteratureReviewItem | null>(null)
const litReviewGenerateSubmitting = ref(false)
const litReviewGenerateUserLitOpen = ref(false)
const litReviewGenerateForm = reactive({
  sourceCodes: ['arxiv'] as string[],
  randomReferences: true,
  humanCheckpoint: true,
  userLiteratureIds: [] as string[],
})
let litReviewPollTimer: ReturnType<typeof setInterval> | null = null

const expPlanItems = ref<PaperExperimentPlanItem[]>([])
const expPlanListTotal = ref(0)
const expPlanListPage = ref(1)
const expPlansLoading = ref(false)
type ExpPlanBundleTabId = 'plan' | 'data' | 'code' | 'results' | 'audit'

const expPlanBundleView = ref<{
  item: PaperExperimentPlanItem
  tab: ExpPlanBundleTabId
  loading: boolean
} | null>(null)
const expPlanBundleDataFile = ref<File | null>(null)
const expPlanBundleDataInputRef = ref<HTMLInputElement | null>(null)
const expPlanBundleDataSubmitting = ref(false)
const expPlanCodeRunning = ref(false)
const expPlanAuditRunning = ref(false)
const expPlanDeletePending = ref<PaperExperimentPlanItem | null>(null)
const expPlanDeleteSubmitting = ref(false)
const expPlanUploadPlanOpen = ref(false)
const expPlanUploadPlanId = ref('')
const expPlanUploadPlanFile = ref<File | null>(null)
const expPlanUploadPlanInputRef = ref<HTMLInputElement | null>(null)
const expPlanUploadPlanSubmitting = ref(false)
let expPlanReloadSeq = 0

type PaperWritePhaseId = 'plan' | 'write' | 'compile' | 'polish' | 'audit'

const PAPER_WRITE_PHASES: { id: PaperWritePhaseId; label: string; skill: string }[] = [
  { id: 'plan', label: '规划', skill: 'paper-plan' },
  { id: 'write', label: '正文', skill: 'paper-write' },
  { id: 'compile', label: '编译', skill: 'paper-compile' },
  { id: 'polish', label: '润色', skill: 'auto-paper-improvement-loop' },
  { id: 'audit', label: '审计', skill: 'paper-claim-audit' },
]

const PAPER_WRITE_OUTLINE_ROWS: {
  section: string
  claim: string
  figures: string
  pages: string
  status: string
  statusTone: 'success' | 'draft' | 'muted'
}[] = [
  {
    section: 'Abstract',
    claim: 'C1 主贡献',
    figures: '—',
    pages: '0.5',
    status: '已定稿',
    statusTone: 'success',
  },
  {
    section: 'Introduction',
    claim: 'C1 + 动机',
    figures: 'Fig.1 总览',
    pages: '1.5',
    status: '草稿',
    statusTone: 'draft',
  },
  {
    section: 'Method',
    claim: 'C1 机制',
    figures: 'Fig.2 架构',
    pages: '2.0',
    status: '待写',
    statusTone: 'muted',
  },
  {
    section: 'Experiments',
    claim: 'C1 证据',
    figures: 'Fig.3–5',
    pages: '3.0',
    status: '待写',
    statusTone: 'muted',
  },
]

const PAPER_WRITE_CLAIM_AUDIT_ROWS = [
  { claim: '准确率 85.3%', evidence: 'results/main.json', verdict: '待审计' },
  { claim: '较基线 +12.8%', evidence: 'tables/ablation.csv', verdict: '待审计' },
] as const

const paperWritePhase = ref<PaperWritePhaseId>('plan')
const paperWriteRunStatus = ref<'idle' | 'running'>('idle')
const paperWriteSelectedSection = ref('Introduction')
const paperWriteCompileAnonymous = ref(true)
const paperWriteDepsLoading = ref(false)

const paperWriteVenueLabel = computed(() => 'NeurIPS / ICLR / ICML')

const paperWriteLitGateOk = computed(() =>
  litReviewItems.value.some((it) => literatureReviewCanDownloadDocuments(it)),
)

const paperWriteExpGateOk = computed(() => expPlanListTotal.value > 0)

const paperWriteFigureReadyCount = computed(() => DEMO_FIGURES.files.length)
const paperWriteFigureTargetCount = 6
const paperWriteFigureGateOk = computed(
  () => paperWriteFigureReadyCount.value >= paperWriteFigureTargetCount,
)

const paperWritePhasePrimaryLabel = computed(() => {
  switch (paperWritePhase.value) {
    case 'plan':
      return '生成大纲'
    case 'write':
      return '续写下一节'
    case 'compile':
      return '编译 PDF'
    case 'polish':
      return '开始润色'
    case 'audit':
      return '开始审计'
    default:
      return '运行'
  }
})

const paperWritePhaseHeading = computed(() => {
  switch (paperWritePhase.value) {
    case 'plan':
      return '规划 · PAPER_PLAN.md'
    case 'write':
      return '正文 · LaTeX 分节'
    case 'compile':
      return '编译 · PDF'
    case 'polish':
      return '润色 · 外审循环'
    case 'audit':
      return '审计 · Claim 核对'
    default:
      return '论文写作'
  }
})

const paperWriteSectionsDone = computed(
  () => DEMO_MANUSCRIPT.sections.filter((s) => s.status === 'done').length,
)

const paperWriteSectionPreview = computed(() => {
  if (paperWriteSelectedSection.value === 'Introduction') {
    return DEMO_MANUSCRIPT.excerpt
  }
  const name = paperWriteSelectedSection.value
  return `\\section{${name}}\n\n（接入 Agent 后将加载对应 .tex 片段）`
})

async function reloadPaperWriteDependencies() {
  const ms = manuscriptId.value
  if (!ms || !/^\d+$/.test(ms)) return
  paperWriteDepsLoading.value = true
  try {
    await Promise.all([
      reloadLiteratureReviews({ silent: true }),
      reloadExperimentPlans({ silent: true, page: 1 }),
    ])
  } finally {
    paperWriteDepsLoading.value = false
  }
}

function setPaperWritePhase(id: PaperWritePhaseId) {
  paperWritePhase.value = id
}

function runPaperWritePhasePrimary() {
  ElMessage.info(`${paperWritePhasePrimaryLabel.value}（接口接入后替换）`)
}

async function runPaperWriteFullPipeline() {
  if (paperWriteRunStatus.value === 'running') return
  paperWriteRunStatus.value = 'running'
  try {
    await reloadPaperWriteDependencies()
    await new Promise((r) => setTimeout(r, 600))
    resultVisible.value = { ...resultVisible.value, 'paper-writing': true }
    ElMessage.success('已触发写作全流程（演示占位）')
  } finally {
    paperWriteRunStatus.value = 'idle'
  }
}

function goToFigureManagement() {
  emit('navigateModule', 'figure-generation')
}

function goToLiteratureReviewModule() {
  emit('navigateModule', 'literature-review')
}

function goToExperimentPlanningModule() {
  emit('navigateModule', 'experiment-planning')
}

watch(
  () => [moduleId.value, manuscriptId.value] as const,
  ([mod, ms]) => {
    if (mod !== 'paper-writing') return
    if (!ms || !/^\d+$/.test(ms)) return
    void reloadPaperWriteDependencies()
  },
  { immediate: true },
)

const EXP_PLAN_BUNDLE_TABS: { id: ExpPlanBundleTabId; label: string }[] = [
  { id: 'plan', label: '方案' },
  { id: 'data', label: '数据' },
  { id: 'code', label: '代码' },
  { id: 'results', label: '结果' },
  { id: 'audit', label: '审计' },
]

const EXP_PLAN_CODE_PLACEHOLDER = `# 实验可复现脚本（接入 storage_uri 后将加载真实代码）
# 原始/输入数据请在「数据」Tab 上传

def main():
    print("实验流水线占位：运行接口接入后在此执行")

if __name__ == "__main__":
    main()
`

let litReviewReloadSeq = 0

async function reloadLiteratureReviews(opts?: { silent?: boolean; page?: number }) {
  const silent = opts?.silent === true
  const ms = manuscriptId.value
  if (!ms || !/^\d+$/.test(ms)) {
    if (!silent) {
      litReviewItems.value = []
      litReviewListTotal.value = 0
    }
    return
  }
  const page = opts?.page ?? litReviewListPage.value
  const seq = ++litReviewReloadSeq
  if (!silent) litReviewsLoading.value = true
  try {
    const data = await fetchLiteratureReviews(ms, page, LITERATURE_REVIEW_LIST_PAGE_SIZE)
    if (seq !== litReviewReloadSeq) return
    litReviewItems.value = data.items ?? []
    litReviewListTotal.value = data.total ?? litReviewItems.value.length
    litReviewListPage.value = data.page ?? page
    if (
      litReviewItems.value.length === 0 &&
      litReviewListTotal.value > 0 &&
      litReviewListPage.value > 1
    ) {
      litReviewListPage.value = Math.max(1, litReviewListPage.value - 1)
      await reloadLiteratureReviews({ silent, page: litReviewListPage.value })
      return
    }
  } catch (e) {
    if (seq !== litReviewReloadSeq) return
    if (!silent) {
      litReviewItems.value = []
      litReviewListTotal.value = 0
      const msg = e instanceof Error ? e.message : '加载文献综述失败'
      ElMessage.error(msg)
    }
  } finally {
    if (seq === litReviewReloadSeq) {
      if (!silent) litReviewsLoading.value = false
      syncLitReviewPollTimer()
    }
  }
}

function onLitReviewListPageChange(page: number) {
  litReviewListPage.value = page
  void reloadLiteratureReviews({ page })
}

/** 进入文献综述 / 切换当前论文 / 刷新：同一条路径拉列表（不依赖父组件 ref 时序） */
watch(
  () => [moduleId.value, manuscriptId.value] as const,
  ([mod, ms]) => {
    if (mod !== 'literature-review') return
    if (!ms || !/^\d+$/.test(ms)) {
      litReviewItems.value = []
      litReviewListTotal.value = 0
      litReviewListPage.value = 1
      return
    }
    litReviewListPage.value = 1
    void reloadLiteratureReviews({ page: 1 })
  },
  { immediate: true },
)

function litReviewHasGeneratingExperimentPlan(): boolean {
  return litReviewItems.value.some((it) => isLiteratureReviewGeneratingExperimentPlan(it.status))
}

function syncLitReviewPollTimer() {
  if (moduleId.value !== 'literature-review') {
    if (litReviewPollTimer) {
      clearInterval(litReviewPollTimer)
      litReviewPollTimer = null
    }
    return
  }
  if (litReviewHasGeneratingExperimentPlan()) {
    if (!litReviewPollTimer) {
      litReviewPollTimer = setInterval(() => {
        void reloadLiteratureReviews({ silent: true })
      }, 4000)
    }
  } else if (litReviewPollTimer) {
    clearInterval(litReviewPollTimer)
    litReviewPollTimer = null
  }
}

watch(litReviewItems, () => syncLitReviewPollTimer(), { deep: true })
watch(moduleId, () => syncLitReviewPollTimer())

onUnmounted(() => {
  if (litReviewPollTimer) {
    clearInterval(litReviewPollTimer)
    litReviewPollTimer = null
  }
})

type LitReviewBundleTabId = 'literature-review' | 'references' | 'ppt'

const LIT_REVIEW_BUNDLE_TABS: { id: LitReviewBundleTabId; label: string }[] = [
  { id: 'literature-review', label: '文献综述' },
  { id: 'ppt', label: '汇报 PPT' },
  { id: 'references', label: '参考文献' },
]

const litReviewBundleView = ref<{
  item: PaperLiteratureReviewItem
  tab: LitReviewBundleTabId
} | null>(null)

const litReviewBundleBibliography = computed(() => {
  const ctx = litReviewBundleView.value
  if (!ctx) return []
  return literatureReviewBibliographyEntries(ctx.item)
})

const litReviewRefDetail = ref<LiteratureReviewBibliographyEntry | null>(null)

function openLitReviewRefDetail(entry: LiteratureReviewBibliographyEntry) {
  litReviewRefDetail.value = entry
}

function closeLitReviewRefDetail() {
  litReviewRefDetail.value = null
}

async function ensureLitReviewDetail(item: PaperLiteratureReviewItem): Promise<PaperLiteratureReviewItem> {
  const ms = manuscriptId.value
  if (item.content_medium?.trim() || !ms || !/^\d+$/.test(ms)) return item
  try {
    return await fetchLiteratureReviewDetail(ms, item.id)
  } catch {
    return item
  }
}

function resetLitReviewGenerateForm() {
  litReviewGenerateForm.sourceCodes = ['arxiv']
  litReviewGenerateForm.randomReferences = true
  litReviewGenerateForm.humanCheckpoint = true
  litReviewGenerateForm.userLiteratureIds = []
}

function isLitReviewGenerateSourceChecked(code: string): boolean {
  return litReviewGenerateForm.sourceCodes.includes(code)
}

function toggleLitReviewGenerateSource(code: string, checked: boolean) {
  const set = new Set(litReviewGenerateForm.sourceCodes)
  if (checked) set.add(code)
  else set.delete(code)
  litReviewGenerateForm.sourceCodes = [...set]
}

function openLitReviewGenerateDialog(item: PaperLiteratureReviewItem) {
  resetLitReviewGenerateForm()
  litReviewGeneratePending.value = item
}

function closeLitReviewGenerateDialog() {
  if (litReviewGenerateSubmitting.value) return
  litReviewGeneratePending.value = null
  litReviewGenerateUserLitOpen.value = false
}

function openLitReviewGenerateUserLiteraturePicker() {
  if (litReviewGenerateSubmitting.value) return
  litReviewGenerateUserLitOpen.value = true
}

function closeLitReviewGenerateUserLiteraturePicker() {
  litReviewGenerateUserLitOpen.value = false
}

function onLitReviewGenerateUserLiteraturePicked(ids: string[]) {
  litReviewGenerateForm.userLiteratureIds = [...ids]
  litReviewGenerateUserLitOpen.value = false
}

async function submitLitReviewGenerateRun() {
  if (!litReviewGeneratePending.value) return
  if (!litReviewGenerateForm.sourceCodes.length) {
    ElMessage.warning('请至少选择一个文献来源')
    return
  }
  litReviewGenerateSubmitting.value = true
  try {
    await new Promise((r) => setTimeout(r, 400))
    ElMessage.info('文献综述生成流程即将接入')
    closeLitReviewGenerateDialog()
    void reloadLiteratureReviews({ silent: true })
  } finally {
    litReviewGenerateSubmitting.value = false
  }
}

function onLitReviewRowPrimaryAction(item: PaperLiteratureReviewItem) {
  if (literatureReviewNeedsGenerateRun(item)) {
    openLitReviewGenerateDialog(item)
    return
  }
  void openLitReviewBundleView(item)
}

async function openLitReviewBundleView(
  item: PaperLiteratureReviewItem,
  tab: LitReviewBundleTabId = 'literature-review',
) {
  const review = await ensureLitReviewDetail(item)
  litReviewBundleView.value = {
    item: review,
    tab,
  }
}

function setLitReviewBundleTab(tab: LitReviewBundleTabId) {
  const ctx = litReviewBundleView.value
  if (!ctx) return
  litReviewBundleView.value = { ...ctx, tab }
}

function closeLitReviewBundleView() {
  litReviewRefDetail.value = null
  litReviewBundleView.value = null
}

function downloadLitReviewBundlePpt() {
  const ctx = litReviewBundleView.value
  if (!ctx) return
  const name = literatureReviewPptFileName(ctx.item)
  ElMessage.info(`下载「${name}」功能即将接入`)
}

function downloadLitReviewBundleLiteratureReview() {
  const ctx = litReviewBundleView.value
  if (!ctx) return
  const pdf = literatureReviewPdfFileName(ctx.item)
  const docx = literatureReviewWordFileName(ctx.item)
  ElMessage.info(`下载「${pdf}」/「${docx}」功能即将接入`)
}

function litReviewBundleHasDownloadAction(tab: LitReviewBundleTabId): boolean {
  const ctx = litReviewBundleView.value
  if (!ctx) return false
  if (tab === 'literature-review') return literatureReviewCanDownloadDocuments(ctx.item)
  if (tab === 'ppt') return literatureReviewCanDownloadPpt(ctx.item)
  return false
}

const litReviewUploadExpPlanOptions = computed(() =>
  litReviewItems.value.map((item) => ({
    value: item.id,
    label: litReviewRowTitle(item),
  })),
)

function resetLitReviewUploadExpPlanInput() {
  litReviewUploadExpPlanFile.value = null
  if (litReviewUploadExpPlanInputRef.value) litReviewUploadExpPlanInputRef.value.value = ''
}

function openLitReviewUploadExpPlanDialog() {
  if (!litReviewItems.value.length) {
    ElMessage.warning('暂无文献综述')
    return
  }
  litReviewUploadExpPlanReviewId.value = litReviewItems.value[0]?.id ?? ''
  resetLitReviewUploadExpPlanInput()
  litReviewUploadExpPlanOpen.value = true
}

function closeLitReviewUploadExpPlanDialog() {
  if (litReviewUploadExpPlanSubmitting.value) return
  litReviewUploadExpPlanOpen.value = false
  resetLitReviewUploadExpPlanInput()
}

function openLitReviewUploadExpPlanPicker() {
  if (litReviewUploadExpPlanSubmitting.value) return
  litReviewUploadExpPlanInputRef.value?.click()
}

function onLitReviewUploadExpPlanFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  litReviewUploadExpPlanFile.value = input.files?.[0] ?? null
}

function onLitReviewUploadExpPlanDrop(ev: DragEvent) {
  if (litReviewUploadExpPlanSubmitting.value) return
  const file = ev.dataTransfer?.files?.[0]
  if (file) litReviewUploadExpPlanFile.value = file
}

async function submitLitReviewUploadExpPlan() {
  const reviewId = litReviewUploadExpPlanReviewId.value.trim()
  const file = litReviewUploadExpPlanFile.value
  const ms = manuscriptId.value
  if (!reviewId || !file || !ms) {
    ElMessage.warning('请选择关联综述与文献综述文件')
    return
  }
  litReviewUploadExpPlanSubmitting.value = true
  try {
    await new Promise((r) => setTimeout(r, 400))
    ElMessage.info(`上传文献综述「${file.name}」功能即将接入`)
    closeLitReviewUploadExpPlanDialog()
  } finally {
    litReviewUploadExpPlanSubmitting.value = false
  }
}

async function tryOpenPendingExperimentPlan() {
  const planId = props.pendingOpenExperimentPlanId?.trim()
  if (moduleId.value !== 'experiment-planning' || !planId) return

  let item = expPlanItems.value.find((it) => it.id === planId)
  if (!item) {
    await reloadExperimentPlans({ silent: true })
    item = expPlanItems.value.find((it) => it.id === planId)
  }
  if (item) {
    void openExpPlanBundleView(item)
    emit('consumedPendingExperimentPlan')
    return
  }

  const ms = manuscriptId.value
  if (!ms) return
  try {
    const detail = await fetchExperimentPlanDetail(ms, planId)
    void openExpPlanBundleView(detail)
    emit('consumedPendingExperimentPlan')
  } catch (e) {
    const msg = e instanceof Error ? e.message : '加载实验方案失败'
    ElMessage.error(msg)
    emit('consumedPendingExperimentPlan')
  }
}

async function reloadExperimentPlans(opts?: { silent?: boolean; page?: number }) {
  const silent = opts?.silent === true
  const ms = manuscriptId.value
  if (!ms || !/^\d+$/.test(ms)) {
    if (!silent) {
      expPlanItems.value = []
      expPlanListTotal.value = 0
    }
    return
  }
  const page = opts?.page ?? expPlanListPage.value
  const seq = ++expPlanReloadSeq
  if (!silent) expPlansLoading.value = true
  try {
    const data = await fetchExperimentPlans(ms, page, EXPERIMENT_PLAN_LIST_PAGE_SIZE)
    if (seq !== expPlanReloadSeq) return
    expPlanItems.value = data.items ?? []
    expPlanListTotal.value = data.total ?? expPlanItems.value.length
    expPlanListPage.value = data.page ?? page
    if (
      expPlanItems.value.length === 0 &&
      expPlanListTotal.value > 0 &&
      expPlanListPage.value > 1
    ) {
      expPlanListPage.value = Math.max(1, expPlanListPage.value - 1)
      await reloadExperimentPlans({ silent, page: expPlanListPage.value })
      return
    }
  } catch (e) {
    if (seq !== expPlanReloadSeq) return
    if (!silent) {
      expPlanItems.value = []
      expPlanListTotal.value = 0
      const msg = e instanceof Error ? e.message : '加载实验方案失败'
      ElMessage.error(msg)
    }
  } finally {
    if (seq === expPlanReloadSeq) {
      if (!silent) expPlansLoading.value = false
      void tryOpenPendingExperimentPlan()
    }
  }
}

function onExpPlanListPageChange(page: number) {
  expPlanListPage.value = page
  void reloadExperimentPlans({ page })
}

/** 进入实验方案 / 切换当前论文 / 刷新：同文献综述，模块内 loading（不走顶栏 generic loading） */
watch(
  () => [moduleId.value, manuscriptId.value] as const,
  ([mod, ms]) => {
    if (mod !== 'experiment-planning') return
    if (!ms || !/^\d+$/.test(ms)) {
      expPlanItems.value = []
      expPlanListTotal.value = 0
      expPlanListPage.value = 1
      return
    }
    expPlanListPage.value = 1
    void reloadExperimentPlans({ page: 1 })
  },
  { immediate: true },
)

watch(
  () => props.pendingOpenExperimentPlanId,
  () => {
    void tryOpenPendingExperimentPlan()
  },
)

const figureDatasetOptions = [
  { value: 'results_main.csv', label: 'results_main.csv（演示）' },
  { value: 'ablation_B.csv', label: 'ablation_B.csv（演示）' },
  { value: 'throughput.csv', label: 'throughput.csv（演示）' },
]

const captionLangOptions = [
  { value: 'en', label: 'English' },
  { value: 'zh', label: '中文' },
]

function toggleChartType(value: string, checked: boolean) {
  const set = new Set(figureForm.chartTypes)
  if (checked) set.add(value)
  else set.delete(value)
  figureForm.chartTypes = [...set]
}

function isChartChecked(value: string) {
  return figureForm.chartTypes.includes(value)
}

function litReviewRowTitle(item: PaperLiteratureReviewItem): string {
  const t = item.title?.trim()
  if (t) return t
  return `文献综述 v${item.version}`
}

function formatLitReviewCreatedAt(iso: string): string {
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

function openLitReviewView(item: PaperLiteratureReviewItem) {
  litReviewViewDownloadKind.value = null
  litReviewViewItem.value = item
}

function closeLitReviewView() {
  litReviewViewItem.value = null
  litReviewViewDownloadKind.value = null
}

async function openLitReviewViewWithDownload(item: PaperLiteratureReviewItem, kind: 'pdf' | 'docx') {
  litReviewViewDownloadKind.value = kind
  const ms = manuscriptId.value
  if (!item.content_medium?.trim() && ms && /^\d+$/.test(ms)) {
    try {
      litReviewViewItem.value = await fetchLiteratureReviewDetail(ms, item.id)
      return
    } catch (e) {
      const msg = e instanceof Error ? e.message : '加载文献综述失败'
      ElMessage.error(msg)
    }
  }
  litReviewViewItem.value = item
}

function litReviewAttachmentKindLabel(kind: LitReviewAttachmentKind): string {
  if (kind === 'ppt') return '汇报 PPT'
  if (kind === 'pdf') return 'PDF'
  return 'Word'
}

async function openLitReviewAttachmentPreview(
  item: PaperLiteratureReviewItem,
  kind: LitReviewAttachmentKind,
) {
  if (kind === 'pdf' || kind === 'docx') {
    await openLitReviewViewWithDownload(item, kind)
    return
  }
  litReviewAttachmentPreview.value = { item, kind }
}

function closeLitReviewAttachmentPreview() {
  litReviewAttachmentPreview.value = null
}

function litReviewViewDownloadLabel(
  item: PaperLiteratureReviewItem,
  kind: 'pdf' | 'docx',
): string {
  return kind === 'pdf' ? '下载 PDF' : '下载 Word'
}

function litReviewViewCanDownloadKind(
  item: PaperLiteratureReviewItem,
  kind: 'pdf' | 'docx',
): boolean {
  const uri = kind === 'pdf' ? item.pdf_uri : item.word_uri
  return literatureReviewHasAttachmentUri(uri)
}

function onLitReviewViewPrimaryAction() {
  const item = litReviewViewItem.value
  const kind = litReviewViewDownloadKind.value
  if (!item || !kind || !litReviewViewCanDownloadKind(item, kind)) return
  const name = literatureReviewAttachmentFileName(item, kind)
  ElMessage.info(`下载「${name}」功能即将接入`)
}

function onLitReviewAttachmentPreviewPrimary() {
  const ctx = litReviewAttachmentPreview.value
  if (!ctx || !literatureReviewCanDownloadPpt(ctx.item)) return
  const name = literatureReviewAttachmentFileName(ctx.item, ctx.kind)
  ElMessage.info(`下载「${name}」功能即将接入`)
}

function patchExpPlanItemInList(updated: PaperExperimentPlanItem) {
  const idx = expPlanItems.value.findIndex((it) => it.id === updated.id)
  if (idx >= 0) {
    expPlanItems.value = [
      ...expPlanItems.value.slice(0, idx),
      updated,
      ...expPlanItems.value.slice(idx + 1),
    ]
  }
  const ctx = expPlanBundleView.value
  if (ctx?.item.id === updated.id) {
    expPlanBundleView.value = { ...ctx, item: updated, loading: false }
  }
}

function resetExpPlanBundleDataInput() {
  expPlanBundleDataFile.value = null
  if (expPlanBundleDataInputRef.value) expPlanBundleDataInputRef.value.value = ''
}

async function openExpPlanBundleView(
  item: PaperExperimentPlanItem,
  tab: ExpPlanBundleTabId = 'plan',
) {
  resetExpPlanBundleDataInput()
  const ms = manuscriptId.value
  const needsDetail =
    (!item.content_medium?.trim() && !item.summary?.trim()) && ms && /^\d+$/.test(ms)
  expPlanBundleView.value = { item: { ...item }, tab, loading: needsDetail }
  if (!needsDetail || !ms) return
  try {
    const detail = await fetchExperimentPlanDetail(ms, item.id)
    const cur = expPlanBundleView.value
    if (!cur || cur.item.id !== item.id) return
    expPlanBundleView.value = { ...cur, item: detail, loading: false }
  } catch (e) {
    const msg = e instanceof Error ? e.message : '加载实验方案失败'
    ElMessage.error(msg)
    const cur = expPlanBundleView.value
    if (cur?.item.id === item.id) {
      expPlanBundleView.value = { ...cur, loading: false }
    }
  }
}

function setExpPlanBundleTab(tab: ExpPlanBundleTabId) {
  const ctx = expPlanBundleView.value
  if (!ctx) return
  expPlanBundleView.value = { ...ctx, tab }
}

function closeExpPlanBundleView() {
  if (expPlanBundleDataSubmitting.value) return
  expPlanBundleView.value = null
  resetExpPlanBundleDataInput()
}

function openExpPlanBundleDataPicker() {
  if (expPlanBundleDataSubmitting.value) return
  expPlanBundleDataInputRef.value?.click()
}

function onExpPlanBundleDataFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  expPlanBundleDataFile.value = input.files?.[0] ?? null
}

function onExpPlanBundleDataDrop(ev: DragEvent) {
  if (expPlanBundleDataSubmitting.value) return
  const file = ev.dataTransfer?.files?.[0]
  if (file) expPlanBundleDataFile.value = file
}

async function submitExpPlanBundleDataUpload() {
  const ctx = expPlanBundleView.value
  const file = expPlanBundleDataFile.value
  const ms = manuscriptId.value
  if (!ctx || !file || !ms) {
    ElMessage.warning('请选择实验数据文件')
    return
  }
  expPlanBundleDataSubmitting.value = true
  try {
    const updated = await uploadExperimentPlanData(ms, ctx.item.id, file)
    ElMessage.success('实验数据已上传')
    patchExpPlanItemInList(updated)
    resetExpPlanBundleDataInput()
  } catch (e) {
    const msg = e instanceof Error ? e.message : '上传失败'
    ElMessage.error(msg)
  } finally {
    expPlanBundleDataSubmitting.value = false
  }
}

async function runExpPlanBundleCode() {
  if (!expPlanBundleView.value) return
  expPlanCodeRunning.value = true
  try {
    await new Promise((r) => setTimeout(r, 500))
    ElMessage.info('代码运行功能即将接入')
  } finally {
    expPlanCodeRunning.value = false
  }
}

async function runExpPlanBundleAudit() {
  if (!expPlanBundleView.value) return
  expPlanAuditRunning.value = true
  try {
    await new Promise((r) => setTimeout(r, 500))
    ElMessage.info('实验审计功能即将接入')
  } finally {
    expPlanAuditRunning.value = false
  }
}

const expPlanUploadPlanOptions = computed(() =>
  expPlanItems.value.map((item) => ({
    value: item.id,
    label: experimentPlanRowTitle(item),
  })),
)

function resetExpPlanUploadPlanInput() {
  expPlanUploadPlanFile.value = null
  if (expPlanUploadPlanInputRef.value) expPlanUploadPlanInputRef.value.value = ''
}

async function openExpPlanUploadPlanDialog() {
  await reloadExperimentPlans({ page: expPlanListPage.value })
  if (!expPlanItems.value.length) {
    ElMessage.warning('暂无实验方案，请先在「文献综述」等待生成完成')
    return
  }
  expPlanUploadPlanId.value = expPlanItems.value[0]?.id ?? ''
  resetExpPlanUploadPlanInput()
  expPlanUploadPlanOpen.value = true
}

function closeExpPlanUploadPlanDialog() {
  if (expPlanUploadPlanSubmitting.value) return
  expPlanUploadPlanOpen.value = false
  resetExpPlanUploadPlanInput()
}

function openExpPlanUploadPlanPicker() {
  if (expPlanUploadPlanSubmitting.value) return
  expPlanUploadPlanInputRef.value?.click()
}

function onExpPlanUploadPlanFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  expPlanUploadPlanFile.value = input.files?.[0] ?? null
}

function onExpPlanUploadPlanDrop(ev: DragEvent) {
  if (expPlanUploadPlanSubmitting.value) return
  const file = ev.dataTransfer?.files?.[0]
  if (file) expPlanUploadPlanFile.value = file
}

async function submitExpPlanUploadPlan() {
  const planId = expPlanUploadPlanId.value.trim()
  const file = expPlanUploadPlanFile.value
  const ms = manuscriptId.value
  if (!planId || !file || !ms) {
    ElMessage.warning('请选择关联实验方案与文件')
    return
  }
  expPlanUploadPlanSubmitting.value = true
  try {
    await new Promise((r) => setTimeout(r, 400))
    ElMessage.info(`上传实验方案「${file.name}」功能即将接入`)
    closeExpPlanUploadPlanDialog()
  } finally {
    expPlanUploadPlanSubmitting.value = false
  }
}

function experimentDataFileLabel(item: PaperExperimentPlanItem): string {
  const uri = item.experiment_data_uri?.trim()
  if (!uri) return ''
  const parts = uri.split('/')
  return parts[parts.length - 1] ?? uri
}

function openDeleteExpPlanConfirm(item: PaperExperimentPlanItem) {
  expPlanDeletePending.value = item
}

function closeDeleteExpPlanConfirm() {
  if (expPlanDeleteSubmitting.value) return
  expPlanDeletePending.value = null
}

async function submitDeleteExpPlanConfirm() {
  const item = expPlanDeletePending.value
  if (!item) return
  expPlanDeleteSubmitting.value = true
  try {
    await new Promise((r) => setTimeout(r, 300))
    ElMessage.info('删除实验方案功能即将接入')
    expPlanDeletePending.value = null
    if (expPlanBundleView.value?.item.id === item.id) closeExpPlanBundleView()
  } finally {
    expPlanDeleteSubmitting.value = false
  }
}

function openDeleteLitReviewConfirm(item: PaperLiteratureReviewItem) {
  litReviewDeletePending.value = item
}

function closeDeleteLitReviewConfirm() {
  if (litReviewDeleteSubmitting.value) return
  litReviewDeletePending.value = null
}

async function submitDeleteLitReviewConfirm() {
  const item = litReviewDeletePending.value
  const ms = manuscriptId.value
  if (!item || !ms) return
  litReviewDeleteSubmitting.value = true
  try {
    await softDeleteLiteratureReview(ms, item.id)
    ElMessage.success('已删除')
    litReviewDeletePending.value = null
    if (litReviewViewItem.value?.id === item.id) closeLitReviewView()
    if (litReviewBundleView.value?.item.id === item.id) closeLitReviewBundleView()
    await reloadLiteratureReviews()
  } catch (e) {
    const msg = e instanceof Error ? e.message : '删除失败'
    ElMessage.error(msg)
  } finally {
    litReviewDeleteSubmitting.value = false
  }
}

async function runModule(id: PaperModuleId): Promise<boolean> {
  if (id === 'literature-review') {
    await reloadLiteratureReviews()
    if (!litReviewItems.value.length) {
      ElMessage.warning('暂无文献综述，请先在「选题发现」完成审计步骤')
      return false
    }
    return true
  }
  if (id === 'experiment-planning') {
    await reloadExperimentPlans({ page: expPlanListPage.value })
    if (!expPlanListTotal.value) {
      ElMessage.warning('暂无实验方案，请先在「文献综述」等待生成完成')
      return false
    }
    return true
  }
  if (id === 'paper-writing') {
    await runPaperWriteFullPipeline()
    return true
  }
  if (id === 'figure-generation' && figureForm.chartTypes.length === 0) {
    ElMessage.warning('请至少选择一种图表类型')
    return false
  }

  await new Promise((r) => setTimeout(r, 750))
  resultVisible.value = { ...resultVisible.value, [id]: true }
  return true
}

defineExpose({
  runModule,
  reloadLiteratureReviews,
  reloadExperimentPlans,
  openLitReviewUploadExpPlanDialog,
  openExpPlanUploadPlanDialog,
})
</script>

<template>
  <div class="wf-root">
  <!-- 文献综述：paper_output_literature_review -->
  <div v-if="moduleId === 'literature-review'" class="wf-stack">
    <section v-if="litReviewsLoading" class="wf-panel">
      <p class="wf-meta">正在加载文献综述…</p>
    </section>

    <section v-else-if="!litReviewItems.length" class="wf-panel">
      <h2 class="wf-title wf-title--module-head">文献综述</h2>
      <p class="wf-lead">
        当前论文尚无文献综述。请先在「选题发现」完成审计步骤（第三步会自动生成并入库）。
      </p>
    </section>

    <section v-else class="wf-panel wf-lit-list-panel">
      <table class="wf-table wf-lit-table wf-lit-table--lit-review">
        <colgroup>
          <col class="wf-lit-col-title-pct" />
          <col class="wf-lit-col-created-pct" />
          <col class="wf-lit-col-from-pct" />
          <col class="wf-lit-col-status-pct" />
          <col class="wf-lit-col-actions-pct" />
        </colgroup>
        <thead>
          <tr>
            <th class="wf-lit-col-title">标题</th>
            <th class="wf-lit-col-meta wf-lit-col-created">创建时间</th>
            <th class="wf-lit-col-meta wf-lit-col-from-head">来源</th>
            <th class="wf-lit-col-meta wf-lit-col-status">状态</th>
            <th class="wf-lit-col-actions">
              <div class="wf-lit-actions">
                <span class="wf-lit-head-slot">
                  <button
                    type="button"
                    tabindex="-1"
                    aria-hidden="true"
                    class="paper-btn-generate paper-btn--compact wf-lit-width-ruler"
                  >
                    生成
                  </button>
                  <span class="wf-lit-col-head-label">操作</span>
                </span>
                <button
                  type="button"
                  tabindex="-1"
                  aria-hidden="true"
                  class="paper-btn-danger paper-btn--compact wf-lit-width-ruler"
                >
                  删除
                </button>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in litReviewItems" :key="item.id">
            <td class="wf-lit-col-title">{{ litReviewRowTitle(item) }}</td>
            <td class="wf-lit-col-meta wf-lit-col-created wf-lit-created-at">
              {{ formatLitReviewCreatedAt(item.created_at) }}
            </td>
            <td class="wf-lit-col-meta wf-lit-col-from-head wf-lit-from">
              <span
                class="wf-lit-from-tag"
                :class="`wf-lit-from-tag--${literatureReviewFromTone(item.from)}`"
              >
                {{ formatLiteratureReviewFrom(item.from) }}
              </span>
            </td>
            <td class="wf-lit-col-meta wf-lit-col-status">
              <span
                class="wf-status-pill"
                :class="`wf-status-pill--${literatureReviewStatusTone(item.status)}`"
              >
                {{ formatLiteratureReviewStatus(item.status) }}
              </span>
            </td>
            <td class="wf-lit-col-actions">
              <div class="wf-lit-actions">
                <button
                  type="button"
                  class="paper-btn--compact"
                  :class="
                    literatureReviewNeedsGenerateRun(item)
                      ? 'paper-btn-generate'
                      : 'paper-btn-primary'
                  "
                  @click="onLitReviewRowPrimaryAction(item)"
                >
                  {{ literatureReviewNeedsGenerateRun(item) ? '生成' : '管理' }}
                </button>
                <button
                  type="button"
                  class="paper-btn-danger paper-btn--compact"
                  @click="openDeleteLitReviewConfirm(item)"
                >
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <PaperListPagination
        :total="litReviewListTotal"
        :page-size="LITERATURE_REVIEW_LIST_PAGE_SIZE"
        :current-page="litReviewListPage"
        @update:current-page="onLitReviewListPageChange"
      />
    </section>

    <Teleport to="body">
      <div
        v-if="litReviewUploadExpPlanOpen"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-lit-upload-exp-plan-title"
        @click.self="closeLitReviewUploadExpPlanDialog"
        @keydown.escape="closeLitReviewUploadExpPlanDialog"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="wf-lit-upload-exp-plan-title" class="paper-modal-title">上传文献综述</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="litReviewUploadExpPlanSubmitting"
              @click="closeLitReviewUploadExpPlanDialog"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <label v-if="litReviewUploadExpPlanOptions.length > 1" class="wf-lit-upload-field">
              <span class="wf-lit-upload-field-label">关联文献综述</span>
              <PaperSelect
                v-model="litReviewUploadExpPlanReviewId"
                :options="litReviewUploadExpPlanOptions"
              />
            </label>
            <p v-else-if="litReviewUploadExpPlanOptions.length === 1" class="wf-meta wf-lit-upload-meta">
              关联：{{ litReviewUploadExpPlanOptions[0]?.label }}
            </p>
            <div
              class="wf-exp-data-drop"
              :class="{
                'wf-exp-data-drop--has-file': !!litReviewUploadExpPlanFile,
                'wf-exp-data-drop--disabled': litReviewUploadExpPlanSubmitting,
              }"
              role="button"
              tabindex="0"
              @click="openLitReviewUploadExpPlanPicker"
              @keydown.enter.prevent="openLitReviewUploadExpPlanPicker"
              @keydown.space.prevent="openLitReviewUploadExpPlanPicker"
              @dragover.prevent
              @drop.prevent="onLitReviewUploadExpPlanDrop"
            >
              <input
                ref="litReviewUploadExpPlanInputRef"
                type="file"
                class="wf-exp-data-file-input"
                accept=".md,.markdown,.pdf,.doc,.docx,.zip"
                :disabled="litReviewUploadExpPlanSubmitting"
                @change="onLitReviewUploadExpPlanFileChange"
              />
              <p v-if="!litReviewUploadExpPlanFile" class="wf-exp-data-drop-hint">
                点击或拖拽文献综述文件（Markdown / PDF / Word / ZIP）
              </p>
              <p v-else class="wf-exp-data-drop-name">{{ litReviewUploadExpPlanFile.name }}</p>
            </div>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="litReviewUploadExpPlanSubmitting || !litReviewUploadExpPlanFile"
              @click="submitLitReviewUploadExpPlan"
            >
              {{ litReviewUploadExpPlanSubmitting ? '上传中…' : '上传' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="litReviewUploadExpPlanSubmitting"
              @click="closeLitReviewUploadExpPlanDialog"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="litReviewGeneratePending"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-lit-generate-run-title"
        @click.self="closeLitReviewGenerateDialog"
        @keydown.escape="closeLitReviewGenerateDialog"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel wf-lit-run-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="wf-lit-generate-run-title" class="paper-modal-title">生成文献综述</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="litReviewGenerateSubmitting"
              @click="closeLitReviewGenerateDialog"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body wf-lit-run-body">
            <p class="wf-meta wf-lit-run-meta">
              {{ litReviewRowTitle(litReviewGeneratePending) }}
            </p>
            <div class="paper-message-box__form wf-lit-run-form">
              <div class="paper-field paper-field--block">
                <span class="paper-label">文献来源 <em class="req">*</em></span>
                <div class="paper-check-group wf-lit-run-check-group">
                  <label
                    v-for="src in LITERATURE_SOURCE_OPTIONS"
                    :key="src.code"
                    class="paper-check paper-check--inline wf-lit-run-check"
                  >
                    <input
                      type="checkbox"
                      :checked="isLitReviewGenerateSourceChecked(src.code)"
                      :disabled="litReviewGenerateSubmitting"
                      @change="
                        toggleLitReviewGenerateSource(
                          src.code,
                          ($event.target as HTMLInputElement).checked,
                        )
                      "
                    />
                    <span>{{ src.label }}</span>
                  </label>
                </div>
                <label class="paper-check paper-check--nested wf-lit-run-check">
                  <input
                    v-model="litReviewGenerateForm.randomReferences"
                    type="checkbox"
                    :disabled="litReviewGenerateSubmitting"
                  />
                  <span class="paper-check-row">
                    <span>每次运行随机参考文献</span>
                    <ElTooltip
                      content="勾选后，每次运行随机获取第三方论文，否则会顺序获取第三方论文多次可能是一样的"
                      placement="top"
                    >
                      <button
                        type="button"
                        class="paper-field-help wf-lit-run-field-help"
                        aria-label="随机参考文献说明"
                        @click.prevent
                      >
                        ?
                      </button>
                    </ElTooltip>
                  </span>
                </label>

                <div class="paper-my-literature-block wf-lit-run-my-lit">
                  <span class="paper-label">我的文献</span>
                  <div class="paper-my-literature-row">
                    <button
                      type="button"
                      class="paper-my-literature-btn"
                      :disabled="litReviewGenerateSubmitting"
                      @click="openLitReviewGenerateUserLiteraturePicker"
                    >
                      选择文献
                    </button>
                    <span
                      v-if="litReviewGenerateForm.userLiteratureIds.length"
                      class="paper-hint paper-my-literature-summary wf-lit-run-hint"
                    >
                      已选 {{ litReviewGenerateForm.userLiteratureIds.length }} 篇文献
                    </span>
                  </div>
                </div>
              </div>

              <div class="paper-field paper-field--block wf-lit-run-checkpoint-block">
                <span class="paper-label">人工检查</span>
                <div class="paper-check-group wf-lit-run-check-group">
                  <label class="paper-check paper-check--inline wf-lit-run-check">
                    <input
                      v-model="litReviewGenerateForm.humanCheckpoint"
                      type="checkbox"
                      :disabled="litReviewGenerateSubmitting"
                    />
                    <span>人工检查点</span>
                  </label>
                </div>
                <span class="paper-hint wf-lit-run-hint">
                  开启后三步各暂停一次：检索入库 → 脑暴（含新颖性） → 审计，每步需确认并继续
                </span>
              </div>
            </div>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="litReviewGenerateSubmitting || !litReviewGenerateForm.sourceCodes.length"
              @click="submitLitReviewGenerateRun"
            >
              {{ litReviewGenerateSubmitting ? '运行中…' : '运行' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="litReviewGenerateSubmitting"
              @click="closeLitReviewGenerateDialog"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

    <PaperTopicUserLiteraturePickDialog
      :open="litReviewGenerateUserLitOpen"
      :initial-selected-ids="litReviewGenerateForm.userLiteratureIds"
      :disabled="litReviewGenerateSubmitting"
      @close="closeLitReviewGenerateUserLiteraturePicker"
      @confirm="onLitReviewGenerateUserLiteraturePicked"
    />

    <Teleport to="body">
      <div
        v-if="litReviewDeletePending"
        class="paper-modal-overlay wf-lit-delete-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-lit-delete-title"
        @click.self="closeDeleteLitReviewConfirm"
        @keydown.escape="closeDeleteLitReviewConfirm"
      >
        <div
          class="paper-message-box paper-message-box--danger paper-modal-panel wf-lit-delete-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="wf-lit-delete-title" class="paper-modal-title">删除文献综述</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="litReviewDeleteSubmitting"
              @click="closeDeleteLitReviewConfirm"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <p class="wf-lit-delete-lead">
              确定删除「{{ litReviewRowTitle(litReviewDeletePending) }}」？
            </p>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-danger wf-lit-delete-dialog-btn"
              :disabled="litReviewDeleteSubmitting"
              @click="submitDeleteLitReviewConfirm"
            >
              {{ litReviewDeleteSubmitting ? '删除中…' : '删除' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="litReviewDeleteSubmitting"
              @click="closeDeleteLitReviewConfirm"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

  </div>

  <!-- 实验方案：paper_output_experiment_plan -->
  <div v-else-if="moduleId === 'experiment-planning'" class="wf-stack">
    <section v-if="expPlansLoading" class="wf-panel">
      <p class="wf-meta">正在加载实验方案…</p>
    </section>

    <section v-else-if="expPlanListTotal === 0" class="wf-panel">
      <h2 class="wf-title wf-title--module-head">实验方案</h2>
      <p class="wf-lead">
        当前论文尚无实验方案。请在「文献综述」中等待方案生成完成后在此管理。
      </p>
    </section>

    <section v-else class="wf-panel wf-lit-list-panel">
      <table class="wf-table wf-lit-table">
        <thead>
          <tr>
            <th>标题</th>
            <th>创建时间</th>
            <th>状态</th>
            <th class="wf-lit-col-actions">
              <div class="wf-lit-actions">
                <span class="wf-lit-head-slot">
                  <button
                    type="button"
                    tabindex="-1"
                    aria-hidden="true"
                    class="paper-btn-primary paper-btn--compact wf-lit-width-ruler"
                  >
                    管理
                  </button>
                  <span class="wf-lit-col-head-label">操作</span>
                </span>
                <button
                  type="button"
                  tabindex="-1"
                  aria-hidden="true"
                  class="paper-btn-danger paper-btn--compact wf-lit-width-ruler"
                >
                  删除
                </button>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in expPlanItems" :key="item.id">
            <td class="wf-lit-col-title">{{ experimentPlanRowTitle(item) }}</td>
            <td class="wf-lit-created-at">{{ formatLitReviewCreatedAt(item.created_at) }}</td>
            <td>
              <span
                class="wf-status-pill"
                :class="`wf-status-pill--${experimentPlanStatusTone(item.status)}`"
              >
                {{ formatExperimentPlanStatus(item.status) }}
              </span>
            </td>
            <td class="wf-lit-col-actions">
              <div class="wf-lit-actions">
                <button
                  type="button"
                  class="paper-btn-primary paper-btn--compact"
                  @click="openExpPlanBundleView(item)"
                >
                  管理
                </button>
                <button
                  type="button"
                  class="paper-btn-danger paper-btn--compact"
                  @click="openDeleteExpPlanConfirm(item)"
                >
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <PaperListPagination
        :total="expPlanListTotal"
        :page-size="EXPERIMENT_PLAN_LIST_PAGE_SIZE"
        :current-page="expPlanListPage"
        @update:current-page="onExpPlanListPageChange"
      />
    </section>

    <Teleport to="body">
      <div
        v-if="expPlanUploadPlanOpen"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-exp-plan-upload-plan-title"
        @click.self="closeExpPlanUploadPlanDialog"
        @keydown.escape="closeExpPlanUploadPlanDialog"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="wf-exp-plan-upload-plan-title" class="paper-modal-title">上传实验方案</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="expPlanUploadPlanSubmitting"
              @click="closeExpPlanUploadPlanDialog"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <label v-if="expPlanUploadPlanOptions.length > 1" class="wf-lit-upload-field">
              <span class="wf-lit-upload-field-label">关联实验方案</span>
              <PaperSelect
                v-model="expPlanUploadPlanId"
                :options="expPlanUploadPlanOptions"
              />
            </label>
            <p v-else-if="expPlanUploadPlanOptions.length === 1" class="wf-meta wf-lit-upload-meta">
              关联：{{ expPlanUploadPlanOptions[0]?.label }}
            </p>
            <div
              class="wf-exp-data-drop"
              :class="{
                'wf-exp-data-drop--has-file': !!expPlanUploadPlanFile,
                'wf-exp-data-drop--disabled': expPlanUploadPlanSubmitting,
              }"
              role="button"
              tabindex="0"
              @click="openExpPlanUploadPlanPicker"
              @keydown.enter.prevent="openExpPlanUploadPlanPicker"
              @keydown.space.prevent="openExpPlanUploadPlanPicker"
              @dragover.prevent
              @drop.prevent="onExpPlanUploadPlanDrop"
            >
              <input
                ref="expPlanUploadPlanInputRef"
                type="file"
                class="wf-exp-data-file-input"
                accept=".md,.markdown,.pdf,.doc,.docx,.zip"
                :disabled="expPlanUploadPlanSubmitting"
                @change="onExpPlanUploadPlanFileChange"
              />
              <p v-if="!expPlanUploadPlanFile" class="wf-exp-data-drop-hint">
                点击或拖拽实验方案文件（Markdown / PDF / Word / ZIP）
              </p>
              <p v-else class="wf-exp-data-drop-name">{{ expPlanUploadPlanFile.name }}</p>
            </div>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="expPlanUploadPlanSubmitting || !expPlanUploadPlanFile"
              @click="submitExpPlanUploadPlan"
            >
              {{ expPlanUploadPlanSubmitting ? '上传中…' : '上传' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="expPlanUploadPlanSubmitting"
              @click="closeExpPlanUploadPlanDialog"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="expPlanDeletePending"
        class="paper-modal-overlay wf-lit-delete-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-exp-plan-delete-title"
        @click.self="closeDeleteExpPlanConfirm"
        @keydown.escape="closeDeleteExpPlanConfirm"
      >
        <div
          class="paper-message-box paper-message-box--danger paper-modal-panel wf-lit-delete-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="wf-exp-plan-delete-title" class="paper-modal-title">删除实验方案</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="expPlanDeleteSubmitting"
              @click="closeDeleteExpPlanConfirm"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <p class="wf-lit-delete-lead">
              确定删除「{{ experimentPlanRowTitle(expPlanDeletePending) }}」？
            </p>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-danger wf-lit-delete-dialog-btn"
              :disabled="expPlanDeleteSubmitting"
              @click="submitDeleteExpPlanConfirm"
            >
              {{ expPlanDeleteSubmitting ? '删除中…' : '删除' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="expPlanDeleteSubmitting"
              @click="closeDeleteExpPlanConfirm"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>
  </div>

  <!-- 论文写作：ARIS Workflow 3（plan → write → compile → polish → audit） -->
  <div v-else-if="moduleId === 'paper-writing'" class="wf-stack wf-paper-write-stack">
    <section class="wf-panel wf-paper-write-head">
      <div class="wf-paper-write-head-row">
        <div class="wf-paper-write-head-main">
          <h2 class="wf-title wf-title--module-head">当前论文 · {{ manuscriptTitle }}</h2>
          <p class="wf-meta wf-paper-write-sub">
            {{ DEMO_MANUSCRIPT.title }} · 引用 {{ DEMO_MANUSCRIPT.cites }} 条（已验证
            {{ DEMO_MANUSCRIPT.verifiedCites }}）
          </p>
        </div>
        <div class="wf-paper-write-head-actions">
          <span class="wf-meta">{{ paperWriteVenueLabel }}</span>
          <span
            class="wf-status-pill wf-status-pill--inline"
            :class="
              paperWriteRunStatus === 'running'
                ? 'wf-status-pill--progress'
                : 'wf-status-pill--muted'
            "
          >
            {{ paperWriteRunStatus === 'running' ? 'running' : 'idle' }}
          </span>
          <button
            type="button"
            class="paper-btn-generate"
            :disabled="paperWriteRunStatus === 'running'"
            @click="runPaperWriteFullPipeline"
          >
            {{ paperWriteRunStatus === 'running' ? '运行中…' : '运行全流程' }}
          </button>
        </div>
      </div>

      <p v-if="paperWriteDepsLoading" class="wf-meta">正在检查前置门禁…</p>
      <div v-else class="wf-paper-write-gates">
        <div
          class="wf-paper-write-gate"
          :class="paperWriteLitGateOk ? 'wf-paper-write-gate--ok' : 'wf-paper-write-gate--warn'"
        >
          <h3 class="wf-paper-write-gate-title">文献综述 · 引用门禁</h3>
          <p class="wf-paper-write-gate-body">
            {{
              paperWriteLitGateOk
                ? '已满足，可写入正文'
                : '需 PDF/Word 就绪后再写 cite'
            }}
          </p>
          <button
            v-if="!paperWriteLitGateOk"
            type="button"
            class="wf-paper-write-gate-link"
            @click="goToLiteratureReviewModule"
          >
            前往文献综述
          </button>
        </div>
        <div
          class="wf-paper-write-gate"
          :class="paperWriteExpGateOk ? 'wf-paper-write-gate--ok' : 'wf-paper-write-gate--warn'"
        >
          <h3 class="wf-paper-write-gate-title">实验方案 · 结果/审计</h3>
          <p class="wf-paper-write-gate-body">
            {{ paperWriteExpGateOk ? '已满足' : '暂无实验方案，请先完成文献综述链路' }}
          </p>
          <button
            v-if="!paperWriteExpGateOk"
            type="button"
            class="wf-paper-write-gate-link"
            @click="goToExperimentPlanningModule"
          >
            前往实验方案
          </button>
        </div>
        <div
          class="wf-paper-write-gate"
          :class="paperWriteFigureGateOk ? 'wf-paper-write-gate--ok' : 'wf-paper-write-gate--warn'"
        >
          <h3 class="wf-paper-write-gate-title">
            图表资产 · {{ paperWriteFigureReadyCount }}/{{ paperWriteFigureTargetCount }} 就绪
          </h3>
          <p class="wf-paper-write-gate-body">
            {{
              paperWriteFigureGateOk
                ? '图表已齐，可插入正文'
                : `缺 ${paperWriteFigureTargetCount - paperWriteFigureReadyCount} 张图`
            }}
          </p>
          <button type="button" class="wf-paper-write-gate-link" @click="goToFigureManagement">
            跳转图表管理
          </button>
        </div>
      </div>
    </section>

    <div class="wf-paper-write-layout">
      <aside class="wf-paper-write-sidebar">
        <section class="wf-panel wf-paper-write-sidebar-panel">
          <h3 class="wf-paper-write-sidebar-title">写作阶段</h3>
          <nav class="wf-paper-write-phase-nav" aria-label="写作阶段">
            <button
              v-for="(p, i) in PAPER_WRITE_PHASES"
              :key="p.id"
              type="button"
              class="wf-paper-write-phase-btn"
              :class="{ 'wf-paper-write-phase-btn--active': paperWritePhase === p.id }"
              @click="setPaperWritePhase(p.id)"
            >
              <span class="wf-paper-write-phase-index">{{ i + 1 }}</span>
              <span class="wf-paper-write-phase-text">
                {{ p.label }}
                <span class="wf-paper-write-phase-skill">{{ p.skill }}</span>
              </span>
            </button>
          </nav>
          <p class="wf-meta wf-paper-write-flow-hint">
            全流程：plan → 图表模块 → write → compile → polish → audit
          </p>
        </section>

        <section class="wf-panel wf-paper-write-sidebar-panel">
          <h3 class="wf-paper-write-sidebar-title">人工检查点</h3>
          <ul class="wf-paper-write-checklist">
            <li class="wf-paper-write-checklist-item wf-paper-write-checklist-item--done">
              大纲冻结后再写 Method
            </li>
            <li class="wf-paper-write-checklist-item">编译通过后再开润色循环</li>
          </ul>
        </section>
      </aside>

      <section class="wf-panel wf-paper-write-main">
        <header class="wf-paper-write-main-head">
          <h3 class="wf-title">{{ paperWritePhaseHeading }}</h3>
          <button
            type="button"
            class="paper-btn-primary"
            :disabled="paperWriteRunStatus === 'running'"
            @click="runPaperWritePhasePrimary"
          >
            {{ paperWritePhasePrimaryLabel }}
          </button>
        </header>

        <div class="wf-paper-write-main-body">
          <template v-if="paperWritePhase === 'plan'">
            <div class="wf-paper-write-stats">
              <div class="wf-paper-write-stat">
                <span class="wf-paper-write-stat-label">主 claim</span>
                <strong class="wf-paper-write-stat-value">2</strong>
              </div>
              <div class="wf-paper-write-stat">
                <span class="wf-paper-write-stat-label">计划页数</span>
                <strong class="wf-paper-write-stat-value">9</strong>
              </div>
              <div class="wf-paper-write-stat">
                <span class="wf-paper-write-stat-label">计划图数</span>
                <strong class="wf-paper-write-stat-value">6</strong>
              </div>
            </div>
            <table class="wf-table wf-table--compact wf-paper-write-outline-table">
              <thead>
                <tr>
                  <th>章节</th>
                  <th>支撑 claim</th>
                  <th>图表</th>
                  <th>页预算</th>
                  <th>状态</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in PAPER_WRITE_OUTLINE_ROWS" :key="row.section">
                  <td>{{ row.section }}</td>
                  <td>{{ row.claim }}</td>
                  <td>{{ row.figures }}</td>
                  <td>{{ row.pages }}</td>
                  <td>
                    <span
                      class="wf-status-pill wf-status-pill--inline"
                      :class="`wf-status-pill--${row.statusTone}`"
                    >
                      {{ row.status }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
            <p class="wf-meta wf-paper-write-hero">
              Hero：在固定 KV 预算下，训练-free 动态路由不低于专用稀疏基线（NARRATIVE_REPORT）
            </p>
          </template>

          <template v-else-if="paperWritePhase === 'write'">
            <div class="wf-paper-write-progress">
              <div class="wf-paper-write-progress-labels">
                <span>章节进度</span>
                <span>{{ paperWriteSectionsDone }} / {{ DEMO_MANUSCRIPT.sections.length }} 完成</span>
              </div>
              <div class="wf-paper-write-progress-track" aria-hidden="true">
                <span
                  class="wf-paper-write-progress-seg wf-paper-write-progress-seg--done"
                  :style="{
                    flex: paperWriteSectionsDone,
                  }"
                />
                <span
                  class="wf-paper-write-progress-seg wf-paper-write-progress-seg--draft"
                  :style="{
                    flex: DEMO_MANUSCRIPT.sections.filter((s) => s.status === 'draft').length,
                  }"
                />
                <span
                  class="wf-paper-write-progress-seg wf-paper-write-progress-seg--todo"
                  :style="{
                    flex: Math.max(
                      0,
                      DEMO_MANUSCRIPT.sections.length -
                        paperWriteSectionsDone -
                        DEMO_MANUSCRIPT.sections.filter((s) => s.status === 'draft').length,
                    ),
                  }"
                />
              </div>
            </div>
            <div class="wf-paper-write-write-grid">
              <div class="wf-paper-write-write-col">
                <h4 class="wf-paper-write-subhead">章节列表</h4>
                <ul class="wf-paper-write-section-list">
                  <li v-for="sec in DEMO_MANUSCRIPT.sections" :key="sec.name">
                    <button
                      type="button"
                      class="wf-paper-write-section-btn"
                      :class="{
                        'wf-paper-write-section-btn--active': paperWriteSelectedSection === sec.name,
                      }"
                      @click="paperWriteSelectedSection = sec.name"
                    >
                      <span>{{ sec.name }}</span>
                      <span
                        class="wf-status-pill wf-status-pill--inline"
                        :class="
                          sec.status === 'done' ? 'wf-status-pill--success' : 'wf-status-pill--draft'
                        "
                      >
                        {{ sec.status === 'done' ? '完成' : '草稿' }}
                      </span>
                    </button>
                  </li>
                </ul>
              </div>
              <div class="wf-paper-write-write-col">
                <h4 class="wf-paper-write-subhead">{{ paperWriteSelectedSection }} 预览</h4>
                <pre class="wf-paper-write-excerpt">{{ paperWriteSectionPreview }}</pre>
              </div>
            </div>
            <p class="wf-meta wf-paper-write-callout">
              引用：DBLP 拉 BibTeX；禁止凭空编造 cite key
            </p>
          </template>

          <template v-else-if="paperWritePhase === 'compile'">
            <div class="wf-paper-write-compile-grid">
              <div class="wf-paper-write-pdf-preview">
                <p class="wf-meta">main.pdf · 9 pages</p>
                <p v-if="resultVisible['paper-writing']" class="wf-paper-write-demo-title">
                  {{ DEMO_MANUSCRIPT.title }}
                </p>
                <p v-else class="wf-meta">编译后将在此预览 PDF</p>
              </div>
              <div class="wf-paper-write-compile-side">
                <h4 class="wf-paper-write-subhead">编译日志</h4>
                <pre class="wf-paper-write-log">latexmk -pdf main.tex</pre>
                <p class="wf-meta">R1/R2 pass — 9 pages (limit 9)</p>
                <p class="wf-meta">Warning: Figure 4 float wide</p>
                <label class="wf-paper-write-checkbox">
                  <input v-model="paperWriteCompileAnonymous" type="checkbox" />
                  匿名投稿（ANONYMOUS）
                </label>
              </div>
            </div>
          </template>

          <template v-else-if="paperWritePhase === 'polish'">
            <div class="wf-paper-write-stats">
              <div class="wf-paper-write-stat">
                <span class="wf-paper-write-stat-label">轮次</span>
                <strong class="wf-paper-write-stat-value">0 / 2</strong>
              </div>
              <div class="wf-paper-write-stat">
                <span class="wf-paper-write-stat-label">上次评分</span>
                <strong class="wf-paper-write-stat-value">—</strong>
              </div>
            </div>
            <table class="wf-table wf-table--compact">
              <thead>
                <tr>
                  <th>轮次</th>
                  <th>外审摘要</th>
                  <th>已改项</th>
                  <th>PDF</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>R1</td>
                  <td>（待运行）</td>
                  <td>—</td>
                  <td>—</td>
                </tr>
              </tbody>
            </table>
          </template>

          <template v-else-if="paperWritePhase === 'audit'">
            <p class="wf-meta wf-paper-write-audit-note">
              与 experiment-audit 区分：核对正文数字 ↔ 原始结果文件（paper-claim-audit，零上下文）
            </p>
            <table class="wf-table wf-table--compact">
              <thead>
                <tr>
                  <th>正文表述</th>
                  <th>证据文件</th>
                  <th>结论</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, idx) in PAPER_WRITE_CLAIM_AUDIT_ROWS" :key="idx">
                  <td>{{ row.claim }}</td>
                  <td>{{ row.evidence }}</td>
                  <td>{{ row.verdict }}</td>
                </tr>
              </tbody>
            </table>
          </template>
        </div>
      </section>
    </div>
  </div>

  <!-- 图表管理 -->
  <section v-else-if="moduleId === 'figure-generation'" class="fig-mgmt-panel">
    <div class="fig-mgmt-tabs" role="tablist" aria-label="图表管理">
      <button
        type="button"
        role="tab"
        class="fig-mgmt-tab"
        :class="{ 'fig-mgmt-tab--active': figureTab === 'upload' }"
        :aria-selected="figureTab === 'upload'"
        @click="figureTab = 'upload'"
      >
        上传图表
      </button>
      <button
        type="button"
        role="tab"
        class="fig-mgmt-tab"
        :class="{ 'fig-mgmt-tab--active': figureTab === 'generate' }"
        :aria-selected="figureTab === 'generate'"
        @click="figureTab = 'generate'"
      >
        一键生成
      </button>
    </div>

    <div v-show="figureTab === 'upload'" class="fig-mgmt-pane" role="tabpanel">
      <PaperFigureUploadPanel
        embedded
        :manuscript-id="manuscriptId"
        :manuscript-title="manuscriptTitle"
      />
    </div>

    <div v-show="figureTab === 'generate'" class="fig-mgmt-pane fig-mgmt-pane--stack" role="tabpanel">
    <section class="wf-panel wf-panel--in-fig-mgmt">
      <h2 class="wf-title">参数</h2>
      <p class="wf-lead">基于实验数据生成可复现统计图（非通用美工工具）。</p>
      <label class="wf-field wf-field--block">
        <span class="wf-label">数据文件</span>
        <PaperSelect v-model="figureForm.datasetLabel" :options="figureDatasetOptions" />
      </label>
      <div class="wf-field wf-field--block">
        <span class="wf-label">图表类型</span>
        <div class="wf-check-group">
          <label v-for="opt in FIGURE_CHART_OPTIONS" :key="opt.value" class="wf-check wf-check--inline">
            <input
              type="checkbox"
              :checked="isChartChecked(opt.value)"
              @change="toggleChartType(opt.value, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </div>
      <label class="wf-check">
        <input v-model="figureForm.includeErrorBars" type="checkbox" />
        <span>误差条 / 多 seed 汇总</span>
      </label>
      <label class="wf-field">
        <span class="wf-label">Caption 语言</span>
        <PaperSelect v-model="figureForm.captionLang" :options="captionLangOptions" />
      </label>
    </section>
    <section
      v-if="resultVisible['figure-generation']"
      class="wf-panel wf-panel--in-fig-mgmt wf-panel--in-fig-mgmt-result"
    >
      <h2 class="wf-title">Figure 产出（演示）</h2>
      <div class="wf-figure-grid">
        <div v-for="fig in DEMO_FIGURES.files" :key="fig.id" class="wf-figure-card">
          <div class="wf-figure-preview" :class="`wf-figure-preview--${fig.type}`">
            <div v-for="n in 5" :key="n" class="wf-bar" :style="{ height: `${30 + n * 12}%` }" />
          </div>
          <p class="wf-figure-cap">{{ fig.title }}</p>
          <code class="wf-figure-file">{{ fig.note }}</code>
        </div>
      </div>
      <table class="wf-table wf-table--compact">
        <thead>
          <tr>
            <th>Method</th>
            <th>Score</th>
            <th>±</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in DEMO_FIGURES.dataPreview" :key="row.method">
            <td>{{ row.method }}</td>
            <td>{{ row.score }}</td>
            <td>{{ row.std }}</td>
          </tr>
        </tbody>
      </table>
    </section>
    </div>
  </section>

  <Teleport to="body">
    <div
      v-if="litReviewBundleView"
      class="paper-modal-overlay wf-lit-view-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="wf-lit-bundle-title"
      @click.self="closeLitReviewBundleView"
      @keydown.escape="closeLitReviewBundleView"
    >
      <div
        class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-view-panel wf-lit-bundle-panel"
        @click.stop
      >
        <header class="paper-modal-header wf-lit-view-header wf-lit-bundle-header">
          <div class="wf-lit-view-heading">
            <h2 id="wf-lit-bundle-title" class="paper-modal-title wf-lit-view-title">
              {{ litReviewRowTitle(litReviewBundleView.item) }}
            </h2>
          </div>
          <button type="button" class="paper-modal-close" aria-label="关闭" @click="closeLitReviewBundleView">
            ×
          </button>
        </header>
        <nav class="wf-lit-bundle-tabs" role="tablist" aria-label="产出切换">
          <button
            v-for="t in LIT_REVIEW_BUNDLE_TABS"
            :key="t.id"
            type="button"
            role="tab"
            class="wf-lit-bundle-tab"
            :class="{ 'wf-lit-bundle-tab--active': litReviewBundleView.tab === t.id }"
            :aria-selected="litReviewBundleView.tab === t.id"
            @click="setLitReviewBundleTab(t.id)"
          >
            {{ t.label }}
          </button>
        </nav>
        <div class="paper-modal-body wf-lit-view-body wf-lit-bundle-body">
          <template v-if="litReviewBundleView.tab === 'literature-review'">
            <p class="wf-meta wf-lit-meta">
              <span>{{ formatLitReviewCreatedAt(litReviewBundleView.item.created_at) }}</span>
              <span>·</span>
              <span
                class="wf-status-pill wf-status-pill--inline"
                :class="`wf-status-pill--${literatureReviewStatusTone(litReviewBundleView.item.status)}`"
              >
                {{ formatLiteratureReviewStatus(litReviewBundleView.item.status) }}
              </span>
            </p>
            <pre
              v-if="litReviewBundleView.item.content_medium?.trim()"
              class="wf-pre wf-pre--lit"
            >{{ litReviewBundleView.item.content_medium }}</pre>
            <p v-else class="wf-artifact-empty">暂无正文（content_medium 为空）</p>
          </template>

          <template v-else-if="litReviewBundleView.tab === 'ppt'">
            <p class="wf-meta wf-lit-meta">
              {{ literatureReviewPptFileName(litReviewBundleView.item) }}
            </p>
            <div class="wf-lit-attachment-preview" aria-label="PPT 预览">
              <p class="wf-lit-attachment-preview-hint">
                幻灯片预览区域（接入存储 URL 后将在此嵌入 PPT/PDF 预览）
              </p>
            </div>
          </template>

          <template v-else-if="litReviewBundleView.tab === 'references'">
            <template v-if="litReviewBundleBibliography.length">
              <p class="wf-meta wf-lit-meta">共 {{ litReviewBundleBibliography.length }} 条</p>
              <table class="wf-table wf-table--compact wf-lit-ref-table">
                <thead>
                  <tr>
                    <th>Cite Key</th>
                    <th>题名</th>
                    <th>作者</th>
                    <th>年份</th>
                    <th class="wf-lit-ref-col-action">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="ref in litReviewBundleBibliography"
                    :key="ref.cite_key"
                    class="wf-lit-ref-row"
                    tabindex="0"
                    @click="openLitReviewRefDetail(ref)"
                    @keydown.enter="openLitReviewRefDetail(ref)"
                  >
                    <td>
                      <code class="wf-lit-ref-key-inline">{{ ref.cite_key }}</code>
                    </td>
                    <td class="wf-lit-ref-col-title">{{ ref.title }}</td>
                    <td class="wf-lit-ref-col-authors">{{ ref.authors }}</td>
                    <td class="wf-lit-ref-col-year">{{ ref.year }}</td>
                    <td class="wf-lit-ref-col-action">
                      <button
                        type="button"
                        class="paper-btn-primary paper-btn--compact"
                        @click.stop="openLitReviewRefDetail(ref)"
                      >
                        查看
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </template>
            <p v-else class="wf-artifact-empty">暂无参考文献</p>
            <p v-if="!litReviewBundleBibliography.length" class="wf-meta wf-lit-ref-hint">
              生成或上传综述后，将在此展示引用条目（cite key、题名、作者等）。
            </p>
          </template>
        </div>
        <footer
          v-if="litReviewBundleHasDownloadAction(litReviewBundleView.tab)"
          class="paper-message-box__btns wf-lit-view-btns"
        >
          <button
            v-if="
              litReviewBundleView.tab === 'literature-review' &&
              literatureReviewCanDownloadDocuments(litReviewBundleView.item)
            "
            type="button"
            class="paper-btn-primary"
            @click="downloadLitReviewBundleLiteratureReview"
          >
            下载
          </button>
          <button
            v-if="
              litReviewBundleView.tab === 'ppt' &&
              literatureReviewCanDownloadPpt(litReviewBundleView.item)
            "
            type="button"
            class="paper-btn-primary"
            @click="downloadLitReviewBundlePpt"
          >
            下载 PPT
          </button>
        </footer>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="litReviewRefDetail"
      class="paper-modal-overlay wf-lit-ref-detail-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="wf-lit-ref-detail-title"
      @click.self="closeLitReviewRefDetail"
      @keydown.escape="closeLitReviewRefDetail"
    >
      <div
        class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-view-panel wf-lit-ref-detail-panel"
        @click.stop
      >
        <header class="paper-modal-header wf-lit-view-header">
          <div class="wf-lit-view-heading">
            <p class="wf-lit-view-kind">参考文献</p>
            <h2 id="wf-lit-ref-detail-title" class="paper-modal-title wf-lit-view-title">
              {{ litReviewRefDetail.title }}
            </h2>
          </div>
          <button type="button" class="paper-modal-close" aria-label="关闭" @click="closeLitReviewRefDetail">
            ×
          </button>
        </header>
        <div class="paper-modal-body wf-lit-view-body">
          <dl class="wf-lit-ref-detail-kv">
            <div class="wf-lit-ref-detail-row">
              <dt>Cite Key</dt>
              <dd><code class="wf-lit-ref-key-inline">{{ litReviewRefDetail.cite_key }}</code></dd>
            </div>
            <div class="wf-lit-ref-detail-row">
              <dt>作者</dt>
              <dd>{{ litReviewRefDetail.authors }}</dd>
            </div>
            <div class="wf-lit-ref-detail-row">
              <dt>年份</dt>
              <dd>{{ litReviewRefDetail.year }}</dd>
            </div>
            <div v-if="litReviewRefDetail.venue?.trim()" class="wf-lit-ref-detail-row">
              <dt>出处</dt>
              <dd>{{ litReviewRefDetail.venue }}</dd>
            </div>
          </dl>
          <p class="wf-lit-ref-detail-citation">{{ litReviewRefDetail.citation }}</p>
          <h3 class="wf-lit-ref-detail-subhead">摘要</h3>
          <p class="wf-lit-ref-detail-abstract">
            {{ litReviewRefDetail.abstract?.trim() || '暂无摘要' }}
          </p>
        </div>
        <footer class="paper-message-box__btns wf-lit-view-btns">
          <button type="button" class="paper-btn-primary" @click="closeLitReviewRefDetail">关闭</button>
        </footer>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="litReviewViewItem"
      class="paper-modal-overlay wf-lit-view-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="wf-lit-view-title"
      @click.self="closeLitReviewView"
      @keydown.escape="closeLitReviewView"
    >
      <div class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-view-panel" @click.stop>
        <header class="paper-modal-header wf-lit-view-header">
          <div class="wf-lit-view-heading">
            <p class="wf-lit-view-kind">文献综述</p>
            <h2 id="wf-lit-view-title" class="paper-modal-title wf-lit-view-title">
              {{ litReviewRowTitle(litReviewViewItem) }}
            </h2>
          </div>
          <button type="button" class="paper-modal-close" aria-label="关闭" @click="closeLitReviewView">
            ×
          </button>
        </header>
        <div class="paper-modal-body wf-lit-view-body">
          <p class="wf-meta wf-lit-meta">
            <span>{{ formatLitReviewCreatedAt(litReviewViewItem.created_at) }}</span>
            <span>·</span>
            <span
              class="wf-status-pill wf-status-pill--inline"
              :class="`wf-status-pill--${literatureReviewStatusTone(litReviewViewItem.status)}`"
            >
              {{ formatLiteratureReviewStatus(litReviewViewItem.status) }}
            </span>
          </p>
          <pre v-if="litReviewViewItem.content_medium?.trim()" class="wf-pre wf-pre--lit">{{
            litReviewViewItem.content_medium
          }}</pre>
          <p v-else class="wf-artifact-empty">暂无正文（content_medium 为空）</p>
        </div>
        <footer class="paper-message-box__btns wf-lit-view-btns">
          <button
            v-if="
              litReviewViewDownloadKind &&
              litReviewViewCanDownloadKind(litReviewViewItem, litReviewViewDownloadKind)
            "
            type="button"
            class="paper-btn-primary"
            @click="onLitReviewViewPrimaryAction"
          >
            {{ litReviewViewDownloadLabel(litReviewViewItem, litReviewViewDownloadKind) }}
          </button>
          <button
            type="button"
            :class="
              litReviewViewDownloadKind &&
              litReviewViewCanDownloadKind(litReviewViewItem, litReviewViewDownloadKind)
                ? 'paper-btn-secondary'
                : 'paper-btn-primary'
            "
            @click="closeLitReviewView"
          >
            关闭
          </button>
        </footer>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="litReviewAttachmentPreview?.kind === 'ppt'"
      class="paper-modal-overlay wf-lit-view-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="wf-lit-attachment-view-title"
      @click.self="closeLitReviewAttachmentPreview"
      @keydown.escape="closeLitReviewAttachmentPreview"
    >
      <div class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-view-panel" @click.stop>
        <header class="paper-modal-header wf-lit-view-header">
          <div class="wf-lit-view-heading">
            <p class="wf-lit-view-kind">
              {{ litReviewAttachmentKindLabel(litReviewAttachmentPreview.kind) }}
            </p>
            <h2 id="wf-lit-attachment-view-title" class="paper-modal-title wf-lit-view-title">
              {{ litReviewRowTitle(litReviewAttachmentPreview.item) }}
            </h2>
          </div>
          <button
            type="button"
            class="paper-modal-close"
            aria-label="关闭"
            @click="closeLitReviewAttachmentPreview"
          >
            ×
          </button>
        </header>
        <div class="paper-modal-body wf-lit-view-body">
          <p class="wf-meta wf-lit-meta">
            <span>{{
              literatureReviewAttachmentFileName(
                litReviewAttachmentPreview.item,
                litReviewAttachmentPreview.kind,
              )
            }}</span>
          </p>
          <div class="wf-lit-attachment-preview" aria-label="文件预览">
            <p class="wf-lit-attachment-preview-hint">
              幻灯片预览区域（接入存储 URL 后将在此嵌入 PPT/PDF 预览）
            </p>
          </div>
        </div>
        <footer class="paper-message-box__btns wf-lit-view-btns">
          <button
            v-if="literatureReviewCanDownloadPpt(litReviewAttachmentPreview.item)"
            type="button"
            class="paper-btn-primary"
            @click="onLitReviewAttachmentPreviewPrimary"
          >
            下载 PPT
          </button>
          <button
            type="button"
            :class="
              literatureReviewCanDownloadPpt(litReviewAttachmentPreview.item)
                ? 'paper-btn-secondary'
                : 'paper-btn-primary'
            "
            @click="closeLitReviewAttachmentPreview"
          >
            关闭
          </button>
        </footer>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="expPlanBundleView"
      class="paper-modal-overlay wf-lit-view-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="wf-exp-plan-bundle-title"
      @click.self="closeExpPlanBundleView"
      @keydown.escape="closeExpPlanBundleView"
    >
      <div
        class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-view-panel wf-lit-bundle-panel"
        @click.stop
      >
        <header class="paper-modal-header wf-lit-view-header wf-lit-bundle-header">
          <div class="wf-lit-view-heading">
            <h2 id="wf-exp-plan-bundle-title" class="paper-modal-title wf-lit-view-title">
              {{ experimentPlanRowTitle(expPlanBundleView.item) }}
            </h2>
          </div>
          <button
            type="button"
            class="paper-modal-close"
            aria-label="关闭"
            :disabled="expPlanBundleDataSubmitting"
            @click="closeExpPlanBundleView"
          >
            ×
          </button>
        </header>
        <nav class="wf-lit-bundle-tabs" role="tablist" aria-label="实验方案管理">
          <button
            v-for="t in EXP_PLAN_BUNDLE_TABS"
            :key="t.id"
            type="button"
            role="tab"
            class="wf-lit-bundle-tab"
            :class="{ 'wf-lit-bundle-tab--active': expPlanBundleView.tab === t.id }"
            :aria-selected="expPlanBundleView.tab === t.id"
            @click="setExpPlanBundleTab(t.id)"
          >
            {{ t.label }}
          </button>
        </nav>
        <div class="paper-modal-body wf-lit-view-body wf-lit-bundle-body">
          <template v-if="expPlanBundleView.tab === 'plan'">
            <p v-if="expPlanBundleView.loading" class="wf-meta">正在加载方案…</p>
            <template v-else>
              <p class="wf-meta wf-lit-meta">
                <span>{{ formatLitReviewCreatedAt(expPlanBundleView.item.created_at) }}</span>
                <span>·</span>
                <span
                  class="wf-status-pill wf-status-pill--inline"
                  :class="`wf-status-pill--${experimentPlanStatusTone(expPlanBundleView.item.status)}`"
                >
                  {{ formatExperimentPlanStatus(expPlanBundleView.item.status) }}
                </span>
              </p>
              <p v-if="expPlanBundleView.item.summary?.trim()" class="wf-lit-summary">
                {{ expPlanBundleView.item.summary }}
              </p>
              <pre
                v-if="expPlanBundleView.item.content_medium?.trim()"
                class="wf-pre wf-pre--lit"
              >{{ expPlanBundleView.item.content_medium }}</pre>
              <p v-else class="wf-artifact-empty">暂无正文（content_medium 为空）</p>
            </template>
          </template>

          <template v-else-if="expPlanBundleView.tab === 'data'">
            <section class="wf-exp-plan-data-section" aria-label="已有实验数据">
              <h3 class="wf-exp-plan-data-subhead">已有数据</h3>
              <div
                v-if="experimentDataFileLabel(expPlanBundleView.item)"
                class="wf-exp-plan-data-current"
              >
                <p class="wf-exp-plan-data-name">
                  {{ experimentDataFileLabel(expPlanBundleView.item) }}
                </p>
                <p
                  v-if="expPlanBundleView.item.experiment_data_uri?.trim()"
                  class="wf-meta wf-exp-plan-data-uri"
                >
                  {{ expPlanBundleView.item.experiment_data_uri.trim() }}
                </p>
              </div>
              <p v-else class="wf-artifact-empty">暂无已上传的实验数据</p>
            </section>
            <section class="wf-exp-plan-data-section" aria-label="上传实验数据">
              <h3 class="wf-exp-plan-data-subhead">上传数据</h3>
              <p class="wf-meta wf-exp-plan-results-hint">
                支持 CSV、ZIP、原始记录等；再次上传将覆盖当前数据文件。
              </p>
              <div
                class="wf-exp-data-drop"
                :class="{
                  'wf-exp-data-drop--has-file': !!expPlanBundleDataFile,
                  'wf-exp-data-drop--disabled': expPlanBundleDataSubmitting,
                }"
                role="button"
                tabindex="0"
                @click="openExpPlanBundleDataPicker"
                @keydown.enter.prevent="openExpPlanBundleDataPicker"
                @keydown.space.prevent="openExpPlanBundleDataPicker"
                @dragover.prevent
                @drop.prevent="onExpPlanBundleDataDrop"
              >
                <input
                  ref="expPlanBundleDataInputRef"
                  type="file"
                  class="wf-exp-data-file-input"
                  :disabled="expPlanBundleDataSubmitting"
                  @change="onExpPlanBundleDataFileChange"
                />
                <p v-if="!expPlanBundleDataFile" class="wf-exp-data-drop-hint">
                  点击或拖拽实验数据文件到此处
                </p>
                <p v-else class="wf-exp-data-drop-name">{{ expPlanBundleDataFile.name }}</p>
              </div>
            </section>
          </template>

          <template v-else-if="expPlanBundleView.tab === 'code'">
            <p class="wf-meta wf-lit-meta">可复现实验脚本（接入后将绑定方案 storage_uri）</p>
            <pre class="wf-pre wf-pre--lit wf-exp-plan-code">{{ EXP_PLAN_CODE_PLACEHOLDER }}</pre>
          </template>

          <template v-else-if="expPlanBundleView.tab === 'results'">
            <p v-if="experimentDataFileLabel(expPlanBundleView.item)" class="wf-meta">
              关联数据：{{ experimentDataFileLabel(expPlanBundleView.item) }}
            </p>
            <p v-else class="wf-artifact-empty">
              请先在「数据」Tab 上传实验数据，或在「代码」Tab 运行产出结果
            </p>
            <p class="wf-meta wf-exp-plan-results-hint">
              接入后将在此预览指标、表格与图表摘要（由运行日志或结果文件解析）。
            </p>
          </template>

          <template v-else-if="expPlanBundleView.tab === 'audit'">
            <p class="wf-meta wf-lit-meta">
              对照方案 claim 与结果，检查诚信与证据是否充分（对齐 experiment-audit / result-to-claim）。
            </p>
            <p class="wf-artifact-empty">暂无审计结论</p>
            <p class="wf-meta wf-exp-plan-results-hint">
              接入后将展示 issues、off-manifold 风险、claim 支撑度与建议补实验项。
            </p>
          </template>
        </div>
        <footer
          v-if="
            expPlanBundleView.tab === 'data' ||
            expPlanBundleView.tab === 'code' ||
            expPlanBundleView.tab === 'results' ||
            expPlanBundleView.tab === 'audit'
          "
          class="paper-message-box__btns wf-lit-view-btns"
        >
          <button
            v-if="expPlanBundleView.tab === 'data'"
            type="button"
            class="paper-btn-primary"
            :disabled="expPlanBundleDataSubmitting || !expPlanBundleDataFile"
            @click="submitExpPlanBundleDataUpload"
          >
            {{ expPlanBundleDataSubmitting ? '上传中…' : '上传' }}
          </button>
          <button
            v-if="expPlanBundleView.tab === 'code'"
            type="button"
            class="paper-btn-primary"
            :disabled="expPlanCodeRunning"
            @click="runExpPlanBundleCode"
          >
            {{ expPlanCodeRunning ? '运行中…' : '运行' }}
          </button>
          <button
            v-if="expPlanBundleView.tab === 'results' || expPlanBundleView.tab === 'audit'"
            type="button"
            class="paper-btn-primary"
            :disabled="expPlanAuditRunning"
            @click="runExpPlanBundleAudit"
          >
            {{
              expPlanAuditRunning
                ? '审计中…'
                : expPlanBundleView.tab === 'audit'
                  ? '开始审计'
                  : '审计'
            }}
          </button>
        </footer>
      </div>
    </div>
  </Teleport>
  </div>
</template>

<style scoped>
.wf-root {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.wf-stack {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.fig-mgmt-panel {
  padding: 20px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.fig-mgmt-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 0 2px;
  margin-bottom: 24px;
  background: transparent;
  border-bottom: 1px solid #e2e8f0;
}

.fig-mgmt-pane--stack {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.fig-mgmt-pane .wf-panel--in-fig-mgmt {
  padding: 0;
  margin: 0;
  background: transparent;
  border: none;
  border-radius: 0;
  box-shadow: none;
}

.fig-mgmt-pane .wf-panel--in-fig-mgmt-result {
  padding-top: 24px;
  margin-top: 24px;
  border-top: 1px solid #e2e8f0;
}

.fig-mgmt-tab {
  position: relative;
  flex: 0 1 auto;
  min-width: 112px;
  padding: 12px 20px;
  margin-bottom: -1px;
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  border-radius: 8px 8px 0 0;
  transition:
    color 0.15s,
    background 0.15s,
    border-color 0.15s;
}

.fig-mgmt-tab:hover:not(.fig-mgmt-tab--active) {
  color: #334155;
  background: #f8fafc;
}

.fig-mgmt-tab--active {
  font-weight: 600;
  color: #1e293b;
  background: #fff;
  border-bottom-color: #6366f1;
  box-shadow: inset 0 -1px 0 #fff;
}

.wf-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.wf-panel--result {
  margin-top: 18px;
  border-color: #ddd6fe;
  background: linear-gradient(180deg, #faf5ff 0%, #fff 120px);
}

.wf-title {
  margin: 0 0 18px;
  font-size: 16px;
  font-weight: 700;
  color: #1e1b4b;
}

.wf-title--tight {
  margin-bottom: 8px;
}

.wf-title--module-head {
  font-size: 20px;
}

.wf-lit-list-panel {
  padding: 20px 22px 24px;
}

.wf-lit-table {
  margin: 0;
  font-size: 14px;
}

.wf-lit-table--lit-review {
  table-layout: fixed;
  width: 100%;
}

.wf-lit-table--lit-review col.wf-lit-col-title-pct {
  width: 40%;
}

.wf-lit-table--lit-review col.wf-lit-col-created-pct {
  width: 18%;
}

.wf-lit-table--lit-review col.wf-lit-col-from-pct {
  width: 12%;
}

.wf-lit-table--lit-review col.wf-lit-col-status-pct {
  width: 12%;
}

.wf-lit-table--lit-review col.wf-lit-col-actions-pct {
  width: 18%;
}

.wf-lit-table--lit-review .wf-lit-col-meta,
.wf-lit-table--lit-review .wf-lit-col-actions {
  box-sizing: border-box;
  overflow: hidden;
  text-overflow: ellipsis;
}

.wf-lit-table--lit-review .wf-lit-col-created {
  width: 18%;
}

.wf-lit-table--lit-review .wf-lit-col-from-head {
  width: 12%;
}

.wf-lit-table--lit-review .wf-lit-col-status {
  width: 12%;
}

.wf-lit-table--lit-review th.wf-lit-col-meta {
  white-space: nowrap;
}

.wf-lit-table--lit-review .wf-lit-col-actions {
  width: 18%;
  min-width: 0;
  text-align: right;
}

.wf-lit-table--lit-review .wf-lit-col-status .wf-status-pill {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: middle;
}

.wf-lit-from-tag {
  display: inline-block;
  padding: 2px 9px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.35;
  white-space: nowrap;
}

.wf-lit-from-tag--upload {
  color: #9a3412;
  background: #ffedd5;
}

.wf-lit-from-tag--generated {
  color: #4338ca;
  background: #e0e7ff;
}

.wf-lit-from-tag--default {
  color: #475569;
  background: #f1f5f9;
}

.wf-table.wf-lit-table th,
.wf-table.wf-lit-table td {
  padding: 13px 12px;
  font-size: 14px;
  line-height: 1.5;
  vertical-align: middle;
}

.wf-lit-upload-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 14px;
}

.wf-lit-upload-field-label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.wf-lit-upload-meta {
  margin: 0 0 12px;
}

.wf-lit-col-title {
  font-size: 14px;
  font-weight: 400;
  color: #1e1b4b;
}

.wf-lit-table--lit-review th.wf-lit-col-title,
.wf-lit-table--lit-review .wf-lit-col-title {
  max-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wf-lit-created-at {
  font-weight: 400;
  color: #334155;
  font-variant-numeric: tabular-nums;
}

.wf-table.wf-lit-table .wf-status-pill {
  font-weight: 500;
}

.wf-status-pill {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.35;
  white-space: nowrap;
}

.wf-status-pill--inline {
  padding: 2px 8px;
  vertical-align: middle;
}

.wf-status-pill--success {
  color: #047857;
  background: #d1fae5;
}

.wf-status-pill--draft {
  color: #b45309;
  background: #fef3c7;
}

.wf-status-pill--progress {
  color: #1d4ed8;
  background: #dbeafe;
}

.wf-status-pill--danger {
  color: #b91c1c;
  background: #fee2e2;
}

.wf-status-pill--muted {
  color: #64748b;
  background: #f1f5f9;
}

.wf-status-pill--default {
  color: #475569;
  background: #e2e8f0;
}

.wf-lit-col-pdf {
  min-width: 200px;
  max-width: 280px;
}

.wf-lit-attachment-stack {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

.wf-lit-list-panel .wf-lit-attachment-link {
  display: block;
  max-width: 100%;
  padding: 0;
  border: none;
  background: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  color: #4338ca;
  text-align: left;
  cursor: pointer;
}

.wf-lit-list-panel .wf-lit-attachment-link:hover {
  color: #3730a3;
  text-decoration: underline;
}

.wf-lit-attachment-preview {
  min-height: 280px;
  margin-top: 12px;
  padding: 24px 20px;
  border: 1px dashed #c7d2fe;
  border-radius: 10px;
  background: #f8fafc;
}

.wf-lit-attachment-preview-hint {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: #64748b;
  text-align: center;
}

.wf-lit-col-actions {
  min-width: 148px;
  text-align: right;
}

.wf-lit-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: flex-end;
}

.wf-lit-list-panel thead .wf-lit-actions {
  flex-wrap: nowrap;
}

.wf-lit-list-panel .wf-lit-actions .paper-btn-primary.paper-btn--compact,
.wf-lit-list-panel .wf-lit-actions .paper-btn-generate.paper-btn--compact,
.wf-lit-list-panel .wf-lit-actions .paper-btn-danger.paper-btn--compact {
  padding: 5px 10px;
  font-size: 13px;
  font-weight: 600;
  border-radius: 7px;
  line-height: 1.4;
}

.wf-lit-head-slot {
  position: relative;
  display: inline-flex;
  flex-shrink: 0;
  min-width: 3.25rem;
}

.wf-lit-width-ruler {
  visibility: hidden;
  pointer-events: none;
}

.wf-lit-action-placeholder {
  visibility: hidden;
  pointer-events: none;
  user-select: none;
}

.wf-lit-col-head-label {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 8px;
  font-size: 14px;
  font-weight: 700;
  color: #64748b;
  white-space: nowrap;
  pointer-events: none;
}

.wf-lit-delete-overlay,
.wf-lit-generate-overlay,
.wf-lit-view-overlay {
  z-index: 3200;
}

.wf-lit-generate-panel {
  width: min(420px, calc(100vw - 32px));
}

.wf-lit-run-panel {
  width: min(520px, calc(100vw - 32px));
}

.wf-lit-run-meta {
  margin: 0 0 4px;
}

.wf-lit-run-form .paper-field--block {
  margin: 0;
}

.wf-lit-run-form .paper-label .req {
  color: #e53935;
  font-style: normal;
}

.wf-lit-run-check-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 20px;
  margin-top: 8px;
}

.wf-lit-run-check {
  display: inline-flex;
  gap: 10px;
  align-items: center;
  margin: 0;
  font-size: 14px;
  color: var(--atm-text, #1e1b4b);
  cursor: pointer;
}

.wf-lit-run-check input[type='checkbox'] {
  margin: 0;
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  accent-color: var(--atm-primary, #7c3aed);
}

.wf-lit-run-check--nested {
  margin-top: 12px;
}

.wf-lit-run-my-lit {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 16px;
}

.wf-lit-run-my-lit .paper-my-literature-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 14px;
  align-items: center;
}

.wf-lit-run-my-lit .paper-my-literature-btn {
  flex-shrink: 0;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.25;
  color: #6366f1;
  cursor: pointer;
  background: #fff;
  border: 1px solid #c7d2fe;
  border-radius: 7px;
}

.wf-lit-run-my-lit .paper-my-literature-btn:hover:not(:disabled) {
  color: #4f46e5;
  background: #eef2ff;
  border-color: #a5b4fc;
}

.wf-lit-run-my-lit .paper-my-literature-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.wf-lit-run-hint {
  display: block;
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--atm-text-muted, #64748b);
}

.wf-lit-run-checkpoint-block {
  margin-top: 20px;
}

.wf-lit-run-field-help {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  padding: 0;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  color: #64748b;
  cursor: help;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 50%;
}

.wf-lit-run-field-help:hover {
  color: #475569;
  background: #e2e8f0;
}

.paper-check-row {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  line-height: 1.4;
}

.wf-lit-delete-panel {
  width: min(420px, calc(100vw - 32px));
}

.wf-lit-delete-panel .paper-message-box__btns {
  flex-direction: row-reverse;
  justify-content: flex-start;
}

.wf-lit-delete-panel .paper-message-box__btns .wf-lit-delete-dialog-btn {
  min-width: 88px;
  box-sizing: border-box;
}

.wf-lit-delete-lead {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
  line-height: 1.5;
}

.wf-exp-data-file-input {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  pointer-events: none;
}

.wf-exp-plan-data-section + .wf-exp-plan-data-section {
  margin-top: 22px;
  padding-top: 18px;
  border-top: 1px solid #e8eaf0;
}

.wf-exp-plan-data-subhead {
  margin: 0 0 10px;
  font-size: 13px;
  font-weight: 700;
  color: var(--atm-text, #1e1b4b);
}

.wf-exp-plan-data-current {
  padding: 12px 14px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.wf-exp-plan-data-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
  word-break: break-all;
}

.wf-exp-plan-data-uri {
  margin: 8px 0 0;
  font-size: 12px;
  word-break: break-all;
}

.wf-exp-plan-results-hint {
  margin: 10px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--atm-text-muted, #64748b);
}

.wf-exp-plan-code {
  max-height: min(360px, 45vh);
}

.wf-exp-data-drop {
  position: relative;
  margin-top: 14px;
  padding: 22px 16px;
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  background: #f8fafc;
  text-align: center;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    background 0.15s ease;
}

.wf-exp-data-drop:hover:not(.wf-exp-data-drop--disabled) {
  border-color: #818cf8;
  background: #f5f3ff;
}

.wf-exp-data-drop--has-file {
  border-style: solid;
  border-color: #a5b4fc;
  background: #fff;
}

.wf-exp-data-drop--disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.wf-exp-data-drop-hint {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #64748b;
}

.wf-exp-data-drop-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
  color: var(--atm-text, #1e1b4b);
  word-break: break-all;
}

.wf-lit-view-panel {
  display: flex;
  flex-direction: column;
  width: min(920px, calc(100vw - 32px));
  max-height: min(90vh, 880px);
}

.wf-lit-view-panel.wf-lit-bundle-panel {
  width: min(920px, calc(100vw - 32px));
  height: min(88vh, 780px);
  min-height: min(88vh, 780px);
  max-height: min(88vh, 780px);
}

.wf-lit-bundle-header {
  flex-shrink: 0;
  padding-bottom: 0;
  margin-bottom: 0;
}

.wf-lit-bundle-tabs {
  display: flex;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 6px;
  padding: 0 22px;
  margin: 22px 0 0;
  background: transparent;
  border-bottom: 1px solid #e2e8f0;
}

.wf-lit-bundle-tab {
  position: relative;
  flex: 0 1 auto;
  min-width: 96px;
  padding: 12px 20px;
  margin-bottom: -1px;
  font-size: 16px;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  border-radius: 8px 8px 0 0;
  transition:
    color 0.15s,
    background 0.15s,
    border-color 0.15s;
}

.wf-lit-bundle-tab:hover:not(.wf-lit-bundle-tab--active) {
  color: #334155;
  background: #f8fafc;
}

.wf-lit-bundle-tab--active {
  font-weight: 600;
  color: #1e293b;
  background: #fff;
  border-bottom-color: #6366f1;
}

.wf-lit-ref-table {
  width: 100%;
  margin-top: 4px;
  font-size: 13px;
  table-layout: fixed;
}

.wf-lit-ref-table th,
.wf-lit-ref-table td {
  vertical-align: middle;
}

.wf-lit-ref-row {
  cursor: pointer;
}

.wf-lit-ref-row:hover {
  background: #f8fafc;
}

.wf-lit-ref-key-inline {
  padding: 2px 6px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
  color: #4338ca;
  word-break: break-all;
  background: #eef2ff;
  border-radius: 4px;
}

.wf-lit-ref-col-title {
  max-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wf-lit-ref-col-authors {
  max-width: 8rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #475569;
}

.wf-lit-ref-col-year {
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.wf-lit-ref-col-action {
  width: 5.5rem;
  text-align: right;
  white-space: nowrap;
}

.wf-lit-ref-hint {
  margin-top: 8px;
}

.wf-lit-ref-detail-overlay {
  z-index: 3300;
}

.wf-lit-ref-detail-panel {
  width: min(560px, calc(100vw - 32px));
}

.wf-lit-ref-detail-kv {
  display: grid;
  gap: 10px;
  margin: 0 0 14px;
}

.wf-lit-ref-detail-row {
  display: grid;
  grid-template-columns: 5rem 1fr;
  gap: 8px;
  font-size: 14px;
}

.wf-lit-ref-detail-row dt {
  margin: 0;
  font-weight: 500;
  color: #64748b;
}

.wf-lit-ref-detail-row dd {
  margin: 0;
  color: #334155;
}

.wf-lit-ref-detail-citation {
  margin: 0 0 16px;
  padding: 12px 14px;
  font-size: 14px;
  line-height: 1.55;
  color: #334155;
  background: #f8fafc;
  border-radius: 8px;
}

.wf-lit-ref-detail-subhead {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
}

.wf-lit-ref-detail-abstract {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: #475569;
}

.wf-lit-bundle-body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  padding-top: 16px;
}

.wf-lit-bundle-panel .wf-lit-view-btns {
  flex-shrink: 0;
}

.wf-lit-bundle-panel .wf-lit-attachment-preview {
  min-height: 320px;
}

.wf-lit-bundle-paper-title {
  margin-top: 0;
}

.wf-lit-bundle-paper-table {
  margin: 12px 0;
}

.wf-lit-bundle-paper-hint {
  margin-top: 12px;
  font-size: 13px;
}

.wf-lit-view-header {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  justify-content: space-between;
}

.wf-lit-view-heading {
  flex: 1;
  min-width: 0;
}

.wf-lit-view-kind {
  margin: 0 0 8px;
  font-size: 17px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--atm-text, #1e1b4b);
  line-height: 1.35;
}

.wf-lit-view-title {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--atm-text, #1e1b4b);
  line-height: 1.35;
}

.wf-lit-view-body {
  flex: 1;
  overflow: auto;
}

.wf-lit-view-btns {
  flex-direction: row-reverse;
  justify-content: flex-start;
}

.wf-lit-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.wf-lit-summary {
  margin: 0 0 16px;
  font-size: 14px;
  line-height: 1.65;
  color: #334155;
}

.wf-pre--lit {
  max-height: min(60vh, 720px);
  overflow: auto;
}

.wf-subtitle {
  margin: 16px 0 8px;
  font-size: 14px;
  font-weight: 700;
  color: #1e1b4b;
}

.wf-pipeline {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 20px;
  font-size: 13px;
  color: #475569;
}

.wf-pipeline > span:not(.wf-pipeline-sep) {
  padding: 6px 12px;
  background: #f1f5f9;
  border-radius: 999px;
}

.wf-pipeline-sep {
  color: #94a3b8;
}

.wf-subhead {
  margin: 22px 0 10px;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
}

.wf-hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.45;
  color: #64748b;
}

.wf-check-group--tight {
  margin-top: 12px;
}

.wf-lead {
  margin: -8px 0 16px;
  font-size: 13px;
  color: #64748b;
}

.wf-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px 20px;
}

@media (max-width: 900px) {
  .wf-grid {
    grid-template-columns: 1fr;
  }
}

.wf-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wf-field--block {
  margin-bottom: 4px;
}

.wf-field--span2 {
  grid-column: 1 / -1;
}

.wf-label {
  font-size: 13px;
  font-weight: 600;
  color: #1e1b4b;
}

.wf-label .req {
  color: #e53935;
  font-style: normal;
}

.wf-input,
.wf-textarea {
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  color: #1e1b4b;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.wf-textarea {
  resize: vertical;
  font-family: inherit;
  line-height: 1.5;
}

.wf-check {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-top: 14px;
  font-size: 14px;
  cursor: pointer;
}

.wf-check--inline {
  margin-top: 0;
}

.wf-check--solo {
  justify-content: flex-end;
  align-self: end;
}

.wf-check-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  margin-top: 8px;
}

.wf-radios {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 20px;
  margin-top: 8px;
  font-size: 14px;
}

.wf-meta {
  margin: 0 0 12px;
  font-size: 13px;
  color: #64748b;
}

.wf-kv {
  margin: 0 0 10px;
  font-size: 14px;
  line-height: 1.5;
}

.wf-kv span {
  display: inline-block;
  min-width: 4em;
  margin-right: 8px;
  font-weight: 700;
  color: #64748b;
}

.wf-list {
  margin: 0;
  padding-left: 20px;
  font-size: 13px;
  line-height: 1.55;
}

.wf-list--ordered {
  margin-top: 12px;
}

.wf-pre {
  margin: 12px 0 0;
  padding: 12px;
  font-size: 12px;
  line-height: 1.5;
  color: #334155;
  white-space: pre-wrap;
  background: #f1f5f9;
  border-radius: 10px;
}

.wf-callout {
  margin: 16px 0 0;
  padding: 12px 14px;
  font-size: 13px;
  line-height: 1.5;
  background: #f0fdf4;
  border-radius: 10px;
}

.wf-callout--warn {
  background: #fffbeb;
}

.wf-artifact-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 8px;
}

.wf-artifact-card {
  padding: 14px 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.wf-artifact-card--wide {
  grid-column: 1 / -1;
}

.wf-artifact-title {
  margin: 0 0 8px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
}

.wf-artifact-body {
  margin: 0 0 6px;
  font-size: 14px;
  line-height: 1.5;
  color: #1e1b4b;
}

.wf-artifact-meta {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.wf-artifact-stat {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 700;
  color: #4f46e5;
}

.wf-artifact-stat span {
  font-size: 13px;
  font-weight: 500;
  color: #64748b;
}

.wf-artifact-empty {
  margin: 0;
  font-size: 13px;
  color: #94a3b8;
}

.wf-list--tight {
  margin-top: 0;
  font-size: 13px;
}

.wf-grid--run {
  margin-top: 18px;
}

.wf-block {
  margin-top: 8px;
}

.wf-scores {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 13px;
  font-weight: 600;
}

.wf-scores span {
  padding: 6px 10px;
  background: #ede9fe;
  border-radius: 8px;
}

.wf-findings {
  margin: 0;
  padding: 0;
  list-style: none;
}

.wf-finding--major,
.wf-finding--minor {
  margin-bottom: 8px;
  padding: 10px 12px;
  font-size: 13px;
  border-radius: 10px;
}

.wf-finding--major {
  background: #fee2e2;
}

.wf-finding--minor {
  background: #f1f5f9;
}

.wf-table {
  width: 100%;
  margin: 12px 0;
  font-size: 13px;
  border-collapse: collapse;
}

.wf-table th,
.wf-table td {
  padding: 8px 10px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.wf-table th {
  font-weight: 700;
  color: #64748b;
}

.wf-tag {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
}

.wf-tag--ok {
  color: #15803d;
}

.wf-tag--draft {
  color: #b45309;
}

.wf-figure-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

@media (max-width: 900px) {
  .wf-figure-grid {
    grid-template-columns: 1fr;
  }
}

.wf-figure-card {
  padding: 12px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.wf-figure-preview {
  display: flex;
  gap: 6px;
  align-items: flex-end;
  justify-content: center;
  height: 100px;
  padding: 8px;
  background: #f8fafc;
  border-radius: 8px;
}

.wf-bar {
  flex: 1;
  max-width: 24px;
  background: linear-gradient(180deg, #7c3aed, #6366f1);
  border-radius: 4px 4px 0 0;
}

.wf-figure-preview--line .wf-bar {
  background: #6366f1;
  opacity: 0.85;
}

.wf-figure-cap {
  margin: 10px 0 4px;
  font-size: 13px;
  font-weight: 600;
}

.wf-figure-file {
  font-size: 11px;
  color: #64748b;
}

.wf-score-big {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  color: #5b21b6;
}

.wf-dim-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  margin: 16px 0;
}

@media (max-width: 700px) {
  .wf-dim-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.wf-dim {
  padding: 12px;
  text-align: center;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.wf-dim span {
  display: block;
  font-size: 11px;
  color: #64748b;
}

.wf-dim strong {
  font-size: 18px;
  color: #1e1b4b;
}

.wf-inline-code {
  padding: 1px 6px;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: #4338ca;
  background: #eef2ff;
  border-radius: 4px;
}

.wf-paper-write-stack {
  gap: 16px;
}

.wf-paper-write-head-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
}

.wf-paper-write-head-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.wf-paper-write-sub {
  margin-top: 6px;
}

.wf-paper-write-gates {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 16px;
}

@media (max-width: 900px) {
  .wf-paper-write-gates {
    grid-template-columns: 1fr;
  }
}

.wf-paper-write-gate {
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

.wf-paper-write-gate--ok {
  background: #ecfdf5;
  border-color: #a7f3d0;
}

.wf-paper-write-gate--warn {
  background: #fffbeb;
  border-color: #fde68a;
}

.wf-paper-write-gate-title {
  margin: 0 0 6px;
  font-size: 13px;
  font-weight: 700;
  color: #1e1b4b;
}

.wf-paper-write-gate-body {
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
  color: #475569;
}

.wf-paper-write-gate-link {
  margin-top: 8px;
  padding: 0;
  font-size: 13px;
  font-weight: 600;
  color: #4338ca;
  background: none;
  border: none;
  cursor: pointer;
  text-decoration: underline;
}

.wf-paper-write-layout {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

@media (max-width: 960px) {
  .wf-paper-write-layout {
    grid-template-columns: 1fr;
  }
}

.wf-paper-write-sidebar {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wf-paper-write-sidebar-panel {
  padding: 16px;
}

.wf-paper-write-sidebar-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 700;
  color: #1e1b4b;
}

.wf-paper-write-phase-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wf-paper-write-phase-btn {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  width: 100%;
  padding: 10px 12px;
  text-align: left;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition:
    border-color 0.15s,
    background 0.15s;
}

.wf-paper-write-phase-btn:hover {
  border-color: #c7d2fe;
  background: #f8fafc;
}

.wf-paper-write-phase-btn--active {
  border-color: #818cf8;
  background: #eef2ff;
}

.wf-paper-write-phase-index {
  flex-shrink: 0;
  width: 22px;
  font-size: 12px;
  font-weight: 700;
  line-height: 22px;
  color: #4338ca;
  text-align: center;
  background: #e0e7ff;
  border-radius: 6px;
}

.wf-paper-write-phase-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
}

.wf-paper-write-phase-skill {
  font-size: 11px;
  font-weight: 500;
  color: #64748b;
}

.wf-paper-write-flow-hint {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.wf-paper-write-checklist {
  margin: 0;
  padding: 0;
  list-style: none;
}

.wf-paper-write-checklist-item {
  position: relative;
  padding-left: 22px;
  font-size: 13px;
  line-height: 1.55;
  color: #475569;
}

.wf-paper-write-checklist-item + .wf-paper-write-checklist-item {
  margin-top: 8px;
}

.wf-paper-write-checklist-item::before {
  content: '○';
  position: absolute;
  left: 0;
  color: #94a3b8;
}

.wf-paper-write-checklist-item--done::before {
  content: '✓';
  color: #059669;
  font-weight: 700;
}

.wf-paper-write-main {
  display: flex;
  flex-direction: column;
  min-height: 420px;
}

.wf-paper-write-main-head {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 16px 16px 0;
}

.wf-paper-write-main-body {
  padding: 16px;
}

.wf-paper-write-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.wf-paper-write-stat {
  min-width: 100px;
  padding: 12px 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.wf-paper-write-stat-label {
  display: block;
  font-size: 11px;
  color: #64748b;
}

.wf-paper-write-stat-value {
  display: block;
  margin-top: 4px;
  font-size: 20px;
  color: #1e1b4b;
}

.wf-paper-write-outline-table {
  margin-top: 4px;
}

.wf-paper-write-hero {
  margin-top: 14px;
  padding: 12px 14px;
  background: #f8fafc;
  border-radius: 8px;
}

.wf-paper-write-progress {
  margin-bottom: 16px;
}

.wf-paper-write-progress-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
  font-size: 12px;
  color: #64748b;
}

.wf-paper-write-progress-track {
  display: flex;
  height: 8px;
  overflow: hidden;
  background: #e2e8f0;
  border-radius: 999px;
}

.wf-paper-write-progress-seg--done {
  background: #22c55e;
}

.wf-paper-write-progress-seg--draft {
  background: #f59e0b;
}

.wf-paper-write-progress-seg--todo {
  background: #cbd5e1;
}

.wf-paper-write-write-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.2fr);
  gap: 16px;
}

@media (max-width: 720px) {
  .wf-paper-write-write-grid {
    grid-template-columns: 1fr;
  }
}

.wf-paper-write-subhead {
  margin: 0 0 10px;
  font-size: 13px;
  font-weight: 700;
  color: #334155;
}

.wf-paper-write-section-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.wf-paper-write-section-btn {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  margin-bottom: 4px;
  padding: 8px 10px;
  font-size: 13px;
  text-align: left;
  color: #334155;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
}

.wf-paper-write-section-btn--active {
  border-color: #818cf8;
  background: #eef2ff;
}

.wf-paper-write-callout {
  margin-top: 14px;
  padding: 10px 12px;
  background: #eff6ff;
  border-radius: 8px;
}

.wf-paper-write-compile-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 260px;
  gap: 16px;
}

@media (max-width: 720px) {
  .wf-paper-write-compile-grid {
    grid-template-columns: 1fr;
  }
}

.wf-paper-write-pdf-preview {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: center;
  justify-content: center;
  min-height: 280px;
  padding: 24px;
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  background: #fafafa;
}

.wf-paper-write-log {
  margin: 0 0 8px;
  padding: 10px 12px;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: #475569;
  background: #f1f5f9;
  border-radius: 8px;
}

.wf-paper-write-checkbox {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-top: 12px;
  font-size: 13px;
  color: #475569;
  cursor: pointer;
}

.wf-paper-write-audit-note {
  margin-bottom: 12px;
  padding: 10px 12px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
}

.wf-paper-write-demo-title {
  margin: 0;
  font-weight: 700;
  color: #1e1b4b;
  text-align: center;
}

.wf-paper-write-excerpt {
  margin: 0;
  padding: 14px 16px;
  overflow: auto;
  max-height: 320px;
  font-size: 13px;
  line-height: 1.55;
  color: #334155;
  white-space: pre-wrap;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}
</style>
