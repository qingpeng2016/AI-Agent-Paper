<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  DEFAULT_ENV_PREFERENCE,
  ENV_PREFERENCE_STORAGE_KEY,
  DEFAULT_TOPIC_DISCOVERY,
  DEMO_PAPER_MANUSCRIPTS,
  DISCIPLINE_OPTIONS,
  LITERATURE_SOURCE_OPTIONS,
  PAPER_MODULE_GROUPS,
  TOPIC_USER_LIBRARY_SOURCE,
  USER_LIBRARY_SOURCE_CODE,
  getLiteratureSourceLabel,
  getUserReferenceUploads,
  appendOperationLog,
  TOPIC_DISCOVERY_FLOW_STEPS,
  getPaperModuleMeta,
  type EnvironmentPreferenceForm,
  type PaperManuscriptItem,
  type PaperModuleId,
  type TopicCheckpointKey,
  type TopicDiscoveryArtifactSnapshot,
  type TopicDiscoveryForm,
  type TopicFlowStepStatus,
} from '@paper/types'
import { DEMO_EXPERIMENT_PLAN } from '@paper/demoModuleOutputs'
import { DEMO_MODULE_TOKEN_ESTIMATES } from '@paper/demoOperationLogs'
import PaperModuleNavIcon from '@paper/PaperModuleNavIcon.vue'
import PaperMyManuscriptsPanel from '@paper/PaperMyManuscriptsPanel.vue'
import PaperPersonalCenterPanel from '@paper/PaperPersonalCenterPanel.vue'
import PaperReferenceLibraryPanel from '@paper/PaperReferenceLibraryPanel.vue'
import PaperSelect from '@paper/PaperSelect.vue'
import PaperWorkflowPanels from '@paper/PaperWorkflowPanels.vue'

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
  ideas_ready: {
    title: '候选选题与新颖性结论',
    lines: [
      'Idea A：稀疏注意力 + 动态路由 — 新颖性：与 Static Sparse 系列有明确区分（中等风险）',
      'Idea B：层级 KV 压缩 — 新颖性：需加强与 H2O / SnapKV 对比（偏高风险）',
      'Idea C：训练无关的 token 合并 — 新颖性：检索命中较少（低风险，建议深检索）',
    ],
  },
}

const LIT_REVIEW_RUN_STORAGE_KEY = 'atm:paper:lit-review-run:v1'
const EXPERIMENT_PLAN_RUN_STORAGE_KEY = 'atm:paper:experiment-plan-run:v1'

function buildDemoExperimentPlanLines(): string[] {
  return [
    DEMO_EXPERIMENT_PLAN.hypothesis,
    `基线：${DEMO_EXPERIMENT_PLAN.baselines.join(' · ')}`,
    `指标：${DEMO_EXPERIMENT_PLAN.metrics.join(' · ')}`,
    `消融：${DEMO_EXPERIMENT_PLAN.ablations.join(' · ')}`,
    ...DEMO_EXPERIMENT_PLAN.steps,
  ]
}

const MANUSCRIPTS_STORAGE_KEY = 'atm:paper:manuscripts:v1'
const LEGACY_PROJECTS_STORAGE_KEY = 'atm:paper:projects:v1'

const activeModule = ref<PaperModuleId>('topic-discovery')

function selectModule(id: PaperModuleId) {
  activeModule.value = id
}

const running = ref(false)
const topicRunToken = ref(0)
const topicRunsByManuscript = ref<Record<string, TopicRunDemo>>({})
const litReviewDoneByManuscript = ref<Record<string, boolean>>({})
const experimentPlanDoneByManuscript = ref<Record<string, boolean>>({})
const workflowPanelsRef = ref<InstanceType<typeof PaperWorkflowPanels> | null>(null)
const figureManagementTab = ref<'upload' | 'generate'>('upload')
const paperMainRef = ref<HTMLElement | null>(null)
const topicFlowPanelRef = ref<HTMLElement | null>(null)

const topicForm = reactive<TopicDiscoveryForm>({ ...DEFAULT_TOPIC_DISCOVERY })
const envPreference = reactive<EnvironmentPreferenceForm>({ ...DEFAULT_ENV_PREFERENCE })

const manuscripts = ref<PaperManuscriptItem[]>([...DEMO_PAPER_MANUSCRIPTS])
const activeManuscriptId = ref<string>(DEMO_PAPER_MANUSCRIPTS[0]?.id ?? '')

function loadEnvFromStorage() {
  try {
    const raw = localStorage.getItem(ENV_PREFERENCE_STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as { preference?: EnvironmentPreferenceForm }
    if (data.preference) Object.assign(envPreference, data.preference)
  } catch {
    /* ignore */
  }
}

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
    if (list?.length) manuscripts.value = list
    if (activeId && manuscripts.value.some((m) => m.id === activeId)) {
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

function onManuscriptChange(id: string) {
  activeManuscriptId.value = id
  persistManuscripts()
}

function onManuscriptsListUpdate(list: PaperManuscriptItem[]) {
  manuscripts.value = list
  persistManuscripts()
}

loadEnvFromStorage()
loadManuscriptsFromStorage()
loadLitReviewFlagsFromStorage()
loadExperimentPlanFlagsFromStorage()
topicForm.disciplineCode = envPreference.disciplineCode
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

const isReferenceLibraryModule = computed(() => activeModule.value === 'reference-library')
const isMyManuscriptsModule = computed(() => activeModule.value === 'my-manuscripts')
const isUtilityModule = computed(
  () => isReferenceLibraryModule.value || isMyManuscriptsModule.value,
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

function onPersonalCenterEnvironmentSaved() {
  topicForm.disciplineCode = envPreference.disciplineCode
  topicForm.venue = envPreference.defaultVenueText || topicForm.venue
  topicForm.intensity = envPreference.intensity
  topicForm.auditLevel = envPreference.auditLevel
  topicForm.humanCheckpoint = envPreference.humanCheckpoint
  recordModuleOperationLog('environment', '保存默认配置')
}

const SECONDARY_WORKFLOW_MODULES = [
  'literature-review',
  'experiment-planning',
  'auto-review',
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

watch(
  activeManuscriptId,
  (id) => {
    if (id) ensureTopicRunForManuscript(id)
  },
  { immediate: true },
)

const currentTopicRun = computed((): TopicRunDemo => {
  const id = activeManuscriptId.value
  if (!id) return createIdleTopicRun()
  return topicRunsByManuscript.value[id] ?? createIdleTopicRun()
})

const topicRunVisible = computed(
  () => isTopicDiscoveryModule.value && currentTopicRun.value.status !== 'idle',
)

const topicRunBusy = computed(
  () => currentTopicRun.value.status === 'running' || running.value,
)

/** 文献综述只读：当前论文最近一次选题 run 的 artifact 快照 */
const topicDiscoveryArtifact = computed((): TopicDiscoveryArtifactSnapshot => {
  const run = currentTopicRun.value
  const retrieve = run.steps.find((s) => s.stageCode === 'retrieve')
  const novelty = run.steps.find((s) => s.stageCode === 'novelty')
  const corpusReady = retrieve?.status === 'completed'
  const runCompleted = run.status === 'completed'
  const sourceLabels = topicForm.sourceCodes.map((c) => getLiteratureSourceLabel(c))
  const direction =
    topicForm.direction.trim() ||
    (corpusReady ? '（本次 run 未保留方向文案 · 演示）' : '（尚未填写研究方向）')

  const noveltyReady =
    novelty?.status === 'completed' ||
    run.status === 'checkpoint' ||
    runCompleted
  const disciplineLabel =
    DISCIPLINE_OPTIONS.find((d) => d.code === topicForm.disciplineCode)?.label ?? topicForm.disciplineCode

  return {
    runStatus: run.status,
    corpusReady,
    runCompleted,
    disciplineLabel,
    direction,
    venue: topicForm.venue,
    sourceLabels,
    literatureHitCount: corpusReady ? 86 : 0,
    verifiedHitCount: corpusReady ? 79 : 0,
    candidateIdeas: noveltyReady ? [...CHECKPOINT_COPY.ideas_ready.lines] : [],
    noveltyLines: noveltyReady ? [...CHECKPOINT_COPY.ideas_ready.lines] : [],
    experimentPlanLines: experimentPlanDoneForManuscript.value ? buildDemoExperimentPlanLines() : [],
  }
})

const litReviewDoneForManuscript = computed(
  () => !!litReviewDoneByManuscript.value[activeManuscriptId.value],
)

const experimentPlanDoneForManuscript = computed(
  () => !!experimentPlanDoneByManuscript.value[activeManuscriptId.value],
)

const topicReadyForLiteratureReview = computed(
  () => currentTopicRun.value.status === 'completed' && !litReviewDoneForManuscript.value,
)

const litReviewReadyForExperimentPlan = computed(
  () =>
    isLiteratureReviewModule.value &&
    litReviewDoneForManuscript.value &&
    !experimentPlanDoneForManuscript.value,
)

const showPrimaryAction = computed(() => {
  if (isUtilityModule.value) return false
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
    if (currentTopicRun.value.status === 'completed') return '再次运行'
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

function delay(ms: number, token: number) {
  return new Promise<void>((resolve, reject) => {
    window.setTimeout(() => {
      if (token !== topicRunToken.value) reject(new Error('aborted'))
      else resolve()
    }, ms)
  })
}

function checkpointView(key: TopicCheckpointKey): TopicCheckpointView {
  const copy = CHECKPOINT_COPY[key]
  return { key, title: copy.title, lines: [...copy.lines] }
}

async function executeTopicFlow(fromIndex: number, token: number) {
  const msId = activeManuscriptId.value
  if (!msId) return

  const humanOn = topicForm.humanCheckpoint
  let run = { ...currentTopicRun.value, steps: currentTopicRun.value.steps.map((s) => ({ ...s })) }

  for (let i = fromIndex; i < run.steps.length; i++) {
    if (token !== topicRunToken.value) return

    const step = run.steps[i]
    step.status = 'running'
    run.status = 'running'
    run.checkpoint = null
    persistTopicRun(msId, run)

    const ms = step.stageCode === 'retrieve' ? 1400 : step.stageCode === 'audit' ? 900 : 1100
    try {
      await delay(ms, token)
    } catch {
      return
    }

    step.status = 'completed'
    persistTopicRun(msId, { ...run })

    const cpKey = step.checkpointKey
    if (humanOn && cpKey) {
      run.status = 'checkpoint'
      run.checkpoint = checkpointView(cpKey)
      persistTopicRun(msId, { ...run })
      ElMessage.info('流程已暂停：请确认检查点内容后再继续')
      return
    }
  }

  run.status = 'completed'
  run.checkpoint = null
  persistTopicRun(msId, run)
  recordModuleOperationLog('topic-discovery', '运行工作流（retrieve → ideas → novelty → audit）')
  ElMessage.success('选题发现已完成：可点顶栏「生成文献综述」继续')
}

function resetTopicRunForAction(msId: string) {
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
  if (!panel) return
  if (scroller) {
    const top =
      panel.getBoundingClientRect().top -
      scroller.getBoundingClientRect().top +
      scroller.scrollTop -
      12
    scroller.scrollTo({ top: Math.max(0, top), behavior: 'smooth' })
    return
  }
  panel.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function markLiteratureReviewDone(msId: string) {
  litReviewDoneByManuscript.value = { ...litReviewDoneByManuscript.value, [msId]: true }
  persistLitReviewFlags()
}

async function runExperimentPlanFromLiteratureReview() {
  const msId = activeManuscriptId.value
  if (!msId) return
  if (!litReviewDoneForManuscript.value) {
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
        `「实验规划」已生成（演示）· 论文「${currentManuscript.value?.title ?? '未命名'}」`,
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
  if (currentTopicRun.value.status !== 'completed') {
    ElMessage.warning('请先完成选题发现（至新颖性检查）')
    return
  }

  selectModule('literature-review')
  running.value = true
  try {
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
  const msId = activeManuscriptId.value
  if (!msId) return

  resetTopicRunForAction(msId)
  const token = topicRunToken.value
  running.value = true

  let run = createIdleTopicRun()
  run.status = 'running'
  persistTopicRun(msId, run)
  await scrollToTopicFlowPanel()

  try {
    await executeTopicFlow(0, token)
  } finally {
    if (token === topicRunToken.value) running.value = false
  }
}

function continueTopicAfterCheckpoint() {
  const msId = activeManuscriptId.value
  if (!msId || currentTopicRun.value.status !== 'checkpoint') return

  const nextIndex = currentTopicRun.value.steps.findIndex((s) => s.status === 'pending')
  if (nextIndex < 0) return

  const token = topicRunToken.value
  running.value = true
  persistTopicRun(msId, {
    ...currentTopicRun.value,
    status: 'running',
    checkpoint: null,
    steps: currentTopicRun.value.steps.map((s) => ({ ...s })),
  })

  executeTopicFlow(nextIndex, token).finally(() => {
    if (token === topicRunToken.value) running.value = false
  })
}

function rejectTopicCheckpoint() {
  const msId = activeManuscriptId.value
  if (!msId) return
  topicRunToken.value += 1
  const run = {
    ...currentTopicRun.value,
    status: 'failed' as const,
    checkpoint: null,
    steps: currentTopicRun.value.steps.map((s) =>
      s.status === 'running' ? { ...s, status: 'failed' as TopicFlowStepStatus } : { ...s },
    ),
  }
  persistTopicRun(msId, run)
  running.value = false
  ElMessage.warning('已驳回：可修改参数后再次运行')
}

function cancelTopicRun() {
  const msId = activeManuscriptId.value
  if (!msId) return
  resetTopicRunForAction(msId)
  running.value = false
  ElMessage.info('已取消运行')
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

const intensityOptions = [
  { value: 'fast', label: '更快' },
  { value: 'balanced', label: 'Balanced（平衡）' },
  { value: 'deep', label: '更深' },
]

const auditOptions = [
  { value: 'standard', label: 'Standard' },
  { value: 'polished', label: 'Polished（精修）' },
  { value: 'strict', label: 'Strict' },
]

const disciplineSelectOptions = DISCIPLINE_OPTIONS.map((d) => ({
  value: d.code,
  label: d.label,
}))

function toggleTopicSource(code: string, checked: boolean) {
  const set = new Set(topicForm.sourceCodes)
  if (checked) set.add(code)
  else set.delete(code)
  topicForm.sourceCodes = [...set]
}

function isTopicSourceChecked(code: string) {
  return topicForm.sourceCodes.includes(code)
}

const userReferenceUploadCount = computed(() => {
  void activeModule.value
  return getUserReferenceUploads(activeManuscriptId.value).length
})

function goToUserReferenceLibrary() {
  selectModule('reference-library')
}

async function onPrimaryAction() {
  if (activeModule.value === 'topic-discovery') {
    if (!topicForm.direction.trim()) {
      ElMessage.warning('请填写研究方向')
      return
    }
    if (topicForm.sourceCodes.length === 0) {
      ElMessage.warning('请至少选择一个文献来源（检索在选题发现完成）')
      return
    }
    if (
      topicForm.sourceCodes.includes(USER_LIBRARY_SOURCE_CODE) &&
      userReferenceUploadCount.value === 0
    ) {
      ElMessage.warning('已勾选「上传文献」，请先在「上传文献」上传至少一篇')
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
        if (mod !== 'personal-center' && mod !== 'reference-library' && mod !== 'my-manuscripts') {
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
</script>

<template>
  <div class="paper-workbench">
    <aside class="paper-sidebar">
      <div class="paper-sidebar-brand">
        <span class="paper-sidebar-logo" aria-hidden="true">◆</span>
        <div>
          <div class="paper-sidebar-title">Research Desktop</div>
          <div class="paper-sidebar-sub">Research workflows</div>
        </div>
      </div>

      <div class="paper-manuscript-switcher">
        <label class="paper-manuscript-label" for="paper-manuscript-select">当前论文</label>
        <select
          id="paper-manuscript-select"
          class="paper-manuscript-select"
          :value="activeManuscriptId"
          @change="onManuscriptChange(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="m in activeManuscripts" :key="m.id" :value="m.id">
            {{ m.title }}{{ m.venueHint ? ` · ${m.venueHint}` : '' }}
          </option>
        </select>
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
          <p v-if="currentManuscript" class="paper-main-manuscript">
            当前论文：<strong>{{ currentManuscript.title }}</strong>
            <span v-if="currentManuscript.venueHint"> · {{ currentManuscript.venueHint }}</span>
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
      <!-- 选题发现（与运行进度同属一块，避免 v-else-if 链误绑） -->
      <template v-if="activeModule === 'topic-discovery'">
      <section class="paper-panel">
        <label class="paper-field paper-field--block">
          <span class="paper-label">研究方向 / 检索主题 <em class="req">*</em></span>
          <textarea
            v-model="topicForm.direction"
            class="paper-textarea"
            rows="4"
            placeholder="输入要探索的研究主题、问题或关键词（也作为检索 query 依据）。"
          />
        </label>

        <div class="paper-field-grid">
          <label class="paper-field">
            <span class="paper-label">学科</span>
            <PaperSelect v-model="topicForm.disciplineCode" :options="disciplineSelectOptions" />
            <span class="paper-hint">与个人中心「默认配置」同一套学科；影响默认文献源与 venue</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">目标会议/期刊</span>
            <input v-model="topicForm.venue" type="text" class="paper-input" />
            <span class="paper-hint">贡献类型、实验门槛、写作调性</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">执行强度</span>
            <PaperSelect v-model="topicForm.intensity" :options="intensityOptions" />
            <span class="paper-hint">检索条数、脑暴轮数、计划深度</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">审计等级</span>
            <PaperSelect v-model="topicForm.auditLevel" :options="auditOptions" />
            <span class="paper-hint">检索校验、选题断言、kill argument</span>
          </label>
        </div>

        <div class="paper-field paper-field--block paper-field--section-gap">
          <span class="paper-label">文献来源（检索用）</span>
          <div class="paper-check-group">
            <label
              v-for="src in LITERATURE_SOURCE_OPTIONS"
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
          <div class="paper-user-lib-source">
            <label class="paper-check paper-check--inline">
              <input
                type="checkbox"
                :checked="isTopicSourceChecked(TOPIC_USER_LIBRARY_SOURCE.code)"
                @change="
                  toggleTopicSource(
                    TOPIC_USER_LIBRARY_SOURCE.code,
                    ($event.target as HTMLInputElement).checked,
                  )
                "
              />
              <span>{{ TOPIC_USER_LIBRARY_SOURCE.label }}</span>
            </label>
            <p class="paper-hint paper-hint--block">
              使用本篇「上传文献」中已上传 PDF / BibTeX 等，供模型阅读与归纳（不走 arXiv 等 API）。
              <template v-if="isTopicSourceChecked(TOPIC_USER_LIBRARY_SOURCE.code)">
                当前已上传 <strong>{{ userReferenceUploadCount }}</strong> 篇。
              </template>
              <button type="button" class="paper-inline-link" @click="goToUserReferenceLibrary">
                去管理
              </button>
            </p>
          </div>
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
              开启后会在 <em>新颖性检查</em> 完成后暂停，确认后再执行选题 audit
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
          <span class="paper-flow-badge" :class="`paper-flow-badge--${currentTopicRun.status}`">
            {{
              currentTopicRun.status === 'checkpoint'
                ? '等待人工确认'
                : currentTopicRun.status === 'running'
                  ? '执行中'
                  : currentTopicRun.status === 'completed'
                    ? '已完成'
                    : currentTopicRun.status === 'failed'
                      ? '已中止'
                      : ''
            }}
          </span>
          <button
            v-if="currentTopicRun.status === 'running' || currentTopicRun.status === 'checkpoint'"
            type="button"
            class="paper-flow-cancel"
            @click="cancelTopicRun"
          >
            取消
          </button>
        </div>
        <p class="paper-section-lead">
          阶段进度由各 <code>paper_output_*</code> 与 <code>paper_manuscript.draft_*</code> 是否存在推导（演示交互，非真实 Agent）。
        </p>

        <ol class="paper-flow-steps">
          <li
            v-for="step in currentTopicRun.steps"
            :key="step.stageCode"
            class="paper-flow-step"
            :class="`paper-flow-step--${step.status}`"
          >
            <span class="paper-flow-step-icon" aria-hidden="true">{{ topicStepIcon(step.status) }}</span>
            <div class="paper-flow-step-body">
              <span class="paper-flow-step-label">{{ step.label }}</span>
              <span class="paper-flow-step-code">{{ step.stageCode }}</span>
              <span v-if="step.checkpointKey && topicForm.humanCheckpoint" class="paper-flow-step-tag">
                检查点 · {{ step.checkpointKey }}
              </span>
            </div>
          </li>
        </ol>

        <div v-if="currentTopicRun.status === 'checkpoint' && currentTopicRun.checkpoint" class="paper-checkpoint">
          <div class="paper-checkpoint-head">
            <span class="paper-checkpoint-pause" aria-hidden="true">⏸</span>
            <div>
              <h3 class="paper-checkpoint-title">人工检查点 · {{ currentTopicRun.checkpoint.key }}</h3>
              <p class="paper-checkpoint-sub">{{ currentTopicRun.checkpoint.title }}</p>
            </div>
          </div>
          <ul class="paper-checkpoint-list">
            <li v-for="(line, i) in currentTopicRun.checkpoint.lines" :key="i">{{ line }}</li>
          </ul>
          <p class="paper-checkpoint-note">
            确认前后续阶段不会继续；驳回后本次 run 标记为失败，可改参数后点「再次运行」。
          </p>
          <div class="paper-checkpoint-actions">
            <button type="button" class="paper-btn-primary" @click="continueTopicAfterCheckpoint">
              确认并继续
            </button>
            <button type="button" class="paper-btn-secondary" @click="rejectTopicCheckpoint">
              驳回并重跑
            </button>
          </div>
        </div>

        <div v-else-if="currentTopicRun.status === 'completed'" class="paper-flow-done">
          产出已写入当前论文（入库语料、候选 idea、新颖性结论 · 演示）。下一步请点顶栏
          <strong>「生成文献综述」</strong>；实验方案请在「实验规划」模块单独制定。
        </div>
      </section>
      </template>

      <PaperMyManuscriptsPanel
        v-else-if="activeModule === 'my-manuscripts'"
        :manuscripts="manuscripts"
        :active-manuscript-id="activeManuscriptId"
        @select="onManuscriptChange"
        @update-manuscripts="onManuscriptsListUpdate"
      />

      <PaperReferenceLibraryPanel
        v-else-if="activeModule === 'reference-library'"
        :manuscript-id="activeManuscriptId"
        :manuscript-title="currentManuscript?.title ?? '未命名'"
      />

      <PaperPersonalCenterPanel
        v-else-if="activeModule === 'personal-center'"
        ref="personalCenterPanelRef"
        :manuscript-id="activeManuscriptId"
        :manuscript-title="currentManuscript?.title ?? '未命名'"
        :env-preference="envPreference"
        @environment-saved="onPersonalCenterEnvironmentSaved"
      />

      <PaperWorkflowPanels
        v-else-if="isSecondaryWorkflowModule"
        :key="activeModule"
        ref="workflowPanelsRef"
        v-model:figure-tab="figureManagementTab"
        :module-id="activeModule"
        :manuscript-id="activeManuscriptId"
        :manuscript-title="currentManuscript?.title ?? '未命名'"
        :topic-artifact="topicDiscoveryArtifact"
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
  color: var(--atm-text, #1e1b4b);
  background: #fff;
  border: 1px solid #e2e8f0;
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

.paper-user-lib-source {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed #e2e8f0;
}

.paper-inline-link {
  margin-left: 6px;
  padding: 0;
  font-size: inherit;
  font-weight: 600;
  color: #6366f1;
  cursor: pointer;
  background: none;
  border: none;
  text-decoration: underline;
}

.paper-inline-link:hover {
  color: #4f46e5;
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
