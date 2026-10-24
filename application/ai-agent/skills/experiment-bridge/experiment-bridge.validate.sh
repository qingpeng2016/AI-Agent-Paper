#!/usr/bin/env bash
# /experiment-bridge — from project root: bash aris-constraints/skills/experiment-bridge/experiment-bridge.validate.sh [root]
# contract-id: experiment-bridge-validate-2026-10-22
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
CONTRACT_TAG="experiment-bridge-validate-2026-10-22"

fail_en() { echo "EXPERIMENT-BRIDGE FAIL [$CONTRACT_TAG]: $*" >&2; exit 1; }

MANIFEST="idea-stage/experiment-bridge.manifest.result.json"
[[ -f "$MANIFEST" ]] || fail_en "$MANIFEST missing"
if [[ -f refine-logs/experiment-bridge.manifest.result.json ]]; then
  fail_en "manifest is under refine-logs/; use idea-stage/experiment-bridge.manifest.result.json"
fi

python3 <<'PY'
import json, os, sys

CONTRACT = "experiment-bridge-validate-2026-10-22"

def fail_en(msg):
    print(f"EXPERIMENT-BRIDGE FAIL [{CONTRACT}]: {msg}", file=sys.stderr)
    sys.exit(1)

mf = json.load(open("idea-stage/experiment-bridge.manifest.result.json", encoding="utf-8"))
if mf.get("aris_phase") != "experiment-bridge":
    fail_en(f'aris_phase must be "experiment-bridge", got {mf.get("aris_phase")!r}')

result = mf.get("result")
if result not in ("success", "in_progress", "failed"):
    fail_en(f'result must be success|in_progress|failed, got {result!r}')

state_path = "refine-logs/BRIDGE_STATE.json"
deliverables = mf.get("deliverables") or {}
results_md = deliverables.get("initial_results_md", "refine-logs/EXPERIMENT_RESULTS.md")

if result == "success":
    if not os.path.isfile(state_path):
        fail_en(f"{state_path} missing for result success")
    st = json.load(open(state_path, encoding="utf-8"))
    if st.get("status") not in ("completed", "done"):
        fail_en(f'BRIDGE_STATE.status must be completed for success, got {st.get("status")!r}')
    if not os.path.isfile(results_md) or os.path.getsize(results_md) < 100:
        fail_en(f"initial results missing or too short: {results_md}")
    tr = mf.get("experiment_tracker_en") or "refine-logs/EXPERIMENT_TRACKER.en.md"
    if not os.path.isfile(tr):
        fail_en(f"tracker missing: {tr}")
elif result == "in_progress":
    if not os.path.isfile(state_path):
        fail_en(f"{state_path} required when result is in_progress")
    st = json.load(open(state_path, encoding="utf-8"))
    if st.get("status") not in ("in_progress", "awaiting_deploy_approval"):
        fail_en(f'BRIDGE_STATE.status must be in_progress or awaiting_deploy_approval, got {st.get("status")!r}')
    phase = st.get("current_phase") or mf.get("current_phase")
    if not phase:
        fail_en("current_phase missing in BRIDGE_STATE or manifest")

print(f"EXPERIMENT-BRIDGE OK [{CONTRACT}] result={result}")
PY
