#!/usr/bin/env bash
# Step 4 — from project root: bash aris-constraints/research-review.validate.example.sh [root]
set -euo pipefail
cd "${1:-.}"
grep -q '^## External Critical Review' idea-stage/IDEA_REPORT.md
python3 <<'PY'
import json
ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
rr = json.load(open("idea-stage/research-review.manifest.result.json"))
assert rr.get("result") == "success"
r1 = next(x for x in ic["ideas"] if x.get("rank") == 1)
rev = (rr.get("reviews") or [None])[0]
assert rev and rev.get("idea_title") == r1["title"]
assert (rev.get("verdict_summary") or "").strip()
assert len(rev.get("review_points") or []) >= 1
print("STEP4 OK")
PY
