#!/usr/bin/env bash
# /data-source-discovery — contract: aris-constraints/data-source-discovery.manifest.example.desc
# Skill: aris-constraints/skills/data-source-discovery/SKILL.md (mirror: project/skills/ after init)
# Run from project root: bash aris-constraints/data-source-discovery.aris.example.sh [project_root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/data-source-discovery "Read RESEARCH_BRIEF.md. mkdir -p data-source if missing. Run aris-constraints/skills/data-source-discovery/SKILL.md Phases 0–5. DELIVERY: aris-constraints/data-source-discovery.manifest.example.desc — ALL under data-source/. Profiles: READ aris-constraints/md_template/DATA_SOURCE_PROFILE.*.example.md ONLY for shape; write data-source/DATA_SOURCE_PROFILE.*.md with NO skill/manifest prose under title (see template line 1–6). Min 1 max 3 ### 2.x; aim for 3. MUST run validate before success." — sources: web, gemini, openalex — probe download: false'
