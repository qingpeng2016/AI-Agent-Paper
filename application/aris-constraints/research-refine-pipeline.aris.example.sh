#!/usr/bin/env bash
# Step 5 — run from project root: bash aris-constraints/research-refine-pipeline.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-refine-pipeline "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/research-refine-pipeline.manifest.result.json and deliverables per aris-constraints/research-refine-pipeline.manifest.example.desc, aris-constraints/research-refine-pipeline.manifest.example.json, and aris-constraints/research-refine-pipeline.explain.example.desc (EXPERIMENT_PLAN first chapter: Data access and workspace layout)." — composed: idea-stage/IDEA_REPORT.md'
