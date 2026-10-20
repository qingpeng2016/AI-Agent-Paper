#!/usr/bin/env bash
# Step 2 — run from project root: bash aris-constraints/idea-creator.run.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/idea-creator "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/idea-creator.manifest.result.json per aris-constraints/idea-creator.manifest.example.desc and aris-constraints/idea-creator.manifest.example.json." — paper library: literature/ — language: en — REVIEWER_MODEL: gpt-6-sol — PILOT_MAX_HOURS: 2 — PILOT_TIMEOUT_HOURS: 3 — MAX_PILOT_IDEAS: 3 — MAX_TOTAL_GPU_HOURS: 8 — OUTPUT_DIR: idea-stage/ — composed: idea-stage/IDEA_REPORT.md'
