import {
  ARIS_IDEA_DISCOVERY_PHASES,
  type ArisIdeaDiscoveryPhaseId,
} from './types'
import type { TopicDiscoveryStepDTO } from './src/api/topicDiscovery'

export type ArisPhaseUiStatus = 'pending' | 'running' | 'waiting_confirm' | 'completed'

export type ArisPhaseUiRow = {
  id: ArisIdeaDiscoveryPhaseId
  status: ArisPhaseUiStatus
}

function ideaDiscoveryStepDto(
  steps: TopicDiscoveryStepDTO[] | undefined,
): TopicDiscoveryStepDTO | undefined {
  return steps?.find(
    (s) => s.stage_code === 'idea_discovery' || s.stage_code === 'topic_discovery',
  )
}

/** 由 idea_discovery 步 result（aris_pause_after_phase / aris_awaiting_checkpoint）与进度推导子 phase 状态 */
export function arisIdeaDiscoveryPhaseRows(
  steps: TopicDiscoveryStepDTO[] | undefined,
  runStatus?: string,
): ArisPhaseUiRow[] {
  const step = ideaDiscoveryStepDto(steps)
  const phases = ARIS_IDEA_DISCOVERY_PHASES

  if (!step || step.status === 'pending') {
    return phases.map((id) => ({ id, status: 'pending' as const }))
  }
  if (step.status === 'completed') {
    return phases.map((id) => ({ id, status: 'completed' as const }))
  }

  const result = (step.result ?? {}) as Record<string, unknown>
  const awaiting =
    result.aris_awaiting_checkpoint === true ||
    (runStatus === 'checkpoint' && String(result.aris_pause_after_phase ?? '').trim() !== '')
  const nextPhase = String(result.aris_pause_after_phase ?? '').trim()
  let nextIdx = nextPhase ? phases.indexOf(nextPhase as ArisIdeaDiscoveryPhaseId) : 0
  if (nextIdx < 0) nextIdx = 0

  if (!nextPhase && step.status === 'running') {
    const pct = typeof step.exec_progress_pct === 'number' ? step.exec_progress_pct : 0
    if (pct > 0) {
      nextIdx = Math.min(
        phases.length - 1,
        Math.max(0, Math.floor((pct / 100) * phases.length)),
      )
    }
  }

  return phases.map((id, i) => {
    if (i < nextIdx) return { id, status: 'completed' as const }
    if (i === nextIdx) {
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
