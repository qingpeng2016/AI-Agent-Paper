#!/usr/bin/env bash
# Step 3 — contract: aris-constraints/novelty-check.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/novelty-check "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/novelty-check.manifest.example.desc only." — REVIEWER_MODEL: gpt-5.4 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
