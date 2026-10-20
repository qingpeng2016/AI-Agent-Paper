#!/usr/bin/env bash
# Step 1 — from project root: bash aris-constraints/research-lit.validate.example.sh [root]
# Contract: EN §2 H2 = ## Literature Landscape ; ZH §2 H2 = ## 文献综述
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"
MANIFEST="idea-stage/research-lit.manifest.result.json"

fail_en() { echo "STEP1 FAIL: $*" >&2; exit 1; }
fail_zh() { echo "STEP1 失败: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f "$MANIFEST" ]] || fail_en "$MANIFEST missing"

if ! grep -q '^## Literature Landscape' "$REPORT"; then
  fail_en "$REPORT has no line starting with: ## Literature Landscape"
fi
if ! grep -q '^## 文献综述' "$ZH_REPORT"; then
  if grep -q '^## Literature Landscape' "$ZH_REPORT"; then
    fail_zh "$ZH_REPORT 中文报告请用契约 H2「## 文献综述」，不要仅用英文「## Literature Landscape」"
  else
    fail_zh "$ZH_REPORT 缺少以「## 文献综述」开头的行（§2 契约 H2；英文版对应为 ## Literature Landscape）"
  fi
fi

python3 <<'PY'
import json, sys

def fail_en(msg):
    print(f"STEP1 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP1 失败: {msg}", file=sys.stderr)
    sys.exit(1)

try:
    m = json.load(open("idea-stage/research-lit.manifest.result.json"))
except Exception as e:
    fail_en(f"research-lit.manifest.result.json invalid: {e}")

if m.get("result") != "success":
    fail_en(f'manifest result={m.get("result")!r}, need "success"')

body = open("idea-stage/IDEA_REPORT.en.md").read()
body_zh = open("idea-stage/IDEA_REPORT.zh.md").read()
papers = m.get("papers") or []
if not papers:
    fail_en("manifest papers[] empty")

def scaffold_leftovers(path, text, zh=False):
    bad = []
    prefixes = (
        "[Step 1:",
        "[Each entry:",
        "[Repeat 2.4",
        "[第 1 步：",
        "[每条：",
        "{Each ",
        "{Repeat ",
        "{Step 1:",
        "{每条：",
    )
    for line in text.splitlines():
        s = line.strip()
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: {s[:80]}")
    return bad

for msg in scaffold_leftovers("IDEA_REPORT.en.md", body):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md", body_zh, zh=True):
    fail_zh(f"§2/§6 仍留模版占位行: {msg}")

for p in papers:
    pid = p.get("id")
    if not (p.get("title") or "").strip():
        fail_en(f"paper {pid!r} missing title")
    if not isinstance(p.get("year"), int):
        fail_en(f"paper {pid!r} year must be int, got {p.get('year')!r}")
    urls, doi, arx = p.get("urls") or [], p.get("doi"), p.get("arxiv_id")
    if not (urls or doi or arx):
        fail_en(f"paper {pid!r} needs urls, doi, or arxiv_id")
    listed = [(u or "").strip() for u in urls if (u or "").strip()]
    if listed and not any(u in body for u in listed):
        fail_en(f"no url from papers[].urls in IDEA_REPORT.en.md for paper {pid!r}: {listed}")
    if listed and not any(u in body_zh for u in listed):
        fail_zh(f"论文 {pid!r} 的 urls[] 中没有任何一条出现在 IDEA_REPORT.zh.md 的 §2.4/§6 中: {listed}")

print("STEP1 OK")
PY
