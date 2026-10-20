#!/usr/bin/env bash
# Step 1 — run from project root: bash aris-constraints/research-lit.aris.example.sh [root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-lit "Read RESEARCH_BRIEF.md. DELIVERY: Write idea-stage/research-lit.manifest.result.json per aris-constraints/research-lit.manifest.example.desc and aris-constraints/research-lit.manifest.example.json." — language: en — paper library: literature/ — sources: all, gemini — ARXIV_DOWNLOAD: false — ARXIV_MAX_DOWNLOAD: 5 — composed: idea-stage/IDEA_REPORT.md'
