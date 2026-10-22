#!/usr/bin/env bash
# Step 2 — contract: aris-constraints/idea-creator.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/idea-creator "Read RESEARCH_BRIEF.md. If data-source/DATA_SOURCE_PROFILE.en.md exists, ideas must fit that dataset (data-anchor.requirement.example.desc); set data_anchor in manifest. DELIVERY: idea-creator.manifest.example.desc only." — paper library: literature/ — language: en — REVIEWER_MODEL: gpt-5.4 — PILOT_MAX_HOURS: 2 — PILOT_TIMEOUT_HOURS: 3 — MAX_PILOT_IDEAS: 3 — MAX_TOTAL_GPU_HOURS: 8 — OUTPUT_DIR: idea-stage/ — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
