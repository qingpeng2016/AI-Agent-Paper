#!/usr/bin/env bash
# /experiment-plan — contract: aris-constraints/experiment-plan.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/experiment-plan "Read RESEARCH_BRIEF.md. If data-source/DATA_SOURCE_PROFILE.en.md exists, acquisition section must follow that profile URLs/paths (data-anchor.requirement.example.desc); set data_anchor. DELIVERY: experiment-plan.manifest.example.desc only."'
