#!/usr/bin/env bash
# /data-source-discovery — contract: aris-constraints/data-source-discovery.manifest.example.desc
# Skill: aris-constraints/skills/data-source-discovery/SKILL.md (mirror: project/skills/ after init)
# Run from project root: bash aris-constraints/data-source-discovery.aris.example.sh [project_root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/data-source-discovery "Read RESEARCH_BRIEF.md. mkdir -p data-source if missing. Run aris-constraints/skills/data-source-discovery/SKILL.md Phases 0–5. DELIVERY: aris-constraints/data-source-discovery.manifest.example.desc — ALL under data-source/. Profiles: READ md_template/DATA_SOURCE_PROFILE.*.example.md; CLONE v2: min 1 max 3 datasets in profile (### 2.1–2.3 + 2.x.1–2.x.4); aim for 3 ranked entries when search allows; datasets[] matches block count; explain in executive_summary if fewer than 3. Directory tree under #### 2.1.2 目录结构. Run data-source-discovery.validate.example.sh before result success." — sources: web, gemini, openalex — probe download: false'
