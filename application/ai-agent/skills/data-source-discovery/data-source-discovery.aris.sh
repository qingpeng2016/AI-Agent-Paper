#!/usr/bin/env bash
# /data-source-discovery — contract: aris-constraints/skills/data-source-discovery/data-source-discovery.desc
# Invoke: Claude CLI (not aris). Skill bundle: aris-constraints/skills/data-source-discovery/
# Run from project root: bash aris-constraints/skills/data-source-discovery/data-source-discovery.aris.sh [project_root]
set -euo pipefail
cd "${1:-.}"

PROMPT='Run Paper skill /data-source-discovery.

Read RESEARCH_BRIEF.md. mkdir -p data-source if missing.
Run aris-constraints/skills/data-source-discovery/SKILL.md Phases 0–5.

DELIVERY: Follow aris-constraints/skills/data-source-discovery/data-source-discovery.desc only — ALL outputs under data-source/.
Profiles: READ aris-constraints/md_template/DATA_SOURCE_PROFILE.*.md ONLY for shape; write data-source/DATA_SOURCE_PROFILE.en.md and .zh.md with NO skill/manifest prose under title (see template line 1–6).
Min 1 max 3 ### 2.x dataset blocks; aim for 3.

Discovery sources: use web retrieval (and openalex / external search if available in your tools); do not full cohort download (probe_download=false — minimal probe only if user overrides).

MUST run: bash aris-constraints/skills/data-source-discovery/data-source-discovery.validate.sh before marking manifest result success.'

exec claude -p "$PROMPT"
