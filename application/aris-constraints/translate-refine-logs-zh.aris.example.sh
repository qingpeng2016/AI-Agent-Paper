#!/usr/bin/env bash
# Optional — from project root: bash aris-constraints/translate-refine-logs-zh.aris.example.sh [root]
# Writes Simplified Chinese copies under refine-logs/*.zh.md (does not replace English sources).
set -euo pipefail
cd "${1:-.}"

exec aris prompt "$(cat <<'EOF'
请阅读 refine-logs/EXPERIMENT_PLAN.md、refine-logs/EXPERIMENT_TRACKER.md、refine-logs/FINAL_PROPOSAL.md、refine-logs/PIPELINE_SUMMARY.md。若某文件不在 refine-logs/ 而在项目根目录，则读根目录同名文件。

将每份全文翻译为简体中文学术文体：保留 Markdown 标题层级与表格；专业术语首次出现用「中文（English）」；统计/方法名（AUC、FC-2、GRAPH、HAM-D 等）保留英文缩写；不要删改 claim 编号与 Phase 结构。

分别写入：
- refine-logs/EXPERIMENT_PLAN.zh.md
- refine-logs/EXPERIMENT_TRACKER.zh.md
- refine-logs/FINAL_PROPOSAL.zh.md
- refine-logs/PIPELINE_SUMMARY.zh.md

不要编造未出现在原文的内容。 — language: cn
EOF
)"
