#!/usr/bin/env bash
# Optional — from project root: bash aris-constraints/translate-refine-logs-zh.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"

fail() { echo "TRANSLATE-ZH FAIL: $*" >&2; exit 1; }

for f in \
  refine-logs/EXPERIMENT_PLAN.zh.md \
  refine-logs/EXPERIMENT_TRACKER.zh.md \
  refine-logs/FINAL_PROPOSAL.zh.md \
  refine-logs/PIPELINE_SUMMARY.zh.md
do
  [[ -f "$f" ]] || fail "$f missing"
  [[ -s "$f" ]] || fail "$f empty"
done

echo "TRANSLATE-ZH OK"
