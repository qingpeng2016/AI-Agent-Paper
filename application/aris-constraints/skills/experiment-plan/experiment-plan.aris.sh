#!/usr/bin/env bash
# /experiment-plan — contract: aris-constraints/skills/experiment-plan/experiment-plan.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/experiment-plan "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/skills/experiment-plan/experiment-plan.desc only."'
