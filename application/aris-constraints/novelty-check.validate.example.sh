#!/usr/bin/env bash
# Step 3 — from project root: bash aris-constraints/novelty-check.validate.example.sh [root]
set -euo pipefail
cd "${1:-.}"
grep -q '^## Novelty Verification' idea-stage/IDEA_REPORT.md
python3 <<'PY'
import json
ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
nc = json.load(open("idea-stage/novelty-check.manifest.result.json"))
assert nc.get("result") == "success"
ideas, checks = ic.get("ideas") or [], nc.get("checks") or []
assert len(checks) == len(ideas)
by_rank = {c["rank"]: c for c in checks}
for idea in ideas:
    r = idea["rank"]
    assert by_rank[r]["title"] == idea["title"], f"rank {r} title mismatch"
print("STEP3 OK")
PY
