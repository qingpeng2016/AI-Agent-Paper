#!/usr/bin/env bash
# /experiment-bridge — contract: aris-constraints/skills/experiment-bridge/experiment-bridge.desc
# Process: upstream experiment-bridge/SKILL.md (Phases 1–6). Paper .desc adds BRIDGE_STATE + resume.
# Run from project root: bash aris-constraints/skills/experiment-bridge/experiment-bridge.aris.sh [project_root]
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/experiment-bridge "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/skills/experiment-bridge/experiment-bridge.desc only. Run skills/experiment-bridge/SKILL.md Phases 1–6. Prefer refine-logs/EXPERIMENT_PLAN.en.md and refine-logs/EXPERIMENT_TRACKER.en.md. Paper data root: data-source/ only — read DATA_SOURCE_PROFILE.*.md and data-source-discovery.manifest.result.json; staged downloads under data-source/raw/; dataloaders must use paths aligned with plan acquisition section. Maintain refine-logs/BRIDGE_STATE.json; if status in_progress or manifest result in_progress RESUME from current_phase — do NOT restart parse_plan unless checkpoint missing. While GPU jobs run use /monitor-experiment until tracker must-runs DONE or upstream stop rules. result success only after Phase 6 and validate.sh pass. If AUTO_DEPLOY false set awaiting_deploy_approval until user approves deploy then continue same workflow."'
