#!/usr/bin/env bash
# Step 4 — contract: aris-constraints/research-review.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-review "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/research-review.manifest.example.desc only." — REVIEWER_MODEL: gpt-5.4 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
