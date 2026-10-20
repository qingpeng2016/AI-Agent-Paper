#!/usr/bin/env bash
# Step 2 — from project root: bash aris-constraints/idea-creator.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.md"
MANIFEST="idea-stage/idea-creator.manifest.result.json"

fail() { echo "STEP2 FAIL: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail "$REPORT missing"
[[ -f "$MANIFEST" ]] || fail "$MANIFEST missing"

if ! grep -q '^## Recommended Ideas (ranked)' "$REPORT"; then
  fail "$REPORT has no line starting with: ## Recommended Ideas (ranked) (hint: grep -n 'Recommended Ideas' $REPORT)"
fi

python3 <<'PY'
import json, sys

def fail(msg):
    print(f"STEP2 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

try:
    m = json.load(open("idea-stage/idea-creator.manifest.result.json"))
except Exception as e:
    fail(f"idea-creator.manifest.result.json invalid: {e}")

if m.get("result") != "success":
    fail(f'manifest result={m.get("result")!r}, need "success"')

ideas = m.get("ideas") or []
if not ideas:
    fail("manifest ideas[] empty, need at least 1")

ev = m.get("ideas_evaluated") or {}
rec = ev.get("recommended")
if rec is not None and rec != len(ideas):
    fail(f"ideas_evaluated.recommended={rec!r} but ideas.length={len(ideas)} (must match if set)")

ranks = [i.get("rank") for i in ideas]
want = list(range(1, len(ideas) + 1))
if sorted(ranks) != want:
    fail(f"ideas[].rank must be exactly {want}, got {ranks}")

body = open("idea-stage/IDEA_REPORT.md").read()
for i in ideas:
    line = f"### Idea {i['rank']}: {i['title']}"
    if line not in body:
        fail(f"IDEA_REPORT missing exact line: {line!r}")

print("STEP2 OK")
PY
