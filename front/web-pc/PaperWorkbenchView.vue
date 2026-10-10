<script setup lang="ts">
import { computed, nextTick, onUnmounted, reactive, ref, watch } from 'vue'
import { loadWorkbenchModuleContent } from '@/composables/loadWorkbenchModule'
import {
  reloadTopicDiscoveryFormOptions,
  useTopicDiscoveryFormOptions,
} from '@/composables/useTopicDiscoveryFormOptions'
import {
  postTopicDiscoveryRun,
  fetchCurrentTopicDiscoveryRun,
  cancelTopicDiscoveryRun,
  commitTopicDiscoveryManuscript,
  parseRetrieveLiteratureLinks,
  applyStoredTopicRunInputToForm,
  isTopicRunTerminalCompleted,
  type TopicDiscoveryRunResponse,
  type TopicDiscoveryStepDTO,
} from '@/api/topicDiscovery'
import {
  fetchPaperManuscripts,
  createPaperManuscript,
  setCurrentPaperManuscript,
  type PaperManuscriptListResponse,
} from '@/api/manuscripts'
import { ElMessage } from 'element-plus'
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
import { DEMO_MODULE_TOKEN_ESTIMATES } from './demoOperationLogs'
import PaperModuleNavIcon from './PaperModuleNavIcon.vue'
import PaperMyManuscriptsPanel from './PaperMyManuscriptsPanel.vue'
import PaperInviteRebatePanel from './PaperInviteRebatePanel.vue'
import PaperPersonalCenterPanel from './PaperPersonalCenterPanel.vue'
import PaperSelect from './PaperSelect.vue'
import PaperWorkflowPanels from './PaperWorkflowPanels.vue'

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
  generate_ideas_ready: {
    title: '候选选题（脑暴）',
    lines: [
      'Idea A：稀疏注意力 + 动态路由 — 一句话 claim + 12 条 reference keys',
      'Idea B：层级 KV 压缩 — 可测指标：吞吐 / 困惑度',
      'Idea C：训练无关 token 合并 — 偏系统向，需补实验资源说明',
    ],
  },
  audit_ready: {
    title: '审查结论与文献综述',
    lines: [
      '贴题复核：第二步脑暴与新颖性是否与方向一致（演示）',
      '审查结论：blocker/major/minor 与 off_topic 说明',
      '已生成文献综述摘要，写入 paper_output_literature_review（演示）',
    ],
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
    if (raw && PAPER_MODULES.some((m) => m.id === raw && m.id !== 'environment')) {
      return raw as PaperModuleId
    }
  } catch {
    /* ignore */
  }
  return 'topic-discovery'
}

function persistActiveModule(id: PaperModuleId) {
  if (id === 'environment') return
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
    activeModule.value !== 'experiment-planning',
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
      const msId = activeManuscriptId.value
      if (msId) await hydrateTopicRunFromServer(msId)
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

function onNavigateModule(id: PaperModuleId, openExperimentPlanId?: string) {
  const planId = openExperimentPlanId?.trim()
  pendingOpenExperimentPlanId.value = planId || null
  selectModule(id)
}

function onConsumedPendingExperimentPlan() {
  pendingOpenExperimentPlanId.value = null
}

const running = ref(false)
const topicTerminateLoading = ref(false)
const topicRunToken = ref(0)
/** POST 新建 run 完成前，背景轮询不应用 GET /current（避免仍读到上一轮 failed） */
const topicRunStartInFlight = ref(false)
/** 检查点「确认并继续」请求进行中 */
const topicContinueLoading = ref(false)
/** continue 后乐观标 running 的阶段，避免 POST/轮询仍返回 checkpoint 把转圈打掉 */
const topicContinueOptimisticStage = ref<string | null>(null)
/** 本轮选题 run 由用户启动/继续，完成后弹窗引导去文献综述 */
const topicRunAwaitingCompletionPrompt = ref(false)
const topicCompletionPromptVisible = ref(false)
const topicRunsByManuscript = ref<Record<string, TopicRunDemo>>({})
const litReviewDoneByManuscript = ref<Record<string, boolean>>({})
const experimentPlanDoneByManuscript = ref<Record<string, boolean>>({})
const workflowPanelsRef = ref<InstanceType<typeof PaperWorkflowPanels> | null>(null)
const figureManagementTab = ref<'upload' | 'generate'>('upload')
const paperMainRef = ref<HTMLElement | null>(null)
const topicFlowPanelRef = ref<HTMLElement | null>(null)

const topicForm = reactive<TopicDiscoveryForm>({ ...DEFAULT_TOPIC_DISCOVERY })
const envPreference = reactive<EnvironmentPreferenceForm>({ ...DEFAULT_ENV_PREFERENCE })

const {
  disciplineSelectOptions,
  intensityOptions,
  auditOptions,
  ready: topicFormOptionsReady,
  applyCatalogIntensityAuditDefaults,
  literatureSourceOptions,
  applyDefaultLiteratureSourceCodes,
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
const createManuscriptForm = reactive<{ disciplineCode: string; title: string; contentLanguage: 'zh' | 'en' }>({
  disciplineCode: '',
  title: '',
  contentLanguage: 'en',
})
let createManuscriptDialogResolve: ((ok: boolean) => void) | null = null

function pickDefaultDisciplineCode(): string {
  return disciplineSelectOptions.value[0]?.value ?? ''
}

async function openCreateManuscriptDialog(): Promise<boolean> {
  await reloadTopicDiscoveryFormOptions()
  createManuscriptForm.disciplineCode = pickDefaultDisciplineCode()
  createManuscriptForm.title = ''
  createManuscriptForm.contentLanguage = 'en'
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
      createManuscriptForm.contentLanguage,
    )
    applyManuscriptListFromApi(data)
    topicForm.disciplineCode = createManuscriptForm.disciplineCode
    closeCreateManuscriptDialog(true)
    ElMessage.success('论文已创建')
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
  if (!ready || envPreferenceFromStorage.value) return
  applyCatalogIntensityAuditDefaults(envPreference)
  applyDefaultLiteratureSourceCodes(envPreference)
  topicForm.intensity = envPreference.intensity
  topicForm.auditLevel = envPreference.auditLevel
  topicForm.sourceCodes = [...envPreference.literatureSourceCodes]
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

function onManuscriptsListUpdate(list: PaperManuscriptItem[]) {
  manuscripts.value = list
  persistManuscripts()
}

loadEnvFromStorage()
loadManuscriptsFromStorage()
loadLitReviewFlagsFromStorage()
loadExperimentPlanFlagsFromStorage()
topicForm.venue = envPreference.defaultVenueText || topicForm.venue
topicForm.sourceCodes = [...envPreference.literatureSourceCodes]
topicForm.intensity = envPreference.intensity
topicForm.auditLevel = envPreference.auditLevel
topicForm.humanCheckpoint = envPreference.humanCheckpoint

const activeManuscripts = computed(() => manuscripts.value.filter((m) => m.status === 'active'))

const currentManuscript = computed(
  () =>
    manuscripts.value.find((m) => m.id === activeManuscriptId.value) ?? activeManuscripts.value[0],
)

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
const isInviteRebateModule = computed(() => activeModule.value === 'invite-rebate')
const isPersonalCenterModule = computed(() => activeModule.value === 'personal-center')
const isUtilityModule = computed(
  () =>
    isMyManuscriptsModule.value ||
    isInviteRebateModule.value ||
    isPersonalCenterModule.value,
)
const isTopicDiscoveryModule = computed(() => activeModule.value === 'topic-discovery')

const isLiteratureReviewModule = computed(() => activeModule.value === 'literature-review')

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
  topicForm.venue = envPreference.defaultVenueText || topicForm.venue
  topicForm.intensity = envPreference.intensity
  topicForm.auditLevel = envPreference.auditLevel
  topicForm.humanCheckpoint = envPreference.humanCheckpoint
  recordModuleOperationLog('environment', '保存环境配置')
}

const SECONDARY_WORKFLOW_MODULES = [
  'literature-review',
  'experiment-planning',
  'paper-writing',
  'manuscript-analysis',
  'figure-generation',
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

const currentTopicRun = computed((): TopicRunDemo => {
  const id = activeManuscriptId.value
  if (!id) return createIdleTopicRun()
  return topicRunsByManuscript.value[id] ?? createIdleTopicRun()
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

/** 人工检查点暂停（无步骤在 running） */
const topicAtHumanPause = computed(() => {
  if (!topicForm.humanCheckpoint) return false
  if (topicShowFailedPanel.value) return false
  if (currentTopicRun.value.status === 'completed') return false
  if (currentTopicRun.value.steps.some((s) => s.status === 'running')) return false
  const raw = topicDiscoveryApiRun.value
  if (raw?.steps.some((s) => s.status === 'running')) return false
  if (raw?.run_status === 'checkpoint') return true
  if (raw?.pause_after_stage?.trim()) return true
  if (topicStepsCompletedThenPending(currentTopicRun.value.steps)) return true
  return false
})

const topicRunBusy = computed(
  () =>
    topicContinueLoading.value ||
    currentTopicRun.value.steps.some((s) => s.status === 'running') ||
    (running.value && !topicAtHumanPause.value) ||
    (currentTopicRun.value.status === 'running' && !topicAtHumanPause.value),
)

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

/** 非失败且在人工暂停点：「终止并返回」+「确认并继续」 */
const topicShowCheckpointPanel = computed(() => {
  if (currentTopicRun.value.steps.some((s) => s.status === 'running')) return false
  if (topicDiscoveryApiRun.value?.steps.some((s) => s.status === 'running')) return false
  return topicAtHumanPause.value
})

function buildTopicCheckpointView(
  pauseStage: string,
  raw: TopicDiscoveryRunResponse | null,
): TopicCheckpointView | null {
  const def = TOPIC_DISCOVERY_FLOW_STEPS.find((s) => s.stageCode === pauseStage)
  if (!def?.checkpointKey) return null
  const dto = raw?.steps.find((s) => s.stage_code === pauseStage)
  const lines = linesFromApiStep(dto)
  return {
    key: def.checkpointKey,
    title: def.label,
    lines: lines.length ? lines : CHECKPOINT_COPY[def.checkpointKey].lines,
  }
}

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

const topicFailedSummary = computed((): { stageLabel: string; stageCode: string; error: string } | null => {
  if (!topicShowFailedPanel.value) return null
  const raw = topicDiscoveryApiRun.value
  const failedDto = raw?.steps.find((s) => s.status === 'failed')
  const stageCode = failedDto?.stage_code ?? 'unknown'
  const def = TOPIC_DISCOVERY_FLOW_STEPS.find((s) => s.stageCode === stageCode)
  const extra = (failedDto?.extra ?? {}) as Record<string, unknown>
  const errRaw = extra.error
  const error =
    typeof errRaw === 'string' && errRaw.trim()
      ? errRaw.trim()
      : '本步执行失败，请查看日志或调整配置后重试。'
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
  if (activeModule.value === 'figure-generation' && figureManagementTab.value === 'upload') return false
  return true
})

const topicPrimaryDisabled = computed(() => {
  if (isUtilityModule.value) return true
  if (isTopicDiscoveryModule.value) {
    return (
      topicRunBusy.value ||
      currentTopicRun.value.status === 'checkpoint' ||
      currentTopicRun.value.status === 'running'
    )
  }
  return running.value
})

const primaryActionLabel = computed(() => {
  if (isTopicDiscoveryModule.value) {
    if (currentTopicRun.value.status === 'checkpoint') return '等待确认'
    if (currentTopicRun.value.status === 'running' || running.value) return '运行中…'
    if (topicReadyForLiteratureReview.value) return '生成文献综述'
    if (topicServerRunCompleted.value) return '再次运行'
    if (currentTopicRun.value.status === 'failed') return '再次运行'
    return '运行'
  }
  if (isLiteratureReviewModule.value) {
    if (running.value) return '运行中…'
    if (litReviewReadyForExperimentPlan.value) return '生成实验计划'
    if (litReviewDoneForManuscript.value && experimentPlanDoneForManuscript.value) return '再次运行'
    return '运行'
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
  generate_ideas: 'AI 脑暴选题与新颖性分析中…',
  audit: 'AI 审查结论与文献综述生成中…',
}

function topicStepRunningHint(stageCode: string) {
  return TOPIC_STEP_RUNNING_HINT[stageCode] ?? 'AI 模型运行中…'
}

const TOPIC_RUN_POLL_MS = 1500

/** 服务端 retrieve / generate_ideas / audit 等 bot 异步步为 running 时需轮询 */
function topicRunHasAsyncWorkOnServer(msId: string): boolean {
  const apiRun = topicLastRunByMs.value[msId]
  if (!apiRun) return false
  return apiRun.steps.some((s) => s.status === 'running')
}

/** 轮询 GET /run/current：PDF 下载（retrieve）、脑暴 LLM（generate_ideas）、审查 LLM（audit）等异步步 */
const topicRunNeedsPoll = computed(() => {
  const msId = activeManuscriptId.value
  if (topicContinueLoading.value) return true
  if (msId && topicRunHasAsyncWorkOnServer(msId)) return true
  if (running.value) return true
  if (!isTopicDiscoveryModule.value || !topicRunVisible.value) return false
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
  const summary = step.summary_text?.trim()
  if (summary) {
    const fromSummary = summary
      .split(/\n+/)
      .map((l) => l.trim())
      .filter(Boolean)
    if (fromSummary.length) return fromSummary.slice(0, 12)
  }
  const result = step.result as Record<string, unknown> | undefined
  if (!result) return summary ? [summary] : []

  if (step.stage_code === 'retrieve') {
    const hits = (result.literature_hits as Array<{ title?: string; external_key?: string }>) ?? []
    const extra = (step.extra ?? {}) as Record<string, unknown>
    const lines: string[] = []
    if (typeof extra.hit_count === 'number') {
      lines.push(`命中 ${extra.hit_count} 篇 · 验真 ${extra.verified_count ?? extra.hit_count} 篇`)
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

  if (step.stage_code === 'generate_ideas') {
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
  const steps: TopicFlowStepRuntime[] = TOPIC_DISCOVERY_FLOW_STEPS.map((def) => {
    const st = data.steps.find((s) => s.stage_code === def.stageCode)
    return {
      stageCode: def.stageCode,
      label: def.label,
      checkpointKey: def.checkpointKey,
      status: st ? stepStatusFromApi(st.status) : 'pending',
    }
  })

  const pauseStage = topicPauseStageFromApi(data)
  const anyStepRunning = steps.some((s) => s.status === 'running')
  if (
    pauseStage &&
    !anyStepRunning &&
    status !== 'failed' &&
    status !== 'completed'
  ) {
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
    venue: topicForm.venue,
    source_codes: [...topicForm.sourceCodes],
    intensity: topicForm.intensity,
    audit_level: topicForm.auditLevel,
    human_checkpoint: topicForm.humanCheckpoint,
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
        applyStoredTopicRunInputToForm(topicForm, data)
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

/** 检查点继续：下一步先标 running，与 retrieve 异步时立刻出现转圈 */
function optimisticMarkNextStepRunning(msId: string) {
  const cur = topicRunsByManuscript.value[msId] ?? createIdleTopicRun()
  const steps = cur.steps.map((s) => ({ ...s }))
  let marked = false
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
  if (!marked) topicContinueOptimisticStage.value = null
  persistTopicRun(msId, {
    status: 'running',
    steps,
    checkpoint: null,
  })
}

function syncRunningFlagFromServer(msId: string) {
  if (topicRunHasAsyncWorkOnServer(msId)) {
    running.value = true
    return
  }
  if (currentTopicRun.value.steps.some((s) => s.status === 'running')) {
    running.value = true
    return
  }
  if (topicAtHumanPause.value) {
    running.value = false
    return
  }
  running.value = false
}

function finalizeTopicDiscoveryRun(_msId: string) {
  recordModuleOperationLog('topic-discovery', '运行工作流（retrieve → ideas → audit）')
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
  const set = new Set(topicForm.sourceCodes)
  if (checked) set.add(code)
  else set.delete(code)
  topicForm.sourceCodes = [...set]
}

function isTopicSourceChecked(code: string) {
  return topicForm.sourceCodes.includes(code)
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
    if (topicForm.sourceCodes.length === 0) {
      ElMessage.warning('请至少选择一个文献来源（检索在选题发现完成）')
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
          mod !== 'my-manuscripts'
        ) {
          recordModuleOperationLog(mod, `运行「${currentMeta.value.label}」`)
        }
        const successHint =
          mod === 'literature-review' && msId && !experimentPlanDoneForManuscript.value
            ? `「${currentMeta.value.label}」已生成 · 可点顶栏「生成实验计划」继续`
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
      <div class="paper-message-box paper-message-box--default paper-modal-panel">
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
              />
            </div>
            <div>
              <span class="paper-label">论文名称</span>
              <input
                v-model="createManuscriptForm.title"
                class="paper-input"
                type="text"
                maxlength="256"
                placeholder="例如：MDD 脑网络拓扑研究"
                @keyup.enter="submitCreateManuscriptDialog"
              />
            </div>
            <div>
              <span class="paper-label">工作语言</span>
              <div class="paper-create-ms-lang" role="radiogroup" aria-label="工作语言">
                <label class="paper-create-ms-lang__option">
                  <input
                    v-model="createManuscriptForm.contentLanguage"
                    type="radio"
                    name="paper-create-ms-lang"
                    value="en"
                  />
                  英文
                </label>
                <label class="paper-create-ms-lang__option">
                  <input
                    v-model="createManuscriptForm.contentLanguage"
                    type="radio"
                    name="paper-create-ms-lang"
                    value="zh"
                  />
                  中文
                </label>
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

      <div class="paper-manuscript-switcher">
        <label class="paper-manuscript-label" for="paper-manuscript-select">当前论文</label>
        <div
          class="paper-manuscript-select-wrap"
          :class="{ 'paper-manuscript-select-wrap--busy': manuscriptSwitching }"
        >
          <select
            id="paper-manuscript-select"
            class="paper-manuscript-select"
            :class="{ 'paper-manuscript-select--empty': !activeManuscripts.length }"
            :value="activeManuscriptId"
            :disabled="!activeManuscripts.length || manuscriptSwitching"
            @change="onManuscriptChange(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="m in activeManuscripts" :key="m.id" :value="m.id">
              {{ m.title }}{{ m.venueHint ? ` · ${m.venueHint}` : '' }}
            </option>
          </select>
          <span
            v-if="manuscriptSwitching"
            class="paper-manuscript-select-loading"
            aria-label="切换中"
          />
        </div>
        <button type="button" class="paper-manuscript-new" @click="onAddManuscriptClick">
          添加论文
        </button>
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
          <p class="paper-main-manuscript">
            当前论文：<strong>{{ currentManuscript?.title ?? '' }}</strong>
            <span v-if="currentManuscript?.venueHint"> · {{ currentManuscript.venueHint }}</span>
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
            <div class="paper-input paper-input--readonly" aria-readonly="true">
              {{ topicDisciplineDisplayLabel }}
            </div>
            <span class="paper-hint">与当前论文创建时选定的学科一致，不可在此修改</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">目标会议/期刊</span>
            <input v-model="topicForm.venue" type="text" class="paper-input" />
            <span class="paper-hint">用于约束贡献类型、实验标准和写作风格。</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">执行强度</span>
            <PaperSelect v-model="topicForm.intensity" :options="intensityOptions" />
            <span class="paper-hint">控制检索数量、迭代轮数和输出深度。</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">审计等级</span>
            <PaperSelect v-model="topicForm.auditLevel" :options="auditOptions" />
            <span class="paper-hint">控制 citation audit、claim audit、kill argument 等门禁强度。</span>
          </label>
        </div>

        <div class="paper-field paper-field--block paper-field--section-gap">
          <span class="paper-label">文献来源（检索用）</span>
          <div class="paper-check-group">
            <label
              v-for="src in literatureSourceOptions"
              :key="src.code"
              class="paper-check paper-check--inline"
            >
              <input
                type="checkbox"
                :checked="isTopicSourceChecked(src.code)"
                @change="toggleTopicSource(src.code, ($event.target as HTMLInputElement).checked)"
              />
              <span>{{ src.label }}</span>
            </label>
          </div>
          <span class="paper-hint">所有结果必须进入真实文献验证流程。</span>
        </div>

        <label class="paper-check">
          <input
            v-model="topicForm.humanCheckpoint"
            type="checkbox"
            :disabled="topicRunBusy || currentTopicRun.status === 'checkpoint'"
          />
          <span>
            <strong>人工检查点</strong>
            <span class="paper-hint paper-hint--inline">
              开启后三步各暂停一次：检索入库 → 脑暴（含新颖性）→ 审计，每步需确认并继续
            </span>
          </span>
        </label>
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
            v-if="topicRunBusy || currentTopicRun.status === 'checkpoint'"
            type="button"
            class="paper-flow-cancel"
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
              <span v-if="step.status === 'running'" class="paper-flow-step-spinner" />
              <template v-else>{{ topicStepIcon(step.status) }}</template>
            </span>
            <div class="paper-flow-step-body">
              <span class="paper-flow-step-label">{{ step.label }}</span>
              <span class="paper-flow-step-code">{{ step.stageCode }}</span>
              <span v-if="step.checkpointKey && topicForm.humanCheckpoint" class="paper-flow-step-tag">
                检查点 · {{ step.checkpointKey }}
              </span>
              <span v-if="step.status === 'running'" class="paper-flow-step-running" role="status">
                {{ topicStepRunningHint(step.stageCode) }}
              </span>
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

        <div v-if="topicShowCheckpointPanel" class="paper-checkpoint">
          <div v-if="topicCheckpointPanel" class="paper-checkpoint-head">
            <span class="paper-checkpoint-pause" aria-hidden="true">⏸</span>
            <div>
              <h3 class="paper-checkpoint-title">人工检查点 · {{ topicCheckpointPanel.key }}</h3>
              <p class="paper-checkpoint-sub">{{ topicCheckpointPanel.title }}</p>
            </div>
          </div>
          <div v-else class="paper-checkpoint-head">
            <span class="paper-checkpoint-pause" aria-hidden="true">⏸</span>
            <h3 class="paper-checkpoint-title">人工检查点</h3>
          </div>
          <ul v-if="topicCheckpointPanel?.lines.length" class="paper-checkpoint-list">
            <li v-for="(line, i) in topicCheckpointPanel.lines" :key="i">{{ line }}</li>
          </ul>
          <p class="paper-checkpoint-note">
            确认后继续下一步；终止会取消本轮进度，刷新后不会再加载。
          </p>
          <div class="paper-checkpoint-actions">
            <button
              type="button"
              class="paper-btn-secondary"
              :disabled="topicTerminateLoading || running"
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
              {{
                topicContinueLoading ||
                currentTopicRun.steps.some((s) => s.status === 'running')
                  ? '执行中…'
                  : '确认并继续'
              }}
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
              <p v-if="topicFailedSummary?.stageCode" class="paper-checkpoint-sub">
                {{ topicFailedSummary.stageCode }}
              </p>
            </div>
          </div>
          <p class="paper-flow-failed-error" role="alert">
            {{ topicFailedSummary?.error ?? '本步执行失败，请查看日志或调整配置后重试。' }}
          </p>
          <p class="paper-checkpoint-note">
            可驳回并返回以取消本轮进度；修正配置后请点顶栏「再次运行」重新开始。
          </p>
          <div class="paper-checkpoint-actions">
            <button
              type="button"
              class="paper-btn-secondary"
              :disabled="topicTerminateLoading || running"
              @click="rejectFailedTopicRun"
            >
              {{ topicTerminateLoading ? '处理中…' : '驳回并返回' }}
            </button>
          </div>
        </div>

        <div v-else-if="currentTopicRun.status === 'completed'" class="paper-flow-done">
          产出已写入当前论文（入库语料、候选 idea、新颖性结论 · 演示）。下一步请点顶栏
          <strong>「生成文献综述」</strong>；实验方案请在「实验方案」模块查看或生成。
        </div>
      </section>
      </template>
      </template>

      <PaperMyManuscriptsPanel
        v-else-if="activeModule === 'my-manuscripts'"
        :key="`my-manuscripts-${moduleContentKey}`"
        :manuscripts="manuscripts"
        :active-manuscript-id="activeManuscriptId"
        @select="onManuscriptChange"
        @update-manuscripts="onManuscriptsListUpdate"
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
        :env-preference="envPreference"
        @environment-saved="onEnvironmentSaved"
      />

      <PaperWorkflowPanels
        v-else-if="isSecondaryWorkflowModule"
        :key="activeModule"
        ref="workflowPanelsRef"
        v-model:figure-tab="figureManagementTab"
        :module-id="activeModule"
        :manuscript-id="activeManuscriptId"
        :manuscript-title="currentManuscript?.title ?? '未命名'"
        :pending-open-experiment-plan-id="pendingOpenExperimentPlanId"
        @navigate-module="onNavigateModule"
        @consumed-pending-experiment-plan="onConsumedPendingExperimentPlan"
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

.paper-manuscript-switcher {
  margin-top: 20px;
  padding: 14px 12px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
}

.paper-manuscript-label {
  display: block;
  margin-bottom: 8px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.65);
}

.paper-manuscript-select-wrap {
  position: relative;
}

.paper-manuscript-select-wrap--busy .paper-manuscript-select {
  opacity: 0.88;
}

.paper-manuscript-select-loading {
  position: absolute;
  top: 50%;
  right: 34px;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(100, 116, 139, 0.35);
  border-top-color: #6366f1;
  border-radius: 50%;
  pointer-events: none;
  animation: paper-manuscript-select-spin 0.65s linear infinite;
}

@keyframes paper-manuscript-select-spin {
  to {
    transform: rotate(360deg);
  }
}

.paper-manuscript-select {
  width: 100%;
  padding: 9px 32px 9px 10px;
  font-size: 13px;
  font-weight: 600;
  color: #1e1b4b;
  background-color: #fff;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='%2364748b' stroke-width='2' stroke-linecap='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  border: none;
  border-radius: 8px;
  appearance: none;
  -webkit-appearance: none;
  cursor: pointer;
}

.paper-manuscript-select:disabled {
  cursor: default;
  opacity: 0.92;
}

.paper-manuscript-select--empty:disabled {
  min-height: 36px;
  color: transparent;
  background-color: rgba(255, 255, 255, 0.92);
}

.paper-manuscript-new {
  width: 100%;
  margin-top: 8px;
  padding: 8px 10px;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  background: transparent;
  border: 1px dashed rgba(255, 255, 255, 0.35);
  border-radius: 8px;
  cursor: pointer;
}

.paper-manuscript-new:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.5);
}

.paper-nav {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
  margin-top: 16px;
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
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--atm-text-muted, #64748b);
}

.paper-main-manuscript strong {
  color: var(--atm-text, #1e1b4b);
  font-weight: 700;
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

.paper-flow-cancel:hover {
  border-color: #cbd5e1;
  color: var(--atm-text, #1e1b4b);
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

.paper-flow-step-code {
  font-size: 11px;
  font-family: ui-monospace, monospace;
  color: #94a3b8;
}

.paper-flow-step-tag {
  font-size: 11px;
  font-weight: 600;
  color: #b45309;
  background: #fffbeb;
  padding: 2px 8px;
  border-radius: 6px;
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

.paper-field--section-gap {
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
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  transition: border-color 0.12s;
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

.paper-input--readonly {
  color: #64748b;
  background: #eef2f6;
  border-color: #e2e8f0;
  cursor: default;
  user-select: none;
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
