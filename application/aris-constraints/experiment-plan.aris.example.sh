#!/usr/bin/env bash
# /experiment-plan — contract: aris-constraints/experiment-plan.manifest.example.desc
set -euo pipefail
cd "${1:-.}"
exec aris prompt '/experiment-plan "Read RESEARCH_BRIEF.md. DELIVERY: Follow aris-constraints/experiment-plan.manifest.example.desc only."'
