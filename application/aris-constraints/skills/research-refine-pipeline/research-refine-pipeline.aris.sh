#!/usr/bin/env bash
# Step 5 — contract: aris-constraints/skills/research-refine-pipeline/research-refine-pipeline.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-refine-pipeline "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/skills/research-refine-pipeline/research-refine-pipeline.desc only." — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
