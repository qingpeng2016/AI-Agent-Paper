# Paper Skill 一览

**状态：** 已有 = `skills/<name>/` 有 bundle；规划 = 尚未建 bundle；上游 = ARIS 安装；平台 = 后端/前端步骤。

---

## 1. 数据与下载

| 命令 / 步骤 | 状态 | 做什么 |
|-------------|------|--------|
| `/data-source-discovery` | 已有 | 找数据、许可、验 URL、写 `data-source/` 画像与 manifest；默认不全量下载 |
| `/data-source-acquire` | 规划 | 按 manifest 把数据拉到 `data-source/raw/` |
| `/data-source-prep` | 规划 | 预处理 → `data-source/derivatives/` |
| `/experiment-plan`（实验数据获取段） | 已有 | 方法一/二写进 plan，与 PROFILE 对齐；不负责大文件下载 |
| 选题文献 PDF 下载 | 平台 | retrieve 时下载 PDF 到 storage |
| 用户文献上传 | 平台 | Corpus / verified 文献 |

---

## 2. 工作流 1（UI：`topic-discovery`）

| 命令 / 步骤 | 状态 | 做什么 |
|-------------|------|--------|
| `/research-lit` | 已有 | Step 1 · 文献矩阵 |
| `/idea-creator` | 已有 | Step 2 · idea |
| `/novelty-check` | 已有 | Step 3 · 新颖性 |
| `/research-review` | 已有 | Step 4 · 评审 |
| `/research-refine-pipeline` | 已有 | Step 5 · refine + plan 一体 |
| `/research-refine` | 已有 | 独立 refine（非 Step5 pipeline） |
| `topic_confirm` | 平台 | 锁定选题快照 |
| `literature-review` 模块 | 平台 | 工作流 2 · 综述正文（选题之后） |

---

## 3. 实验（UI：`experiment-planning`）

| 命令 | 状态 | 做什么 |
|------|------|--------|
| `/experiment-plan` | 已有 | 四文件 EXPERIMENT_PLAN + TRACKER |
| `/experiment-bridge` | 已有 | 写码、审码、sanity、deploy、收结果 |
| `/run-experiment` | 上游 | 启动训练 / 批处理（可单独调用） |
| `/monitor-experiment` | 上游 | 查日志与结果 JSON |

---

## 4. 工作流 3 — 准备写论文（UI：`paper-writing`）

整包：`/paper-writing`（规划）。或分步：

| 命令 | 状态 | 做什么 |
|------|------|--------|
| `/paper-plan` | 规划 | 大纲、章节与 claim、图表占位 |
| `/paper-figure` | 规划 | 图表规格/生成（或与「图表管理」模块配合） |
| `/paper-write` | 规划 | 正文 Md/LaTeX |
| `/paper-compile` | 规划 | 模板编译 PDF |
| `/auto-paper-improvement-loop` | 规划 | 按审计多轮润色 |
| `/paper-claim-audit` | 规划 | 引用 / 论断 / 驳论审计（前端「审计」步） |

---

## 5. ARIS `SKILL.md` 正文小节（参考）

上游 ARIS 无强制枚举；下表为 `skills/*/SKILL.md` 里 **反复出现的 H2/H3 类型**（非 YAML 字段）。写 Paper bundle 时可对照；**不是每个 skill 段落齐全**。

| 章节关键词 | 常见度 | 职责 |
|------------|--------|------|
| **Overview** | 高 | 本 skill 在整条研究链里 **干什么、输入输出关系** |
| **Constants** | 高 | **默认配置**（路径、上限、开关）；`> Override:` / `— key: value` 可改 |
| **Inputs** | 中高 | **开工前要读的文件**；缺失时 fallback / 问用户 |
| **Context: $ARGUMENTS** | 中 | 当次用户参数 **注入上下文**（多见于 loop 类） |
| **When to Use This Skill** | 中 | 与 **其它 skill 的分工**（何时用、何时不用） |
| **Workflow** | 极高 | **主流程**；其下为 Phase / Step |
| **Phase N / Step N** | 极高 | **单步**：读啥、写啥、调哪个 `/skill` 或 Bash |
| **Key Rules** | 高 | **硬约束** ALWAYS / NEVER（诚信、GPU、勿套 `/loop` 等） |
| **Output Protocols** | 高 | **交付契约**：路径、命名、manifest、append log |
| **Composing with Other Skills** | 中高 | **上下游 slash**：先谁后谁、invoke 谁 |
| **Deliverables** / **Output Files** | 中 | 产出 **清单或目录树**（有时并入 Output Protocols） |
| **Next Step(s)** | 中 | 结束后 **建议下一个 skill** |
| **Pre-flight Check** | 低中 | 主流程前 **环境 / 文件 / GPU** 检查 |
| **State Persistence** | 中 | **断点续跑**、COMPACT 恢复 |
| **Reviewer Calling Convention** | 审查类 | **Codex/MCP** 调用约定 |
| **Pipeline** / **Pipeline Summary** | 复合类 | 内嵌 **多子阶段**（如 paper-writing） |
| **CLAUDE.md Example** | 执行类 | **算力/环境** 配置示例 |
| **Loop (repeat up to …)** | loop 类 | **Workflow 内多轮**：`repeat up to MAX_ROUNDS`；每轮 Phase A→B→C…，**非** OS/Cursor `/loop` |
| **Initialization**（resume） | loop 类 | 读 **`REVIEW_STATE.json`** 等；`in_progress` 续跑 / `completed` 新开 |
| **Stop-Evaluation Gate** / **STOP CONDITION** | loop 类 | **何时退出循环**（分数+verdict 与 helper 脚本判定，勿模型臆断） |
| **Phase B.6 Debate Protocol** 等 | loop 类 | 高难度审查：**多轮记忆、反驳**（`REVIEWER_DIFFICULTY`） |

**Loop 类 Constants / 关键词（正文里，非 YAML）：**

| 关键词 | 职责 |
|--------|------|
| **`MAX_ROUNDS`** / **`MAX_ITERATIONS`** | 轮数硬上限（如 review 4 轮、润色 2 轮） |
| **`POSITIVE_THRESHOLD`** | 达标停：`score` + `verdict` **同时**满足才停 |
| **`HUMAN_CHECKPOINT`** | 每轮审完后 **等人** 再继续或提前停 |
| **`REVIEW_STATE.json`** / **`status: completed`** | 断点续跑；跑完标记 completed |
| **`threadId` / `codex-reply`** | 同 reviewer **跨轮记忆**（外裹 `/loop` 会重置） |
| **Do not wrap `/loop` or Cron** | 内环已循环；外再套 **无新信号、白烧 token** |
| **invoke `/strategy-code` ↔ `/strategy-review`** | 编排型 loop：**只调原子 skill**，自己汇总 manifest |

**上游名字含 `loop` 的 skill：** `/auto-review-loop`、`/auto-paper-improvement-loop`、`/dse-loop`（及 llm/minimax 变体）。**Pipeline/bridge/queue** 串多步但不一定叫 loop。

**常抄骨架：** `Overview` → `Constants` → `Inputs`（可选）→ `Workflow` → `Output Protocols` → `Key Rules` → `Composing`（可选）。文首 **`>` 引用块**（reviewer、external-cadence）等同加强版 Key Rules。

**开放标准（全球）** 仅要求 frontmatter 的 **`name` + `description`**；上表均为 ARIS 正文习惯。
