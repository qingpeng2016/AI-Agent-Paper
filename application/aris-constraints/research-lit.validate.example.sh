#!/usr/bin/env bash
# Step 1 — from project root: bash aris-constraints/research-lit.validate.example.sh [root]
set -euo pipefail
cd "${1:-.}"
grep -q '^## Literature Landscape' idea-stage/IDEA_REPORT.md
python3 <<'PY'
import json

m = json.load(open("idea-stage/research-lit.manifest.result.json"))
assert m.get("result") == "success", m.get("result")
body = open("idea-stage/IDEA_REPORT.md").read()
papers = m.get("papers") or []
assert papers, "papers array empty"
for p in papers:
    assert (p.get("title") or "").strip()
    assert isinstance(p.get("year"), int)
    urls, doi, arx = p.get("urls") or [], p.get("doi"), p.get("arxiv_id")
    assert urls or doi or arx, p.get("id")
    listed = [(u or "").strip() for u in urls if (u or "").strip()]
    if listed:
        assert any(u in body for u in listed), (
            f"no url from papers[].urls found in IDEA_REPORT for paper {p.get('id')!r}: {listed}"
        )
print("STEP1 OK")
PY
