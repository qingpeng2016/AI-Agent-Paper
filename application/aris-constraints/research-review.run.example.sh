#!/usr/bin/env bash
# Step 4 — run from project root: bash aris-constraints/research-review.run.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-review "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/research-review.manifest.result.json per aris-constraints/research-review.manifest.example.desc and aris-constraints/research-review.manifest.example.json." — REVIEWER_MODEL: gpt-6-sol — composed: idea-stage/IDEA_REPORT.md'
