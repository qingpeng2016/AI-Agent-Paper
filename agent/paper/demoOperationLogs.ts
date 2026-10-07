import type { PaperModuleId, PaperOperationLogEntry } from './types'

/** 演示：各模块一次「运行」的 token 估算（接入计费 API 后替换） */
export const DEMO_MODULE_TOKEN_ESTIMATES: Partial<
  Record<PaperModuleId, { prompt: number; completion: number }>
> = {
  'topic-discovery': { prompt: 18_400, completion: 6_200 },
  'literature-review': { prompt: 9_800, completion: 4_100 },
  'experiment-planning': { prompt: 5_200, completion: 2_400 },
  'auto-review': { prompt: 7_600, completion: 3_300 },
  'paper-writing': { prompt: 22_000, completion: 11_500 },
  'figure-generation': { prompt: 3_400, completion: 1_200 },
  'manuscript-analysis': { prompt: 14_200, completion: 5_800 },
  environment: { prompt: 120, completion: 80 },
}

export const DEMO_OPERATION_LOG_SEED: PaperOperationLogEntry[] = [
  {
    id: 'op-seed-1',
    occurredAt: '2026-10-05T09:12:04.000Z',
    moduleId: 'topic-discovery',
    moduleLabel: '选题发现',
    action: '运行工作流（retrieve → ideas → novelty → audit）',
    tokensPrompt: 18_400,
    tokensCompletion: 6_200,
    tokensTotal: 24_600,
    manuscriptId: 'ms-demo-1',
    manuscriptTitle: '稀疏注意力机制',
    status: 'success',
  },
  {
    id: 'op-seed-2',
    occurredAt: '2026-10-05T10:48:31.000Z',
    moduleId: 'literature-review',
    moduleLabel: '文献综述',
    action: '合并选题 artifact 生成 literature_review',
    tokensPrompt: 9_800,
    tokensCompletion: 4_100,
    tokensTotal: 13_900,
    manuscriptId: 'ms-demo-1',
    manuscriptTitle: '稀疏注意力机制',
    status: 'success',
  },
  {
    id: 'op-seed-3',
    occurredAt: '2026-10-06T14:22:18.000Z',
    moduleId: 'paper-writing',
    moduleLabel: '论文写作',
    action: '分节撰写（Introduction + Related Work）',
    tokensPrompt: 12_400,
    tokensCompletion: 6_800,
    tokensTotal: 19_200,
    manuscriptId: 'ms-demo-2',
    manuscriptTitle: '多模态 RAG 审计',
    status: 'success',
  },
  {
    id: 'op-seed-4',
    occurredAt: '2026-10-06T16:05:02.000Z',
    moduleId: 'auto-review',
    moduleLabel: '结果审查',
    action: '运行审查（实验数据未上传）',
    tokensPrompt: 0,
    tokensCompletion: 0,
    tokensTotal: 0,
    manuscriptId: 'ms-demo-2',
    manuscriptTitle: '多模态 RAG 审计',
    status: 'failed',
    note: '缺少上传数据，未调用模型',
  },
]
