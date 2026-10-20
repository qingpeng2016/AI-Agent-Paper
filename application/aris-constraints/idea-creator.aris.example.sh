#!/usr/bin/env bash
# Step 2 — run from project root: bash aris-constraints/idea-creator.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/idea-creator "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/idea-creator.manifest.result.json per aris-constraints/idea-creator.manifest.example.desc and aris-constraints/idea-creator.manifest.example.json. Composed reports: fill §3 in idea-stage/IDEA_REPORT.en.md and idea-stage/IDEA_REPORT.zh.md (see aris-constraints/md_template/IDEA_REPORT.en.example.md and aris-constraints/md_template/IDEA_REPORT.zh.example.md; ### Idea {rank}: {title} byte-identical in both files)." — paper library: literature/ — language: en — REVIEWER_MODEL: gpt-5.4 — PILOT_MAX_HOURS: 2 — PILOT_TIMEOUT_HOURS: 3 — MAX_PILOT_IDEAS: 3 — MAX_TOTAL_GPU_HOURS: 8 — OUTPUT_DIR: idea-stage/ — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
