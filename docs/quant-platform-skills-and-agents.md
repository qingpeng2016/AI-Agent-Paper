# 量化平台 — Skill 与 Agent 对照

## 1. Skill 清单（共 14）

| 序号 | 命令 | 类型 | 职责（管什么） | 主要输出 | 链路上位置 |
|------|------|------|----------------|----------|------------|
| 1 | `/quant-orchestrate` | 编排 | 读 Brief；选 `hypothesis_id` 开 trial；派单、失败路由、Token/family 上限 | `orchestrator.log`、`handoff.json` | 主链中枢 |
| 2 | `/market-data-profile` | 原子 | 在数据中心绑定 `dataset_id` + `snapshot_id`；写字段、区间、PIT、缺口 | `MARKET_DATA_PROFILE.md`、market-data manifest | 主链首段（Brief 后） |
| 3 | `/strategy-hypothesis` | 原子 | 写策略方向 MD + `hypotheses.json`；Trial Log / 因子库去重 | `STRATEGY_HYPOTHESIS.md`、`hypotheses.json` | 主链；外环每轮 |
| 4 | `/strategy-code` | 原子 | 单轮：**写 Strategy 策略代码** + `IMPLEMENTATION_NOTES.md` | 插件路径、code manifest | 仅由 code-loop 按轮调用（或人工单轮调试） |
| 5 | `/strategy-code-review` | 原子 | 单轮：**审核代码**（对照策略文档与实现说明）→ `approved` / `rejected` + `issues[]` | review manifest | 仅由 code-loop 按轮调用 |
| 6 | `/strategy-code-loop` | Loop | **循环编排—策略代码构建 + 策略代码审核** | code-loop 总 manifest、`code-loop/round-n/` | 主链（每个 trial） |
| 7 | `/strategy-risk-control` | 原子 | G2 通过后、回测前：Brief forbidden、集中度、尾部暴露等 **设计层** 风控 | `STRATEGY_RISK_REVIEW.md`、risk manifest | 主链（code-loop 与回测启动之间） |
| 8 | `/strategy-backtest-run` | 原子 | **回测启动**：触发固件取数 → 跑插件 → validation → Gate（G3–G8） | `metrics.json`、`gate.json` | 主链（风控通过后） |
| 9 | `/strategy-backtest-analysis` | 原子 | 对照 Brief + Gate + 文档 + 代码；写是否达标与原因 | `BACKTEST_ANALYSIS.md` | 主链；外环每轮末 |
| 10 | `/strategy-backtest-analysis-loop` | Loop | 不达标时多轮：#3 → #6 → #7 → #8 → #9 直至达标或上限 | backtest-loop 总 manifest、`backtest-loop/round-n/` | 主链（条件触发） |
| 11 | `/factor-mining` | 原子 | 挖掘/整理因子定义，写入因子库 | `factor_library/<id>.md`、`index.json` | **侧链**（供 #3 只读引用） |
| 12 | `/strategy-report` | 原子 | 达标后聚合 Brief、假设、metrics、Gate、分析 → 标准报告 | `STRATEGY_REPORT.html`（+ 可选 PDF） | 主链末段 |
| 13 | `/virtual-paper-watch` | 原子 | 人工批准虚拟盘后：对比虚拟盘 vs 回测，超阈告警 | `VIRTUAL_PAPER_WATCH.md` | 主链可选（上线观察） |
| 14 | `/strategy-live-rolling-backtest-compare` | 原子 | **同一策略版本**下：对比 **实盘在跑** vs **滚动/实时回测**；偏差过大且 **回测更优** → 写入 **待人工审核—部署最新代码** 队列 | `LIVE_VS_ROLLING_BACKTEST.md`、compare manifest、**`deployment_review_pending`** 记录 | **上线后**（实盘已挂载该 `strategy_version_id`） |

**类型：** **原子** = 单步交付；**Loop** = 多轮 invoke 其他 Skill 并汇总 manifest；**编排** = 状态机派单，不代替子 Skill 写码/回测。

**默认主链顺序（摘要）：**

```text
Brief → /market-data-profile → /strategy-hypothesis → /quant-orchestrate
  → /strategy-code-loop → /strategy-risk-control → /strategy-backtest-run
  → /strategy-backtest-analysis
       ├→ 达标 → /strategy-report →（人工虚拟盘/实盘审批）→ 可选 /virtual-paper-watch
       └→ 不达标 → /strategy-backtest-analysis-loop（每轮 #3→#6→#7→#8→#9）

上线后（后端已跑实盘 + 滚动回测）→ 定时或事件触发 /strategy-live-rolling-backtest-compare
       └→ 满足规则 → 待人工审核（部署替换最新策略代码）；**不自动下单、不自动换版**
```

### `/strategy-live-rolling-backtest-compare` 要点

| 项 | 说明 |
|----|------|
| **对齐键** | 同一 **`strategy_version_id`**（策略描述/契约一致）；实盘与滚动回测均须绑定该版本元数据 |
| **输入（后端 Tool API）** | 实盘滚动指标序列；同版本下的 **实时/滚动回测** 指标序列；可选「候选最新已通过审查的 plugin 版本」仅作 **部署建议**，不在此 Skill 内改 live |
| **触发条件（须同时）** | ① 偏差超过 `config/live_rolling_compare.yaml` 阈值；② **滚动回测表现优于实盘**（规则可配置：Sharpe、回撤、收益等） |
| **Skill 产出** | 对比报告 MD + manifest；**写入待人工审核队列**（部署：用 **最新策略代码** 替换当前实盘挂载） |
| **禁止** | 自动部署、自动换插件、自动停实盘；仅 **告警 + 入队** |
| **与后端分工** | **后端**：实盘执行、滚动回测 job、队列库表、人工审批 UI、批准后的 **发布/换版 API**；**本 Skill**：读数、比对、写报告与 **pending** 状态 |

---

## 2. Agent 角色与 Skill 映射（共 7 个逻辑角色）

| Agent 角色 | 承担的 Skill | 说明 |
|------------|--------------|------|
| **1 · 任务协调 Agent**（Orchestrator） | `/quant-orchestrate` | 派单、失败路由、Token/family；不写插件、不改 Gate |
| **2 · 数据锚点 Agent** | `/market-data-profile` | 绑定数据中心 snapshot；不联网拉数、不跑回测 |
| **3 · 策略研究 Agent**（Researcher） | `/strategy-hypothesis`、`/strategy-code`、`/strategy-code-review`、`/strategy-code-loop` | 假设 MD、写策略代码、**审核代码**、**内环** code-loop（构建+审码） |
| **4 · 代码审查 Agent**（Reviewer） | `/strategy-code-review` | 与 #3 **同一 Skill**、**异模型/异上下文** 执行审码；仅由 code-loop 按轮 invoke，不写码 |
| **5 · 验证与分析 Agent**（Validator） | `/strategy-backtest-run`、`/strategy-backtest-analysis`、`/strategy-backtest-analysis-loop` | **回测启动** → **结果分析**；不达标时 **外环 loop** 再编排整链（loop 内仍会 invoke 研究/风控等 Skill） |
| **6 · 因子与风控 Agent**（Factor） | `/factor-mining`、`/strategy-risk-control` | **侧链**：因子库；**主链**：G2 后、回测前研究风控，reject → 回 code-loop |
| **7 · 交付与运行对齐 Agent**（Reporting） | `/strategy-report`、`/virtual-paper-watch`（可选）、`/strategy-live-rolling-backtest-compare` | 研究报告；虚拟盘观察；**实盘 vs 滚动回测** 偏差 → **待人工部署审核** |




1. 我们需要能跑出要求的策略代码（比如达到某个年化率）

2. 但是验收标准，是什么需要讨论下

3. ？？ 可扩展性，以及fred做的对不对？