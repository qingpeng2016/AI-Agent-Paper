<script setup lang="ts">
import { computed, onUnmounted, reactive, ref, toRefs, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchLiteratureReviewDetail,
  fetchLiteratureReviews,
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
  softDeleteLiteratureReview,
  type LitReviewAttachmentKind,
  type LiteratureReviewBibliographyEntry,
  type PaperLiteratureReviewItem,
} from '@/api/literatureReviews'
import {
  experimentPlanRowTitle,
  fetchExperimentPlanDetail,
  fetchExperimentPlans,
  formatExperimentPlanStatus,
  experimentPlanStatusTone,
  uploadExperimentPlanData,
  type PaperExperimentPlanItem,
} from '@/api/experimentPlans'
import {
  DEMO_FIGURES,
  DEMO_MANUSCRIPT,
  DEMO_MANUSCRIPT_ANALYSIS,
} from './demoModuleOutputs'
import PaperFigureUploadPanel from './PaperFigureUploadPanel.vue'
import PaperSelect from './PaperSelect.vue'
import {
  DEFAULT_FIGURE_GENERATION,
  DEFAULT_MANUSCRIPT_ANALYSIS,
  DEFAULT_PAPER_WRITING,
  FIGURE_CHART_OPTIONS,
  PAPER_WRITING_SECTION_OPTIONS,
  type FigureGenerationForm,
  type ManuscriptAnalysisForm,
  type PaperModuleId,
  type PaperWritingForm,
} from './types'

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

const writeForm = reactive<PaperWritingForm>({
  ...DEFAULT_PAPER_WRITING,
  sections: [...DEFAULT_PAPER_WRITING.sections],
})
const figureForm = reactive<FigureGenerationForm>({
  ...DEFAULT_FIGURE_GENERATION,
  chartTypes: [...DEFAULT_FIGURE_GENERATION.chartTypes],
})
const analysisForm = reactive<ManuscriptAnalysisForm>({ ...DEFAULT_MANUSCRIPT_ANALYSIS })

const resultVisible = ref<Partial<Record<PaperModuleId, boolean>>>({})

const litReviewItems = ref<PaperLiteratureReviewItem[]>([])
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
let litReviewPollTimer: ReturnType<typeof setInterval> | null = null

const expPlanItems = ref<PaperExperimentPlanItem[]>([])
const expPlansLoading = ref(false)
const expPlanViewItem = ref<PaperExperimentPlanItem | null>(null)
const expPlanLitReviewLoading = ref(false)
const expPlanUploadPending = ref<PaperExperimentPlanItem | null>(null)
const expPlanUploadFile = ref<File | null>(null)
const expPlanUploadInputRef = ref<HTMLInputElement | null>(null)
const expPlanUploadSubmitting = ref(false)
let expPlanReloadSeq = 0

let litReviewReloadSeq = 0

async function reloadLiteratureReviews(opts?: { silent?: boolean }) {
  const silent = opts?.silent === true
  const ms = manuscriptId.value
  if (!ms || !/^\d+$/.test(ms)) {
    if (!silent) litReviewItems.value = []
    return
  }
  const seq = ++litReviewReloadSeq
  if (!silent) litReviewsLoading.value = true
  try {
    const data = await fetchLiteratureReviews(ms)
    if (seq !== litReviewReloadSeq) return
    litReviewItems.value = data.items ?? []
  } catch (e) {
    if (seq !== litReviewReloadSeq) return
    if (!silent) {
      litReviewItems.value = []
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

/** 进入文献综述 / 切换当前论文 / 刷新：同一条路径拉列表（不依赖父组件 ref 时序） */
watch(
  () => [moduleId.value, manuscriptId.value] as const,
  ([mod, ms]) => {
    if (mod !== 'literature-review') return
    if (!ms || !/^\d+$/.test(ms)) {
      litReviewItems.value = []
      return
    }
    void reloadLiteratureReviews()
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

function literatureReviewHasExperimentPlan(item: PaperLiteratureReviewItem): boolean {
  const id = item.experiment_plan_id?.trim()
  return !!id && id !== '0'
}

type LitReviewBundleTabId =
  | 'literature-review'
  | 'references'
  | 'experiment-plan'
  | 'ppt'
  | 'paper'

const LIT_REVIEW_BUNDLE_TABS: { id: LitReviewBundleTabId; label: string }[] = [
  { id: 'literature-review', label: '文献综述' },
  { id: 'experiment-plan', label: '实验方案' },
  { id: 'ppt', label: '汇报 PPT' },
  { id: 'paper', label: '论文' },
  { id: 'references', label: '参考文献' },
]

const litReviewBundleView = ref<{
  item: PaperLiteratureReviewItem
  tab: LitReviewBundleTabId
  expPlan: PaperExperimentPlanItem | null
  expPlanLoading: boolean
  expPlanError: string | null
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

async function loadLitReviewBundleExpPlan(item: PaperLiteratureReviewItem) {
  const planId = item.experiment_plan_id?.trim()
  const ms = manuscriptId.value
  const ctx = litReviewBundleView.value
  if (!ctx || ctx.item.id !== item.id) return
  if (!planId || planId === '0' || !ms) {
    litReviewBundleView.value = { ...ctx, expPlanLoading: false, expPlan: null }
    return
  }
  litReviewBundleView.value = { ...ctx, expPlanLoading: true, expPlanError: null }
  try {
    const expPlan = await fetchExperimentPlanDetail(ms, planId)
    const cur = litReviewBundleView.value
    if (!cur || cur.item.id !== item.id) return
    litReviewBundleView.value = { ...cur, expPlan, expPlanLoading: false, expPlanError: null }
  } catch (e) {
    const msg = e instanceof Error ? e.message : '加载实验方案失败'
    const cur = litReviewBundleView.value
    if (!cur || cur.item.id !== item.id) return
    litReviewBundleView.value = { ...cur, expPlanLoading: false, expPlanError: msg }
  }
}

async function openLitReviewBundleView(
  item: PaperLiteratureReviewItem,
  tab: LitReviewBundleTabId = 'literature-review',
) {
  const review = await ensureLitReviewDetail(item)
  const hasPlan = literatureReviewHasExperimentPlan(review)
  litReviewBundleView.value = {
    item: review,
    tab,
    expPlan: null,
    expPlanLoading: hasPlan,
    expPlanError: null,
  }
  if (hasPlan) void loadLitReviewBundleExpPlan(review)
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

function aiGenerateLitReviewBundlePpt() {
  ElMessage.info('AI 生成汇报 PPT 功能即将接入')
}

function downloadLitReviewBundleLiteratureReview() {
  const ctx = litReviewBundleView.value
  if (!ctx) return
  const pdf = literatureReviewPdfFileName(ctx.item)
  const docx = literatureReviewWordFileName(ctx.item)
  ElMessage.info(`下载「${pdf}」/「${docx}」功能即将接入`)
}

function aiGenerateLitReviewBundleDocuments() {
  ElMessage.info('AI 生成综述功能即将接入')
}

function onLitReviewBundleLiteratureReviewPrimary() {
  const ctx = litReviewBundleView.value
  if (!ctx) return
  if (literatureReviewCanDownloadDocuments(ctx.item)) {
    downloadLitReviewBundleLiteratureReview()
  } else {
    aiGenerateLitReviewBundleDocuments()
  }
}

function onLitReviewBundlePptPrimary() {
  const ctx = litReviewBundleView.value
  if (!ctx) return
  if (literatureReviewCanDownloadPpt(ctx.item)) {
    downloadLitReviewBundlePpt()
  } else {
    aiGenerateLitReviewBundlePpt()
  }
}

function litReviewBundleHasDownloadAction(tab: LitReviewBundleTabId): boolean {
  return tab === 'literature-review' || tab === 'ppt'
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
    openExpPlanView(item)
    emit('consumedPendingExperimentPlan')
    return
  }

  const ms = manuscriptId.value
  if (!ms) return
  try {
    const detail = await fetchExperimentPlanDetail(ms, planId)
    openExpPlanView(detail)
    emit('consumedPendingExperimentPlan')
  } catch (e) {
    const msg = e instanceof Error ? e.message : '加载实验方案失败'
    ElMessage.error(msg)
    emit('consumedPendingExperimentPlan')
  }
}

async function reloadExperimentPlans(opts?: { silent?: boolean }) {
  const silent = opts?.silent === true
  const ms = manuscriptId.value
  if (!ms || !/^\d+$/.test(ms)) {
    if (!silent) expPlanItems.value = []
    return
  }
  const seq = ++expPlanReloadSeq
  if (!silent) expPlansLoading.value = true
  try {
    const data = await fetchExperimentPlans(ms)
    if (seq !== expPlanReloadSeq) return
    expPlanItems.value = data.items ?? []
  } catch (e) {
    if (seq !== expPlanReloadSeq) return
    if (!silent) {
      expPlanItems.value = []
      const msg = e instanceof Error ? e.message : '加载实验方案失败'
      ElMessage.error(msg)
    }
  } finally {
    if (seq === expPlanReloadSeq && !silent) {
      expPlansLoading.value = false
      void tryOpenPendingExperimentPlan()
    }
  }
}

/** 进入实验方案 / 切换当前论文 / 刷新：同文献综述，模块内 loading（不走顶栏 generic loading） */
watch(
  () => [moduleId.value, manuscriptId.value] as const,
  ([mod, ms]) => {
    if (mod !== 'experiment-planning') return
    if (!ms || !/^\d+$/.test(ms)) {
      expPlanItems.value = []
      return
    }
    void reloadExperimentPlans()
  },
  { immediate: true },
)

watch(
  () => props.pendingOpenExperimentPlanId,
  () => {
    void tryOpenPendingExperimentPlan()
  },
)

const auditOptions = [
  { value: 'polished', label: '基础' },
  { value: 'standard', label: '标准' },
  { value: 'strict', label: '严格' },
]

const writeToneOptions = [
  { value: 'concise', label: '简洁' },
  { value: 'standard', label: '标准 conference' },
]

const figureDatasetOptions = [
  { value: 'results_main.csv', label: 'results_main.csv（演示）' },
  { value: 'ablation_B.csv', label: 'ablation_B.csv（演示）' },
  { value: 'throughput.csv', label: 'throughput.csv（演示）' },
]

const captionLangOptions = [
  { value: 'en', label: 'English' },
  { value: 'zh', label: '中文' },
]

function toggleWritingSection(value: string, checked: boolean) {
  const set = new Set(writeForm.sections)
  if (checked) set.add(value)
  else set.delete(value)
  writeForm.sections = [...set]
}

function isWritingSectionChecked(value: string) {
  return writeForm.sections.includes(value)
}

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
  const uri = kind === 'pdf' ? item.pdf_uri : item.word_uri
  if (!literatureReviewHasAttachmentUri(uri)) return 'AI生成综述'
  return kind === 'pdf' ? '下载 PDF' : '下载 Word'
}

function onLitReviewViewPrimaryAction() {
  const item = litReviewViewItem.value
  const kind = litReviewViewDownloadKind.value
  if (!item || !kind) return
  const uri = kind === 'pdf' ? item.pdf_uri : item.word_uri
  if (!literatureReviewHasAttachmentUri(uri)) {
    aiGenerateLitReviewBundleDocuments()
    return
  }
  const name = literatureReviewAttachmentFileName(item, kind)
  ElMessage.info(`下载「${name}」功能即将接入`)
}

function onLitReviewAttachmentPreviewPrimary() {
  const ctx = litReviewAttachmentPreview.value
  if (!ctx) return
  if (ctx.kind === 'ppt' && !literatureReviewCanDownloadPpt(ctx.item)) {
    aiGenerateLitReviewBundlePpt()
    return
  }
  const name = literatureReviewAttachmentFileName(ctx.item, ctx.kind)
  ElMessage.info(`下载「${name}」功能即将接入`)
}

async function openExpPlanView(item: PaperExperimentPlanItem) {
  const ms = manuscriptId.value
  const needsDetail = !item.content_medium?.trim() && ms && /^\d+$/.test(ms)
  if (needsDetail) {
    try {
      expPlanViewItem.value = { ...item }
      const detail = await fetchExperimentPlanDetail(ms, item.id)
      expPlanViewItem.value = detail
      return
    } catch (e) {
      const msg = e instanceof Error ? e.message : '加载实验方案失败'
      ElMessage.error(msg)
    }
  }
  expPlanViewItem.value = item
}

function closeExpPlanView() {
  expPlanViewItem.value = null
}

function resetExpPlanUploadInput() {
  expPlanUploadFile.value = null
  if (expPlanUploadInputRef.value) expPlanUploadInputRef.value.value = ''
}

function openExpPlanUploadDialog(item: PaperExperimentPlanItem) {
  expPlanUploadPending.value = item
  resetExpPlanUploadInput()
}

function closeExpPlanUploadDialog() {
  if (expPlanUploadSubmitting.value) return
  expPlanUploadPending.value = null
  resetExpPlanUploadInput()
}

function openExpPlanUploadPicker() {
  if (expPlanUploadSubmitting.value) return
  expPlanUploadInputRef.value?.click()
}

function onExpPlanUploadFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  expPlanUploadFile.value = input.files?.[0] ?? null
}

function onExpPlanUploadDrop(ev: DragEvent) {
  if (expPlanUploadSubmitting.value) return
  const file = ev.dataTransfer?.files?.[0]
  if (file) expPlanUploadFile.value = file
}

async function submitExpPlanUpload() {
  const item = expPlanUploadPending.value
  const file = expPlanUploadFile.value
  const ms = manuscriptId.value
  if (!item || !file || !ms) {
    ElMessage.warning('请选择文件')
    return
  }
  expPlanUploadSubmitting.value = true
  try {
    const updated = await uploadExperimentPlanData(ms, item.id, file)
    ElMessage.success('实验数据已上传')
    const idx = expPlanItems.value.findIndex((it) => it.id === item.id)
    if (idx >= 0) {
      expPlanItems.value = [
        ...expPlanItems.value.slice(0, idx),
        updated,
        ...expPlanItems.value.slice(idx + 1),
      ]
    }
    closeExpPlanUploadDialog()
  } catch (e) {
    const msg = e instanceof Error ? e.message : '上传失败'
    ElMessage.error(msg)
  } finally {
    expPlanUploadSubmitting.value = false
  }
}

function experimentDataFileLabel(item: PaperExperimentPlanItem): string {
  const uri = item.experiment_data_uri?.trim()
  if (!uri) return ''
  const parts = uri.split('/')
  return parts[parts.length - 1] ?? uri
}

async function openLinkedLiteratureReviewFromExpPlan(item: PaperExperimentPlanItem) {
  const lrId = item.literature_review_id?.trim()
  const ms = manuscriptId.value
  if (!lrId || !ms) {
    ElMessage.warning('未关联文献综述')
    return
  }
  expPlanLitReviewLoading.value = true
  try {
    const detail = await fetchLiteratureReviewDetail(ms, lrId)
    openLitReviewView(detail)
  } catch (e) {
    const msg = e instanceof Error ? e.message : '加载文献综述失败'
    ElMessage.error(msg)
  } finally {
    expPlanLitReviewLoading.value = false
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
    await reloadExperimentPlans()
    if (!expPlanItems.value.length) {
      ElMessage.warning('暂无实验方案，请先在「文献综述」等待生成完成后点击「查看实验方案」')
      return false
    }
    return true
  }
  if (id === 'paper-writing' && writeForm.sections.length === 0) {
    ElMessage.warning('请至少选择一个章节')
    return false
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
                  class="paper-btn-primary paper-btn--compact"
                  @click="openLitReviewBundleView(item)"
                >
                  管理
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

    <section v-else-if="!expPlanItems.length" class="wf-panel">
      <h2 class="wf-title wf-title--module-head">实验方案</h2>
      <p class="wf-lead">
        当前论文尚无实验方案。请在「文献综述」中等待方案生成后，点击「查看实验方案」进入。
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
                    上传实验数据
                  </button>
                  <span class="wf-lit-col-head-label">操作</span>
                </span>
                <button
                  type="button"
                  tabindex="-1"
                  aria-hidden="true"
                  class="paper-btn-primary paper-btn--compact wf-lit-width-ruler"
                >
                  查看实验方案
                </button>
                <button
                  type="button"
                  tabindex="-1"
                  aria-hidden="true"
                  class="paper-btn-primary paper-btn--compact wf-lit-width-ruler"
                >
                  查看文献综述
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
                  @click="openExpPlanUploadDialog(item)"
                >
                  上传实验数据
                </button>
                <button
                  type="button"
                  class="paper-btn-primary paper-btn--compact"
                  @click="openExpPlanView(item)"
                >
                  查看实验方案
                </button>
                <button
                  type="button"
                  class="paper-btn-primary paper-btn--compact"
                  :disabled="expPlanLitReviewLoading || !item.literature_review_id?.trim()"
                  @click="openLinkedLiteratureReviewFromExpPlan(item)"
                >
                  查看文献综述
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </section>

    <Teleport to="body">
      <div
        v-if="expPlanUploadPending"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-exp-plan-upload-title"
        @click.self="closeExpPlanUploadDialog"
        @keydown.escape="closeExpPlanUploadDialog"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="wf-exp-plan-upload-title" class="paper-modal-title">上传实验数据</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="expPlanUploadSubmitting"
              @click="closeExpPlanUploadDialog"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <p class="wf-lit-delete-lead">
              「{{ experimentPlanRowTitle(expPlanUploadPending) }}」
            </p>
            <p v-if="experimentDataFileLabel(expPlanUploadPending)" class="wf-meta">
              已上传：{{ experimentDataFileLabel(expPlanUploadPending) }}（再次上传将覆盖）
            </p>
            <div
              class="wf-exp-data-drop"
              :class="{
                'wf-exp-data-drop--has-file': !!expPlanUploadFile,
                'wf-exp-data-drop--disabled': expPlanUploadSubmitting,
              }"
              role="button"
              tabindex="0"
              @click="openExpPlanUploadPicker"
              @keydown.enter.prevent="openExpPlanUploadPicker"
              @keydown.space.prevent="openExpPlanUploadPicker"
              @dragover.prevent
              @drop.prevent="onExpPlanUploadDrop"
            >
              <input
                ref="expPlanUploadInputRef"
                type="file"
                class="wf-exp-data-file-input"
                :disabled="expPlanUploadSubmitting"
                @change="onExpPlanUploadFileChange"
              />
              <p v-if="!expPlanUploadFile" class="wf-exp-data-drop-hint">点击或拖拽文件到此处</p>
              <p v-else class="wf-exp-data-drop-name">{{ expPlanUploadFile.name }}</p>
            </div>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="expPlanUploadSubmitting || !expPlanUploadFile"
              @click="submitExpPlanUpload"
            >
              {{ expPlanUploadSubmitting ? '上传中…' : '上传' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="expPlanUploadSubmitting"
              @click="closeExpPlanUploadDialog"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>
  </div>

  <!-- 论文写作 -->
  <div v-else-if="moduleId === 'paper-writing'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">目标 venue</span>
          <input v-model="writeForm.venue" type="text" class="wf-input" />
        </label>
        <label class="wf-field">
          <span class="wf-label">文风</span>
          <PaperSelect v-model="writeForm.tone" :options="writeToneOptions" />
        </label>
      </div>
      <div class="wf-field wf-field--block">
        <span class="wf-label">生成章节</span>
        <div class="wf-check-group">
          <label v-for="opt in PAPER_WRITING_SECTION_OPTIONS" :key="opt.value" class="wf-check wf-check--inline">
            <input
              type="checkbox"
              :checked="isWritingSectionChecked(opt.value)"
              @change="toggleWritingSection(opt.value, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </div>
      <label class="wf-check">
        <input v-model="writeForm.useLitReview" type="checkbox" />
        <span>引用已有文献综述 artifact</span>
      </label>
      <label class="wf-check">
        <input v-model="writeForm.useReviewArtifact" type="checkbox" />
        <span>参考写前审查意见修订 Experiments 表述</span>
      </label>
    </section>
    <section v-if="resultVisible['paper-writing']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">稿件（演示）</h2>
      <p class="wf-kv"><span>标题</span>{{ DEMO_MANUSCRIPT.title }}</p>
      <table class="wf-table">
        <thead>
          <tr>
            <th>章节</th>
            <th>字数</th>
            <th>状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in DEMO_MANUSCRIPT.sections" :key="s.name">
            <td>{{ s.name }}</td>
            <td>{{ s.words }}</td>
            <td>
              <span class="wf-tag" :class="s.status === 'done' ? 'wf-tag--ok' : 'wf-tag--draft'">{{
                s.status
              }}</span>
            </td>
          </tr>
        </tbody>
      </table>
      <p class="wf-meta">引用 {{ DEMO_MANUSCRIPT.cites }} · 已验证 {{ DEMO_MANUSCRIPT.verifiedCites }}</p>
      <pre class="wf-pre">{{ DEMO_MANUSCRIPT.excerpt }}</pre>
    </section>
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

  <!-- 论文审查 -->
  <div v-else-if="moduleId === 'manuscript-analysis'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <p class="wf-lead">写作完成后对全文做投稿前全面分析。</p>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">分析深度</span>
          <PaperSelect v-model="analysisForm.depth" :options="auditOptions" />
        </label>
      </div>
      <label class="wf-check">
        <input v-model="analysisForm.includeFigures" type="checkbox" />
        <span>纳入图表与表（{{ DEMO_FIGURES.files.length }} 个 demo figure）</span>
      </label>
      <label class="wf-check">
        <input v-model="analysisForm.includeBib" type="checkbox" />
        <span>纳入参考文献门禁结果</span>
      </label>
    </section>
    <section v-if="resultVisible['manuscript-analysis']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">论文审查报告（演示）</h2>
      <p class="wf-score-big">总评 {{ DEMO_MANUSCRIPT_ANALYSIS.overall }} / 10</p>
      <p class="wf-meta">{{ DEMO_MANUSCRIPT_ANALYSIS.recommendation }}</p>
      <div class="wf-dim-grid">
        <div v-for="d in DEMO_MANUSCRIPT_ANALYSIS.dimensions" :key="d.name" class="wf-dim">
          <span>{{ d.name }}</span>
          <strong>{{ d.score }}</strong>
        </div>
      </div>
      <h3 class="wf-subtitle">Must fix</h3>
      <ul class="wf-list">
        <li v-for="(m, i) in DEMO_MANUSCRIPT_ANALYSIS.mustFix" :key="i">{{ m }}</li>
      </ul>
      <p class="wf-callout wf-callout--warn"><strong>Kill argument：</strong>{{ DEMO_MANUSCRIPT_ANALYSIS.kill }}</p>
    </section>
  </div>

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

          <template v-else-if="litReviewBundleView.tab === 'experiment-plan'">
            <p v-if="litReviewBundleView.expPlanLoading" class="wf-meta">正在加载实验方案…</p>
            <p v-else-if="litReviewBundleView.expPlanError" class="wf-artifact-empty">
              {{ litReviewBundleView.expPlanError }}
            </p>
            <p
              v-else-if="isLiteratureReviewGeneratingExperimentPlan(litReviewBundleView.item.status)"
              class="wf-artifact-empty"
            >
              实验方案生成中，请稍候刷新
            </p>
            <p
              v-else-if="!literatureReviewHasExperimentPlan(litReviewBundleView.item)"
              class="wf-artifact-empty"
            >
              暂无关联实验方案
            </p>
            <template v-else-if="litReviewBundleView.expPlan">
              <p class="wf-meta wf-lit-meta">
                <span>{{ formatLitReviewCreatedAt(litReviewBundleView.expPlan.created_at) }}</span>
                <span>·</span>
                <span
                  class="wf-status-pill wf-status-pill--inline"
                  :class="`wf-status-pill--${experimentPlanStatusTone(litReviewBundleView.expPlan.status)}`"
                >
                  {{ formatExperimentPlanStatus(litReviewBundleView.expPlan.status) }}
                </span>
              </p>
              <p v-if="litReviewBundleView.expPlan.summary?.trim()" class="wf-lit-summary">
                {{ litReviewBundleView.expPlan.summary }}
              </p>
              <pre
                v-if="litReviewBundleView.expPlan.content_medium?.trim()"
                class="wf-pre wf-pre--lit"
              >{{ litReviewBundleView.expPlan.content_medium }}</pre>
              <p v-else class="wf-artifact-empty">暂无正文（content_medium 为空）</p>
            </template>
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

          <template v-else-if="litReviewBundleView.tab === 'paper'">
            <p class="wf-kv wf-lit-bundle-paper-title">
              <span>标题</span>{{ DEMO_MANUSCRIPT.title }}
            </p>
            <table class="wf-table wf-table--compact wf-lit-bundle-paper-table">
              <thead>
                <tr>
                  <th>章节</th>
                  <th>字数</th>
                  <th>状态</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="s in DEMO_MANUSCRIPT.sections" :key="s.name">
                  <td>{{ s.name }}</td>
                  <td>{{ s.words }}</td>
                  <td>{{ s.status === 'done' ? '已完成' : '草稿' }}</td>
                </tr>
              </tbody>
            </table>
            <p class="wf-meta">
              引用 {{ DEMO_MANUSCRIPT.cites }} · 已验证 {{ DEMO_MANUSCRIPT.verifiedCites }}
            </p>
            <pre class="wf-pre wf-pre--lit">{{ DEMO_MANUSCRIPT.excerpt }}</pre>
            <p class="wf-artifact-empty wf-lit-bundle-paper-hint">论文正文接入后将展示当前论文稿件</p>
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
        <footer class="paper-message-box__btns wf-lit-view-btns">
          <button
            v-if="litReviewBundleView.tab === 'literature-review'"
            type="button"
            class="paper-btn-primary"
            @click="onLitReviewBundleLiteratureReviewPrimary"
          >
            {{
              literatureReviewCanDownloadDocuments(litReviewBundleView.item)
                ? '下载'
                : 'AI生成综述'
            }}
          </button>
          <button
            v-if="litReviewBundleView.tab === 'ppt'"
            type="button"
            class="paper-btn-primary"
            @click="onLitReviewBundlePptPrimary"
          >
            {{
              literatureReviewCanDownloadPpt(litReviewBundleView.item)
                ? '下载 PPT'
                : 'AI生成汇报 PPT'
            }}
          </button>
          <button
            type="button"
            :class="
              litReviewBundleHasDownloadAction(litReviewBundleView.tab)
                ? 'paper-btn-secondary'
                : 'paper-btn-primary'
            "
            @click="closeLitReviewBundleView"
          >
            关闭
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
            v-if="litReviewViewDownloadKind"
            type="button"
            class="paper-btn-primary"
            @click="onLitReviewViewPrimaryAction"
          >
            {{
              litReviewViewDownloadLabel(litReviewViewItem, litReviewViewDownloadKind)
            }}
          </button>
          <button
            type="button"
            :class="litReviewViewDownloadKind ? 'paper-btn-secondary' : 'paper-btn-primary'"
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
          <button type="button" class="paper-btn-primary" @click="onLitReviewAttachmentPreviewPrimary">
            {{
              literatureReviewCanDownloadPpt(litReviewAttachmentPreview.item)
                ? '下载 PPT'
                : 'AI生成汇报 PPT'
            }}
          </button>
          <button type="button" class="paper-btn-secondary" @click="closeLitReviewAttachmentPreview">
            关闭
          </button>
        </footer>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="expPlanViewItem"
      class="paper-modal-overlay wf-lit-view-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="wf-exp-plan-view-title"
      @click.self="closeExpPlanView"
      @keydown.escape="closeExpPlanView"
    >
      <div class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-view-panel" @click.stop>
        <header class="paper-modal-header wf-lit-view-header">
          <div class="wf-lit-view-heading">
            <p class="wf-lit-view-kind">实验方案</p>
            <h2 id="wf-exp-plan-view-title" class="paper-modal-title wf-lit-view-title">
              {{ experimentPlanRowTitle(expPlanViewItem) }}
            </h2>
          </div>
          <button type="button" class="paper-modal-close" aria-label="关闭" @click="closeExpPlanView">
            ×
          </button>
        </header>
        <div class="paper-modal-body wf-lit-view-body">
          <p class="wf-meta wf-lit-meta">
            <span>{{ formatLitReviewCreatedAt(expPlanViewItem.created_at) }}</span>
            <span>·</span>
            <span
              class="wf-status-pill wf-status-pill--inline"
              :class="`wf-status-pill--${experimentPlanStatusTone(expPlanViewItem.status)}`"
            >
              {{ formatExperimentPlanStatus(expPlanViewItem.status) }}
            </span>
          </p>
          <p v-if="expPlanViewItem.summary?.trim()" class="wf-lit-summary">
            {{ expPlanViewItem.summary }}
          </p>
          <pre v-if="expPlanViewItem.content_medium?.trim()" class="wf-pre wf-pre--lit">{{
            expPlanViewItem.content_medium
          }}</pre>
          <p v-else class="wf-artifact-empty">暂无正文（content_medium 为空）</p>
        </div>
        <footer class="paper-message-box__btns wf-lit-view-btns">
          <button type="button" class="paper-btn-primary" @click="closeExpPlanView">关闭</button>
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
</style>
