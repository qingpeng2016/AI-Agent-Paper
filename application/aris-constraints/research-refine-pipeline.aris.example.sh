#!/usr/bin/env bash
# Step 5 — run from project root: bash aris-constraints/research-refine-pipeline.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt "$(cat <<'EOF'
Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/research-refine-pipeline.manifest.result.json and deliverables per aris-constraints/research-refine-pipeline.manifest.example.desc, aris-constraints/research-refine-pipeline.manifest.example.json, and aris-constraints/research-refine-pipeline.explain.example.desc (EXPERIMENT_PLAN first chapter: Data access and workspace layout). Composed: idea-stage/IDEA_REPORT.md.

After each English markdown above exists, write a Simplified Chinese translation alongside (same directory, basename + .zh.md before .md): refine-logs/FINAL_PROPOSAL.zh.md, refine-logs/EXPERIMENT_PLAN.zh.md, refine-logs/EXPERIMENT_TRACKER.zh.md; if you write refine-logs/PIPELINE_SUMMARY.md also write refine-logs/PIPELINE_SUMMARY.zh.md; and idea-stage/IDEA_REPORT.zh.md for the full composed IDEA_REPORT.md. Do not modify English sources. — language: cn
EOF
)"
