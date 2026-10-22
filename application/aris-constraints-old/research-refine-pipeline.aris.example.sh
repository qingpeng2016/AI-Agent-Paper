#!/usr/bin/env bash
# Step 5 — contract: aris-constraints/research-refine-pipeline.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-refine-pipeline "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/research-refine-pipeline.manifest.example.desc only." — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
