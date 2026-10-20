#!/usr/bin/env bash
# Step 4 — run from project root: bash aris-constraints/research-review.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-review "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/research-review.manifest.result.json per aris-constraints/research-review.manifest.example.desc and aris-constraints/research-review.manifest.example.json. Composed reports: fill §5 in idea-stage/IDEA_REPORT.en.md and idea-stage/IDEA_REPORT.zh.md (see aris-constraints/md_template/IDEA_REPORT.en.example.md and aris-constraints/md_template/IDEA_REPORT.zh.example.md; rank-1 idea only)." — REVIEWER_MODEL: gpt-5.4 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
