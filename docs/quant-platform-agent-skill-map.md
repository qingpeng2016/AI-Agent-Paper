# 量化研究平台 — Agent 与 Skill 映射

> **范围：** 仅 **Agent 层** 的「谁是谁、由哪些 Skill 组成」。  
> **不含：** 调用顺序、失败路由、契约字段、Tool API — 见 [`quant-platform-agents.md`](./quant-platform-agents.md)。  
> **实盘 / 执行链 Agent** 见 [`quant-platform-live-agents.md`](./quant-platform-live-agents.md)。

**日期：** 2026-10-23  
**关联：** [`quant-platform-agents.md`](./quant-platform-agents.md) §0–§3、§5 起

---

## 1. 术语（避免把 Skill 和 Agent 混为一谈）

| 概念 | 含义 | 工程产物 |
|------|------|----------|
| **Skill** | 一个 **`/slash-command`** 步骤；有固定交付与 `*.manifest.result.json` | `agents/<skill-name>/SKILL.md` + `contracts/<skill-name>.*` |
| **Loop Skill** | **编排型 Skill**：按规则 **invoke** 其他 Skill，汇总 manifest；**不**代替子 Skill 写码、审码、跑 gate | 同上；实现多为 Shell/Python，可选轻量 LLM 只做摘要 |
| **Agent（逻辑角色）** | 架构上的 **职责角色**（Orchestrator、Reviewer 等）；**由 1–N 个 Skill 承担**，不是第四个工程目录类型 | 文档与编排配置中的 **角色名**；运行时 = 调用对应 Skill |

**约定：**

- 下文 **「Skill 清单」** 列的是 **13 个 slash-command**（10 原子 + 3 Loop）。  
- 下文 **「Agent 角色」** 列的是 **8 个逻辑角色**；其中 Researcher 等 **跨多个 Skill**，由 `/quant-orchestrate` 或 Loop Skill 串联。  
- **固件**（回测引擎、G1–G8、`gate.judge`）**不是 Skill**，也 **不** 算作 Agent 层 slash-command；`/strategy-validate-run` 只是 **触发固件** 的 Skill。

---

## 2. 数量摘要

| 类别 | 数量 | 说明 |
|------|------|------|
| **原子 Skill** | **10** | 单轮职责：数据锚点、假设、写码、审查、风控、回测触发、分析、因子、报告、虚拟盘观察 |
| **Loop / 编排 Skill** | **3** | 任务协调、代码审查闭环、策略研究外环 |
| **slash-command 合计** | **13** | 每个对应一行契约与（通常）一个 `agents/<name>/` 目录 |
| **逻辑 Agent 角色** | **8** | Orchestrator、Data Anchor、Researcher、Reviewer、Research Risk、Validator、Factor、Reporting |
| **独立侧链** | **1** | Factor（`/factor-mining`）默认 **不** 进 `/quant-orchestrate` 主链 |

---

## 3. Agent 角色 ↔ 组成的 Skill

一个 **Agent 角色** 至少对应 **一个** Skill；Researcher / Validator 等 **多个 Skill 同属同一角色**，由编排器按阶段启用。

| Agent 角色 | 职责（一句话） | 组成的 Skill（1–N） | 备注 |
|------------|----------------|---------------------|------|
| **Orchestrator** | 读 Brief、选 `hypothesis_id`、派单、Token/族上限 | **`/quant-orchestrate`** | Loop Skill；失败路由见主文档 §5.1 |
| **Data Anchor** | 把任务锚定到数据中心 **已有** dataset + snapshot | **`/market-data-profile`** | 不拉数、不写库 |
| **Researcher** | 假设、实现插件、不达标时换方向再试 | **`/strategy-hypothesis`**、**`/strategy-code`**、**`/strategy-code-loop`**、**`/strategy-backtest-analysis-loop`** | `code` 通常 **只** 由 `code-loop` 按轮 invoke；外环 loop invoke 假设 + code-loop + 下游 |
| **Reviewer** | 与写码 **隔离** 的代码/文档审查 | **`/strategy-review`** | 由 **`/strategy-code-loop`** 按轮 invoke |
| **Research Risk** | G2 通过后、回测前的 **策略风险** 设计审查 | **`/strategy-risk-control`** | 不替代 G3–G8、不替代实盘 L1–L5 |
| **Validator** | 确定性回测 + **解读** 是否满足 Brief/Gate | **`/strategy-validate-run`**、**`/strategy-backtest-analysis`** | Gate **判定** 在固件；分析 Skill **只解释** |
| **Factor** | 维护因子库，供假设去重与引用 | **`/factor-mining`** | **独立任务**；Brief 未要求时不进主链 |
| **Reporting** | 达标后出报告；可选虚拟盘偏离观察 | **`/strategy-report`**、**`/virtual-paper-watch`**（可选） | 报告仅 **`meets_requirements: true`** 后 |

**Loop Skill 与角色的关系（不重复计为 Agent）：**

| Loop Skill | 编排对象 | 对应角色组合 |
|------------|----------|--------------|
| **`/strategy-code-loop`** | `/strategy-code` ↔ `/strategy-review` | Researcher + Reviewer |
| **`/strategy-backtest-analysis-loop`** | `/strategy-hypothesis` → `/strategy-code-loop` → `/strategy-risk-control` → `/strategy-validate-run` → `/strategy-backtest-analysis` | Researcher + Reviewer + Research Risk + Validator |
| **`/quant-orchestrate`** | 主链各步 + 上述 Loop | 全链（Factor 除外除非 Brief 指定） |

---

## 4. Skill 清单（13 个 slash-command）

建议目录（每个 Skill 一套）：

`quant-platform/agents/<skill-name>/SKILL.md`  
`quant-platform/contracts/<skill-name>.manifest.desc`  
`quant-platform/contracts/<skill-name>.validate.sh`  
`quant-platform/contracts/<skill-name>.invoke.example.sh`（可选）

### 4.1 编排类（3）

| # | Skill 名 | 命令 | 类型 | 主要输出 |
|---|----------|------|------|----------|
| 0 | 任务协调 | `/quant-orchestrate` | Loop | `runs/<run_id>/orchestrator.log`、`handoff.json` |
| 5 | 策略代码审查闭环 | `/strategy-code-loop` | Loop | `strategy-code-loop.manifest.result.json` + `code-loop/` |
| 9 | 策略研究外环闭环 | `/strategy-backtest-analysis-loop` | Loop | `strategy-backtest-analysis-loop.manifest.result.json` + `backtest-loop/` |

### 4.2 原子类（10）

| # | Skill 名 | 命令 | 推荐模型 / 执行体 | 所属 Agent 角色 | 主要输出 |
|---|----------|------|-------------------|-----------------|----------|
| 1 | 市场数据锚点 | `/market-data-profile` | 规则/脚本为主；缺口语化可用 Claude | Data Anchor | `data/MARKET_DATA_PROFILE.md` + manifest |
| 2 | 假设与方向挖掘 | `/strategy-hypothesis` | Claude | Researcher | `STRATEGY_HYPOTHESIS.md`、`hypotheses.json` + manifest |
| 3 | 策略代码（单轮） | `/strategy-code` | Claude Agent SDK | Researcher | 插件 `.py` + `IMPLEMENTATION_NOTES.md` + manifest |
| 4 | 代码审查（单轮） | `/strategy-review` | **与 Code 异模型** | Reviewer | `review.manifest.result.json` |
| 6 | 研究风控（单轮） | `/strategy-risk-control` | **与 Review 异模型/会话** | Research Risk | `STRATEGY_RISK_REVIEW.md` + manifest |
| 7 | 回测与验证执行 | `/strategy-validate-run` | **固件为主**（Skill 只触发 Tool API） | Validator | `metrics.json`、`gate.json` + manifest |
| 8 | 回测结果分析（单轮） | `/strategy-backtest-analysis` | Claude | Validator | `BACKTEST_ANALYSIS.md` + manifest |
| 10 | 因子挖掘 | `/factor-mining` | Claude + 检索 | Factor | `factor_library/…` + manifest |
| 11 | 策略研究报告 | `/strategy-report` | Claude + 固件模板 | Reporting | `STRATEGY_REPORT.html` + manifest |
| 12 | 虚拟盘观察（可选） | `/virtual-paper-watch` | 规则为主 | Reporting | `VIRTUAL_PAPER_WATCH.md` + manifest |

**禁止边界（各 Skill 共性摘要）：** 见主文档 §4 原表「禁止」列及 §6 契约；编排类 **禁止** 代替子 Skill 改 gate、手算净值、直连改数据中心。

---

## 5. 主链与 Agent 的对应（速查）

```text
（侧链，可选）
  Factor Agent:  /factor-mining  →  factor_library/

主链（Orchestrator 驱动）:
  Data Anchor:     /market-data-profile
  Researcher:      /strategy-hypothesis
  Orchestrator:    /quant-orchestrate 选定 hypothesis_id
       Researcher+Reviewer:     /strategy-code-loop
       Research Risk:           /strategy-risk-control
       Validator:               /strategy-validate-run → /strategy-backtest-analysis
       Researcher+…（外环）:    /strategy-backtest-analysis-loop（不达标时）
       Reporting:               /strategy-report → 可选 /virtual-paper-watch
```

完整顺序、Phase 0–3、失败路由：**[`quant-platform-agents.md` §5、§5.1](./quant-platform-agents.md)**。

---

## 6. 实现阶段与 Skill 数量（与主文档一致）

| 阶段 | 建议先具备的 Skill 数 | 包含 |
|------|------------------------|------|
| **Phase 0（MVP）** | 4 + 编排脚本 | `/strategy-code`、`/strategy-review`、`/strategy-code-loop` + 简易 orchestrator |
| **Phase 1（全自动主链）** | +6（累计约 10 原子中 6 个 + orchestrate） | 数据锚点、假设、风控、validate-run、backtest-analysis、**`/quant-orchestrate`** |
| **Phase 2** | +1 | **`/factor-mining`** |
| **Phase 3** | +2 | **`/strategy-report`**、可选 **`/virtual-paper-watch`** |

（Loop **`/strategy-backtest-analysis-loop`** 属 Phase 1 外环，与上表「主链全自动」一并启用。）

---

## 7. 变更记录

| 日期 | 说明 |
|------|------|
| 2026-10-23 | 从 `quant-platform-agents.md` §4 拆出；区分 **Agent 角色** 与 **13 个 Skill** |
