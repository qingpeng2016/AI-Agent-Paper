#!/usr/bin/env bash
# /research-refine — contract: aris-constraints/research-refine.manifest.example.desc
# Process: upstream research-refine/SKILL.md (Phases 0–5). Paper .desc adds paths + manifest only.
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-refine "Read RESEARCH_BRIEF.md. If data-source/DATA_SOURCE_PROFILE.en.md exists, refine rank-1 method for that dataset (data-anchor.requirement.example.desc); set data_anchor. DELIVERY: research-refine.manifest.example.desc — standalone /research-refine; outputs under refine-logs/. Phase 5 before manifest success." — REVIEWER_MODEL: gpt-5.4'
