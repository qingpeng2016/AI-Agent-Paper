#!/usr/bin/env bash
# Step 3 — run from project root: bash aris-constraints/novelty-check.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/novelty-check "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/novelty-check.manifest.result.json per aris-constraints/novelty-check.manifest.example.desc and aris-constraints/novelty-check.manifest.example.json." — REVIEWER_MODEL: gpt-5.4 — composed: idea-stage/IDEA_REPORT.md'
