import type {
  IdeaDiscoveryUIGroupDTO,
  TopicDiscoveryRunResponse,
  TopicDiscoveryStepDTO,
} from './src/api/topicDiscovery'
import {
  ARIS_IDEA_DISCOVERY_PHASES,
  IDEA_DISCOVERY_PHASE_STATUS,
  type IdeaDiscoveryPhaseStatus,
} from './types'

export type ArisPhaseUiStatus = 'pending' | 'running' | 'waiting_confirm' | 'awaiting_continue' | 'completed'

export type ArisPhaseUiRow = {
  /** UI 段 index（与 idea_discovery_ui_groups[].index 一致） */
  id: string
  label: string
  status: ArisPhaseUiStatus
}

/** 与 conf 默认一致；API 未带 idea_discovery_ui_groups 时的兜底 */
const DEFAULT_IDEA_DISCOVERY_UI_GROUPS: IdeaDiscoveryUIGroupDTO[] = [
  { index: 0, phase_ids: ['research-lit', 'idea-creator'], label_zh: '文献调研、选题大纲' },
  { index: 1, phase_ids: ['novelty-check'], label_zh: '新颖性检验' },
  { index: 2, phase_ids: ['research-review'], label_zh: '外部审查' },
  { index: 3, phase_ids: ['research-refine-pipeline'], label_zh: '文献综述与实验规划' },
]

function normalizePhaseStatusValue(v: unknown): IdeaDiscoveryPhaseStatus {
  if (typeof v === 'string') {
    const s = v.trim().toLowerCase()
    if (s === IDEA_DISCOVERY_PHASE_STATUS.not_started || s === 'pending') {
      return IDEA_DISCOVERY_PHASE_STATUS.not_started
    }
    if (s === IDEA_DISCOVERY_PHASE_STATUS.awaiting_confirm || s === 'waiting_confirm') {
      return IDEA_DISCOVERY_PHASE_STATUS.awaiting_confirm
    }
    if (
      s === IDEA_DISCOVERY_PHASE_STATUS.executing ||
      s === 'running' ||
      s === 'confirmed' ||
      s === 'true' ||
      s === '1'
    ) {
      return IDEA_DISCOVERY_PHASE_STATUS.executing
    }
    if (s === IDEA_DISCOVERY_PHASE_STATUS.completed || s === 'done') {
      return IDEA_DISCOVERY_PHASE_STATUS.completed
    }
    return IDEA_DISCOVERY_PHASE_STATUS.not_started
  }
  if (v === true || v === 1) return IDEA_DISCOVERY_PHASE_STATUS.executing
  return IDEA_DISCOVERY_PHASE_STATUS.not_started
}

/** runtime_state.stage（序号 1..n）或旧版扁平 phase key */
function parsePhaseStatus(result: Record<string, unknown>, phase: string): IdeaDiscoveryPhaseStatus {
  const p = phase.trim()
  if (!p) return IDEA_DISCOVERY_PHASE_STATUS.not_started
  const stage = result.stage
  if (stage && typeof stage === 'object' && !Array.isArray(stage)) {
    const slots = Object.keys(stage as object).sort((a, b) => Number(a) - Number(b))
    for (const slot of slots) {
      const entry = (stage as Record<string, unknown>)[slot]
      if (!entry || typeof entry !== 'object') continue
      const o = entry as Record<string, unknown>
        if (String(o.name ?? '').trim() === p) {
        return normalizePhaseStatusValue(o.status)
      }
    }
  }
  return normalizePhaseStatusValue(result[p])
}

function ideaDiscoveryAsyncLLMQueued(result: Record<string, unknown>): boolean {
  const other = result.other
  if (other && typeof other === 'object' && !Array.isArray(other)) {
    if ((other as Record<string, unknown>).async_llm === true) return true
  }
  return result.async_llm === true
}

/** 全 pipeline 中第一个 awaiting_confirm（仅它展示「待确认」） */
export function firstIdeaDiscoveryPhaseAwaitingConfirm(
  result: Record<string, unknown>,
): string | null {
  for (const id of ARIS_IDEA_DISCOVERY_PHASES) {
    if (parsePhaseStatus(result, id) === IDEA_DISCOVERY_PHASE_STATUS.awaiting_confirm) {
      return id
    }
  }
  return null
}

function parseUiGroupsFromStepResult(result: Record<string, unknown> | undefined): IdeaDiscoveryUIGroupDTO[] {
  const raw = result?.aris_ui_groups
  if (!Array.isArray(raw) || raw.length === 0) return []
  const out: IdeaDiscoveryUIGroupDTO[] = []
  for (const item of raw) {
    if (!item || typeof item !== 'object') continue
    const o = item as Record<string, unknown>
    const index = typeof o.index === 'number' ? o.index : Number(o.index)
    const phaseIds = Array.isArray(o.phase_ids)
      ? o.phase_ids.map((x) => String(x).trim()).filter(Boolean)
      : []
    const labelZh = String(o.label_zh ?? '').trim()
    if (!Number.isFinite(index) || !labelZh) continue
    out.push({ index, phase_ids: phaseIds, label_zh: labelZh })
  }
  out.sort((a, b) => a.index - b.index)
  return out
}

/** 优先 run.idea_discovery_ui_groups，其次 step.result.aris_ui_groups，最后默认分段 */
export function resolveIdeaDiscoveryUIGroups(
  run: TopicDiscoveryRunResponse | null | undefined,
  steps: TopicDiscoveryStepDTO[] | undefined,
): IdeaDiscoveryUIGroupDTO[] {
  const fromRun = run?.idea_discovery_ui_groups
  if (fromRun?.length) return fromRun
  const step = ideaDiscoveryStepDto(steps)
  const fromDb = parseUiGroupsFromStepResult((step?.result ?? {}) as Record<string, unknown>)
  if (fromDb.length) return fromDb
  return DEFAULT_IDEA_DISCOVERY_UI_GROUPS
}

export function ideaDiscoveryUiGroupsLabelChain(groups: IdeaDiscoveryUIGroupDTO[]): string {
  return groups.map((g) => g.label_zh).join(' → ')
}

function ideaDiscoveryStepDto(
  steps: TopicDiscoveryStepDTO[] | undefined,
): TopicDiscoveryStepDTO | undefined {
  return steps?.find(
    (s) => s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery',
  )
}

function uiGroupPhaseStatuses(
  group: IdeaDiscoveryUIGroupDTO,
  result: Record<string, unknown>,
): IdeaDiscoveryPhaseStatus[] {
  return group.phase_ids.map((id) => parsePhaseStatus(result, id))
}

function uiGroupToRowStatus(
  group: IdeaDiscoveryUIGroupDTO,
  result: Record<string, unknown>,
  firstAwaiting: string | null,
): ArisPhaseUiStatus {
  const statuses = uiGroupPhaseStatuses(group, result)
  if (statuses.length === 0) return 'pending'
  if (statuses.every((s) => s === IDEA_DISCOVERY_PHASE_STATUS.completed)) {
    return 'completed'
  }
  if (statuses.some((s) => s === IDEA_DISCOVERY_PHASE_STATUS.executing)) {
    return 'running'
  }
  if (
    firstAwaiting &&
    group.phase_ids.includes(firstAwaiting) &&
    statuses.some((s) => s === IDEA_DISCOVERY_PHASE_STATUS.awaiting_confirm)
  ) {
    return 'waiting_confirm'
  }
  if (statuses.some((s) => s === IDEA_DISCOVERY_PHASE_STATUS.awaiting_confirm)) {
    return 'pending'
  }
  if (statuses.every((s) => s === IDEA_DISCOVERY_PHASE_STATUS.not_started)) {
    return 'pending'
  }
  if (statuses.some((s) => s === IDEA_DISCOVERY_PHASE_STATUS.completed)) {
    return 'running'
  }
  return 'pending'
}

/** 是否有 ARIS 段处于 executing（用户已确认、bot/CLI 仍在跑） */
export function isAnyIdeaDiscoveryPhaseExecuting(
  steps: TopicDiscoveryStepDTO[] | undefined,
): boolean {
  const step = ideaDiscoveryStepDto(steps)
  if (!step) return false
  const result = (step.result ?? {}) as Record<string, unknown>
  for (const id of ARIS_IDEA_DISCOVERY_PHASES) {
    if (parsePhaseStatus(result, id) === IDEA_DISCOVERY_PHASE_STATUS.executing) {
      return true
    }
  }
  return false
}

/** 段间人工检查点：存在 awaiting_confirm，且没有 executing / bot 仍在排队 */
export function isIdeaDiscoveryArisCheckpoint(
  steps: TopicDiscoveryStepDTO[] | undefined,
  runStatus?: string,
): boolean {
  void runStatus
  const step = ideaDiscoveryStepDto(steps)
  if (!step) return false
  const result = (step.result ?? {}) as Record<string, unknown>
  if (ideaDiscoveryAsyncLLMQueued(result)) return false
  if (isAnyIdeaDiscoveryPhaseExecuting(steps)) return false
  return firstIdeaDiscoveryPhaseAwaitingConfirm(result) != null
}

/** 当前待用户确认的 UI 段中文名（无则 null） */
export function arisPhaseAwaitingConfirmLabel(
  steps: TopicDiscoveryStepDTO[] | undefined,
  runStatus?: string,
  run?: TopicDiscoveryRunResponse | null,
): string | null {
  const waiting = arisIdeaDiscoveryPhaseRows(steps, runStatus, run).find(
    (r) => r.status === 'waiting_confirm',
  )
  return waiting?.label ?? null
}

/** 由 runtime_state 各 phase 四态 + UI 段推导展示行 */
export function arisIdeaDiscoveryPhaseRows(
  steps: TopicDiscoveryStepDTO[] | undefined,
  runStatus?: string,
  run?: TopicDiscoveryRunResponse | null,
): ArisPhaseUiRow[] {
  const uiGroups = resolveIdeaDiscoveryUIGroups(run, steps)
  const step = ideaDiscoveryStepDto(steps)

  if (!uiGroups.length) return []

  if (!step || step.status === 'pending') {
    return uiGroups.map((g) => ({
      id: String(g.index),
      label: g.label_zh,
      status: 'pending' as const,
    }))
  }
  if (step.status === 'completed') {
    return uiGroups.map((g) => ({
      id: String(g.index),
      label: g.label_zh,
      status: 'completed' as const,
    }))
  }

  const result = (step.result ?? {}) as Record<string, unknown>
  const firstAwaiting = firstIdeaDiscoveryPhaseAwaitingConfirm(result)

  return uiGroups.map((g) => ({
    id: String(g.index),
    label: g.label_zh,
    status: uiGroupToRowStatus(g, result, firstAwaiting),
  }))
}

export function arisPhaseStatusLabel(status: ArisPhaseUiStatus): string {
  switch (status) {
    case 'completed':
      return '已完成'
    case 'running':
      return '运行中'
    case 'waiting_confirm':
      return '待确认'
    case 'awaiting_continue':
      return '待继续'
    default:
      return '未开始'
  }
}

/** @deprecated 使用 row.label */
export function arisPhaseIdLabel(labelOrLegacyId: string): string {
  return labelOrLegacyId
}
