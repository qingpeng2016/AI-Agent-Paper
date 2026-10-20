#!/usr/bin/env bash
# Step 5 — from project root: bash aris-constraints/research-refine-pipeline.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.md"

fail() { echo "STEP5 FAIL: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail "$REPORT missing"
[[ -f idea-stage/research-refine-pipeline.manifest.result.json ]] || fail "research-refine-pipeline.manifest.result.json missing"
[[ -f idea-stage/idea-creator.manifest.result.json ]] || fail "idea-creator.manifest.result.json missing"

if ! grep -q '^## Executive Summary' "$REPORT"; then
  fail "$REPORT has no line starting with: ## Executive Summary"
fi

python3 <<'PY'
import json, os, sys

def fail(msg):
    print(f"STEP5 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

try:
    ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
except Exception as e:
    fail(f"idea-creator.manifest.result.json invalid: {e}")

try:
    rf = json.load(open("idea-stage/research-refine-pipeline.manifest.result.json"))
except Exception as e:
    fail(f"research-refine-pipeline.manifest.result.json invalid: {e}")

if rf.get("result") != "success":
    fail(f'manifest result={rf.get("result")!r}, need "success"')

ideas = ic.get("ideas") or []
r1_list = [x for x in ideas if x.get("rank") == 1]
if not r1_list:
    fail("idea-creator manifest has no idea with rank 1")
r1 = r1_list[0]

if rf.get("idea_title") != r1["title"]:
    fail(
        f"idea_title mismatch:\n"
        f"  idea-creator rank 1: {r1['title']!r}\n"
        f"  refine manifest:       {rf.get('idea_title')!r}"
    )

if not (rf.get("executive_summary") or "").strip():
    fail("executive_summary empty")

d = rf.get("deliverables") or {}
for k in ("final_proposal", "experiment_plan", "experiment_tracker"):
    p = d.get(k)
    if not p:
        fail(f"deliverables.{k} path missing")
    if not os.path.isfile(p):
        fail(f"deliverables.{k} not a file: {p!r}")
    if os.path.getsize(p) <= 0:
        fail(f"deliverables.{k} empty file: {p!r}")

print("STEP5 OK")
PY
