#!/usr/bin/env bash
# Step 3 — from project root: bash aris-constraints/novelty-check.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.md"

fail() { echo "STEP3 FAIL: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail "$REPORT missing"
[[ -f idea-stage/novelty-check.manifest.result.json ]] || fail "novelty-check.manifest.result.json missing"
[[ -f idea-stage/idea-creator.manifest.result.json ]] || fail "idea-creator.manifest.result.json missing"

if ! grep -q '^## Novelty Verification' "$REPORT"; then
  fail "$REPORT has no line starting with: ## Novelty Verification"
fi

python3 <<'PY'
import json, sys

def fail(msg):
    print(f"STEP3 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
nc = json.load(open("idea-stage/novelty-check.manifest.result.json"))

if nc.get("result") != "success":
    fail(f'novelty manifest result={nc.get("result")!r}, need "success"')

ideas, checks = ic.get("ideas") or [], nc.get("checks") or []
if len(checks) != len(ideas):
    fail(f"checks.length={len(checks)} but idea-creator ideas.length={len(ideas)}")

by_rank = {c.get("rank"): c for c in checks}
for idea in ideas:
    r = idea["rank"]
    if r not in by_rank:
        fail(f"checks missing rank {r}")
    got = by_rank[r].get("title")
    want = idea["title"]
    if got != want:
        fail(f"rank {r} title mismatch:\n  idea-creator: {want!r}\n  novelty-check:  {got!r}")

print("STEP3 OK")
PY
