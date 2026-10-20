#!/usr/bin/env bash
# Optional — from project root: bash aris-constraints/translate-refine-logs-zh.aris.example.sh [root]
# Writes Simplified Chinese copies under refine-logs/*.zh.md (does not replace English sources).
set -euo pipefail
cd "${1:-.}"

exec aris prompt "$(cat <<'EOF'
请阅读 refine-logs/EXPERIMENT_PLAN.md、refine-logs/EXPERIMENT_TRACKER.md、refine-logs/FINAL_PROPOSAL.md、refine-logs/PIPELINE_SUMMARY.md，分别译为简体中文，写入 refine-logs/EXPERIMENT_PLAN.zh.md、refine-logs/EXPERIMENT_TRACKER.zh.md、refine-logs/FINAL_PROPOSAL.zh.md、refine-logs/PIPELINE_SUMMARY.zh.md。不要修改英文原版。 — language: cn
EOF
)"
