#!/usr/bin/env bash
# Step 4 — contract: aris-constraints/research-review.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-review "Read RESEARCH_BRIEF.md. If data-source/DATA_SOURCE_PROFILE.en.md exists, review rank-1 against that dataset (data-anchor.requirement.example.desc); set data_anchor. DELIVERY: research-review.manifest.example.desc only." — REVIEWER_MODEL: gpt-5.4 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
