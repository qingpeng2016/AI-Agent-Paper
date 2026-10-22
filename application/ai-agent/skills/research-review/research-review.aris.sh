#!/usr/bin/env bash
# Step 4 — contract: aris-constraints/skills/research-review/research-review.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-review "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/skills/research-review/research-review.desc only." — REVIEWER_MODEL: gpt-5.4 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
