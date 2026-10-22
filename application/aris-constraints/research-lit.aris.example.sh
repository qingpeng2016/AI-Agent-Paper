#!/usr/bin/env bash
# Step 1 — contract: aris-constraints/research-lit.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-lit "Read RESEARCH_BRIEF.md. If data-source/DATA_SOURCE_PROFILE.en.md exists, constrain §2 to that dataset (data-anchor.requirement.example.desc); data_anchor must be JSON object with primary_dataset_id d1. DELIVERY: research-lit.manifest.example.desc — manifest ONLY idea-stage/research-lit.manifest.result.json (NOT aris-constraints/*.json); edit §2+§6 only in idea-stage/IDEA_REPORT.*.md; run research-lit.validate.example.sh before success." — language: en — paper library: literature/ — sources: all, gemini — ARXIV_DOWNLOAD: false — ARXIV_MAX_DOWNLOAD: 5 — composed: idea-stage/IDEA_REPORT.en.md,idea-stage/IDEA_REPORT.zh.md'
