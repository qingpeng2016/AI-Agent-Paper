#!/usr/bin/env bash
# /research-refine — contract: aris-constraints/skills/research-refine/research-refine.desc
# Process: upstream research-refine/SKILL.md (Phases 0–5). Paper .desc adds paths + manifest only.
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/research-refine "Read RESEARCH_BRIEF.md from project root. DELIVERY: Follow aris-constraints/skills/research-refine/research-refine.desc only — standalone /research-refine, NOT /research-refine-pipeline. Write ALL outputs under project-root refine-logs/ (never idea-stage/refine-logs/). Run skills/research-refine/SKILL.md Phases 0–5; hard cap 5 review rounds; Phase 5 before manifest success. If refine-logs/FINAL_PROPOSAL.md missing, execute the workflow — do NOT answer from aris memory or old manifest alone. aris_phase in manifest must be research-refine. Resume refine-logs/REFINE_STATE.json if in_progress." — REVIEWER_MODEL: gpt-5.4'
