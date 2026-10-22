#!/usr/bin/env bash
# Step 3 — contract: aris-constraints/novelty-check.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/novelty-check "Read RESEARCH_BRIEF.md. If data-source/DATA_SOURCE_PROFILE.en.md exists, assess novelty on that locked dataset (data-anchor.requirement.example.desc); set data_anchor. DELIVERY: novelty-check.manifest.example.desc only." — REVIEWER_MODEL: gpt-5.4 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
