<script setup lang="ts">
import { computed, nextTick, onUnmounted, reactive, ref, watch } from 'vue'
import { loadWorkbenchModuleContent } from '@/composables/loadWorkbenchModule'
import {
  reloadTopicDiscoveryFormOptions,
  useTopicDiscoveryFormOptions,
} from '@/composables/useTopicDiscoveryFormOptions'
import {
  postTopicDiscoveryRun,
  postTopicDiscoveryConfirmRound,
  fetchCurrentTopicDiscoveryRun,
  cancelTopicDiscoveryRun,
  commitTopicDiscoveryManuscript,
  parseRetrieveLiteratureLinks,
  applyStoredTopicRunInputToForm,
  formatVenueTextFromCodes,
  inferVenueCodesFromText,
  isTopicRunTerminalCompleted,
  topicConfirmLiteratureReviewId,
  experimentPlanPreviewFromPayload,
  type TopicDiscoveryRunResponse,
  type TopicDiscoveryStepDTO,
} from '@/api/topicDiscovery'
import {
  fetchPaperManuscripts,
  createPaperManuscript,
  setCurrentPaperManuscript,
  type PaperManuscriptListResponse,
} from '@/api/manuscripts'
import { ElMessage, ElTooltip } from 'element-plus'
import { paperConfirm } from '@/utils/paperDialog'
import {
  DEFAULT_ENV_PREFERENCE,
  ENV_PREFERENCE_STORAGE_KEY,
  DEFAULT_TOPIC_DISCOVERY,
  formatTopicDirectionText,
  parseTopicKeywords,
  PAPER_MODULE_GROUPS,
  PAPER_MODULES,
  appendOperationLog,
  TOPIC_DISCOVERY_FLOW_STEPS,
  getPaperModuleMeta,
  type EnvironmentPreferenceForm,
  type PaperManuscriptItem,
  type PaperModuleId,
  type TopicCheckpointKey,
  type TopicDiscoveryForm,
  type TopicFlowStepStatus,
} from './types'
import {
  arisIdeaDiscoveryPhaseRows,
  ideaDiscoveryUiGroupsLabelChain,
  resolveIdeaDiscoveryUIGroups,
  arisPhaseStatusLabel,
  arisPhaseAwaitingConfirmLabel,
  isAnyIdeaDiscoveryPhaseExecuting,
  isIdeaDiscoveryArisCheckpoint,
  type ArisPhaseUiRow,
} from './arisIdeaDiscoveryPhases'
import { DEMO_MODULE_TOKEN_ESTIMATES } from './demoOperationLogs'
import PaperModuleNavIcon from './PaperModuleNavIcon.vue'
import PaperMyManuscriptsPanel from './PaperMyManuscriptsPanel.vue'
import PaperUserLiteraturePanel from './PaperUserLiteraturePanel.vue'
import PaperFigureManagementPanel from './PaperFigureManagementPanel.vue'
import PaperInviteRebatePanel from './PaperInviteRebatePanel.vue'
import PaperPersonalCenterPanel from './PaperPersonalCenterPanel.vue'
import PaperEnvironmentPanel from './PaperEnvironmentPanel.vue'
import PaperCourseReportPanel from './PaperCourseReportPanel.vue'
import PaperSubmissionHubPanel from './PaperSubmissionHubPanel.vue'
import PaperSelect from './PaperSelect.vue'
import PaperMultiSelect from './PaperMultiSelect.vue'
import PaperWorkflowPanels from './PaperWorkflowPanels.vue'
import PaperTopicUserLiteraturePickDialog from './PaperTopicUserLiteraturePickDialog.vue'

type TopicFlowStepRuntime = {
  stageCode: string
  label: string
  checkpointKey?: TopicCheckpointKey
  status: TopicFlowStepStatus
}

type TopicCheckpointView = {
  key: TopicCheckpointKey
  title: string
  lines: string[]
}

type TopicRunDemo = {
  status: 'idle' | 'running' | 'checkpoint' | 'completed' | 'failed'
  steps: TopicFlowStepRuntime[]
  checkpoint: TopicCheckpointView | null
}

function createIdleTopicRun(): TopicRunDemo {
  return {
    status: 'idle',
    steps: TOPIC_DISCOVERY_FLOW_STEPS.map((s) => ({
      stageCode: s.stageCode,
      label: s.label,
      checkpointKey: s.checkpointKey,
      status: 'pending' as TopicFlowStepStatus,
    })),
    checkpoint: null,
  }
}

const CHECKPOINT_COPY: Record<TopicCheckpointKey, Omit<TopicCheckpointView, 'key'>> = {
  retrieve_ready: {
    title: '检索与验真入库结果',
    lines: [
      '命中 86 篇 · 验真通过 79 篇 · 待补全 7 篇（演示）',
      '来源：arXiv / OpenAlex / Semantic Scholar 已去重',
      '覆盖缺口：2025 Q1 同方向预印本偏少，可加深检索或补 Crossref',
    ],
  },
  idea_discovery_ready: {
    title: 'Idea Discovery 段间检查点',
    lines: [],
  },
  topic_discovery_ready: {
    title: '选题发现产出',
    lines: ['（旧版 run）'],
  },
  topic_confirm_ready: {
    title: '确认并入库',
    lines: ['将 refine 产物写入文献综述与实验方案表'],
  },
}

const LIT_REVIEW_RUN_STORAGE_KEY = 'atm:paper:lit-review-run:v1'
const EXPERIMENT_PLAN_RUN_STORAGE_KEY = 'atm:paper:experiment-plan-run:v1'

const MANUSCRIPTS_STORAGE_KEY = 'atm:paper:manuscripts:v1'
const LEGACY_PROJECTS_STORAGE_KEY = 'atm:paper:projects:v1'
const ACTIVE_MODULE_STORAGE_KEY = 'atm:paper:active-module:v1'

function loadStoredActiveModule(): PaperModuleId {
  try {
    const raw = localStorage.getItem(ACTIVE_MODULE_STORAGE_KEY)?.trim()
    if (raw && PAPER_MODULES.some((m) => m.id === raw)) {
      return raw as PaperModuleId
    }
  } catch {
    /* ignore */
  }
  return 'topic-discovery'
}

function persistActiveModule(id: PaperModuleId) {
  try {
    localStorage.setItem(ACTIVE_MODULE_STORAGE_KEY, id)
  } catch {
    /* ignore */
  }
}

const topicLastRunByMs = ref<Record<string, TopicDiscoveryRunResponse>>({})

const activeModule = ref<PaperModuleId>(loadStoredActiveModule())

/** 首屏与 activeModule immediate prepare 对齐，避免子面板抢先出现模块内 loading */
const moduleContentLoading = ref(true)
const moduleContentKey = ref(0)
let moduleSwitchSeq = 0
/** 仅在不同模块间切换时递增 key，避免刷新同一模块时 remount 子组件导致重复拉数 */
let lastPreparedModuleId: PaperModuleId | undefined

const showGenericModuleLoading = computed(
  () =>
    moduleContentLoading.value &&
    activeModule.value !== 'topic-discovery' &&
    activeModule.value !== 'literature-review' &&
    activeModule.value !== 'experiment-planning' &&
    activeModule.value !== 'paper-writing' &&
    activeModule.value !== 'paper-submission' &&
    activeModule.value !== 'user-literature' &&
    activeModule.value !== 'figure-generation' &&
    activeModule.value !== 'course-report',
)

const topicDiscoveryPageLoading = computed(
  () => activeModule.value === 'topic-discovery' && moduleContentLoading.value,
)

let manuscriptsBootstrapOnce: Promise<void> | null = null

function ensureManuscriptsReady(): Promise<void> {
  if (!manuscriptsBootstrapOnce) {
    manuscriptsBootstrapOnce = (async () => {
      const ok = await loadManuscriptsFromServer(undefined, { initial: true })
      if (!ok) loadManuscriptsFromStorage()
    })()
  }
  return manuscriptsBootstrapOnce
}

async function prepareModuleContent(moduleId: PaperModuleId) {
  const seq = ++moduleSwitchSeq
  moduleContentLoading.value = true
  try {
    await ensureManuscriptsReady()
    await loadWorkbenchModuleContent(moduleId, {
      reloadManuscriptsFromStorage: loadManuscriptsFromStorage,
    })
    if (moduleId === 'topic-discovery') {
      const msId =
        (await ensureManuscriptWhenEmpty()) ?? activeManuscriptId.value
      if (msId && /^\d+$/.test(msId)) await hydrateTopicRunFromServer(msId)
    }
  } finally {
    if (seq === moduleSwitchSeq) {
      if (lastPreparedModuleId !== undefined && lastPreparedModuleId !== moduleId) {
        moduleContentKey.value++
      }
      lastPreparedModuleId = moduleId
      moduleContentLoading.value = false
    }
  }
  if (seq !== moduleSwitchSeq) return
  await nextTick()
  if (moduleId === 'personal-center') {
    await personalCenterPanelRef.value?.reloadFromMenu?.()
  }
}

function selectModule(id: PaperModuleId) {
  activeModule.value = id
  persistActiveModule(id)
}

/** 文献综述「查看实验方案」：切模块并打开对应方案弹窗 */
const pendingOpenExperimentPlanId = ref<string | null>(null)
const pendingOpenLiteratureReviewId = ref<string | null>(null)

function onNavigateModule(id: PaperModuleId, openExperimentPlanId?: string) {
  const planId = openExperimentPlanId?.trim()
  pendingOpenExperimentPlanId.value = planId || null
  selectModule(id)
}

function onConsumedPendingExperimentPlan() {
  pendingOpenExperimentPlanId.value = null
}

function onConsumedPendingLiteratureReview() {
  pendingOpenLiteratureReviewId.value = null
}

function openLiteratureReviewAfterTopicConfirm(data: TopicDiscoveryRunResponse) {
  const reviewId = topicConfirmLiteratureReviewId(data)
  pendingOpenLiteratureReviewId.value = reviewId
  selectModule('literature-review')
}

const running = ref(false)
const topicTerminateLoading = ref(false)
const topicRunToken = ref(0)
/** POST 新建 run 完成前，背景轮询不应用 GET /current（避免仍读到上一轮 failed） */
const topicRunStartInFlight = ref(false)
/** 检查点「确认并继续」请求进行中 */
const topicContinueLoading = ref(false)
const topicConfirmRoundLoading = ref(false)
const topicSelectedRound = ref<number | null>(null)
/** continue 后乐观标 running 的阶段，避免 POST/轮询仍返回 checkpoint 把转圈打掉 */
const topicContinueOptimisticStage = ref<string | null>(null)
/** 本轮选题 run 由用户启动/继续，完成后弹窗引导去文献综述 */
const topicRunAwaitingCompletionPrompt = ref(false)
const topicCompletionPromptVisible = ref(false)
const topicRunsByManuscript = ref<Record<string, TopicRunDemo>>({})
const litReviewDoneByManuscript = ref<Record<string, boolean>>({})
const experimentPlanDoneByManuscript = ref<Record<string, boolean>>({})
const workflowPanelsRef = ref<InstanceType<typeof PaperWorkflowPanels> | null>(null)
const userLiteraturePanelRef = ref<InstanceType<typeof PaperUserLiteraturePanel> | null>(null)
const figureManagementPanelRef = ref<InstanceType<typeof PaperFigureManagementPanel> | null>(null)
const paperSubmissionHubRef = ref<InstanceType<typeof PaperSubmissionHubPanel> | null>(null)
const paperMainRef = ref<HTMLElement | null>(null)
const topicFlowPanelRef = ref<HTMLElement | null>(null)

const topicForm = reactive<TopicDiscoveryForm>({ ...DEFAULT_TOPIC_DISCOVERY })

const topicUserLitPickerOpen = ref(false)

const envPreference = reactive<EnvironmentPreferenceForm>({ ...DEFAULT_ENV_PREFERENCE })

const {
  disciplineSelectOptions,
  intensityOptions,
  auditOptions,
  venueOptions,
  venueSelectOptions,
  ready: topicFormOptionsReady,
  applyCatalogIntensityAuditDefaults,
  literatureSourceOptions,
  applyDefaultLiteratureSourceCodes,
  applyDefaultLiteratureSources,
  validateTopicLiteratureSourceCodes,
} = useTopicDiscoveryFormOptions()

const envPreferenceFromStorage = ref(false)

const manuscripts = ref<PaperManuscriptItem[]>([])
const activeManuscriptId = ref<string>('')
const manuscriptsLoading = ref(false)
/** 切换当前论文请求进行中；阻止列表刷新把选中项打回服务端旧 is_current */
const manuscriptSwitching = ref(false)
const manuscriptSwitchPendingId = ref<string | null>(null)

const createManuscriptDialogVisible = ref(false)
const createManuscriptSubmitting = ref(false)
const createManuscriptForm = reactive<{ disciplineCode: string; title: string; language: 'zh' | 'en' }>({
  disciplineCode: '',
  title: '',
  language: 'en',
})
let createManuscriptDialogResolve: ((ok: boolean) => void) | null = null

const createManuscriptUploadFile = ref<File | null>(null)
const createManuscriptUploadInputRef = ref<HTMLInputElement | null>(null)

function resetCreateManuscriptUpload() {
  createManuscriptUploadFile.value = null
  if (createManuscriptUploadInputRef.value) createManuscriptUploadInputRef.value.value = ''
}

function openCreateManuscriptUploadPicker() {
  if (createManuscriptSubmitting.value) return
  createManuscriptUploadInputRef.value?.click()
}

function onCreateManuscriptUploadChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  createManuscriptUploadFile.value = file
  if (file && !createManuscriptForm.title.trim()) {
    createManuscriptForm.title = file.name.replace(/\.[^.]+$/, '')
  }
}

function onCreateManuscriptUploadDrop(e: DragEvent) {
  if (createManuscriptSubmitting.value) return
  e.preventDefault()
  const file = e.dataTransfer?.files?.[0] ?? null
  if (!file) return
  createManuscriptUploadFile.value = file
  if (!createManuscriptForm.title.trim()) {
    createManuscriptForm.title = file.name.replace(/\.[^.]+$/, '')
  }
}

async function openCreateManuscriptDialog(): Promise<boolean> {
  await reloadTopicDiscoveryFormOptions()
  createManuscriptForm.disciplineCode = ''
  createManuscriptForm.title = ''
  createManuscriptForm.language = 'en'
  resetCreateManuscriptUpload()
  createManuscriptDialogVisible.value = true
  return new Promise((resolve) => {
    createManuscriptDialogResolve = resolve
  })
}

function onAddManuscriptClick() {
  void openCreateManuscriptDialog()
}

function closeCreateManuscriptDialog(ok: boolean) {
  createManuscriptDialogVisible.value = false
  createManuscriptDialogResolve?.(ok)
  createManuscriptDialogResolve = null
}

async function submitCreateManuscriptDialog() {
  const title = createManuscriptForm.title.trim()
  if (!title) {
    ElMessage.warning('请填写论文名称')
    return
  }
  if (!createManuscriptForm.disciplineCode) {
    ElMessage.warning('请选择学科')
    return
  }
  createManuscriptSubmitting.value = true
  try {
    const data = await createPaperManuscript(
      title,
      createManuscriptForm.disciplineCode,
      createManuscriptForm.language,
    )
    applyManuscriptListFromApi(data)
    topicForm.disciplineCode = createManuscriptForm.disciplineCode
    const pendingUpload = createManuscriptUploadFile.value
    closeCreateManuscriptDialog(true)
    ElMessage.success('论文已创建')
    if (pendingUpload) {
      ElMessage.info(`上传论文「${pendingUpload.name}」功能即将接入`)
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '创建论文失败')
  } finally {
    createManuscriptSubmitting.value = false
  }
}

let ensureManuscriptInFlight: Promise<string | null> | null = null

async function fetchManuscriptListOrNull(): Promise<PaperManuscriptListResponse | null> {
  try {
    return await fetchPaperManuscripts()
  } catch {
    return null
  }
}

/** 无论文则弹「创建论文」；有论文则同步 is_current 到侧栏 */
async function ensureManuscriptWhenEmpty(options?: { warnIfUnset?: boolean }): Promise<string | null> {
  if (ensureManuscriptInFlight) return ensureManuscriptInFlight

  ensureManuscriptInFlight = (async () => {
    let list = await fetchManuscriptListOrNull()

    if (list?.items?.length) {
      applyManuscriptListFromApi(list)
    } else if (!createManuscriptDialogVisible.value) {
      const ok = await openCreateManuscriptDialog()
      if (!ok) return null
      list = await fetchManuscriptListOrNull()
      if (!list?.items?.length) return null
      applyManuscriptListFromApi(list)
    } else {
      return null
    }

    const msId = activeManuscriptId.value
    if (!/^\d+$/.test(msId)) {
      if (options?.warnIfUnset) ElMessage.warning('请选择当前论文')
      return null
    }
    return msId
  })()

  try {
    return await ensureManuscriptInFlight
  } finally {
    ensureManuscriptInFlight = null
  }
}

const retrieveLiteratureLinks = computed(() =>
  parseRetrieveLiteratureLinks(topicLastRunByMs.value[activeManuscriptId.value]),
)

function loadEnvFromStorage() {
  try {
    const raw = localStorage.getItem(ENV_PREFERENCE_STORAGE_KEY)
    if (!raw) return
    envPreferenceFromStorage.value = true
    const data = JSON.parse(raw) as { preference?: EnvironmentPreferenceForm }
    if (data.preference) Object.assign(envPreference, data.preference)
  } catch {
    /* ignore */
  }
}

watch(topicFormOptionsReady, (ready) => {
  if (!ready) return
  if (!envPreferenceFromStorage.value) {
    applyCatalogIntensityAuditDefaults(envPreference)
    applyDefaultLiteratureSourceCodes(envPreference)
    topicForm.intensity = envPreference.intensity
    topicForm.auditLevel = envPreference.auditLevel
  }
  if (topicForm.sourceCodes.length === 0) {
    if (envPreference.literatureSourceCodes.length > 0) {
      topicForm.sourceCodes = [...envPreference.literatureSourceCodes]
    } else {
      applyDefaultLiteratureSources(topicForm)
    }
  }
})

function loadLitReviewFlagsFromStorage() {
  try {
    const raw = localStorage.getItem(LIT_REVIEW_RUN_STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as { byManuscript?: Record<string, boolean> }
    if (data.byManuscript) litReviewDoneByManuscript.value = data.byManuscript
  } catch {
    /* ignore */
  }
}

function persistLitReviewFlags() {
  localStorage.setItem(
    LIT_REVIEW_RUN_STORAGE_KEY,
    JSON.stringify({ byManuscript: litReviewDoneByManuscript.value }),
  )
}

function loadExperimentPlanFlagsFromStorage() {
  try {
    const raw = localStorage.getItem(EXPERIMENT_PLAN_RUN_STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as { byManuscript?: Record<string, boolean> }
    if (data.byManuscript) experimentPlanDoneByManuscript.value = data.byManuscript
  } catch {
    /* ignore */
  }
}

function persistExperimentPlanFlags() {
  localStorage.setItem(
    EXPERIMENT_PLAN_RUN_STORAGE_KEY,
    JSON.stringify({ byManuscript: experimentPlanDoneByManuscript.value }),
  )
}

function clearExperimentPlanDone(msId: string) {
  if (!experimentPlanDoneByManuscript.value[msId]) return
  experimentPlanDoneByManuscript.value = { ...experimentPlanDoneByManuscript.value, [msId]: false }
  persistExperimentPlanFlags()
}

function loadManuscriptsFromStorage() {
  try {
    let raw = localStorage.getItem(MANUSCRIPTS_STORAGE_KEY)
    if (!raw) raw = localStorage.getItem(LEGACY_PROJECTS_STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as {
      manuscripts?: PaperManuscriptItem[]
      activeManuscriptId?: string
      projects?: PaperManuscriptItem[]
      activeProjectId?: string
    }
    const list = data.manuscripts ?? data.projects
    const activeId = data.activeManuscriptId ?? data.activeProjectId
    const serverLike = list?.filter((m) => /^\d+$/.test(m.id)) ?? []
    if (serverLike.length) manuscripts.value = serverLike
    if (activeId && /^\d+$/.test(activeId) && manuscripts.value.some((m) => m.id === activeId)) {
      activeManuscriptId.value = activeId
    }
  } catch {
    /* ignore */
  }
}

function persistManuscripts() {
  localStorage.setItem(
    MANUSCRIPTS_STORAGE_KEY,
    JSON.stringify({
      manuscripts: manuscripts.value,
      activeManuscriptId: activeManuscriptId.value,
    }),
  )
}

function applyManuscriptListFromApi(data: PaperManuscriptListResponse, preferId?: string) {
  manuscripts.value = (data.items ?? []).map((m) => ({
    id: m.id,
    title: m.title,
    venueHint: '',
    status: m.status,
    isCurrent: m.is_current,
    disciplineCode: m.discipline_code,
    disciplineLabel: m.discipline_label,
    draftVersion: m.draft_version,
  }))

  if (!manuscripts.value.length) {
    activeManuscriptId.value = ''
    persistManuscripts()
    return
  }

  const fromCurrent = data.current_manuscript_id ? String(data.current_manuscript_id) : ''
  const fromFlag = data.items?.find((m) => m.is_current)?.id ?? ''
  const pending = manuscriptSwitchPendingId.value
  let nextId = preferId || fromCurrent || fromFlag || manuscripts.value[0].id
  if (pending && manuscripts.value.some((m) => m.id === pending)) {
    nextId = pending
  }
  if (manuscripts.value.some((m) => m.id === nextId)) {
    activeManuscriptId.value = nextId
    ensureTopicRunForManuscript(nextId)
  } else {
    activeManuscriptId.value = manuscripts.value[0].id
    ensureTopicRunForManuscript(activeManuscriptId.value)
  }
  syncTopicDisciplineFromCurrentManuscript()
  persistManuscripts()
}

function syncTopicDisciplineFromCurrentManuscript() {
  const m = currentManuscript.value
  if (m?.disciplineCode) {
    topicForm.disciplineCode = m.disciplineCode
  }
}

async function loadManuscriptsFromServer(
  preferId?: string,
  opts?: { initial?: boolean },
) {
  const initial = opts?.initial === true
  if (initial) manuscriptsLoading.value = true
  try {
    const data = await fetchPaperManuscripts()
    applyManuscriptListFromApi(data, preferId)
    return true
  } catch {
    /* 刷新失败时保留已有列表，避免一点击下拉就被清空 */
    if (initial && manuscripts.value.length === 0) {
      activeManuscriptId.value = ''
    }
    return false
  } finally {
    if (initial) manuscriptsLoading.value = false
  }
}

function applyOptimisticCurrentManuscript(id: string) {
  activeManuscriptId.value = id
  manuscripts.value = manuscripts.value.map((m) => ({
    ...m,
    isCurrent: m.id === id,
  }))
  persistManuscripts()
  syncTopicDisciplineFromCurrentManuscript()
  ensureTopicRunForManuscript(id)
}

async function onManuscriptChange(id: string) {
  if (!id || id === activeManuscriptId.value) return
  if (!/^\d+$/.test(id)) {
    ElMessage.warning('请从列表中选择已创建的论文')
    return
  }
  const previousId = activeManuscriptId.value
  manuscriptSwitchPendingId.value = id
  manuscriptSwitching.value = true
  applyOptimisticCurrentManuscript(id)

  try {
    const data = await setCurrentPaperManuscript(id)
    applyManuscriptListFromApi(data, id)
    ElMessage.success('已切换当前论文')
  } catch (e) {
    applyOptimisticCurrentManuscript(previousId)
    if (isTopicDiscoveryModule.value && previousId) {
      void refreshTopicDiscoveryForManuscript(previousId)
    }
    const msg = e instanceof Error ? e.message : '切换当前论文失败'
    ElMessage.error(msg)
  } finally {
    manuscriptSwitchPendingId.value = null
    manuscriptSwitching.value = false
  }
}

async function onStartWritingFromMyManuscripts(manuscriptId: string) {
  const id = manuscriptId.trim()
  if (!id || !/^\d+$/.test(id)) {
    ElMessage.warning('请从列表中选择已创建的论文')
    return
  }
  if (manuscriptSwitching.value) return
  if (id !== activeManuscriptId.value) {
    await onManuscriptChange(id)
    if (activeManuscriptId.value !== id) return
  }
  selectModule('paper-writing')
}

async function onStartSubmissionFromMyManuscripts(manuscriptId: string) {
  const id = manuscriptId.trim()
  if (!id || !/^\d+$/.test(id)) {
    ElMessage.warning('请从列表中选择已创建的论文')
    return
  }
  if (manuscriptSwitching.value) return
  if (id !== activeManuscriptId.value) {
    await onManuscriptChange(id)
    if (activeManuscriptId.value !== id) return
  }
  selectModule('paper-submission')
}

function onStartRevisionWritingFromSubmission(_payload: { submissionId: string }) {
  selectModule('paper-writing')
  ElMessage.info('请在论文写作中修改正文；修回记录仍在「论文投稿」')
}

function onManuscriptsListUpdate(list: PaperManuscriptItem[]) {
  manuscripts.value = list
  persistManuscripts()
}

loadEnvFromStorage()
loadManuscriptsFromStorage()
loadLitReviewFlagsFromStorage()
loadExperimentPlanFlagsFromStorage()
{
  const fromEnv = inferVenueCodesFromText(
    envPreference.defaultVenueText,
    venueOptions.value,
  )
  if (fromEnv.length) topicForm.venueCodes = fromEnv
}
topicForm.intensity = envPreference.intensity
topicForm.auditLevel = envPreference.auditLevel
topicForm.humanCheckpoint = true

const activeManuscripts = computed(() => manuscripts.value.filter((m) => m.status === 'active'))

const currentManuscript = computed(
  () =>
    manuscripts.value.find((m) => m.id === activeManuscriptId.value) ?? activeManuscripts.value[0],
)

const topicDisciplineUnset = computed(() => {
  const m = currentManuscript.value
  if (!m) return true
  if (m.disciplineLabel?.trim()) return false
  if (m.disciplineCode?.trim()) return false
  return true
})

const topicDisciplineDisplayLabel = computed(() => {
  const m = currentManuscript.value
  if (m?.disciplineLabel) return m.disciplineLabel
  if (m?.disciplineCode) {
    return (
      disciplineSelectOptions.value.find((d) => d.value === m.disciplineCode)?.label ??
      m.disciplineCode
    )
  }
  return '未设置学科'
})

const isMyManuscriptsModule = computed(() => activeModule.value === 'my-manuscripts')
const isUserLiteratureModule = computed(() => activeModule.value === 'user-literature')
const isFigureGenerationModule = computed(() => activeModule.value === 'figure-generation')
const isInviteRebateModule = computed(() => activeModule.value === 'invite-rebate')
const isCourseReportModule = computed(() => activeModule.value === 'course-report')
const isPersonalCenterModule = computed(() => activeModule.value === 'personal-center')
const isEnvironmentModule = computed(() => activeModule.value === 'environment')
const isUtilityModule = computed(
  () =>
    isMyManuscriptsModule.value ||
    isUserLiteratureModule.value ||
    isFigureGenerationModule.value ||
    isEnvironmentModule.value ||
    isInviteRebateModule.value ||
    isCourseReportModule.value ||
    isPaperSubmissionModule.value ||
    isPersonalCenterModule.value,
)
const isTopicDiscoveryModule = computed(() => activeModule.value === 'topic-discovery')

const isLiteratureReviewModule = computed(() => activeModule.value === 'literature-review')
const isExperimentPlanningModule = computed(() => activeModule.value === 'experiment-planning')
const isPaperWritingModule = computed(() => activeModule.value === 'paper-writing')
const isPaperSubmissionModule = computed(() => activeModule.value === 'paper-submission')

/** 仅本篇工作流模块展示顶栏「当前论文」切换（全局模块如我的论文/文献管理不展示） */
const MANUSCRIPT_CONTEXT_MODULE_IDS = [
  'topic-discovery',
  'literature-review',
  'experiment-planning',
  'paper-writing',
  'paper-submission',
] as const satisfies readonly PaperModuleId[]

const showManuscriptContextInHeader = computed(() =>
  (MANUSCRIPT_CONTEXT_MODULE_IDS as readonly PaperModuleId[]).includes(activeModule.value),
)

const manuscriptSelectOptions = computed(() =>
  activeManuscripts.value.map((m) => ({
    value: m.id,
    label: `${m.title}${m.venueHint ? ` · ${m.venueHint}` : ''}`,
  })),
)

const personalCenterPanelRef = ref<InstanceType<typeof PaperPersonalCenterPanel> | null>(null)

function recordModuleOperationLog(
  moduleId: PaperModuleId,
  action: string,
  status: 'success' | 'failed' = 'success',
  note?: string,
) {
  const meta = getPaperModuleMeta(moduleId)
  const est = DEMO_MODULE_TOKEN_ESTIMATES[moduleId]
  const tokensPrompt = status === 'success' ? (est?.prompt ?? 0) : 0
  const tokensCompletion = status === 'success' ? (est?.completion ?? 0) : 0
  appendOperationLog({
    moduleId,
    moduleLabel: meta.label,
    action,
    tokensPrompt,
    tokensCompletion,
    manuscriptId: activeManuscriptId.value,
    manuscriptTitle: currentManuscript.value?.title,
    status,
    note,
  })
  personalCenterPanelRef.value?.reloadLogs()
}

function onEnvironmentSaved() {
  {
    const fromEnv = inferVenueCodesFromText(
      envPreference.defaultVenueText,
      venueOptions.value,
    )
    if (fromEnv.length) topicForm.venueCodes = fromEnv
  }
  topicForm.intensity = envPreference.intensity
  topicForm.auditLevel = envPreference.auditLevel
  topicForm.humanCheckpoint = true
  recordModuleOperationLog('environment', '保存环境配置')
}

const SECONDARY_WORKFLOW_MODULES = [
  'literature-review',
  'experiment-planning',
  'paper-writing',
] as const

const isSecondaryWorkflowModule = computed(() =>
  (SECONDARY_WORKFLOW_MODULES as readonly string[]).includes(activeModule.value),
)

function ensureTopicRunForManuscript(id: string) {
  if (!topicRunsByManuscript.value[id]) {
    topicRunsByManuscript.value = { ...topicRunsByManuscript.value, [id]: createIdleTopicRun() }
  }
}

watch(activeManuscriptId, (id, prev) => {
  if (!id) return
  ensureTopicRunForManuscript(id)
  if (activeModule.value !== 'topic-discovery') return
  // 首次从空写入 id（bootstrap）已在 prepareModuleContent 里 hydrate，避免第二次 loading
  if (!prev) return
  if (id === prev) return
  if (moduleContentLoading.value) {
    void hydrateTopicRunFromServer(id)
    return
  }
  void refreshTopicDiscoveryForManuscript(id)
})

function clampTopicConfirmWhileIdeaDiscoveryActive(
  steps: TopicFlowStepRuntime[],
): TopicFlowStepRuntime[] {
  const idea = steps.find(
    (s) => s.stageCode === 'idea_discovery' || s.stageCode === 'topic_discovery',
  )
  if (!idea || idea.status === 'completed') return steps
  return steps.map((s) =>
    s.stageCode === 'topic_confirm' && s.status === 'running'
      ? { ...s, status: 'pending' as TopicFlowStepStatus }
      : s,
  )
}

const currentTopicRun = computed((): TopicRunDemo => {
  const id = activeManuscriptId.value
  if (!id) return createIdleTopicRun()
  const run = topicRunsByManuscript.value[id] ?? createIdleTopicRun()
  return {
    ...run,
    steps: clampTopicConfirmWhileIdeaDiscoveryActive(run.steps),
  }
})

const topicRunVisible = computed(
  () => isTopicDiscoveryModule.value && currentTopicRun.value.status !== 'idle',
)

function topicStepsCompletedThenPending(
  steps: Array<{ status: TopicFlowStepStatus }>,
): boolean {
  for (let i = 0; i < steps.length - 1; i++) {
    if (steps[i].status === 'completed' && steps[i + 1].status === 'pending') {
      return true
    }
  }
  return false
}

function topicPauseStageFromApi(data: TopicDiscoveryRunResponse): string {
  const fromApi = data.pause_after_stage?.trim()
  if (fromApi) return fromApi
  const disc = data.steps.find(
    (s) => s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery',
  )
  const conf = data.steps.find((s) => s.stage_code === 'topic_confirm')
  if (disc?.status === 'completed' && conf?.status === 'pending') {
    return 'topic_confirm'
  }
  if (!data.human_checkpoint) return ''
  const ordered = TOPIC_DISCOVERY_FLOW_STEPS.map((def) => {
    const st = data.steps.find((s) => s.stage_code === def.stageCode)
    return st ? stepStatusFromApi(st.status) : ('pending' as TopicFlowStepStatus)
  })
  for (let i = 0; i < ordered.length - 1; i++) {
    if (ordered[i] === 'completed' && ordered[i + 1] === 'pending') {
      return TOPIC_DISCOVERY_FLOW_STEPS[i].stageCode
    }
  }
  return ''
}

/** 人工检查点暂停（含 idea_discovery 仍为 running 的 ARIS 段间检查点） */
const topicAtHumanPause = computed(() => {
  if (!topicForm.humanCheckpoint) return false
  if (topicShowFailedPanel.value) return false
  if (currentTopicRun.value.status === 'completed') return false
  const raw = topicDiscoveryApiRun.value
  const ideaDiscActive = raw?.steps.some(
    (s) =>
      (s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery') &&
      s.status === 'running',
  )
  if (ideaDiscActive && isAnyIdeaDiscoveryPhaseExecuting(raw?.steps)) return false
  if (
    topicArisPhaseRows.value.some((r) => r.status === 'waiting_confirm') &&
    !isAnyIdeaDiscoveryPhaseExecuting(raw?.steps)
  ) {
    return true
  }
  if (isIdeaDiscoveryArisCheckpoint(raw?.steps, raw?.run_status)) return true
  if (currentTopicRun.value.status === 'checkpoint') {
    if (ideaDiscActive && isAnyIdeaDiscoveryPhaseExecuting(raw?.steps)) return false
    return true
  }
  if (currentTopicRun.value.steps.some((s) => s.status === 'running')) return false
  if (raw?.steps.some((s) => s.status === 'running')) return false
  if (raw?.run_status === 'checkpoint') return true
  if (raw?.pause_after_stage?.trim()) return true
  if (topicStepsCompletedThenPending(currentTopicRun.value.steps)) return true
  return false
})

const topicRunBusy = computed(() => {
  if (topicAtHumanPause.value) {
    return topicContinueLoading.value
  }
  return (
    topicContinueLoading.value ||
    currentTopicRun.value.steps.some((s) => s.status === 'running') ||
    running.value ||
    currentTopicRun.value.status === 'running'
  )
})

const topicDiscoveryApiRun = computed(
  () => topicLastRunByMs.value[activeManuscriptId.value] ?? null,
)

/** 任一步骤 status=failed（后端落盘） */
const topicHasFailedStep = computed(() =>
  currentTopicRun.value.steps.some((s) => s.status === 'failed'),
)

/** 失败：仅「驳回并返回」 */
const topicShowFailedPanel = computed(
  () => topicHasFailedStep.value || currentTopicRun.value.status === 'failed',
)

/** 第二步完成、等待用户选轮（第三步） */
const topicAwaitingRoundConfirm = computed(() => {
  const raw = topicDiscoveryApiRun.value
  if (!raw) return false
  const disc = raw.steps.find(
    (s) => s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery',
  )
  const conf = raw.steps.find((s) => s.stage_code === 'topic_confirm')
  return disc?.status === 'completed' && conf?.status === 'pending'
})

type TopicDiscoveryIterationView = {
  round: number
  passed: boolean
  auditSummary: string
  litTitle: string
  litSummary: string
  planTitle: string
  planSummary: string
  hasExperimentPlan: boolean
  ideaTitles: string[]
}

const topicDiscoveryIterations = computed((): TopicDiscoveryIterationView[] => {
  const raw = topicDiscoveryApiRun.value
  const disc = raw?.steps.find(
    (s) => s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery',
  )
  const result = disc?.result as Record<string, unknown> | undefined
  const iters = (result?.iterations as unknown[]) ?? []
  const out: TopicDiscoveryIterationView[] = []
  for (const item of iters) {
    if (!item || typeof item !== 'object') continue
    const m = item as Record<string, unknown>
    const round = typeof m.round === 'number' ? m.round : Number(m.round)
    if (!Number.isFinite(round) || round < 1) continue
    const passed = m.passed === true
    const audit = (m.audit ?? {}) as Record<string, unknown>
    const gen = (m.generate ?? {}) as Record<string, unknown>
    const lr = (gen.literature_review ?? {}) as Record<string, unknown>
    const plan = (gen.experiment_plan ?? {}) as Record<string, unknown>
    const ideas = (gen.ideas as Array<{ title?: string }>) ?? []
    const ideaTitles = ideas.map((i) => String(i.title ?? '').trim()).filter(Boolean)
    const planPreview = experimentPlanPreviewFromPayload(plan, ideaTitles)
    out.push({
      round,
      passed,
      auditSummary: String(audit.summary ?? '').trim(),
      litTitle: String(lr.title ?? '').trim(),
      litSummary: String(lr.summary ?? '').trim(),
      planTitle: planPreview.title,
      planSummary: planPreview.summary,
      hasExperimentPlan: planPreview.hasPlan,
      ideaTitles,
    })
  }
  out.sort((a, b) => a.round - b.round)
  return out
})

watch(topicDiscoveryIterations, (rows) => {
  if (!rows.length) {
    topicSelectedRound.value = null
    return
  }
  if (topicSelectedRound.value != null && rows.some((r) => r.round === topicSelectedRound.value)) {
    return
  }
  const passedLast = [...rows].reverse().find((r) => r.passed)
  topicSelectedRound.value = passedLast?.round ?? rows[rows.length - 1].round
})

/** 非失败且在人工暂停点：「终止并返回」+「确认并继续」 */
const topicShowCheckpointPanel = computed(() => {
  if (topicShowFailedPanel.value) return false
  if (topicAwaitingRoundConfirm.value) return false
  const raw = topicDiscoveryApiRun.value
  const ideaDisc = raw?.steps.some(
    (s) => s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery',
  )
  if (ideaDisc) {
    if (!topicArisPhaseRows.value.some((r) => r.status === 'waiting_confirm')) return false
    if (isAnyIdeaDiscoveryPhaseExecuting(raw?.steps)) return false
    return true
  }
  const arisPause = isIdeaDiscoveryArisCheckpoint(raw?.steps, raw?.run_status)
  if (!arisPause) {
    if (currentTopicRun.value.steps.some((s) => s.status === 'running')) return false
    if (raw?.steps.some((s) => s.status === 'running')) return false
  }
  return topicAtHumanPause.value
})

const topicShowRoundConfirmPanel = computed(
  () =>
    topicAwaitingRoundConfirm.value &&
    !currentTopicRun.value.steps.some((s) => s.status === 'running') &&
    !topicDiscoveryApiRun.value?.steps.some((s) => s.status === 'running'),
)

function buildTopicCheckpointView(
  pauseStage: string,
  raw: TopicDiscoveryRunResponse | null,
): TopicCheckpointView | null {
  const def = TOPIC_DISCOVERY_FLOW_STEPS.find((s) => s.stageCode === pauseStage)
  if (!def?.checkpointKey) return null
  const dto = raw?.steps.find((s) => s.stage_code === pauseStage)
  const lines = linesFromApiStep(dto)
  const base = lines.length ? lines : CHECKPOINT_COPY[def.checkpointKey].lines
  if (pauseStage === 'idea_discovery' || pauseStage === 'topic_discovery') {
    const phaseLines = arisIdeaDiscoveryPhaseRows(raw?.steps, raw?.run_status, raw)
      .filter((r) => r.status === 'waiting_confirm' || r.status === 'completed')
      .map((r) => `${r.label}：${arisPhaseStatusLabel(r.status)}`)
    if (phaseLines.length) {
      return {
        key: def.checkpointKey,
        title: 'Idea Discovery · 段间检查点',
        lines: [...phaseLines, ...base].slice(0, 12),
      }
    }
  }
  return {
    key: def.checkpointKey,
    title: def.label,
    lines: base,
  }
}

const topicArisPhaseRows = computed((): ArisPhaseUiRow[] =>
  arisIdeaDiscoveryPhaseRows(
    topicDiscoveryApiRun.value?.steps,
    topicDiscoveryApiRun.value?.run_status,
    topicDiscoveryApiRun.value,
  ),
)

const topicIdeaDiscoveryUiGroupLabels = computed(() =>
  ideaDiscoveryUiGroupsLabelChain(
    resolveIdeaDiscoveryUIGroups(
      topicDiscoveryApiRun.value,
      topicDiscoveryApiRun.value?.steps,
    ),
  ),
)

const topicCheckpointPanel = computed((): TopicCheckpointView | null => {
  if (!topicShowCheckpointPanel.value) return null
  if (currentTopicRun.value.checkpoint) return currentTopicRun.value.checkpoint
  const raw = topicDiscoveryApiRun.value
  const fromApi = raw?.pause_after_stage?.trim()
  if (fromApi) {
    return buildTopicCheckpointView(fromApi, raw)
  }
  const steps = currentTopicRun.value.steps
  for (let i = 0; i < steps.length - 1; i++) {
    if (steps[i].status === 'completed' && steps[i + 1].status === 'pending') {
      return buildTopicCheckpointView(steps[i].stageCode, raw)
    }
  }
  return null
})

/** 检查点标题后缀：待确认的子阶段中文名（如文献调研），或主流程步骤 label */
const topicCheckpointHeadingSuffix = computed((): string => {
  const raw = topicDiscoveryApiRun.value
  const phaseLabel = arisPhaseAwaitingConfirmLabel(raw?.steps, raw?.run_status, raw)
  if (phaseLabel) return phaseLabel
  const pause = raw?.pause_after_stage?.trim()
  if (pause) {
    const def = TOPIC_DISCOVERY_FLOW_STEPS.find((s) => s.stageCode === pause)
    if (def?.label) return def.label
  }
  const title = topicCheckpointPanel.value?.title?.trim()
  if (title && !title.includes('段间检查点')) return title
  return title || '选题发现-全流程'
})

const topicFailedSummary = computed((): { stageLabel: string; stageCode: string; error: string } | null => {
  if (!topicShowFailedPanel.value) return null
  const raw = topicDiscoveryApiRun.value
  const failedDto = raw?.steps.find((s) => s.status === 'failed')
  const stageCode = failedDto?.stage_code ?? 'unknown'
  const def = TOPIC_DISCOVERY_FLOW_STEPS.find((s) => s.stageCode === stageCode)
  const error =
    failedDto?.error_msg?.trim() ||
    '本步执行失败，请查看日志或调整配置后重试。'
  return {
    stageLabel: def?.label ?? stageCode,
    stageCode,
    error,
  }
})

const topicServerRunCompleted = computed(() =>
  isTopicRunTerminalCompleted(topicDiscoveryApiRun.value),
)

const litReviewDoneForManuscript = computed(
  () => !!litReviewDoneByManuscript.value[activeManuscriptId.value],
)

const experimentPlanDoneForManuscript = computed(
  () => !!experimentPlanDoneByManuscript.value[activeManuscriptId.value],
)

const topicReadyForLiteratureReview = computed(
  () => topicServerRunCompleted.value && !litReviewDoneForManuscript.value,
)

const litReviewReadyForExperimentPlan = computed(
  () =>
    isLiteratureReviewModule.value &&
    litReviewDoneForManuscript.value &&
    !experimentPlanDoneForManuscript.value,
)

const showPrimaryAction = computed(() => {
  if (isUtilityModule.value) return false
  if (isLiteratureReviewModule.value) return false
  if (isExperimentPlanningModule.value) return false
  if (isPaperSubmissionModule.value) return false
  return true
})

const topicPrimaryDisabled = computed(() => {
  if (isUtilityModule.value) return true
  if (isTopicDiscoveryModule.value) {
    return (
      topicShowFailedPanel.value ||
      topicRunBusy.value ||
      currentTopicRun.value.status === 'checkpoint' ||
      (currentTopicRun.value.status === 'running' && !topicAtHumanPause.value)
    )
  }
  return running.value
})

const primaryActionLabel = computed(() => {
  if (isTopicDiscoveryModule.value) {
    if (topicAtHumanPause.value || currentTopicRun.value.status === 'checkpoint') return '等待确认'
    // 本步已失败：下方「驳回并返回」处理；顶栏保持与执行中一致，避免误点「再次运行」
    if (topicShowFailedPanel.value || currentTopicRun.value.status === 'failed') return '运行中…'
    if (currentTopicRun.value.status === 'running' || running.value) return '运行中…'
    if (topicReadyForLiteratureReview.value) return '生成文献综述'
    if (topicServerRunCompleted.value) return '再次运行'
    return '运行'
  }
  if (isLiteratureReviewModule.value) {
    if (running.value) return '运行中…'
    if (litReviewReadyForExperimentPlan.value) return '生成实验计划'
    if (litReviewDoneForManuscript.value && experimentPlanDoneForManuscript.value) return '再次运行'
    return '运行'
  }
  if (isPaperWritingModule.value) {
    return running.value ? '生成中…' : '生成论文'
  }
  return running.value ? '运行中…' : '运行'
})

function persistTopicRun(msId: string, patch: TopicRunDemo) {
  topicRunsByManuscript.value = { ...topicRunsByManuscript.value, [msId]: patch }
}

function mergeTopicRunFromApi(msId: string, data: TopicDiscoveryRunResponse): TopicRunDemo {
  const fromApi = topicRunFromApi(data)
  const optStage = topicContinueOptimisticStage.value
  if (!optStage) return fromApi

  const apiStep = fromApi.steps.find((s) => s.stageCode === optStage)
  const localStep = topicRunsByManuscript.value[msId]?.steps.find((s) => s.stageCode === optStage)
  const apiStillPending = apiStep?.status === 'pending'
  const localRunning = localStep?.status === 'running'

  if (apiStep?.status === 'running' || apiStep?.status === 'completed' || apiStep?.status === 'failed') {
    topicContinueOptimisticStage.value = null
    return fromApi
  }

  if (!localRunning || !apiStillPending) return fromApi

  const steps = fromApi.steps.map((s) =>
    s.stageCode === optStage ? { ...s, status: 'running' as TopicFlowStepStatus } : s,
  )
  let status: TopicRunDemo['status'] = fromApi.status
  if (status === 'checkpoint') status = 'running'
  return { ...fromApi, status, steps, checkpoint: null }
}

/** 与后端 run_status 落盘值一一对应，不做推断 */
function runStatusFromApi(runStatus: string): TopicRunDemo['status'] {
  switch (runStatus) {
    case 'running':
    case 'checkpoint':
    case 'completed':
    case 'failed':
      return runStatus
    default:
      return 'idle'
  }
}

/** 与后端 steps[].status 落盘值一一对应，不做推断 */
function stepStatusFromApi(raw: string): TopicFlowStepStatus {
  switch (raw) {
    case 'pending':
    case 'running':
    case 'completed':
    case 'failed':
      return raw
    default:
      return 'pending'
  }
}

const TOPIC_STEP_RUNNING_HINT: Record<string, string> = {
  retrieve: '文献 PDF 下载与入库中…',
  idea_discovery: 'Idea Discovery…',
  topic_discovery: 'Idea Discovery…',
  topic_confirm: '等待确认并入库…',
  generate_ideas: 'AI 脑暴选题与新颖性分析中…',
  audit: 'AI 审查结论与文献综述生成中…',
}

function topicStepRunningHint(stageCode: string) {
  if (stageCode === 'idea_discovery' || stageCode === 'topic_discovery') {
    const running = topicArisPhaseRows.value.find((r) => r.status === 'running')
    const waiting = topicArisPhaseRows.value.find((r) => r.status === 'waiting_confirm')
    const awaitingContinue = topicArisPhaseRows.value.find((r) => r.status === 'awaiting_continue')
    if (waiting) return `段间暂停 · 下一段「${waiting.label}」待确认并继续`
    if (awaitingContinue) return `本轮 ARIS 已结束 · 点击「继续」启动下一段「${awaitingContinue.label}」`
    if (running) return `当前段「${running.label}」运行中…`
    const chain = topicIdeaDiscoveryUiGroupLabels.value
    if (chain) return `Idea Discovery：${chain}…`
    const dto = topicDiscoveryApiRun.value?.steps.find(
      (s) => s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery',
    )
    const pct = dto?.exec_progress_pct
    if (typeof pct === 'number' && pct > 0) return `Idea Discovery 总进度约 ${pct}%`
    return TOPIC_STEP_RUNNING_HINT.idea_discovery
  }
  const base = TOPIC_STEP_RUNNING_HINT[stageCode] ?? 'AI 模型运行中…'
  const dto = topicDiscoveryApiRun.value?.steps.find((s) => s.stage_code === stageCode)
  const pct = dto?.exec_progress_pct
  if (typeof pct === 'number' && pct > 0) {
    return `${base} ${pct}%`
  }
  return base
}

const TOPIC_RUN_POLL_MS = 1500

/** 服务端 retrieve / generate_ideas / audit 等 bot 异步步为 running 时需轮询 */
function topicRunHasAsyncWorkOnServer(msId: string): boolean {
  const apiRun = topicLastRunByMs.value[msId]
  if (!apiRun) return false
  if (isIdeaDiscoveryArisCheckpoint(apiRun.steps, apiRun.run_status)) return false
  return apiRun.steps.some((s) => s.status === 'running')
}

/** 轮询 GET /run/current：PDF 下载（retrieve）、脑暴 LLM（generate_ideas）、审查 LLM（audit）等异步步 */
const topicRunNeedsPoll = computed(() => {
  const msId = activeManuscriptId.value
  if (topicContinueLoading.value) return true
  if (msId && topicRunHasAsyncWorkOnServer(msId)) return true
  if (running.value) return true
  if (!isTopicDiscoveryModule.value || !topicRunVisible.value) return false
  if (topicAtHumanPause.value) return false
  if (currentTopicRun.value.status === 'running') return true
  return currentTopicRun.value.steps.some((s) => s.status === 'running')
})

const topicRunBackgroundPollStop = ref<(() => void) | null>(null)

watch(
  topicRunNeedsPoll,
  (need) => {
    if (topicRunBackgroundPollStop.value) {
      topicRunBackgroundPollStop.value()
      topicRunBackgroundPollStop.value = null
    }
    if (!need) return
    const msId = activeManuscriptId.value
    if (!msId) return
    topicRunBackgroundPollStop.value = startTopicRunProgressPoll(msId, topicRunToken.value)
  },
  { flush: 'post', immediate: true },
)

onUnmounted(() => {
  topicRunBackgroundPollStop.value?.()
})

function startTopicRunProgressPoll(msId: string, token: number): () => void {
  const tick = async () => {
    if (token !== topicRunToken.value) return
    if (topicRunStartInFlight.value) return
    try {
      const data = await fetchCurrentTopicDiscoveryRun(msId)
      if (!data || token !== topicRunToken.value) return
      syncTopicRunUiAfterServer(msId, data)
      if (data.manuscript_id > 0) {
        void loadManuscriptsFromServer(String(data.manuscript_id))
      }
    } catch {
      /* 忽略轮询失败 */
    }
  }
  void tick()
  const timer = window.setInterval(() => void tick(), TOPIC_RUN_POLL_MS)
  return () => window.clearInterval(timer)
}

function noveltyLinesFromResultObj(r: Record<string, unknown>): string[] {
  const lines = (r.lines as string[]) ?? []
  if (lines.length) return lines.slice(0, 10)
  const risks = (r.risks as Array<{ idea_title?: string; risk?: string; note?: string }>) ?? []
  return risks.slice(0, 8).map((item) => {
    const t = item.idea_title ?? '选题'
    return `${t}：${item.risk ?? '—'}${item.note ? ` · ${item.note}` : ''}`
  })
}

function linesFromApiStep(step?: TopicDiscoveryStepDTO): string[] {
  if (!step) return []
  const result = step.result as Record<string, unknown> | undefined
  if (!result) {
    const err = step.error_msg?.trim()
    if (err && step.status === 'failed') return [err]
    return []
  }

  if (step.stage_code === 'retrieve') {
    const hits = (result.literature_hits as Array<{ title?: string; external_key?: string }>) ?? []
    const lines: string[] = []
    if (typeof result.hit_count === 'number') {
      lines.push(`命中 ${result.hit_count} 篇 · 验真 ${result.verified_count ?? result.hit_count} 篇`)
    }
    for (const h of hits.slice(0, 6)) {
      const row = h as { title?: string; external_key?: string; url?: string; meta?: { url?: string } }
      const url = row.url?.trim() || row.meta?.url?.trim()
      if (row.title && url) {
        lines.push(`· ${row.title} — ${url}`)
      } else if (row.title) {
        lines.push(`· ${row.title}${row.external_key ? ` (${row.external_key})` : ''}`)
      }
    }
    return lines.length ? lines : ['检索已完成']
  }

  if (
    step.stage_code === 'idea_discovery' ||
    step.stage_code === 'topic_discovery' ||
    step.stage_code === 'generate_ideas'
  ) {
    if (step.stage_code === 'idea_discovery' || step.stage_code === 'topic_discovery') {
      const passed = result.passed
      const iters = (result.iterations as unknown[]) ?? []
      const lines: string[] = []
      if (typeof passed === 'boolean') {
        lines.push(passed ? `门禁通过（${iters.length} 轮）` : `未完全通过（${iters.length} 轮，保留最后一版）`)
      }
      const lr = result.literature_review as { title?: string; summary?: string } | undefined
      if (lr?.title?.trim()) lines.push(`文献综述：${lr.title.trim()}`)
      const ideas = (result.ideas as Array<{ title?: string }>) ?? []
      for (const idea of ideas.slice(0, 5)) {
        if (idea.title?.trim()) lines.push(`· ${idea.title.trim()}`)
      }
      const finalAudit = result.final_audit as { summary?: string; issues?: Array<{ severity?: string; claim?: string }> } | undefined
      if (finalAudit?.summary?.trim()) lines.push(finalAudit.summary.trim())
      if (finalAudit?.issues?.length) {
        lines.push(
          ...finalAudit.issues.slice(0, 4).map((iss) => `[${iss.severity ?? 'issue'}] ${iss.claim ?? ''}`.trim()),
        )
      }
      if (lines.length) return lines.slice(0, 12)
    }
    const ideas = (result.ideas as Array<{ title?: string; problem?: string }>) ?? []
    const ideaLines = ideas.slice(0, 8).map((idea, i) => {
      const t = idea.title?.trim() || `Idea ${i + 1}`
      const p = idea.problem?.trim()
      return p ? `${t} — ${p}` : t
    })
    const nov = result.novelty
    if (nov && typeof nov === 'object') {
      const novLines = noveltyLinesFromResultObj(nov as Record<string, unknown>)
      if (novLines.length) {
        return [...ideaLines, '— 新颖性 —', ...novLines.slice(0, 6)]
      }
    }
    return ideaLines
  }

  if (step.stage_code === 'novelty') {
    return noveltyLinesFromResultObj(result)
  }

  if (step.stage_code === 'audit') {
    const lines: string[] = []
    const lr = result.literature_review as { title?: string; summary?: string } | undefined
    if (lr?.title?.trim()) {
      lines.push(`文献综述：${lr.title.trim()}`)
    }
    if (lr?.summary?.trim()) {
      lines.push(lr.summary.trim())
    }
    const rounds = (result.rounds as Array<{
      issues?: Array<{ severity?: string; claim?: string }>
      summary?: string
      off_topic?: { generate_ideas_on_topic?: boolean; notes?: string }
    }>) ?? []
    const last = rounds[rounds.length - 1]
    const auditBlock = last as Record<string, unknown> | undefined
    const issues =
      last?.issues ??
      (auditBlock?.audit as { issues?: Array<{ severity?: string; claim?: string }> } | undefined)?.issues ??
      (result.issues as Array<{ severity?: string; claim?: string }>)
    if (last?.summary?.trim()) {
      lines.push(last.summary.trim())
    }
    const off = last?.off_topic
    if (off?.notes?.trim()) {
      lines.push(`贴题复核：${off.notes.trim()}`)
    } else if (off && typeof off.generate_ideas_on_topic === 'boolean') {
      lines.push(`第二步贴题：${off.generate_ideas_on_topic ? '是' : '否'}`)
    }
    if (issues?.length) {
      lines.push(
        ...issues.slice(0, 6).map((iss) => `[${iss.severity ?? 'issue'}] ${iss.claim ?? ''}`.trim()),
      )
    }
    if (lines.length) return lines.slice(0, 12)
  }

  try {
    return [JSON.stringify(result).slice(0, 500)]
  } catch {
    return ['（已返回结构化结果）']
  }
}

/** 展示态：run_status + steps；人工暂停时统一为 checkpoint（后端可能仍标 running） */
function topicRunFromApi(data: TopicDiscoveryRunResponse): TopicRunDemo {
  let status = runStatusFromApi(data.run_status)
  let steps: TopicFlowStepRuntime[] = TOPIC_DISCOVERY_FLOW_STEPS.map((def) => {
    const st = data.steps.find((s) => s.stage_code === def.stageCode)
    return {
      stageCode: def.stageCode,
      label: def.label,
      checkpointKey: def.checkpointKey,
      status: st ? stepStatusFromApi(st.status) : 'pending',
    }
  })
  steps = clampTopicConfirmWhileIdeaDiscoveryActive(steps)

  const pauseStage = topicPauseStageFromApi(data)
  const anyStepRunning = steps.some((s) => s.status === 'running')
  const arisPause = isIdeaDiscoveryArisCheckpoint(data.steps, data.run_status)
  const ideaExecuting = isAnyIdeaDiscoveryPhaseExecuting(data.steps)
  if (
    pauseStage &&
    (arisPause || (!anyStepRunning && !ideaExecuting)) &&
    status !== 'failed' &&
    status !== 'completed'
  ) {
    status = 'checkpoint'
  }
  if (arisPause && !ideaExecuting && status !== 'failed' && status !== 'completed') {
    status = 'checkpoint'
  }

  let checkpoint: TopicCheckpointView | null = null
  if (status === 'checkpoint' && pauseStage) {
    const def = TOPIC_DISCOVERY_FLOW_STEPS.find((s) => s.stageCode === pauseStage)
    const step = data.steps.find((s) => s.stage_code === pauseStage)
    if (def && step) {
      const lines = linesFromApiStep(step)
      checkpoint = {
        key: def.checkpointKey,
        title: def.label,
        lines: lines.length ? lines : CHECKPOINT_COPY[def.checkpointKey].lines,
      }
    }
  }

  return { status, steps, checkpoint }
}

function markTopicRunForCompletionPrompt() {
  topicRunAwaitingCompletionPrompt.value = true
}

function maybeShowTopicDiscoveryCompletionPrompt() {
  if (!topicRunAwaitingCompletionPrompt.value) return
  topicRunAwaitingCompletionPrompt.value = false
  topicCompletionPromptVisible.value = true
}

function dismissTopicCompletionPromptToLiteratureReview() {
  if (!topicCompletionPromptVisible.value) return
  topicCompletionPromptVisible.value = false
  selectModule('literature-review')
}

function syncTopicRunUiAfterServer(msId: string, data: TopicDiscoveryRunResponse) {
  topicLastRunByMs.value = { ...topicLastRunByMs.value, [msId]: data }
  if (isTopicRunTerminalCompleted(data)) {
    persistTopicRun(msId, createIdleTopicRun())
    running.value = false
    topicContinueOptimisticStage.value = null
    maybeShowTopicDiscoveryCompletionPrompt()
    return
  }
  persistTopicRun(msId, mergeTopicRunFromApi(msId, data))
  syncRunningFlagFromServer(msId)
}

function buildTopicRunRequest(action: 'start' | 'continue') {
  const direction = formatTopicDirectionText(topicForm)
  const msNum = Number(activeManuscriptId.value)
  return {
    manuscript_id: msNum,
    manuscript_title: currentManuscript.value?.title ?? '',
    discipline_code: topicForm.disciplineCode,
    keywords: parseTopicKeywords(topicForm.keywords),
    description: topicForm.description.trim(),
    direction: direction || undefined,
    venue: formatVenueTextFromCodes(topicForm.venueCodes, venueOptions.value),
    venue_codes: [...topicForm.venueCodes],
    source_codes: [...topicForm.sourceCodes],
    intensity: topicForm.intensity,
    audit_level: topicForm.auditLevel,
    human_checkpoint: topicForm.humanCheckpoint,
    random_references: topicForm.randomReferences,
    user_literature_ids:
      topicForm.userLiteratureIds.length > 0 ? [...topicForm.userLiteratureIds] : undefined,
    action,
  }
}

let hydrateTopicRunInflight: Promise<void> | null = null
let hydrateTopicRunInflightMs = ''

async function hydrateTopicRunFromServer(msId: string): Promise<void> {
  if (hydrateTopicRunInflight && hydrateTopicRunInflightMs === msId) {
    return hydrateTopicRunInflight
  }
  hydrateTopicRunInflightMs = msId
  hydrateTopicRunInflight = (async () => {
    try {
      const data = await fetchCurrentTopicDiscoveryRun(msId)
      if (!data) {
        const next = { ...topicLastRunByMs.value }
        delete next[msId]
        topicLastRunByMs.value = next
        persistTopicRun(msId, createIdleTopicRun())
        return
      }
      if (!isTopicRunTerminalCompleted(data)) {
        applyStoredTopicRunInputToForm(topicForm, data, venueOptions.value)
      }
      syncTopicDisciplineFromCurrentManuscript()
      syncTopicRunUiAfterServer(msId, data)
      if (!isTopicRunTerminalCompleted(data)) {
        requestScrollToTopicFlowIfNeeded()
      }
    } catch {
      /* 未登录或无 run 时忽略 */
    }
  })()
  try {
    await hydrateTopicRunInflight
  } finally {
    if (hydrateTopicRunInflightMs === msId) {
      hydrateTopicRunInflight = null
      hydrateTopicRunInflightMs = ''
    }
  }
}

/** 选题页：单次 loading 壳 + hydrate（切换论文 / 论文 id 就绪后补拉） */
async function refreshTopicDiscoveryForManuscript(msId: string) {
  if (!msId || activeModule.value !== 'topic-discovery') return
  if (moduleContentLoading.value) {
    await hydrateTopicRunFromServer(msId)
    return
  }
  moduleContentLoading.value = true
  try {
    await hydrateTopicRunFromServer(msId)
  } finally {
    moduleContentLoading.value = false
  }
}

async function invokeTopicDiscoveryRun(action: 'start' | 'continue', token: number) {
  const msId = activeManuscriptId.value
  if (!msId) return

  const stopPoll = startTopicRunProgressPoll(msId, token)
  let data: TopicDiscoveryRunResponse | undefined
  try {
    data = await postTopicDiscoveryRun(buildTopicRunRequest(action))
  } catch (e) {
    if (token === topicRunToken.value) {
      await hydrateTopicRunFromServer(msId)
    }
    throw e
  } finally {
    stopPoll()
  }
  if (token !== topicRunToken.value || !data) return

  if (isTopicRunTerminalCompleted(data)) {
    syncTopicRunUiAfterServer(msId, data)
    finalizeTopicDiscoveryRun(msId)
    return
  }

  topicLastRunByMs.value = { ...topicLastRunByMs.value, [msId]: data }
  const run = mergeTopicRunFromApi(msId, data)
  persistTopicRun(msId, run)
  syncRunningFlagFromServer(msId)

  if (run.status === 'checkpoint' && run.checkpoint) {
    ElMessage.info(`流程已暂停：请确认「${run.checkpoint.title}」后再继续`)
  } else if (run.status === 'failed') {
    const errLine = topicFailedSummary.value?.error
    ElMessage.error(errLine && errLine.length < 120 ? errLine : '选题发现运行失败，请配置 LLM 或改参数后重试')
  }
}

/** 新开 run：第一步立刻标 running，避免 POST 返回前只有「提交中」无转圈 */
function optimisticMarkFirstStepRunning(msId: string) {
  const steps = TOPIC_DISCOVERY_FLOW_STEPS.map((def, i) => ({
    stageCode: def.stageCode,
    label: def.label,
    checkpointKey: def.checkpointKey,
    status: (i === 0 ? 'running' : 'pending') as TopicFlowStepStatus,
  }))
  persistTopicRun(msId, {
    status: 'running',
    steps,
    checkpoint: null,
  })
}

/** 检查点继续：原生 idea_discovery 段内 continue 仍标 idea_discovery；全流程完成后才标 topic_confirm */
function optimisticMarkNextStepRunning(msId: string) {
  const cur = topicRunsByManuscript.value[msId] ?? createIdleTopicRun()
  const steps = cur.steps.map((s) => ({ ...s }))
  let marked = false
  const ideaIdx = steps.findIndex(
    (s) => s.stageCode === 'idea_discovery' || s.stageCode === 'topic_discovery',
  )

  if (ideaIdx >= 0 && steps[ideaIdx].status !== 'completed') {
    steps[ideaIdx].status = 'running'
    topicContinueOptimisticStage.value = steps[ideaIdx].stageCode
    marked = true
  } else {
    const pendingIdx = steps.findIndex((s) => s.status === 'pending')
    if (pendingIdx >= 0) {
      steps[pendingIdx].status = 'running'
      topicContinueOptimisticStage.value = steps[pendingIdx].stageCode
      marked = true
    } else {
      for (let i = 0; i < steps.length - 1; i++) {
        if (steps[i].status === 'completed' && steps[i + 1].status === 'pending') {
          steps[i + 1].status = 'running'
          topicContinueOptimisticStage.value = steps[i + 1].stageCode
          marked = true
          break
        }
      }
    }
  }
  if (!marked) topicContinueOptimisticStage.value = null
  persistTopicRun(msId, {
    status: 'running',
    steps: clampTopicConfirmWhileIdeaDiscoveryActive(steps),
    checkpoint: null,
  })
}

function syncRunningFlagFromServer(msId: string) {
  if (topicAtHumanPause.value) {
    running.value = false
    return
  }
  if (topicRunHasAsyncWorkOnServer(msId)) {
    running.value = true
    return
  }
  if (currentTopicRun.value.steps.some((s) => s.status === 'running')) {
    running.value = true
    return
  }
  running.value = false
}

function finalizeTopicDiscoveryRun(_msId: string) {
  recordModuleOperationLog('topic-discovery', '运行工作流（retrieve → topic_discovery）')
}

function resetTopicRunForAction(msId: string) {
  topicRunAwaitingCompletionPrompt.value = false
  topicCompletionPromptVisible.value = false
  topicRunToken.value += 1
  persistTopicRun(msId, createIdleTopicRun())
  if (litReviewDoneByManuscript.value[msId]) {
    litReviewDoneByManuscript.value = { ...litReviewDoneByManuscript.value, [msId]: false }
    persistLitReviewFlags()
  }
  clearExperimentPlanDone(msId)
}

async function scrollToTopicFlowPanel() {
  await nextTick()
  await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()))
  const panel = topicFlowPanelRef.value
  const scroller = paperMainRef.value
  if (!panel) return false
  if (scroller) {
    const top =
      panel.getBoundingClientRect().top -
      scroller.getBoundingClientRect().top +
      scroller.scrollTop -
      12
    scroller.scrollTo({ top: Math.max(0, top), behavior: 'smooth' })
    return true
  }
  panel.scrollIntoView({ behavior: 'smooth', block: 'start' })
  return true
}

let scrollToTopicFlowSeq = 0

/** 等模块 DOM（含 v-else 解除 loading）挂载后再滚，避免刷新时 hydrate 先于 form-options 完成 */
async function scrollToTopicFlowPanelWhenReady() {
  const seq = ++scrollToTopicFlowSeq
  const deadline = Date.now() + 4000
  while (Date.now() < deadline) {
    if (seq !== scrollToTopicFlowSeq) return
    if (!isTopicDiscoveryModule.value || !topicRunVisible.value) return
    if (topicDiscoveryPageLoading.value) {
      await new Promise<void>((r) => window.setTimeout(r, 50))
      continue
    }
    if (await scrollToTopicFlowPanel()) return
    await new Promise<void>((r) => window.setTimeout(r, 50))
  }
}

function requestScrollToTopicFlowIfNeeded() {
  if (!isTopicDiscoveryModule.value || !topicRunVisible.value) return
  void scrollToTopicFlowPanelWhenReady()
}

/** 刷新 hydrate、切回选题模块、或 loading 结束且仍有 run 时滚到「运行进度」 */
watch(
  () =>
    [
      topicRunVisible.value,
      topicDiscoveryPageLoading.value,
      activeModule.value,
      moduleContentKey.value,
    ] as const,
  ([visible, loading, module]) => {
    if (module !== 'topic-discovery' || !visible || loading) return
    void scrollToTopicFlowPanelWhenReady()
  },
  { flush: 'post' },
)

async function markLiteratureReviewDone(msId: string) {
  litReviewDoneByManuscript.value = { ...litReviewDoneByManuscript.value, [msId]: true }
  persistLitReviewFlags()
}

async function runExperimentPlanFromLiteratureReview(opts?: { skipLitReviewDoneCheck?: boolean }) {
  const msId = activeManuscriptId.value
  if (!msId) return
  if (!opts?.skipLitReviewDoneCheck && !litReviewDoneForManuscript.value) {
    ElMessage.warning('请先生成文献综述')
    return
  }

  selectModule('experiment-planning')
  running.value = true
  try {
    await nextTick()
    const ok = await workflowPanelsRef.value?.runModule('experiment-planning')
    if (ok) {
      experimentPlanDoneByManuscript.value = { ...experimentPlanDoneByManuscript.value, [msId]: true }
      persistExperimentPlanFlags()
      recordModuleOperationLog('experiment-planning', '从文献综述续跑 · 生成实验计划')
      ElMessage.success(
        `「实验方案」已生成（演示）· 论文「${currentManuscript.value?.title ?? '未命名'}」`,
      )
    }
  } finally {
    running.value = false
  }
}

async function rerunLiteratureReviewModule() {
  const msId = activeManuscriptId.value
  if (!msId) return
  litReviewDoneByManuscript.value = { ...litReviewDoneByManuscript.value, [msId]: false }
  persistLitReviewFlags()
  clearExperimentPlanDone(msId)
  running.value = true
  try {
    const ok = await workflowPanelsRef.value?.runModule('literature-review')
    if (ok) {
      await markLiteratureReviewDone(msId)
      recordModuleOperationLog('literature-review', '再次运行 · 生成文献综述')
      ElMessage.success(
        `「文献综述」已重新生成（演示）· 论文「${currentManuscript.value?.title ?? '未命名'}」`,
      )
    }
  } finally {
    running.value = false
  }
}

async function runLiteratureReviewFromTopic() {
  const msId = activeManuscriptId.value
  if (!msId) return
  if (!topicServerRunCompleted.value) {
    ElMessage.warning('请先完成选题发现（三步含审计）')
    return
  }

  selectModule('literature-review')
  running.value = true
  try {
    const committed = await commitTopicDiscoveryManuscript(currentManuscript.value?.title)
    syncTopicRunUiAfterServer(msId, committed)
    await nextTick()
    const ok = await workflowPanelsRef.value?.runModule('literature-review')
    if (ok) {
      clearExperimentPlanDone(msId)
      await markLiteratureReviewDone(msId)
      recordModuleOperationLog('literature-review', '从选题发现续跑 · 生成文献综述')
      ElMessage.success(
        `「文献综述」已生成 · 可点顶栏「生成实验计划」继续（演示）`,
      )
    }
  } finally {
    running.value = false
  }
}

async function startTopicDiscoveryRun() {
  const msId = await ensureManuscriptWhenEmpty({ warnIfUnset: true })
  if (!msId) return

  resetTopicRunForAction(msId)
  markTopicRunForCompletionPrompt()
  const token = topicRunToken.value
  topicRunStartInFlight.value = true
  running.value = true
  optimisticMarkFirstStepRunning(msId)
  await scrollToTopicFlowPanel()

  try {
    await invokeTopicDiscoveryRun('start', token)
  } catch (e) {
    if (token !== topicRunToken.value) return
    await hydrateTopicRunFromServer(msId)
    const msg = e instanceof Error ? e.message : '运行失败'
    ElMessage.error(msg.includes('401') ? '请先登录后再运行选题发现' : msg)
  } finally {
    if (token === topicRunToken.value) {
      topicRunStartInFlight.value = false
      syncRunningFlagFromServer(msId)
    }
  }
}

async function confirmTopicSelectedRound() {
  const msId = activeManuscriptId.value
  if (!msId || !topicShowRoundConfirmPanel.value) return
  const round = topicSelectedRound.value
  if (round == null || round < 1) {
    ElMessage.warning('请先选择一轮迭代结果')
    return
  }
  if (topicConfirmRoundLoading.value) return
  const token = topicRunToken.value
  topicConfirmRoundLoading.value = true
  try {
    const data = await postTopicDiscoveryConfirmRound({
      manuscript_id: Number(msId),
      run_version: topicDiscoveryApiRun.value?.run_version,
      selected_round: round,
    })
    if (token !== topicRunToken.value) return
    topicRunAwaitingCompletionPrompt.value = false
    syncTopicRunUiAfterServer(msId, data)
    await markLiteratureReviewDone(msId)
    openLiteratureReviewAfterTopicConfirm(data)
    ElMessage.success(`已确认第 ${round} 轮，正在打开文献综述…`)
  } catch (e) {
    if (token !== topicRunToken.value) return
    const msg = e instanceof Error ? e.message : '确认失败'
    ElMessage.error(msg.includes('401') ? '请先登录' : msg)
  } finally {
    if (token === topicRunToken.value) {
      topicConfirmRoundLoading.value = false
      syncRunningFlagFromServer(msId)
    }
  }
}

function continueTopicAfterCheckpoint() {
  const msId = activeManuscriptId.value
  if (!msId || !topicShowCheckpointPanel.value) return
  if (topicContinueLoading.value) return

  const token = topicRunToken.value
  markTopicRunForCompletionPrompt()
  optimisticMarkNextStepRunning(msId)
  running.value = true
  topicContinueLoading.value = true
  void scrollToTopicFlowPanelWhenReady()

  void invokeTopicDiscoveryRun('continue', token)
    .catch(async (e) => {
      if (token !== topicRunToken.value) return
      topicContinueOptimisticStage.value = null
      await hydrateTopicRunFromServer(msId)
      const msg = e instanceof Error ? e.message : '继续失败'
      if (currentTopicRun.value.status !== 'failed') {
        ElMessage.error(msg.includes('401') ? '请先登录后再运行选题发现' : msg)
      }
    })
    .finally(() => {
      if (token !== topicRunToken.value) return
      topicContinueLoading.value = false
      if (!topicContinueOptimisticStage.value) {
        syncRunningFlagFromServer(msId)
        return
      }
      const apiRun = topicLastRunByMs.value[msId]
      const stage = topicContinueOptimisticStage.value
      const apiStep = apiRun?.steps.find((s) => s.stage_code === stage)
      if (apiStep && apiStep.status !== 'pending') {
        topicContinueOptimisticStage.value = null
      }
      syncRunningFlagFromServer(msId)
    })
}

async function abandonCurrentTopicRun(msId: string) {
  topicRunToken.value += 1
  running.value = false
  await cancelTopicDiscoveryRun(msId)
  const next = { ...topicLastRunByMs.value }
  delete next[msId]
  topicLastRunByMs.value = next
  persistTopicRun(msId, createIdleTopicRun())
}

async function abandonTopicRunWithConfirm(
  body: string,
  dialogTitle: string,
  confirmButtonText: string,
  doneMessage: string,
) {
  const msId = activeManuscriptId.value
  if (!msId || topicTerminateLoading.value) return
  try {
    await paperConfirm(body, dialogTitle, {
      confirmButtonText,
      cancelButtonText: '取消',
      variant: 'danger',
    })
  } catch {
    return
  }
  topicTerminateLoading.value = true
  try {
    await abandonCurrentTopicRun(msId)
    ElMessage.info(doneMessage)
  } catch (e) {
    const msg = e instanceof Error ? e.message : '操作失败'
    ElMessage.error(msg)
  } finally {
    topicTerminateLoading.value = false
  }
}

async function rejectTopicCheckpoint() {
  await abandonTopicRunWithConfirm(
    '终止后本轮进度将作废，确定终止吗？',
    '终止并返回',
    '确定终止',
    '已终止本轮',
  )
}

async function rejectFailedTopicRun() {
  await abandonTopicRunWithConfirm(
    '驳回后本轮进度将作废，确定返回吗？',
    '驳回并返回',
    '确定驳回',
    '已驳回本轮',
  )
}

async function cancelTopicRun() {
  const msId = activeManuscriptId.value
  if (!msId) return
  try {
    await abandonCurrentTopicRun(msId)
    ElMessage.info('已取消运行')
  } catch (e) {
    const msg = e instanceof Error ? e.message : '取消失败'
    ElMessage.error(msg)
  }
}

function topicStepIcon(status: TopicFlowStepStatus) {
  if (status === 'completed') return '✓'
  if (status === 'running') return '…'
  if (status === 'waiting_human') return '⏸'
  if (status === 'failed') return '✕'
  if (status === 'skipped') return '–'
  return '○'
}

watch(activeManuscriptId, () => {
  running.value = false
})

const currentMeta = computed(() => getPaperModuleMeta(activeModule.value))

function toggleTopicSource(code: string, checked: boolean) {
  if (checked) {
    const err = validateTopicLiteratureSourceCodes([code])
    if (err) {
      if (err.includes('暂不支持')) ElMessage.error(err)
      else ElMessage.warning(err)
      return
    }
  }
  const set = new Set(topicForm.sourceCodes)
  if (checked) set.add(code)
  else set.delete(code)
  topicForm.sourceCodes = [...set]
}

function isTopicSourceChecked(code: string) {
  return topicForm.sourceCodes.includes(code)
}

function openTopicUserLiteraturePicker() {
  if (topicRunBusy.value || currentTopicRun.value.status === 'checkpoint') return
  topicUserLitPickerOpen.value = true
}

function closeTopicUserLiteraturePicker() {
  topicUserLitPickerOpen.value = false
}

function onTopicUserLiteraturePicked(ids: string[]) {
  topicForm.userLiteratureIds = [...ids]
}

function onUploadExperimentPlanPrimary() {
  workflowPanelsRef.value?.openLitReviewUploadExpPlanDialog()
}

function onUserLiteratureUploadPrimary() {
  userLiteraturePanelRef.value?.openUploadDialog()
}

function onFigureUploadPrimary() {
  figureManagementPanelRef.value?.openUploadDialog()
}

function onExpPlanUploadPlanPrimary() {
  void workflowPanelsRef.value?.openExpPlanUploadPlanDialog()
}

function onPaperSubmissionNewPrimary() {
  if (!activeManuscriptId.value) {
    ElMessage.warning('请先在顶栏选择当前论文')
    return
  }
  paperSubmissionHubRef.value?.openWizard()
}

async function onPrimaryAction() {
  if (activeModule.value === 'topic-discovery') {
    if (parseTopicKeywords(topicForm.keywords).length === 0) {
      ElMessage.warning('请填写检索关键词（可多个，逗号分隔）')
      return
    }
    if (!topicForm.description.trim()) {
      ElMessage.warning('请填写详细描述')
      return
    }
    if (topicForm.venueCodes.length === 0) {
      ElMessage.warning('请至少选择一个目标会议/期刊')
      return
    }
    const litSrcErr = validateTopicLiteratureSourceCodes(topicForm.sourceCodes)
    if (litSrcErr) {
      if (litSrcErr.includes('暂不支持') || litSrcErr.includes('已停用')) ElMessage.error(litSrcErr)
      else ElMessage.warning(litSrcErr)
      return
    }
  }
  if (activeModule.value === 'topic-discovery') {
    if (topicReadyForLiteratureReview.value) {
      await runLiteratureReviewFromTopic()
    } else {
      await startTopicDiscoveryRun()
    }
    return
  }

  if (activeModule.value === 'literature-review') {
    if (litReviewReadyForExperimentPlan.value) {
      await runExperimentPlanFromLiteratureReview()
      return
    }
    if (litReviewDoneForManuscript.value && experimentPlanDoneForManuscript.value) {
      await rerunLiteratureReviewModule()
      return
    }
  }

  if (isSecondaryWorkflowModule.value) {
    running.value = true
    try {
      const msId = activeManuscriptId.value
      const ok = await workflowPanelsRef.value?.runModule(activeModule.value)
      if (ok) {
        const mod = activeModule.value
        if (mod === 'literature-review' && msId) {
          clearExperimentPlanDone(msId)
          await markLiteratureReviewDone(msId)
        }
        if (mod === 'experiment-planning' && msId) {
          experimentPlanDoneByManuscript.value = {
            ...experimentPlanDoneByManuscript.value,
            [msId]: true,
          }
          persistExperimentPlanFlags()
        }
        if (
          mod !== 'personal-center' &&
          mod !== 'invite-rebate' &&
          mod !== 'environment' &&
          mod !== 'my-manuscripts' &&
          mod !== 'user-literature' &&
          mod !== 'figure-generation' &&
          mod !== 'course-report' &&
          mod !== 'paper-submission'
        ) {
          recordModuleOperationLog(mod, `运行「${currentMeta.value.label}」`)
        }
        const successHint =
          mod === 'literature-review' && msId && !experimentPlanDoneForManuscript.value
            ? `「${currentMeta.value.label}」已生成 · 可点顶栏「生成实验计划」继续`
            : mod === 'paper-writing'
              ? `论文「${currentManuscript.value?.title ?? '未命名'}」生成任务已提交（演示）`
              : `「${currentMeta.value.label}」已完成演示运行 · 论文「${currentManuscript.value?.title ?? '未命名'}」`
        ElMessage.success(successHint)
      }
    } finally {
      running.value = false
    }
    return
  }

  running.value = true
  try {
    await new Promise((r) => setTimeout(r, 500))
    ElMessage.success(
      `「${currentMeta.value.label}」已在论文「${currentManuscript.value?.title ?? '未命名'}」下启动（演示）`,
    )
  } finally {
    running.value = false
  }
}

// 须在 loadManuscriptsFromServer / hydrateTopicRunFromServer 等定义之后注册，immediate 否则会 TDZ 导致整页不发请求
watch(
  activeModule,
  (id) => {
    persistActiveModule(id)
    void prepareModuleContent(id)
  },
  { immediate: true },
)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="createManuscriptDialogVisible"
      class="paper-modal-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="paper-create-ms-title"
    >
      <div class="paper-message-box paper-message-box--default paper-modal-panel paper-create-ms-panel">
        <header class="paper-modal-header">
          <h2 id="paper-create-ms-title" class="paper-modal-title">创建论文</h2>
          <button
            type="button"
            class="paper-modal-close"
            aria-label="关闭"
            @click="closeCreateManuscriptDialog(false)"
          >
            ×
          </button>
        </header>
        <div class="paper-modal-body">
          <div class="paper-message-box__form">
            <div>
              <span class="paper-label">学科</span>
              <PaperSelect
                v-model="createManuscriptForm.disciplineCode"
                :options="disciplineSelectOptions"
                placeholder="请选择"
                filterable
                filter-placeholder="输入学科名称筛选"
              />
            </div>
            <div>
              <span class="paper-label">论文名称</span>
              <input
                v-model="createManuscriptForm.title"
                class="paper-input"
                type="text"
                maxlength="256"
                placeholder="请输入论文名称"
                @keyup.enter="submitCreateManuscriptDialog"
              />
            </div>
            <div>
              <span class="paper-label">论文语言</span>
              <div class="paper-create-ms-lang" role="radiogroup" aria-label="论文语言">
                <label class="paper-create-ms-lang__option">
                  <input
                    v-model="createManuscriptForm.language"
                    type="radio"
                    name="paper-create-ms-lang"
                    value="en"
                  />
                  英文
                </label>
                <label class="paper-create-ms-lang__option">
                  <input
                    v-model="createManuscriptForm.language"
                    type="radio"
                    name="paper-create-ms-lang"
                    value="zh"
                  />
                  中文
                </label>
              </div>
            </div>
            <div>
              <span class="paper-label">上传论文（可选）</span>
              <p class="paper-create-ms-upload-lead">
                支持 PDF、Word、LaTeX 源码包等；选中后可用文件名预填论文名称。
              </p>
              <div
                class="paper-create-ms-upload-drop"
                :class="{
                  'paper-create-ms-upload-drop--has-file': !!createManuscriptUploadFile,
                  'paper-create-ms-upload-drop--disabled': createManuscriptSubmitting,
                }"
                role="button"
                tabindex="0"
                @click="openCreateManuscriptUploadPicker"
                @keydown.enter.prevent="openCreateManuscriptUploadPicker"
                @keydown.space.prevent="openCreateManuscriptUploadPicker"
                @dragover.prevent
                @drop.prevent="onCreateManuscriptUploadDrop"
              >
                <input
                  ref="createManuscriptUploadInputRef"
                  type="file"
                  class="paper-create-ms-upload-file-input"
                  accept=".pdf,.doc,.docx,.tex,.zip,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                  :disabled="createManuscriptSubmitting"
                  @change="onCreateManuscriptUploadChange"
                />
                <p v-if="!createManuscriptUploadFile" class="paper-create-ms-upload-hint">
                  点击或拖拽文件到此处
                </p>
                <p v-else class="paper-create-ms-upload-name">{{ createManuscriptUploadFile.name }}</p>
              </div>
            </div>
          </div>
        </div>
        <footer class="paper-message-box__btns">
          <el-button
            class="paper-message-box__confirm"
            type="primary"
            :loading="createManuscriptSubmitting"
            @click="submitCreateManuscriptDialog"
          >
            创建
          </el-button>
          <el-button class="paper-message-box__cancel" @click="closeCreateManuscriptDialog(false)">
            取消
          </el-button>
        </footer>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="topicCompletionPromptVisible"
      class="paper-modal-overlay paper-topic-complete-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="paper-topic-complete-title"
      @keydown.escape="dismissTopicCompletionPromptToLiteratureReview"
    >
      <div
        class="paper-message-box paper-message-box--default paper-modal-panel paper-topic-complete-panel"
        @click.stop
      >
        <header class="paper-modal-header">
          <h2 id="paper-topic-complete-title" class="paper-modal-title">选题发现已完成</h2>
        </header>
        <div class="paper-modal-body">
          <p class="paper-topic-complete-lead">文献综述已生成完成。</p>
        </div>
        <footer class="paper-message-box__btns paper-topic-complete-btns">
          <el-button
            class="paper-message-box__confirm"
            type="primary"
            @click="dismissTopicCompletionPromptToLiteratureReview"
          >
            确认
          </el-button>
        </footer>
      </div>
    </div>
  </Teleport>

  <div class="paper-workbench">
    <aside class="paper-sidebar">
      <div class="paper-sidebar-brand">
        <span class="paper-sidebar-logo" aria-hidden="true">◆</span>
        <div>
          <div class="paper-sidebar-title">AIRS</div>
          <div class="paper-sidebar-sub">AI Research Studio</div>
        </div>
      </div>

      <nav class="paper-nav" aria-label="科研工作流">
        <div v-for="group in PAPER_MODULE_GROUPS" :key="group.id" class="paper-nav-group">
          <div class="paper-nav-group-head">
            <span class="paper-nav-group-chevron" aria-hidden="true">▾</span>
            <span class="paper-nav-group-label">{{ group.label }}</span>
          </div>
          <div class="paper-nav-group-items">
            <button
              v-for="moduleId in group.moduleIds"
              :key="moduleId"
              type="button"
              class="paper-nav-item paper-nav-item--child"
              :class="{ 'paper-nav-item--active': activeModule === moduleId }"
              @click="selectModule(moduleId)"
            >
              <PaperModuleNavIcon :module-id="moduleId" class="paper-nav-item-icon" />
              <span class="paper-nav-item-label">{{ getPaperModuleMeta(moduleId).label }}</span>
            </button>
          </div>
        </div>
      </nav>
    </aside>

    <main ref="paperMainRef" class="paper-main">
      <header class="paper-main-head">
        <div>
          <h1 class="paper-main-title">{{ currentMeta.label }}</h1>
          <p v-if="showManuscriptContextInHeader" class="paper-main-manuscript">
            <span class="paper-main-manuscript-label">当前论文：</span>
            <span
              class="paper-main-manuscript-picker"
              :class="{ 'paper-main-manuscript-picker--busy': manuscriptSwitching }"
            >
              <PaperSelect
                :model-value="activeManuscriptId"
                :options="manuscriptSelectOptions"
                variant="inline"
                placeholder="请选择论文"
                filterable
                filter-placeholder="搜索论文标题"
                :disabled="!activeManuscripts.length || manuscriptSwitching"
                @update:model-value="onManuscriptChange"
              />
              <span
                v-if="manuscriptSwitching"
                class="paper-main-manuscript-picker-loading"
                aria-label="切换中"
              />
            </span>
          </p>
          <p class="paper-main-desc">{{ currentMeta.description }}</p>
        </div>
        <button
          v-if="showPrimaryAction"
          type="button"
          class="paper-run-btn"
          :disabled="topicPrimaryDisabled"
          @click="onPrimaryAction"
        >
          <span class="paper-run-icon" aria-hidden="true">▶</span>
          {{ primaryActionLabel }}
        </button>
        <button
          v-else-if="isExperimentPlanningModule"
          type="button"
          class="paper-run-btn"
          @click="onExpPlanUploadPlanPrimary"
        >
          <span class="paper-run-icon paper-run-icon--upload" aria-hidden="true">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M12 3v12" />
              <path d="m7 8 5-5 5 5" />
              <path d="M5 21h14" />
            </svg>
          </span>
          上传实验方案
        </button>
        <button
          v-else-if="isPaperSubmissionModule"
          type="button"
          class="paper-run-btn"
          @click="onPaperSubmissionNewPrimary"
        >
          <span class="paper-run-icon paper-run-icon--add" aria-hidden="true">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M12 5v14" />
              <path d="M5 12h14" />
            </svg>
          </span>
          新建投递
        </button>
        <button
          v-else-if="isLiteratureReviewModule"
          type="button"
          class="paper-run-btn"
          @click="onUploadExperimentPlanPrimary"
        >
          <span class="paper-run-icon paper-run-icon--upload" aria-hidden="true">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M12 3v12" />
              <path d="m7 8 5-5 5 5" />
              <path d="M5 21h14" />
            </svg>
          </span>
          上传文献综述
        </button>
        <button
          v-else-if="isUserLiteratureModule"
          type="button"
          class="paper-run-btn"
          @click="onUserLiteratureUploadPrimary"
        >
          <span class="paper-run-icon paper-run-icon--upload" aria-hidden="true">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M12 3v12" />
              <path d="m7 8 5-5 5 5" />
              <path d="M5 21h14" />
            </svg>
          </span>
          上传文献
        </button>
        <button
          v-else-if="isFigureGenerationModule"
          type="button"
          class="paper-run-btn"
          @click="onFigureUploadPrimary"
        >
          <span class="paper-run-icon paper-run-icon--upload" aria-hidden="true">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M12 3v12" />
              <path d="m7 8 5-5 5 5" />
              <path d="M5 21h14" />
            </svg>
          </span>
          上传图表
        </button>
        <button
          v-else-if="isMyManuscriptsModule"
          type="button"
          class="paper-run-btn"
          @click="onAddManuscriptClick"
        >
          <span class="paper-run-icon paper-run-icon--add" aria-hidden="true">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M12 5v14" />
              <path d="M5 12h14" />
            </svg>
          </span>
          添加论文
        </button>
      </header>

      <div class="paper-gate" role="status">
        <span class="paper-gate-icon" aria-hidden="true">📖</span>
        <p>
          <strong>参考文献门禁已启用：</strong>
          未验证文献不能进入最终引用、BibTeX 或论文正文。
        </p>
      </div>

      <div class="paper-module-body">
      <div
        v-if="showGenericModuleLoading"
        class="paper-module-loading"
        role="status"
        aria-live="polite"
      >
        <div class="paper-module-loading-spinner" aria-hidden="true" />
        <p class="paper-module-loading-text">加载中…</p>
      </div>

      <!-- 选题发现（与运行进度同属一块，避免 v-else-if 链误绑） -->
      <template v-else-if="activeModule === 'topic-discovery'">
      <section
        v-if="topicDiscoveryPageLoading"
        class="paper-wf-panel"
        role="status"
        aria-live="polite"
      >
        <p class="paper-wf-meta">正在加载选题发现…</p>
      </section>
      <template v-else>
      <section class="paper-panel">
        <label class="paper-field paper-field--block">
          <span class="paper-label">检索关键词 <em class="req">*</em></span>
          <input
            v-model="topicForm.keywords"
            type="text"
            class="paper-input"
            placeholder="多个关键词用逗号分隔"
          />
        </label>

        <label class="paper-field paper-field--block paper-field--after-keywords">
          <span class="paper-label">详细描述 <em class="req">*</em></span>
          <textarea
            v-model="topicForm.description"
            class="paper-textarea"
            rows="5"
            placeholder="研究的背景，研究的内容"
          />
        </label>

        <div class="paper-field-grid">
          <label class="paper-field">
            <span class="paper-label">学科</span>
            <div
              class="paper-input paper-input--like-discipline paper-input--readonly"
              :class="{ 'paper-input--discipline-unset': topicDisciplineUnset }"
              aria-readonly="true"
            >
              {{ topicDisciplineDisplayLabel }}
            </div>
            <span class="paper-hint">与当前论文创建时选定的学科一致，不可在此修改</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">目标会议/期刊 <em class="req">*</em></span>
            <PaperMultiSelect
              v-model="topicForm.venueCodes"
              :options="venueSelectOptions"
              placeholder="请选择（可多选）"
              filterable
              :disabled="topicRunBusy || currentTopicRun.status === 'checkpoint'"
            />
            <span class="paper-hint">用于约束贡献类型、实验标准和写作风格；可多选。</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">执行强度</span>
            <PaperSelect v-model="topicForm.intensity" :options="intensityOptions" />
            <span class="paper-hint">控制 /idea-discovery 深度与迭代轮数。</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">审计等级</span>
            <PaperSelect v-model="topicForm.auditLevel" :options="auditOptions" />
            <span class="paper-hint">控制 citation audit、claim audit、kill argument 等门禁强度。</span>
          </label>
        </div>

        <div class="paper-field paper-field--block paper-field--section-gap">
          <span class="paper-label">文献来源</span>
          <div class="paper-check-group">
            <label
              v-for="src in literatureSourceOptions"
              :key="src.code"
              class="paper-check paper-check--inline"
            >
              <input
                type="checkbox"
                :checked="isTopicSourceChecked(src.code)"
                :disabled="
                  topicRunBusy ||
                  currentTopicRun.status === 'checkpoint' ||
                  src.status !== 'active'
                "
                @change="toggleTopicSource(src.code, ($event.target as HTMLInputElement).checked)"
              />
              <span>{{ src.label }}</span>
            </label>
          </div>
          <label class="paper-check paper-check--nested">
            <input
              v-model="topicForm.randomReferences"
              type="checkbox"
              :disabled="topicRunBusy || currentTopicRun.status === 'checkpoint'"
            />
            <span class="paper-check-row">
              <span>每次运行随机参考文献</span>
              <ElTooltip
                content="勾选后，每次随机获取第三方论文，否则会顺序获取，可能多次获取论文是一样"
                placement="top"
              >
                <button
                  type="button"
                  class="paper-field-help"
                  aria-label="随机参考文献说明"
                  @click.prevent
                >
                  ?
                </button>
              </ElTooltip>
            </span>
          </label>

          <div class="paper-my-literature-block">
            <span class="paper-label">我的文献</span>
            <div class="paper-my-literature-row">
              <button
                type="button"
                class="paper-my-literature-btn"
                :disabled="topicRunBusy || currentTopicRun.status === 'checkpoint'"
                @click="openTopicUserLiteraturePicker"
              >
                选择文献
              </button>
              <span
                v-if="topicForm.userLiteratureIds.length"
                class="paper-hint paper-my-literature-summary"
              >
                已选 {{ topicForm.userLiteratureIds.length }} 篇文献
              </span>
            </div>
          </div>
        </div>

        <PaperTopicUserLiteraturePickDialog
          :open="topicUserLitPickerOpen"
          :initial-selected-ids="topicForm.userLiteratureIds"
          :disabled="topicRunBusy || currentTopicRun.status === 'checkpoint'"
          @close="closeTopicUserLiteraturePicker"
          @confirm="onTopicUserLiteraturePicked"
        />

        <div class="paper-field paper-field--block paper-field--after-sources">
          <span class="paper-label">人工检查</span>
          <div class="paper-check-group">
            <label class="paper-check paper-check--inline paper-check--locked">
              <input
                type="checkbox"
                checked
                disabled
                aria-checked="true"
              />
              <span>人工确认关键步骤</span>
            </label>
          </div>
        </div>
      </section>

      <section
        v-if="topicRunVisible"
        ref="topicFlowPanelRef"
        class="paper-panel paper-panel--flow"
      >
        <div class="paper-flow-head">
          <h2 class="paper-panel-title paper-panel-title--tight">运行进度</h2>
          <span
            class="paper-flow-badge"
            :class="
              topicAtHumanPause
                ? 'paper-flow-badge--checkpoint'
                : topicRunBusy && currentTopicRun.status !== 'failed'
                  ? 'paper-flow-badge--running'
                  : `paper-flow-badge--${currentTopicRun.status}`
            "
          >
            {{
              topicAtHumanPause
                ? '等待人工确认'
                : topicRunBusy && currentTopicRun.status !== 'completed' && currentTopicRun.status !== 'failed'
                  ? '执行中'
                  : currentTopicRun.status === 'checkpoint'
                    ? '等待人工确认'
                    : currentTopicRun.status === 'running'
                      ? '执行中'
                      : currentTopicRun.status === 'completed'
                        ? '已完成'
                        : currentTopicRun.status === 'failed'
                          ? '执行失败'
                          : ''
            }}
          </span>
          <button
            v-if="topicAtHumanPause || topicRunBusy || currentTopicRun.status === 'checkpoint'"
            type="button"
            class="paper-flow-cancel"
            :disabled="topicTerminateLoading || topicContinueLoading"
            @click="cancelTopicRun"
          >
            取消
          </button>
        </div>
        <ol class="paper-flow-steps">
          <li
            v-for="step in currentTopicRun.steps"
            :key="step.stageCode"
            class="paper-flow-step"
            :class="`paper-flow-step--${step.status}`"
          >
            <span class="paper-flow-step-icon" aria-hidden="true">
              <span
                v-if="step.status === 'running' && !(step.stageCode === 'idea_discovery' && topicAtHumanPause)"
                class="paper-flow-step-spinner"
              />
              <template v-else>{{ topicStepIcon(step.status) }}</template>
            </span>
            <div class="paper-flow-step-body">
              <span class="paper-flow-step-label">{{ step.label }}</span>
              <span
                v-if="step.status === 'running' && !(step.stageCode === 'idea_discovery' && topicAtHumanPause)"
                class="paper-flow-step-running"
                role="status"
              >
                {{ topicStepRunningHint(step.stageCode) }}
              </span>
              <ul
                v-if="step.stageCode === 'idea_discovery' || step.stageCode === 'topic_discovery'"
                class="paper-flow-subphases"
                aria-label="Idea Discovery 分段进度"
              >
                <li
                  v-for="row in topicArisPhaseRows"
                  :key="row.id"
                  class="paper-flow-subphase"
                  :class="`paper-flow-subphase--${row.status}`"
                >
                  <span class="paper-flow-subphase-main">
                    <span
                      v-if="row.status === 'waiting_confirm' || row.status === 'running'"
                      class="paper-flow-subphase-spinner"
                      aria-hidden="true"
                    />
                    <span class="paper-flow-subphase-name">{{ row.label }}</span>
                  </span>
                  <span class="paper-flow-subphase-status">{{ arisPhaseStatusLabel(row.status) }}</span>
                </li>
              </ul>
            </div>
          </li>
        </ol>

        <div
          v-if="retrieveLiteratureLinks.length"
          class="paper-retrieve-links"
        >
          <h3 class="paper-retrieve-links-title">检索文献（{{ retrieveLiteratureLinks.length }} 篇，可打开）</h3>
          <ul class="paper-retrieve-links-list">
            <li v-for="(item, i) in retrieveLiteratureLinks" :key="item.external_key ?? item.url ?? i">
              <a :href="item.url" target="_blank" rel="noopener noreferrer">{{ item.title || item.url }}</a>
              <span v-if="item.source_code" class="paper-retrieve-links-src">{{ item.source_code }}</span>
            </li>
          </ul>
        </div>

        <div v-if="topicShowRoundConfirmPanel" class="paper-checkpoint paper-round-confirm">
          <div class="paper-checkpoint-head">
            <span class="paper-checkpoint-pause" aria-hidden="true">⏸</span>
            <div>
              <h3 class="paper-checkpoint-title">第三步 · 生成文献综述、实验方案</h3>
              <p class="paper-checkpoint-sub">
                各轮含生成包与门禁审计；确认后将写入文献综述与实验方案表。
              </p>
            </div>
          </div>
          <div v-if="!topicDiscoveryIterations.length" class="paper-checkpoint-note">
            暂无迭代结果，请刷新或重新运行选题发现。
          </div>
          <ul v-else class="paper-round-list">
            <li
              v-for="item in topicDiscoveryIterations"
              :key="item.round"
              class="paper-round-card"
              :class="{ 'paper-round-card--selected': topicSelectedRound === item.round }"
            >
              <label class="paper-round-card-label">
                <input
                  v-model="topicSelectedRound"
                  type="radio"
                  name="topic-round"
                  :value="item.round"
                />
                <span class="paper-round-card-title">
                  第 {{ item.round }} 轮
                  <span
                    class="paper-round-badge"
                    :class="item.passed ? 'paper-round-badge--ok' : 'paper-round-badge--warn'"
                  >
                    {{ item.passed ? '门禁通过' : '未完全通过' }}
                  </span>
                </span>
              </label>
              <p v-if="item.auditSummary" class="paper-round-line">
                <strong>审计：</strong>{{ item.auditSummary }}
              </p>
              <p v-if="item.litTitle || item.litSummary" class="paper-round-line">
                <strong>综述：</strong>
                {{ item.litTitle || '（无标题）' }}
                <span v-if="item.litSummary"> — {{ item.litSummary.slice(0, 160) }}</span>
              </p>
              <p v-if="item.hasExperimentPlan || item.planTitle || item.planSummary" class="paper-round-line">
                <strong>实验：</strong>
                {{ item.planTitle || '（实验方案）' }}
                <span v-if="item.planSummary"> — {{ item.planSummary.slice(0, 120) }}</span>
              </p>
              <p
                v-else-if="!item.hasExperimentPlan && !item.planTitle && !item.planSummary"
                class="paper-round-line paper-round-muted"
              >
                实验：本方向无需实验或未生成方案
              </p>
              <ul v-if="item.ideaTitles.length" class="paper-round-ideas">
                <li v-for="(t, j) in item.ideaTitles.slice(0, 4)" :key="j">{{ t }}</li>
              </ul>
            </li>
          </ul>
          <div class="paper-checkpoint-actions">
            <button
              type="button"
              class="paper-btn-secondary"
              :disabled="topicTerminateLoading || topicConfirmRoundLoading"
              @click="rejectTopicCheckpoint"
            >
              {{ topicTerminateLoading ? '终止中…' : '终止并返回' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary"
              :disabled="
                topicTerminateLoading ||
                topicConfirmRoundLoading ||
                !topicDiscoveryIterations.length
              "
              @click="confirmTopicSelectedRound"
            >
              {{ topicConfirmRoundLoading ? '写入中…' : '确认并写入' }}
            </button>
          </div>
        </div>

        <div v-else-if="topicShowCheckpointPanel" class="paper-checkpoint">
          <div v-if="topicCheckpointPanel" class="paper-checkpoint-head">
            <span class="paper-checkpoint-pause" aria-hidden="true">⏸</span>
            <div>
              <h3 class="paper-checkpoint-title">人工检查点 · {{ topicCheckpointHeadingSuffix }}</h3>
              <p
                v-if="
                  topicCheckpointPanel.title &&
                  !topicCheckpointPanel.title.includes('段间检查点')
                "
                class="paper-checkpoint-sub"
              >
                {{ topicCheckpointPanel.title }}
              </p>
            </div>
          </div>
          <div v-else class="paper-checkpoint-head">
            <span class="paper-checkpoint-pause" aria-hidden="true">⏸</span>
            <h3 class="paper-checkpoint-title">人工检查点</h3>
          </div>
          <ul v-if="topicCheckpointPanel?.lines.length" class="paper-checkpoint-list">
            <li v-for="(line, i) in topicCheckpointPanel.lines" :key="i">{{ line }}</li>
          </ul>
          <div class="paper-checkpoint-actions">
            <button
              type="button"
              class="paper-btn-secondary"
              :disabled="topicTerminateLoading || topicContinueLoading"
              @click="rejectTopicCheckpoint"
            >
              {{ topicTerminateLoading ? '终止中…' : '终止并返回' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary"
              :disabled="topicTerminateLoading || topicContinueLoading"
              @click="continueTopicAfterCheckpoint"
            >
              {{ topicContinueLoading ? '执行中…' : '确认并继续' }}
            </button>
          </div>
        </div>

        <div v-else-if="topicShowFailedPanel" class="paper-checkpoint paper-checkpoint--failed">
          <div class="paper-checkpoint-head">
            <span class="paper-checkpoint-pause paper-checkpoint-pause--failed" aria-hidden="true">✕</span>
            <div>
              <h3 class="paper-checkpoint-title">
                步骤失败 · {{ topicFailedSummary?.stageLabel ?? '未知步骤' }}
              </h3>
            </div>
          </div>
          <p class="paper-flow-failed-error" role="alert">
            {{ topicFailedSummary?.error ?? '本步执行失败，请查看日志或调整配置后重试。' }}
          </p>
          <p class="paper-checkpoint-note">
            日志路径：<code>inout/user_&lt;uid&gt;/manuscript_&lt;id&gt;/run_&lt;v&gt;/paper/.paper/invoke/logs/</code>
          </p>
          <div class="paper-checkpoint-actions">
            <button
              type="button"
              class="paper-btn-primary"
              :disabled="topicTerminateLoading"
              @click="rejectFailedTopicRun"
            >
              {{ topicTerminateLoading ? '处理中…' : '驳回并返回' }}
            </button>
          </div>
        </div>

        <div v-else-if="currentTopicRun.status === 'completed'" class="paper-flow-done">
          三步已完成：语料已入库，所选轮次的文献综述与实验方案已写入产出表。下一步请点顶栏
          <strong>「生成文献综述」</strong>；实验方案请在「实验方案」模块查看。
        </div>
      </section>
      </template>
      </template>

      <PaperMyManuscriptsPanel
        v-else-if="activeModule === 'my-manuscripts'"
        :key="`my-manuscripts-${moduleContentKey}`"
        :active-manuscript-id="activeManuscriptId"
        :manuscript-switching="manuscriptSwitching"
        @start-writing="onStartWritingFromMyManuscripts($event.id)"
        @start-submission="onStartSubmissionFromMyManuscripts($event.id)"
      />

      <PaperSubmissionHubPanel
        v-else-if="activeModule === 'paper-submission'"
        ref="paperSubmissionHubRef"
        :key="`paper-submission-${moduleContentKey}`"
        :active-manuscript-id="activeManuscriptId"
        :manuscript-title="currentManuscript?.title ?? '未命名'"
        :draft-version="currentManuscript?.draftVersion"
        @manuscript-summary-changed="void loadManuscriptsFromServer()"
        @start-revision-writing="onStartRevisionWritingFromSubmission"
      />

      <PaperUserLiteraturePanel
        v-else-if="activeModule === 'user-literature'"
        ref="userLiteraturePanelRef"
        :key="`user-literature-${moduleContentKey}`"
      />

      <PaperFigureManagementPanel
        v-else-if="activeModule === 'figure-generation'"
        ref="figureManagementPanelRef"
        :key="`figure-generation-${moduleContentKey}`"
      />

      <PaperEnvironmentPanel
        v-else-if="activeModule === 'environment'"
        :key="`environment-${moduleContentKey}`"
        :env-preference="envPreference"
        @environment-saved="onEnvironmentSaved"
      />

      <PaperCourseReportPanel
        v-else-if="activeModule === 'course-report'"
        :key="`course-report-${moduleContentKey}`"
      />

      <PaperInviteRebatePanel
        v-else-if="activeModule === 'invite-rebate'"
        :key="`invite-rebate-${moduleContentKey}`"
      />

      <PaperPersonalCenterPanel
        v-else-if="activeModule === 'personal-center'"
        :key="`personal-center-${moduleContentKey}`"
        ref="personalCenterPanelRef"
        :manuscript-id="activeManuscriptId"
        :manuscript-title="currentManuscript?.title ?? '未命名'"
      />

      <PaperWorkflowPanels
        v-else-if="isSecondaryWorkflowModule"
        :key="activeModule"
        ref="workflowPanelsRef"
        :module-id="activeModule"
        :manuscript-id="activeManuscriptId"
        :manuscript-title="currentManuscript?.title ?? '未命名'"
        :manuscript-draft-version="currentManuscript?.draftVersion"
        :pending-open-experiment-plan-id="pendingOpenExperimentPlanId"
        :pending-open-literature-review-id="pendingOpenLiteratureReviewId"
        @navigate-module="onNavigateModule"
        @consumed-pending-experiment-plan="onConsumedPendingExperimentPlan"
        @consumed-pending-literature-review="onConsumedPendingLiteratureReview"
      />
      </div>
    </main>
  </div>
</template>

<style scoped>
.paper-workbench {
  display: flex;
  height: 100vh;
  max-height: 100vh;
  overflow: hidden;
  background: var(--atm-bg, #f5f3ff);
}

.paper-sidebar {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  width: 220px;
  min-height: 0;
  padding: 20px 12px 16px;
  color: #fff;
  background: linear-gradient(180deg, #1e1b4b 0%, #5b21b6 55%, #6366f1 100%);
  overflow: hidden;
}

.paper-sidebar-brand {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 4px 10px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
}

.paper-sidebar-logo {
  font-size: 18px;
  color: #c4b5fd;
}

.paper-sidebar-title {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.paper-sidebar-sub {
  margin-top: 2px;
  font-size: 11px;
  opacity: 0.75;
}

.paper-nav {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
  margin-top: 20px;
  overflow-x: hidden;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.paper-nav-group {
  --nav-label-inset: 28px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.paper-nav-group-head {
  display: grid;
  grid-template-columns: 12px 1fr;
  column-gap: 6px;
  align-items: center;
  padding: 9px 10px;
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.03em;
  color: #fff;
  user-select: none;
}

.paper-nav-group-chevron {
  flex-shrink: 0;
  width: 12px;
  font-size: 10px;
  line-height: 1;
  color: rgba(255, 255, 255, 0.55);
}

.paper-nav-group-label {
  min-width: 0;
}

.paper-nav-group-items {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-left: 0;
}

.paper-nav-item {
  display: flex;
  gap: 8px;
  align-items: center;
  padding: 10px 12px;
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.82);
  text-align: left;
  background: transparent;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.12s;
}

.paper-nav-item-icon {
  flex-shrink: 0;
  color: rgba(255, 255, 255, 0.72);
}

.paper-nav-item--active .paper-nav-item-icon {
  color: #fff;
}

.paper-nav-item-label {
  min-width: 0;
}

.paper-nav-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.paper-nav-item--active {
  color: #fff;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.14);
  box-shadow: inset 2px 0 0 #fff;
}

.paper-nav-item--active:hover {
  background: rgba(255, 255, 255, 0.18);
}

.paper-nav-item--child {
  padding: 10px 12px 10px var(--nav-label-inset);
  font-size: 14px;
}

.paper-main {
  flex: 1;
  min-width: 0;
  min-height: 0;
  padding: 28px 32px 40px;
  overflow-x: hidden;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.paper-module-body {
  display: flex;
  flex-direction: column;
  gap: 18px;
  min-height: 200px;
}

.paper-wf-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.paper-wf-meta {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.paper-module-loading {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 14px;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  color: var(--atm-text-muted, #64748b);
}

.paper-module-loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(124, 58, 237, 0.12);
  border-top-color: var(--atm-primary, #7c3aed);
  border-radius: 50%;
  animation: paper-module-spin 0.75s linear infinite;
}

.paper-module-loading-text {
  margin: 0;
  font-size: 14px;
}

@keyframes paper-module-spin {
  to {
    transform: rotate(360deg);
  }
}

.paper-main-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px 24px;
  margin-bottom: 20px;
}

.paper-main-title {
  margin: 0;
  font-size: 26px;
  font-weight: 800;
  color: var(--atm-text, #1e1b4b);
  letter-spacing: -0.03em;
}

.paper-main-manuscript {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 8px;
  align-items: center;
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--atm-text-muted, #64748b);
}

.paper-main-manuscript-label {
  flex-shrink: 0;
  font-weight: 500;
}

.paper-main-manuscript-picker {
  position: relative;
  display: inline-flex;
  max-width: min(100%, 560px);
}

.paper-main-manuscript-picker--busy :deep(.paper-select-trigger) {
  opacity: 0.88;
}

.paper-main-manuscript-picker-loading {
  position: absolute;
  top: 50%;
  right: -22px;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(100, 116, 139, 0.35);
  border-top-color: #6366f1;
  border-radius: 50%;
  pointer-events: none;
  transform: translateY(-50%);
  animation: paper-main-manuscript-picker-spin 0.65s linear infinite;
}

@keyframes paper-main-manuscript-picker-spin {
  to {
    transform: translateY(-50%) rotate(360deg);
  }
}

.paper-main-desc {
  margin: 8px 0 0;
  max-width: 640px;
  font-size: 14px;
  line-height: 1.55;
  color: var(--atm-text-muted, #64748b);
}

.paper-run-btn {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  padding: 12px 22px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  background: var(--atm-gradient, linear-gradient(135deg, #7c3aed, #6366f1));
  border: none;
  border-radius: 12px;
  cursor: pointer;
  box-shadow: 0 8px 24px rgba(91, 33, 182, 0.28);
  transition:
    transform 0.12s,
    opacity 0.12s;
}

.paper-run-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.paper-run-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.paper-run-icon {
  font-size: 12px;
}

.paper-run-icon--upload svg,
.paper-run-icon--add svg {
  display: block;
  width: 16px;
  height: 16px;
}

.paper-gate {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 22px;
  padding: 14px 16px;
  font-size: 14px;
  line-height: 1.5;
  color: #713f12;
  background: linear-gradient(90deg, #fef9c3 0%, #fef3c7 100%);
  border: 1px solid #fde68a;
  border-radius: 12px;
}

.paper-gate-icon {
  flex-shrink: 0;
  font-size: 18px;
}

.paper-gate p {
  margin: 0;
}

.paper-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.paper-panel-title {
  margin: 0 0 20px;
  font-size: 16px;
  font-weight: 700;
  color: var(--atm-text, #1e1b4b);
}

.paper-panel-title--tight {
  margin-bottom: 0;
}

.paper-panel--flow {
  scroll-margin-top: 24px;
  margin-top: 18px;
}

.paper-flow-head {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 14px;
  align-items: center;
  margin-bottom: 8px;
}

.paper-flow-badge {
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 700;
  border-radius: 999px;
}

.paper-flow-badge--running {
  color: #5b21b6;
  background: #ede9fe;
}

.paper-flow-badge--checkpoint {
  color: #b45309;
  background: #fef3c7;
}

.paper-flow-badge--completed {
  color: #15803d;
  background: #dcfce7;
}

.paper-flow-badge--failed {
  color: #b91c1c;
  background: #fee2e2;
}

.paper-flow-cancel {
  margin-left: auto;
  padding: 6px 12px;
  font-size: 13px;
  color: var(--atm-text-muted, #64748b);
  background: transparent;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
}

.paper-flow-cancel:hover:not(:disabled) {
  border-color: #cbd5e1;
  color: var(--atm-text, #1e1b4b);
}

.paper-flow-cancel:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.paper-flow-steps {
  margin: 20px 0 0;
  padding: 0;
  list-style: none;
}

.paper-flow-step {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  padding: 12px 0;
  border-left: 2px solid #e2e8f0;
  margin-left: 11px;
  padding-left: 22px;
  position: relative;
}

.paper-flow-step:last-child {
  border-left-color: transparent;
}

.paper-flow-step-icon {
  position: absolute;
  left: -12px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  font-size: 11px;
  font-weight: 800;
  color: #94a3b8;
  background: #fff;
  border: 2px solid #e2e8f0;
  border-radius: 50%;
}

.paper-flow-step--running .paper-flow-step-icon {
  color: #7c3aed;
  border-color: #7c3aed;
  animation: paper-flow-pulse 1s ease-in-out infinite;
}

.paper-flow-step-spinner {
  display: block;
  width: 12px;
  height: 12px;
  border: 2px solid #e9d5ff;
  border-top-color: #7c3aed;
  border-radius: 50%;
  animation: paper-flow-spin 0.75s linear infinite;
}

.paper-flow-step-running {
  flex: 1 1 100%;
  font-size: 13px;
  font-weight: 500;
  color: #7c3aed;
  animation: paper-flow-pulse 1.2s ease-in-out infinite;
}

.paper-flow-step--completed .paper-flow-step-icon {
  color: #fff;
  background: #7c3aed;
  border-color: #7c3aed;
}

.paper-flow-step--failed .paper-flow-step-icon {
  color: #fff;
  background: #dc2626;
  border-color: #dc2626;
}

.paper-flow-step-body {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 10px;
  align-items: center;
  min-height: 22px;
}

.paper-flow-step-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
}

.paper-flow-subphases {
  list-style: none;
  margin: 0.35rem 0 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  width: 100%;
}

.paper-flow-subphase {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  font-size: 0.875rem;
  line-height: 1.5;
  padding: 0.36rem 0.55rem;
  border-radius: 4px;
  background: var(--paper-surface-muted, rgba(0, 0, 0, 0.04));
  color: var(--paper-text-secondary, #666);
}

.paper-flow-subphase-main {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  min-width: 0;
  flex: 1;
}

.paper-flow-subphase-spinner {
  flex-shrink: 0;
  display: block;
  width: 12px;
  height: 12px;
  border: 2px solid #e9d5ff;
  border-top-color: #7c3aed;
  border-radius: 50%;
  animation: paper-flow-spin 0.75s linear infinite;
}

.paper-flow-subphase-name {
  word-break: break-all;
}

.paper-flow-subphase-status {
  flex-shrink: 0;
  font-size: 0.8125rem;
}

.paper-flow-subphase--completed {
  opacity: 0.85;
}

.paper-flow-subphase--running {
  background: rgba(99, 102, 241, 0.12);
  color: var(--paper-text, #222);
}

.paper-flow-subphase--waiting_confirm {
  background: rgba(245, 158, 11, 0.15);
  color: #b45309;
}

.paper-flow-subphase--awaiting_continue {
  background: rgba(59, 130, 246, 0.12);
  color: #1d4ed8;
}

.paper-retrieve-links {
  margin: 16px 0 0;
  padding: 14px 16px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
}

.paper-retrieve-links-title {
  margin: 0 0 10px;
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
}

.paper-retrieve-links-list {
  margin: 0;
  padding-left: 1.1rem;
  font-size: 13px;
  line-height: 1.55;
}

.paper-retrieve-links-list a {
  color: #2563eb;
  text-decoration: none;
}

.paper-retrieve-links-list a:hover {
  text-decoration: underline;
}

.paper-retrieve-links-src {
  margin-left: 8px;
  font-size: 11px;
  color: #94a3b8;
  font-family: ui-monospace, monospace;
}

.paper-checkpoint {
  margin-top: 20px;
  padding: 18px 20px;
  background: linear-gradient(180deg, #fffbeb 0%, #fff 40%);
  border: 1px solid #fde68a;
  border-radius: 14px;
}

.paper-checkpoint-head {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.paper-checkpoint-pause {
  font-size: 22px;
  line-height: 1;
}

.paper-checkpoint-title {
  margin: 0;
  font-size: 15px;
  font-weight: 800;
  color: #92400e;
}

.paper-checkpoint-sub {
  margin: 4px 0 0;
  font-size: 13px;
  color: #78716c;
}

.paper-checkpoint-list {
  margin: 14px 0 0;
  padding-left: 20px;
  font-size: 13px;
  line-height: 1.55;
  color: var(--atm-text, #1e1b4b);
}

.paper-checkpoint-note {
  margin: 12px 0 0;
  font-size: 12px;
  color: var(--atm-text-muted, #64748b);
}

.paper-checkpoint--failed {
  border-color: rgba(220, 38, 38, 0.35);
  background: rgba(254, 242, 242, 0.65);
}

.paper-checkpoint-pause--failed {
  color: #dc2626;
}

.paper-flow-failed-error {
  margin: 0 0 0.75rem;
  padding: 0.65rem 0.85rem;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(220, 38, 38, 0.2);
  font-size: 0.875rem;
  line-height: 1.45;
  color: #991b1b;
  word-break: break-word;
}

.paper-round-list {
  list-style: none;
  margin: 0 0 1rem;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.paper-round-card {
  border: 1px solid var(--paper-border, #e5e7eb);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  background: var(--paper-surface, #fff);
}

.paper-round-card--selected {
  border-color: var(--el-color-primary, #409eff);
  box-shadow: 0 0 0 1px var(--el-color-primary-light-7, #c6e2ff);
}

.paper-round-card-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-weight: 600;
}

.paper-round-card-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.paper-round-badge {
  font-size: 0.75rem;
  font-weight: 500;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
}

.paper-round-badge--ok {
  background: #ecfdf5;
  color: #047857;
}

.paper-round-badge--warn {
  background: #fffbeb;
  color: #b45309;
}

.paper-round-line {
  margin: 0.35rem 0 0;
  font-size: 0.875rem;
  line-height: 1.45;
}

.paper-round-muted {
  color: var(--paper-text-muted, #6b7280);
}

.paper-round-ideas {
  margin: 0.35rem 0 0;
  padding-left: 1.25rem;
  font-size: 0.8125rem;
}

.paper-checkpoint-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 16px;
}

.paper-btn-primary {
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: var(--atm-gradient, linear-gradient(135deg, #7c3aed, #6366f1));
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.paper-btn-secondary {
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #4338ca;
  background: var(--atm-primary-light, #ede9fe);
  border: 1px solid #c4b5fd;
  border-radius: 10px;
  cursor: pointer;
}

.paper-flow-done {
  margin-top: 16px;
  padding: 12px 14px;
  font-size: 13px;
  color: #15803d;
  background: #f0fdf4;
  border-radius: 10px;
}

@keyframes paper-flow-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes paper-flow-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(124, 58, 237, 0.35);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(124, 58, 237, 0);
  }
}

.paper-panel--flow code {
  font-size: 11px;
  padding: 1px 5px;
  background: #f1f5f9;
  border-radius: 4px;
}

.paper-field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px 20px;
  margin-top: 18px;
}

@media (max-width: 900px) {
  .paper-field-grid {
    grid-template-columns: 1fr;
  }

  .paper-workbench {
    flex-direction: column;
  }

  .paper-sidebar {
    flex-shrink: 0;
    width: 100%;
    max-height: 42vh;
  }

  .paper-main {
    flex: 1;
    min-height: 0;
  }
}

.paper-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.paper-field--block {
  margin-bottom: 4px;
}

.paper-field--after-keywords {
  margin-top: 16px;
}

.paper-field--section-gap,
.paper-field--after-sources {
  margin-top: 28px;
}

.paper-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
}

.paper-label .req {
  color: #e53935;
  font-style: normal;
}

.paper-input,
.paper-textarea {
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  color: var(--atm-text, #1e1b4b);
  appearance: none;
  -webkit-appearance: none;
  background-color: #fff;
  border: 1px solid var(--paper-field-border, #e2e8f0);
  border-radius: 10px;
  transition: border-color 0.12s, background-color 0.12s;
}

.paper-input:-webkit-autofill,
.paper-input:-webkit-autofill:focus,
.paper-textarea:-webkit-autofill,
.paper-textarea:-webkit-autofill:focus {
  -webkit-text-fill-color: var(--atm-text, #1e1b4b);
  box-shadow: 0 0 0 1000px #fff inset;
  transition: background-color 99999s ease-out 0s;
}

.paper-textarea {
  resize: vertical;
  min-height: 100px;
  font-family: inherit;
  line-height: 1.5;
}

.paper-input:focus,
.paper-textarea:focus {
  outline: none;
  border-color: #94a3b8;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.22);
}

.paper-input--readonly.paper-input--like-discipline {
  background-color: var(--paper-discipline-field-bg, #eef2f6);
  border-color: var(--paper-field-border, #e2e8f0);
}

.paper-input--readonly {
  cursor: default;
  user-select: none;
}

.paper-input--like-discipline:not(.paper-input--discipline-unset) {
  color: var(--atm-text, #1e1b4b);
}

.paper-input--discipline-unset {
  color: #94a3b8;
}

.paper-panel :deep(.paper-select-trigger) {
  background-color: #fff;
  border-color: var(--paper-field-border, #e2e8f0);
}

.paper-hint {
  font-size: 12px;
  line-height: 1.45;
  color: var(--atm-text-muted, #64748b);
}

.paper-hint--inline {
  display: block;
  margin-top: 2px;
  font-weight: 400;
}

.paper-hint--block {
  margin: 10px 0 0;
}

.paper-check {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-top: 22px;
  font-size: 14px;
  color: var(--atm-text, #1e1b4b);
  cursor: pointer;
}

.paper-check input {
  margin-top: 3px;
  width: 16px;
  height: 16px;
  accent-color: var(--atm-primary, #7c3aed);
}

.paper-panel--placeholder {
  color: var(--atm-text-muted, #64748b);
}

.paper-placeholder-lead {
  margin: 0 0 16px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--atm-text, #1e1b4b);
}

.paper-placeholder-list {
  margin: 0;
  padding-left: 20px;
  font-size: 14px;
  line-height: 1.7;
}

.paper-panel--env {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.paper-section-lead {
  margin: -8px 0 16px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--atm-text-muted, #64748b);
}

.paper-divider {
  margin: 22px 0;
  border: none;
  border-top: 1px solid #e8eaf0;
}

.paper-check-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 20px;
  margin-top: 8px;
}

.paper-check--inline {
  margin-top: 0;
  align-items: center;
}

.paper-check--inline input[type='checkbox'] {
  margin-top: 0;
  flex-shrink: 0;
}

.paper-check--inline > span {
  line-height: 16px;
}

.paper-check--locked {
  opacity: 0.72;
  cursor: not-allowed;
  user-select: none;
}

.paper-check--nested {
  margin-top: 12px;
  align-items: center;
}

.paper-check--nested input[type='checkbox'] {
  margin-top: 0;
  flex-shrink: 0;
}

.paper-my-literature-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 16px;
}

.paper-my-literature-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 14px;
  align-items: center;
}

.paper-my-literature-btn {
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

.paper-my-literature-btn:hover:not(:disabled) {
  color: #4f46e5;
  background: #eef2ff;
  border-color: #a5b4fc;
}

.paper-my-literature-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.paper-my-literature-summary {
  margin: 0;
  flex: 1;
  min-width: 200px;
}

.paper-check-row {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  line-height: 1.4;
}

.paper-check-row :deep(.el-tooltip__trigger) {
  display: inline-flex;
  align-items: center;
  line-height: 0;
}

.paper-field-help {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  padding: 0;
  border: 1px solid #cbd5e1;
  border-radius: 50%;
  background: #fff;
  color: #64748b;
  font-size: 12px;
  line-height: 1;
  cursor: help;
}

.paper-field-help:hover {
  border-color: var(--atm-primary, #7c3aed);
  color: var(--atm-primary, #7c3aed);
}

.paper-field--span2 {
  grid-column: 1 / -1;
}

@media (min-width: 900px) {
  .paper-field--span2 {
    grid-column: span 2;
  }
}
</style>
