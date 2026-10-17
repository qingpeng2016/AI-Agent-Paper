import type {
  IdeaDiscoveryUIGroupDTO,
  TopicDiscoveryRunResponse,
  TopicDiscoveryStepDTO,
} from './src/api/topicDiscovery'

export type ArisPhaseUiStatus = 'pending' | 'running' | 'waiting_confirm' | 'completed'

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

function uiGroupIndexForPhase(groups: IdeaDiscoveryUIGroupDTO[], phase: string): number {
  const p = phase.trim()
  if (!p) return 0
  for (const g of groups) {
    if (g.phase_ids.includes(p)) return g.index
  }
  return 0
}

function ideaDiscoveryStepDto(
  steps: TopicDiscoveryStepDTO[] | undefined,
): TopicDiscoveryStepDTO | undefined {
  return steps?.find(
    (s) => s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery',
  )
}

/** idea_discovery 步仍为 running，但 ARIS 段间检查点等人确认（与 bot 是否在跑无关） */
export function isIdeaDiscoveryArisCheckpoint(
  steps: TopicDiscoveryStepDTO[] | undefined,
  runStatus?: string,
): boolean {
  if (runStatus === 'checkpoint') return true
  const step = ideaDiscoveryStepDto(steps)
  if (!step) return false
  const result = (step.result ?? {}) as Record<string, unknown>
  return result.aris_awaiting_checkpoint === true
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

/** 由 run_state 进度 + conf 合并 UI 段推导展示行 */
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
  const awaiting =
    result.aris_awaiting_checkpoint === true ||
    (runStatus === 'checkpoint' && String(result.aris_pause_after_phase ?? '').trim() !== '')
  const pausePhase = String(result.aris_pause_after_phase ?? '').trim()
  const pendingPhase = String(result.aris_pending_phase ?? '').trim()
  const nextPhase = awaiting ? pausePhase : pendingPhase || pausePhase

  let uiNextIdx = uiGroupIndexForPhase(uiGroups, nextPhase)
  if (!nextPhase && step.status === 'running') {
    const fromMeta = result.aris_ui_group_index
    if (typeof fromMeta === 'number' && Number.isFinite(fromMeta)) {
      uiNextIdx = fromMeta
    } else {
      const completed = typeof result.aris_ui_groups_completed === 'number' ? result.aris_ui_groups_completed : 0
      uiNextIdx = Math.min(uiGroups.length - 1, Math.max(0, completed))
    }
  }

  return uiGroups.map((g) => {
    const i = g.index
    if (i < uiNextIdx) {
      return { id: String(i), label: g.label_zh, status: 'completed' as const }
    }
    if (i === uiNextIdx) {
      if (awaiting) {
        return { id: String(i), label: g.label_zh, status: 'waiting_confirm' as const }
      }
      if (step.status === 'running') {
        return { id: String(i), label: g.label_zh, status: 'running' as const }
      }
      return { id: String(i), label: g.label_zh, status: 'pending' as const }
    }
    return { id: String(i), label: g.label_zh, status: 'pending' as const }
  })
}

export function arisPhaseStatusLabel(status: ArisPhaseUiStatus): string {
  switch (status) {
    case 'completed':
      return '已完成'
    case 'running':
      return '运行中'
    case 'waiting_confirm':
      return '待确认'
    default:
      return '未开始'
  }
}

/** @deprecated 使用 row.label */
export function arisPhaseIdLabel(labelOrLegacyId: string): string {
  return labelOrLegacyId
}
