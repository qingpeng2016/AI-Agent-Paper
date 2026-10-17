# Paper Agent 研究工作流策略（PRD 骨架）

本文档描述从参考文献入库到选题发现、实验、正文写作与查重的 **策略流程**、**三门禁**、**迭代规则** 与 **实现注意事项**。与当前代码实现对齐情况见文末「实现差距」。

---

## 1. 术语

| 术语 | 含义 |
|------|------|
| **Corpus** | 本篇论文已入库、可引用的文献集合（检索命中 + 用户上传，经校验后带 `ref_id` / `external_key`） |
| **verified** | 文献已通过平台校验，可进入引用门禁 |
| **ref_id / reference_keys** | 正文与 JSON 中引用文献的稳定键，必须可追溯到 Corpus |
| **Venue 约束** | 贡献类型、实验标准等（`paper_ref_venue.llm_description_codes` → 2xxxx） |
| **写作风格** | 学术表述习惯（Related Work 写法等）；**不等于** LaTeX 字体/页边距 |
| **版式** | 页数、字体、模板 — 由 **人工配置模板 + 编译** 完成，非 LLM 直接交付 |
| **SOTA** | State of the Art，当前最优/领先水平；论断门禁需核查是否夸大、是否有基线支撑 |
| **audit_report** | 结构化审计结果：`gates`、`issues[]`、`passed` 等 |

---

## 2. 三门禁（全局定义）

三门禁在 **审计步骤** 中正式判定；生成步骤仅通过 **输出约束 + prompt 自检提示** 引导模型，**以审计输出为准**。

### 2.1 引用门禁（Citation）

- 论断是否能在 **Corpus / 已 verified 文献** 中找到依据  
- 引用是否乱编、`ref_id` 是否合法  
- idea / 综述中的 `reference_keys` 是否可追溯  
- **未 verified 的文献不得进入可引用集合**（与 UI「参考文献门禁」一致）

### 2.2 论断门禁（Claim）

- 贡献、因果、实验结论、SOTA 表述等是否被证据支撑  
- 是否缺基线/对照、是否夸大  
- 是否满足目标 venue 的 **贡献类型 / 实验标准**（管理配置）

### 2.3 驳论门禁（Kill argument）

- 审稿人视角的一票否决理由：致命 flaw、与文献矛盾、不可复现、venue 明显不合适等  
- 强度可配置（`paper_llm_constraint_gate` 档位 + `paper_llm_description`）

### 2.4 过关判定（建议默认）

| 配置项 | 建议 |
|--------|------|
| **通过** | 三门禁均无 `blocker` 级 issue（`major` 是否允许带警告通过可配置） |
| **迭代终止** | 达到 `max_revision_rounds` 仍不通过 → 保留「blocker 最少」的一版 + 完整 `audit_report`，可选人工 **waived** |

---

## 3. 参考文献（阶段零）

在进入选题发现 LLM 之前完成；产出 **Corpus 快照**（含 `corpus_revision` 版本号）。

### 3.1 检索入库

1. 根据 **关键词、详细描述** 检索文献（数量可配置，如 `paper_exec_intensity_topic.max_papers`）  
2. **下载原文** 至 storage（如 `paper-files/`），retrieve 步落库 hits / brief  
3. 可选：**用户上传** 参考文献，合并进 Corpus（须走校验 → verified）

### 3.2 随机化（仅 **新 run**）

目的：避免每次选题 run 按固定顺序拿到相同文献。

- 方案：多次查询凑齐数量；**平台随机**；**分页大小与页码随机**  
- **同一 run 内**（含选题生成↔审计多轮迭代）：**禁止**重新随机或整库重搜，除非触发 **Corpus 增补**（见 §5.4）

### 3.3 输入 LLM 时的 Corpus 表述（实现建议）

文档与产品层可称「参考文献」；**实现上**建议：

- 默认：`ref_id`、标题、摘要、年份、verified 状态、本地 PDF 路径索引  
- **非必要不将 PDF 全文塞入 prompt**；按 ref 按需截取片段  
- 与用户上传、retrieve 步 `literature_hits` / 后续 verified Corpus 对齐

---

## 4. 选题发现

绑定 `paper_manuscript`（含 **论文语言** `language` zh/en）。

### 4.1 流程概览

```text
[阶段零] Corpus 快照 (corpus_revision = 1)
    ↓
[1] 生成 — LLM 输出四块内容（带 venue 输出约束，非正式门禁）
    ↓
[2] 审计 — LLM 三门禁判定 + audit_report
    ↓
  通过？ —否→ [1]+[2] 迭代（Corpus 不变，带上一版全文 + audit_report）
    ↓ 是
锁定版本 → 进入「实验计划 / 实验数据 / 正文」等后续阶段
```

### 4.2 步骤 1：生成

| 项 | 说明 |
|----|------|
| **输入** | 关键词、详细描述、目标期刊/会议、Corpus（索引/简报/按需原文）、论文语言 |
| **输出约束（非门禁）** | 见下表 |
| **输出** | idea、新颖性、文献综述、实验计划（可为空） |

**输出约束：**

| 块 | 约束来源 |
|----|----------|
| idea + 新颖性 + 文献综述 | venue **贡献类型**等（`paper_llm_description` constraint_venue 文案） |
| 实验计划 | venue **实验标准**；**允许不输出**（理论/综述等无需实验时，prompt 必须说明可为 null） |

> **注意：** 步骤 1 **不**产出正式「是否通过三门禁」结论；若 prompt 含自检提示，仅作辅助。

### 4.3 步骤 2：审计

| 项 | 说明 |
|----|------|
| **输入** | 同步骤 1 的上下文 + 步骤 1 的 **完整四块输出** |
| **审计** | 引用门禁、论断门禁、驳论门禁（对四块内容分别或汇总判定） |
| **输出** | `passed`（或分门禁 passed）、结构化 **`audit_report`**（不满足时：gate、severity、location、claim、fix） |

实验计划为空时：论断门禁 **跳过** 与实验设计相关的检查项（配置化）。

### 4.4 步骤 3：多轮迭代（步骤 1 + 2）

| 规则 | 说明 |
|------|------|
| **触发** | 步骤 2 未通过且 `round < max_revision_rounds` |
| **Corpus** | **必须与首轮相同**（同一 `corpus_revision`）；保证引用门禁可对照 |
| **带入步骤 1 的额外输入** | ① 上一轮 **完整四块输出**（仅保留迭代链上的 **最新一版** 全文即可） ② 上一轮 **`audit_report`**（结构化；不必堆叠每一版失败全文） |
| **步骤 1 模式** | **修订（revise）**：在保持研究方向前提下最小改动过关；禁止编造新 ref |
| **步骤 2** | 对 **本轮步骤 1 新输出** 重新审计；仍带 Corpus + venue 约束 |
| **轮数** | 配置项 `max_revision_rounds`（与 audit 档位解耦，独立策略或 env） |

### 4.5 Corpus 增补（例外）

仅在以下情况 **增量检索**，`corpus_revision += 1`：

- `audit_report` 明确标记 `needs_more_literature` / 覆盖不足  
- 用户修改关键词或详细描述（视为新意图，可新 run 或增补）  
- 用户手动「扩充文献」

增补后：audit 须说明 **仅 revision ≥ N 的 ref 可用于新增引用**。

---

## 5. 阶段 × 输入 × 输出 × 门禁 × 迭代（总表）

| 阶段 | 生成/动作 | 主要输入 | 输出 | 正式门禁 | 迭代 |
|------|-----------|----------|------|----------|------|
| 零 | 检索+下载+上传校验 | 关键词、描述、来源、数量配置 | Corpus 快照 | — | 新 run 可随机；run 内冻结 |
| 选题-1 | 生成四块 | Corpus、venue、语言、约束模板 | idea、novelty、综述、实验计划? | 无（仅约束） | 见 §4.4 |
| 选题-2 | 审计 | 同上 + 四块全文 | audit_report、passed | 引用+论断+驳论 | 与选题-1 成对迭代 |
| 实验数据 | 校验数据 | 用户上传数据 + 选题锁定快照 | 是否满足计划/论断；原因 | 论断+驳论 | 可配置；换数据 → 新 experiment_run |
| 实验数据 | 生成图表 | 数据 + 快照 | 常规图/表（复杂图人工/外部） | 可选 | — |
| 正文-1 | 写全文 | Corpus、快照、**实验结果**、语言 | 论文全文（Md/Tex 草稿） | 无（venue 写作风格约束） | 见 §6.3 |
| 正文-2 | 审计全文 | 同上 + 全文 | audit_report | 引用+论断+驳论 | 与正文-1 成对迭代 |
| 查重 | 查重 API | 全文 | 相似度报告 | — | — |
| 查重后 | 分段改写 | 报告 + 全文 | 降重表述 | **建议再跑论断门禁** | 限轮；禁止改事实与结论 |

**选题通过后锁定快照（建议字段）：** `manuscript_id`、`corpus_revision`、idea/novelty/review/plan 的版本 id、venue、语言。

---

## 6. 实验计划（阶段）

### 6.1 实验数据 — AI 校验

| 项 | 说明 |
|----|------|
| **输入** | 用户实验数据 + 选题锁定快照（关键词、描述、venue、Corpus 索引、idea、novelty、综述、实验计划） |
| **门禁** | 论断门禁、驳论门禁（对照实验计划与 idea，不单看文件格式） |
| **输出** | 是否满足要求；不满足则原因 |

### 6.2 实验数据 — 生成图表

- 输入：实验数据 + 同上快照  
- 输出：常规统计图/表；**复杂定制图**需人工或外部工具  
- 若实验结果 **推翻** 选题结论：产品策略二选一或 configurable — **回到选题迭代** 或 **仅在正文/驳论中标注限制**

---

## 7. 论文正文

### 7.1 步骤 1：生成全文

| 项 | 说明 |
|----|------|
| **输入** | 关键词、描述、venue、Corpus、选题快照、**实验结果**（含图表描述或路径）、论文语言 |
| **输出约束** | venue **写作风格**（ prose ）；**版式** 走模板配置 + 编译，不由 LLM 保证字体页边距 |
| **输出** | 论文全文 |

### 7.2 步骤 2：审计全文

| 项 | 说明 |
|----|------|
| **输入** | §7.1 输入 + 全文 |
| **门禁** | 引用、论断、驳论 |
| **输出** | `audit_report`、`passed` |

### 7.3 多轮迭代（步骤 1 + 2）

规则同 §4.4：

- **Corpus 不变**（同一 `corpus_revision`）  
- **实验结果** 迭代内一般 **冻结**；更换上传数据应新开 experiment_run  
- 带入 **上一版全文 + audit_report**

---

## 8. 论文查重（查虫）

1. 调用 **查重 API**，输出相似度报告  
2. 若不满足要求：按段请 AI **在不改变逻辑与结论** 的前提下调整表述  
3. **改写后建议再跑论断门禁**（及必要处引用门禁），防止改表述引入夸大或新 claim  
4. 合规：产品定位宜为「合规表述优化 + 自查」，非单纯规避检测

---

## 9. 配置与管理后台（汇总）

| 配置 | 表/位置 | 作用阶段 |
|------|---------|----------|
| 执行强度 | `paper_exec_intensity_topic` | `exec_rounds`（生成↔审计迭代轮数）、`max_papers`、`max_ideas` |
| 三门禁档位 | `paper_llm_constraint_gate`（`prompt_scene`=topic/write_paper） | `llm_description_*_check_code` → `description_code`（topic 10000+ / write_paper 11000+；0=关该 gate） |
| LLM 文案库 | `paper_llm_description` | `description_code`（按 prompt_code 分段）、`prompt_name` / `prompt_desc`、`model_config_id` |
| Venue 约束 | `paper_ref_venue` | 每 venue 一条 `llm_description_codes`（默认 20000–20002）；选题 run `input_params.venue_codes` |
| 选题环节话术 | `paper_llm_description` + `paper_llm_template_topic` | stage_code → `llm_description_code`（30000 段） |
| 迭代上限 | 待增或 env | `max_revision_rounds`（选题/正文可分开） |
| 查重阈值 | 待增 | 相似度上限 |

---

## 10. 与当前代码实现的差距（2025-10 快照）

便于研发排期；**不代表流程设计错误**。

| 设计 | 现状简述 |
|------|----------|
| 两步 UI：检索 + 选题发现 | `retrieve` + `topic_discovery`（bot `llm_topic_discovery`） |
| 生成包 ↔ 审计迭代 | `exec_rounds` 驱动；修订轮带上一版 bundle + audit_report |
| 实验计划 | 含于生成包；不再单独 `experiment_plan_llm` bot |
| 三门禁强度 | DB 有字段，**未**接入 audit prompt/规则 |
| Venue 约束注入 | 多为 venue 字符串进 prompt；**待**按 primary venue → profile → description 加载 |
| 用户上传 + 随机检索 | 需扩展 retrieve |
| 引用门禁台账表 | 已移除；Corpus / retrieve 与 audit JSON 承载 |
| 正文 / 实验数据 AI / 查重 | 部分 UI 演示或未全链路 |

---

## 11. 修订记录

| 日期 | 说明 |
|------|------|
| 2025-10-11 | 初版：合并产品流程、三门禁、迭代与 Corpus 冻结规则 |
