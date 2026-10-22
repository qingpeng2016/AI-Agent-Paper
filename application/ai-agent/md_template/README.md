# Markdown templates (idea discovery)

Steps 1–5: each step’s contract lives in **`aris-constraints/skills/<step-name>/<step-name>.desc`** (field tables + **Example JSON** + report rules). The co-located **`*.aris.sh`** loads only that `.desc`.

## Naming (do not mix)

| What | Path |
|------|------|
| **English composed report** | `idea-stage/IDEA_REPORT.en.md` |
| **Chinese composed report** | `idea-stage/IDEA_REPORT.zh.md` |
| English scaffold copied at init | from `md_template/IDEA_REPORT.en.md` → `idea-stage/IDEA_REPORT.en.md` |
| Chinese scaffold copied at init | from `md_template/IDEA_REPORT.zh.md` → `idea-stage/IDEA_REPORT.zh.md` |

| Template (`md_template/`) | Output path | Language |
|------|-------------|----------|
| `IDEA_REPORT.en.md` | `idea-stage/IDEA_REPORT.en.md` | English (contract / validate) |
| `IDEA_REPORT.zh.md` | `idea-stage/IDEA_REPORT.zh.md` | Simplified Chinese |
| `EXPERIMENT_PLAN.en.md` | `refine-logs/EXPERIMENT_PLAN.en.md` (`/experiment-plan`) | Claim Map, 实验数据获取, Experiment Blocks, milestones — see **`skills/experiment-plan/experiment-plan.desc`** |
| `EXPERIMENT_PLAN.zh.md` | `refine-logs/EXPERIMENT_PLAN.zh.md` | 与 EN 同结构；中文专业术语 |
| `EXPERIMENT_TRACKER.en.md` | `refine-logs/EXPERIMENT_TRACKER.en.md` | Run execution table + optional gates/budget |
| `EXPERIMENT_TRACKER.zh.md` | `refine-logs/EXPERIMENT_TRACKER.zh.md` | 与 EN 相同 Run ID / Status |
| `DATA_SOURCE_PROFILE.en.md` | `data-source/DATA_SOURCE_PROFILE.en.md` | `/data-source-discovery` |
| `DATA_SOURCE_PROFILE.zh.md` | `data-source/DATA_SOURCE_PROFILE.zh.md` | 与 EN 同结构 |

Step 5 (`research-refine-pipeline`) writes **`refine-logs/EXPERIMENT_PLAN.md`** (single EN file) — layout in **`skills/research-refine-pipeline/research-refine-pipeline.explain.desc`**, not the `/experiment-plan` templates above.

Standalone **`/research-refine`**: **`skills/research-refine/research-refine.desc`** — **`refine-logs/FINAL_PROPOSAL.md`**, **`REVIEW_SUMMARY.md`**, **`REFINEMENT_REPORT.md`**, **`REFINE_STATE.json`**.

Standalone **`/experiment-plan`**: **`skills/experiment-plan/experiment-plan.desc`** — four deliverables under **`refine-logs/`** using the four `EXPERIMENT_*` templates in this folder. Recommended after **`/research-refine`** when `FINAL_PROPOSAL.md` is ready.

Standalone **`/data-source-discovery`**: contract + launcher + validate under **`aris-constraints/skills/data-source-discovery/`** (`data-source-discovery.desc`, `*.aris.sh`, `*.validate.sh`) — all **outputs** under **`data-source/`** (profiles + **`data-source-discovery.manifest.result.json`**). **Clone** **`DATA_SOURCE_PROFILE.*.md`** (v2: **1–3** **`### 2.x`**, **prefer filling toward 3** alternates; no workflow prose in deliverables). Recommended **before** **`/research-lit`** when the paper depends on public downloadable data.

Placeholder style matches platform narrative templates (e.g. Workflow 3): **`[Fill-in instruction]`** on its own line or after a label — not sample content to keep.

**目录格式（中文层级列表，带 `#sec-*` 链接）** 见 `IDEA_REPORT.zh.md` 文首 `## 目录`（**6 章**：执行摘要、文献综述、推荐研究方案、新颖性评估、外部审稿意见、参考文献；§2 含 2.1–2.4，§4–5 含子节）。中文用语以模板为准（学术报告体，非直译）。英文模版 `IDEA_REPORT.en.md` 为同一 `sec-*` 锚点的 English TOC。

Same `sec-*` anchor ids in EN/ZH files.

### Contract H2 (## level)

| § | English (`IDEA_REPORT.en.md`) | Chinese (`IDEA_REPORT.zh.md`) |
|---|-------------------------------|-------------------------------|
| 1 | `## Executive Summary` | `## 执行摘要` |
| 2 | `## Literature Landscape` | `## 文献综述` |
| 3 | `## Recommended Ideas (ranked)` | `## 推荐研究方案` |
| 4 | `## Novelty Verification` | `## 新颖性评估` |
| 5 | `## External Critical Review` | `## 外部审稿意见` |

Step validators check the matching H2 per file. `### Idea n: title` lines stay **English title bytes** in both files.

Steps 1–5: compose **English** report from the EN template (or merge sections incrementally).  
Step 1 also composes matching **§2 and §6** in the Chinese file. Later steps may use `translate-idea-report-full-zh.aris.sh` where needed.

### Scaffold vs deliverable

Shipped under `aris-constraints/md_template/` (copied into each project at init). Init seeds `idea-stage/IDEA_REPORT.en.md` and `IDEA_REPORT.zh.md` from those files. Working copies start as **scaffolds**:

| In template | Meaning for agents |
|-------------|-------------------|
| First lines after `# Title` | Workflow instructions (like official “Template for … Fill in all sections”) — may stay or be trimmed; not part of report body. |
| Lines wholly or mainly `[…]` | **Replace** with real content when that step runs; do **not** ship bracket placeholders in finished sections. |
| `### Idea 1: [Title …]` | Contract shape — deliver as `### Idea 1: <manifest title bytes>`. |

§6 must list **only** papers from `research-lit.manifest.result.json`. Partial steps (e.g. Step 1) must replace **entire** §2 and §6 scaffolds in both EN and ZH even if other chapters still contain `[…]` until later steps.
