#!/usr/bin/env bash
# Run all step validators — from project root: bash aris-constraints/validate-idea-discovery.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
for s in research-lit idea-creator novelty-check research-review research-refine-pipeline; do
  bash "$DIR/${s}.validate.example.sh" "$ROOT"
done
echo "ALL STEPS OK"
