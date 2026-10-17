import {
  ARIS_IDEA_DISCOVERY_PHASES,
  ARIS_IDEA_DISCOVERY_UI_PHASES,
  type ArisIdeaDiscoveryPhaseId,
  type ArisIdeaDiscoveryUiPhaseId,
} from './types'
import type { TopicDiscoveryStepDTO } from './src/api/topicDiscovery'

export type ArisPhaseUiStatus = 'pending' | 'running' | 'waiting_confirm' | 'completed'

export type ArisPhaseUiRow = {
  id: ArisIdeaDiscoveryUiPhaseId
  status: ArisPhaseUiStatus
}

/** run_state 子阶段中文名（与后端 / Lark 内部 id 一致） */
export const ARIS_PHASE_LABEL_ZH: Record<ArisIdeaDiscoveryPhaseId, string> = {
  'research-lit': '文献调研',
  'idea-creator': '论文大纲',
  'novelty-check': '新颖性检验',
  'research-review': '外部审查',
  'research-refine-pipeline': '文献综述与实验规划',
}

/** 工作台子步骤展示名（4 步） */
export const ARIS_UI_PHASE_LABEL_ZH: Record<ArisIdeaDiscoveryUiPhaseId, string> = {
  'lit-idea-creator': '文献调研与选题大纲',
  'novelty-check': '新颖性检验',
  'research-review': '外部审查',
  'research-refine-pipeline': '文献综述与实验规划',
}

export function arisPhaseIdLabel(id: ArisIdeaDiscoveryUiPhaseId): string {
  return ARIS_UI_PHASE_LABEL_ZH[id] ?? id
}

export const ARIS_IDEA_DISCOVERY_PHASES_LABEL_ZH = ARIS_IDEA_DISCOVERY_UI_PHASES.map(
  (id) => ARIS_UI_PHASE_LABEL_ZH[id],
).join(' → ')

const RUN_PHASE_COUNT = ARIS_IDEA_DISCOVERY_PHASES.length

function internalRunPhaseIndex(phase: string): number {
  const p = phase.trim()
  if (!p) return 0
  const idx = ARIS_IDEA_DISCOVERY_PHASES.indexOf(p as ArisIdeaDiscoveryPhaseId)
  return idx < 0 ? 0 : idx
}

/** 将 run_state 的 pending phase 映射到 UI 四步中的下标 */
export function arisRunPhaseToUiNextIndex(runPhase: string): number {
  const idx = internalRunPhaseIndex(runPhase)
  if (idx <= 1) return 0
  return idx - 1
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

/** 当前待用户确认的 ARIS 子阶段中文名（无则 null） */
export function arisPhaseAwaitingConfirmLabel(
  steps: TopicDiscoveryStepDTO[] | undefined,
  runStatus?: string,
): string | null {
  const waiting = arisIdeaDiscoveryPhaseRows(steps, runStatus).find(
    (r) => r.status === 'waiting_confirm',
  )
  if (!waiting) return null
  return arisPhaseIdLabel(waiting.id)
}

/** 由 idea_discovery 步 result（aris_pause_after_phase / aris_awaiting_checkpoint）与进度推导子 phase 状态 */
export function arisIdeaDiscoveryPhaseRows(
  steps: TopicDiscoveryStepDTO[] | undefined,
  runStatus?: string,
): ArisPhaseUiRow[] {
  const step = ideaDiscoveryStepDto(steps)
  const uiPhases = ARIS_IDEA_DISCOVERY_UI_PHASES

  if (!step || step.status === 'pending') {
    return uiPhases.map((id) => ({ id, status: 'pending' as const }))
  }
  if (step.status === 'completed') {
    return uiPhases.map((id) => ({ id, status: 'completed' as const }))
  }

  const result = (step.result ?? {}) as Record<string, unknown>
  const awaiting =
    result.aris_awaiting_checkpoint === true ||
    (runStatus === 'checkpoint' && String(result.aris_pause_after_phase ?? '').trim() !== '')
  const pausePhase = String(result.aris_pause_after_phase ?? '').trim()
  const pendingPhase = String(result.aris_pending_phase ?? '').trim()
  const nextPhase =
    awaiting ? pausePhase : pendingPhase || pausePhase
  let uiNextIdx = arisRunPhaseToUiNextIndex(nextPhase)

  if (!nextPhase && step.status === 'running') {
    const pct = typeof step.exec_progress_pct === 'number' ? step.exec_progress_pct : 0
    if (pct > 0) {
      const internalIdx = Math.min(
        RUN_PHASE_COUNT - 1,
        Math.max(0, Math.floor((pct / 100) * RUN_PHASE_COUNT)),
      )
      uiNextIdx = internalIdx <= 1 ? 0 : Math.min(uiPhases.length - 1, internalIdx - 1)
    }
  }

  return uiPhases.map((id, i) => {
    if (i < uiNextIdx) return { id, status: 'completed' as const }
    if (i === uiNextIdx) {
      if (awaiting) return { id, status: 'waiting_confirm' as const }
      if (step.status === 'running') return { id, status: 'running' as const }
      return { id, status: 'pending' as const }
    }
    return { id, status: 'pending' as const }
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
