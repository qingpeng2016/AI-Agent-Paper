#!/usr/bin/env bash
# Step 3 — run from project root: bash aris-constraints/novelty-check.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/novelty-check "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/novelty-check.manifest.result.json per aris-constraints/novelty-check.manifest.example.desc and aris-constraints/novelty-check.manifest.example.json. Composed reports: fill §4 in idea-stage/IDEA_REPORT.en.md and idea-stage/IDEA_REPORT.zh.md (see aris-constraints/md_template/IDEA_REPORT.en.example.md and aris-constraints/md_template/IDEA_REPORT.zh.example.md)." — REVIEWER_MODEL: gpt-5.4 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
