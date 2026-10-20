#!/usr/bin/env bash
# Step 1 — from project root: bash aris-constraints/research-lit.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.md"
MANIFEST="idea-stage/research-lit.manifest.result.json"

fail() { echo "STEP1 FAIL: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail "$REPORT missing"
[[ -f "$MANIFEST" ]] || fail "$MANIFEST missing"

if ! grep -q '^## Literature Landscape' "$REPORT"; then
  fail "$REPORT has no line starting with: ## Literature Landscape"
fi

python3 <<'PY'
import json, sys

def fail(msg):
    print(f"STEP1 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

try:
    m = json.load(open("idea-stage/research-lit.manifest.result.json"))
except Exception as e:
    fail(f"research-lit.manifest.result.json invalid: {e}")

if m.get("result") != "success":
    fail(f'manifest result={m.get("result")!r}, need "success"')

body = open("idea-stage/IDEA_REPORT.md").read()
papers = m.get("papers") or []
if not papers:
    fail("manifest papers[] empty")

for p in papers:
    pid = p.get("id")
    if not (p.get("title") or "").strip():
        fail(f"paper {pid!r} missing title")
    if not isinstance(p.get("year"), int):
        fail(f"paper {pid!r} year must be int, got {p.get('year')!r}")
    urls, doi, arx = p.get("urls") or [], p.get("doi"), p.get("arxiv_id")
    if not (urls or doi or arx):
        fail(f"paper {pid!r} needs urls, doi, or arxiv_id")
    listed = [(u or "").strip() for u in urls if (u or "").strip()]
    if listed and not any(u in body for u in listed):
        fail(f"no url from papers[].urls in IDEA_REPORT for paper {pid!r}: {listed}")

print("STEP1 OK")
PY
