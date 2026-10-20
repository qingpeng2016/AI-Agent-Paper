#!/usr/bin/env bash
# Step 5 — from project root: bash aris-constraints/research-refine-pipeline.validate.example.sh [root]
set -euo pipefail
cd "${1:-.}"
grep -q '^## Executive Summary' idea-stage/IDEA_REPORT.md
python3 <<'PY'
import json, os
ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
rf = json.load(open("idea-stage/research-refine-pipeline.manifest.result.json"))
assert rf.get("result") == "success"
r1 = next(x for x in ic["ideas"] if x.get("rank") == 1)
assert rf.get("idea_title") == r1["title"]
assert (rf.get("executive_summary") or "").strip()
d = rf.get("deliverables") or {}
for k in ("final_proposal", "experiment_plan", "experiment_tracker"):
    p = d.get(k)
    assert p and os.path.isfile(p) and os.path.getsize(p) > 0, k
print("STEP5 OK")
PY
