#!/usr/bin/env bash
# Step 4 — from project root: bash aris-constraints/research-review.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.md"

fail() { echo "STEP4 FAIL: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail "$REPORT missing"
[[ -f idea-stage/research-review.manifest.result.json ]] || fail "research-review.manifest.result.json missing"
[[ -f idea-stage/idea-creator.manifest.result.json ]] || fail "idea-creator.manifest.result.json missing"

if ! grep -q '^## External Critical Review' "$REPORT"; then
  fail "$REPORT has no line starting with: ## External Critical Review"
fi

python3 <<'PY'
import json, sys

def fail(msg):
    print(f"STEP4 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

try:
    ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
except Exception as e:
    fail(f"idea-creator.manifest.result.json invalid: {e}")

try:
    rr = json.load(open("idea-stage/research-review.manifest.result.json"))
except Exception as e:
    fail(f"research-review.manifest.result.json invalid: {e}")

if rr.get("result") != "success":
    fail(f'manifest result={rr.get("result")!r}, need "success"')

ideas = ic.get("ideas") or []
r1_list = [x for x in ideas if x.get("rank") == 1]
if not r1_list:
    fail("idea-creator manifest has no idea with rank 1")
r1 = r1_list[0]

rev = (rr.get("reviews") or [None])[0]
if not rev:
    fail("reviews[0] missing")

if rev.get("idea_title") != r1["title"]:
    fail(
        f"reviews[0].idea_title mismatch:\n"
        f"  idea-creator rank 1: {r1['title']!r}\n"
        f"  research-review:     {rev.get('idea_title')!r}"
    )

if not (rev.get("verdict_summary") or "").strip():
    fail("reviews[0].verdict_summary empty")

print("STEP4 OK")
PY
