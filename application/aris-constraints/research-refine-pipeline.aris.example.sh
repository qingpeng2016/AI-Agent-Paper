#!/usr/bin/env bash
# Step 5 — run from project root: bash aris-constraints/research-refine-pipeline.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-refine-pipeline "Read RESEARCH_BRIEF.md. DELIVERY (paths are mandatory): (1) Manifest JSON at idea-stage/research-refine-pipeline.manifest.result.json — NOT under refine-logs/. (2) refine-logs/FINAL_PROPOSAL.md, EXPERIMENT_PLAN.md, EXPERIMENT_TRACKER.md per aris-constraints/research-refine-pipeline.manifest.example.desc, .json, and research-refine-pipeline.explain.example.desc (EXPERIMENT_PLAN ch.1: data access). (3) Fill §1 in BOTH idea-stage/IDEA_REPORT.en.md (## Executive Summary) AND idea-stage/IDEA_REPORT.zh.md (## 执行摘要): replace all placeholders including [Step 5: / [第 5 步： lines; do NOT strip report title/TOC or edit §2–§6. Rank-1 title from idea-creator.manifest.result.json." — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
