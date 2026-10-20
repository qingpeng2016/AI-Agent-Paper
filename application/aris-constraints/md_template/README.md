# Markdown templates (idea discovery)

Steps 1–5: **`aris-constraints/*.manifest.example.desc`** is the single contract (field tables + **Example JSON** + report rules). **`*.aris.example.sh`** loads only the `.desc`.

## Naming (do not mix)

| What | Path |
|------|------|
| **English composed report** | `idea-stage/IDEA_REPORT.en.md` |
| **Chinese composed report** | `idea-stage/IDEA_REPORT.zh.md` |
| English scaffold copied at init | from `IDEA_REPORT.en.example.md` → `IDEA_REPORT.en.md` |
| Chinese scaffold copied at init | from `IDEA_REPORT.zh.example.md` → `IDEA_REPORT.zh.md` |

| File | Output path | Language |
|------|-------------|----------|
| `IDEA_REPORT.en.example.md` | `idea-stage/IDEA_REPORT.en.md` | English (contract / validate) |
| `IDEA_REPORT.zh.example.md` | `idea-stage/IDEA_REPORT.zh.md` | Simplified Chinese |

Placeholder style matches platform narrative templates (e.g. Workflow 3): **`[Fill-in instruction]`** on its own line or after a label — not sample content to keep.

**目录格式（中文层级列表，带 `#sec-*` 链接）** 见 `IDEA_REPORT.zh.example.md` 文首 `## 目录`（**6 章**：执行摘要、文献综述、推荐研究方案、新颖性评估、外部审稿意见、参考文献；§2 含 2.1–2.4，§4–5 含子节）。中文用语以模板为准（学术报告体，非直译）。英文模版 `IDEA_REPORT.en.example.md` 为同一 `sec-*` 锚点的 English TOC。

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
Step 1 also composes matching **§2 and §6** in the Chinese file. Later steps may use `translate-idea-report-full-zh.aris.example.sh` where needed.

### Scaffold vs deliverable

Shipped under `aris-constraints/md_template/` (copied into each project at init). Init seeds `idea-stage/IDEA_REPORT.en.md` and `IDEA_REPORT.zh.md` from those files. Working copies start as **scaffolds**:

| In template | Meaning for agents |
|-------------|-------------------|
| First lines after `# Title` | Workflow instructions (like official “Template for … Fill in all sections”) — may stay or be trimmed; not part of report body. |
| Lines wholly or mainly `[…]` | **Replace** with real content when that step runs; do **not** ship bracket placeholders in finished sections. |
| `### Idea 1: [Title …]` | Contract shape — deliver as `### Idea 1: <manifest title bytes>`. |

§6 must list **only** papers from `research-lit.manifest.result.json`. Partial steps (e.g. Step 1) must replace **entire** §2 and §6 scaffolds in both EN and ZH even if other chapters still contain `[…]` until later steps.
