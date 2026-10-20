#!/usr/bin/env bash
# Step 5 — run from project root: bash aris-constraints/research-refine-pipeline.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-refine-pipeline "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/research-refine-pipeline.manifest.result.json and all Step 5 outputs per aris-constraints/research-refine-pipeline.manifest.example.desc and aris-constraints/research-refine-pipeline.manifest.example.json. Composed: fill §1 in idea-stage/IDEA_REPORT.en.md and idea-stage/IDEA_REPORT.zh.md (rank-1)." — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
