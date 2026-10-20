#!/usr/bin/env bash
# Step 2 — from project root: bash aris-constraints/idea-creator.validate.example.sh [root]
set -euo pipefail
cd "${1:-.}"
grep -q '^## Recommended Ideas (ranked)' idea-stage/IDEA_REPORT.md
python3 <<'PY'
import json
m = json.load(open("idea-stage/idea-creator.manifest.result.json"))
assert m.get("result") == "success"
ideas = m.get("ideas") or []
assert len(ideas) == 3
ev = m.get("ideas_evaluated") or {}
assert ev.get("recommended") == 3 and ev.get("generated", 0) >= 8
body = open("idea-stage/IDEA_REPORT.md").read()
for i in ideas:
    line = f"### Idea {i['rank']}: {i['title']}"
    assert line in body, f"missing {line!r}"
print("STEP2 OK")
PY
