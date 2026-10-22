#!/usr/bin/env bash
# /data-source-discovery — contract: aris-constraints/data-source-discovery.manifest.example.desc
# Skill: aris-constraints/skills/data-source-discovery/SKILL.md (mirror: project/skills/ after init)
# Run from project root: bash aris-constraints/data-source-discovery.aris.example.sh [project_root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/data-source-discovery "Read RESEARCH_BRIEF.md. mkdir -p data-source if missing. Run aris-constraints/skills/data-source-discovery/SKILL.md Phases 0–5. DELIVERY: Follow aris-constraints/data-source-discovery.manifest.example.desc only — ALL outputs under data-source/ (manifest + DATA_SOURCE_PROFILE.en.md + .zh.md). Verify free open download URLs and document BIDS/file structure for primary dataset." — sources: web, gemini, openalex — probe download: false'
