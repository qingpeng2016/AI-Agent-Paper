/** 各工作流模块演示产出（前端写死，接 Agent 后替换为 API） */

export const DEMO_LIT_REVIEW = {
  retrieved: 86,
  verified: 79,
  /** 合并选题 artifact 后的综述大纲（演示） */
  outline: [
    '1. 领域背景（来自 retrieve + 研究方向）',
    '2. 分主题 Related Work（入库 79 篇归纳）',
    '3. 候选 idea 与新颖性对照（选题 checkpoint 产出）',
    '4. 实验语境与已有基线（实验计划摘要）',
    '5. Research gap → 支撑 Introduction 贡献表述',
  ],
  sections: [
    {
      title: '静态与动态稀疏注意力',
      papers: [
        'Chen et al., 2024 — SparseAttn (arXiv:2403.xxxxx) ✓',
        'Liu et al., 2023 — Dynamic KV Pruning (OpenAlex) ✓',
      ],
    },
    {
      title: '长上下文与 KV 压缩',
      papers: [
        'Zhang et al., 2024 — H2O (Semantic Scholar) ✓',
        'Wang et al., 2024 — SnapKV ✓',
      ],
    },
  ],
  gap: '现有工作较少在统一预算下同时比较训练-free 路由与可学习路由；跨 32k+ 任务的系统消融仍不足。',
  excerpt:
    'We organize prior work into three lines: fixed sparsity patterns, input-dependent routing, and cache compression…',
  unifiedExcerpt:
    '【综述正文 · 演示】\n\nBackground. Motivated by the topic direction on budget-aware sparse attention…\n\nRelated Work. (1) Static/dynamic sparsity… (2) KV compression…\n\nPositioning. Among candidate ideas, Idea A (dynamic routing) remains best supported by novelty check…\n\nGap. Prior work rarely compares training-free routers under a unified KV budget…',
}

export const DEMO_EXPERIMENT_PLAN = {
  hypothesis: '在固定 KV 预算 B 下，训练-free 动态路由不低于专用稀疏训练基线，且吞吐优于 Full Attention。',
  baselines: ['Full Attention', 'StreamingLLM', 'H2O', 'SparseAttn (repro)'],
  metrics: ['LongBench avg', 'Passkey', '吞吐 tok/s', '峰值显存 GB'],
  ablations: ['路由粒度', '预算 B', '是否重排 KV'],
  steps: [
    'Week 1–2：复现 H2O / SparseAttn 官方配置',
    'Week 3：主实验 LongBench + 自建 32k',
    'Week 4：消融 + 3 seed 汇总',
  ],
}

export const DEMO_AUTO_REVIEW = {
  target: '实验方案 + 上传数据 results_main.csv',
  scores: { rigor: 6.5, completeness: 5.5, claimSupport: 6.0 },
  findings: [
    { level: 'major', text: '缺少 StreamingLLM 对照（plan 中列出但 CSV 无对应列）' },
    { level: 'major', text: '主结果仅 1 seed，无 stderr；与 NeurIPS 实验门槛不一致' },
    { level: 'minor', text: 'Passkey 提升 2.1% 但摘要写 5%（与表 2 不一致）' },
  ],
  kill:
    '若审稿人认为 routing 开销抵消吞吐收益，需提供端到端 latency 分解；当前数据不足以反驳。',
}

export const DEMO_MANUSCRIPT = {
  title: 'Budget-Aware Training-Free Sparse Attention for Long Context',
  sections: [
    { name: 'Abstract', words: 168, status: 'done' },
    { name: 'Introduction', words: 820, status: 'done' },
    { name: 'Related Work', words: 640, status: 'done' },
    { name: 'Method', words: 1100, status: 'done' },
    { name: 'Experiments', words: 950, status: 'draft' },
  ],
  excerpt: `We study training-free sparse attention under a fixed KV budget B. Our router assigns token pairs without gradient updates…`,
  cites: 42,
  verifiedCites: 39,
}

export const DEMO_FIGURES = {
  files: [
    { id: 'fig1', title: 'LongBench 主结果', type: 'bar', note: 'demo_main_results.png' },
    { id: 'fig2', title: '预算 B 消融', type: 'line', note: 'demo_ablation_B.png' },
    { id: 'fig3', title: '吞吐 vs 准确率', type: 'scatter', note: 'demo_throughput.png' },
  ],
  dataPreview: [
    { method: 'Ours', score: 72.4, std: 0.8 },
    { method: 'H2O', score: 70.1, std: 1.1 },
    { method: 'Full', score: 73.0, std: 0.6 },
  ],
}

export const DEMO_MANUSCRIPT_ANALYSIS = {
  overall: 7.2,
  recommendation: 'Weak Accept → 需补实验与改 claim 后可再投',
  dimensions: [
    { name: 'Novelty', score: 7.5 },
    { name: 'Evidence', score: 6.0 },
    { name: 'Clarity', score: 8.0 },
    { name: 'Venue fit', score: 7.0 },
  ],
  mustFix: [
    'Experiments：补 baseline 与 3 seeds',
    'Abstract：数字与 Table 2 对齐',
    'Related Work：2 处引用待 verification（门禁已标 pending）',
  ],
  kill: '训练-free 是否仅在长序列成立？需在 ≤8k 任务报告是否退化。',
}
