#!/usr/bin/env bash
# Step 1 — from project root: bash aris-constraints/research-lit.validate.example.sh [root]
set -euo pipefail
cd "${1:-.}"
grep -q '^## Literature Landscape' idea-stage/IDEA_REPORT.md
python3 <<'PY'
import json
m = json.load(open("idea-stage/research-lit.manifest.result.json"))
assert m.get("result") == "success", m.get("result")
papers = m.get("papers") or []
assert len(papers) >= 8, f"papers.length={len(papers)}"
for p in papers:
    assert (p.get("title") or "").strip()
    assert isinstance(p.get("year"), int)
    urls, doi, arx = p.get("urls") or [], p.get("doi"), p.get("arxiv_id")
    assert urls or doi or arx, p.get("id")
print("STEP1 OK")
PY
