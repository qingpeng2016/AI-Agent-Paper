#!/usr/bin/env bash
# Step 5 — contract: aris-constraints/research-refine-pipeline.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-refine-pipeline "Read RESEARCH_BRIEF.md. If data-source/DATA_SOURCE_PROFILE.en.md exists, Step 5 plan §1 and phases must use that dataset (data-anchor.requirement.example.desc); set data_anchor. DELIVERY: research-refine-pipeline.manifest.example.desc only." — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
