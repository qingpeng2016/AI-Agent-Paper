# Markdown templates (idea discovery)

| File | Output path | Language |
|------|-------------|----------|
| `IDEA_REPORT.en.example.md` | `idea-stage/IDEA_REPORT.md` | English (contract / validate) |
| `IDEA_REPORT.zh.example.md` | `idea-stage/IDEA_REPORT.zh.md` | Simplified Chinese |

**目录格式（中文层级列表，带 `#sec-*` 链接）** 见 `IDEA_REPORT.zh.example.md` 文首 `## 目录`（**6 章**：执行摘要、文献综述、推荐研究方案、新颖性评估、外部审稿意见、参考文献；§2 含 2.1–2.4，§4–5 含子节）。中文用语以模板为准（学术报告体，非直译）。英文模版 `IDEA_REPORT.en.example.md` 为同一 `sec-*` 锚点的 English TOC。

Same `sec-*` anchor ids in EN/ZH files.

Steps 1–5: compose **English** report from the EN template (or merge sections incrementally).  
Translate to **Chinese** using the ZH template and `translate-idea-report-full-zh.aris.example.sh`.
