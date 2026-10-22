#!/usr/bin/env bash
# Step 1 — contract: aris-constraints/skills/research-lit/research-lit.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-lit "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/skills/research-lit/research-lit.desc only." — language: en — paper library: literature/ — sources: all, gemini — ARXIV_DOWNLOAD: false — ARXIV_MAX_DOWNLOAD: 5 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
