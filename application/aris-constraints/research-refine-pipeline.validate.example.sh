#!/usr/bin/env bash
# Step 5 — from project root: bash aris-constraints/research-refine-pipeline.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"

fail_en() { echo "STEP5 FAIL: $*" >&2; exit 1; }
fail_zh() { echo "STEP5 失败: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f idea-stage/research-refine-pipeline.manifest.result.json ]] || fail_en "research-refine-pipeline.manifest.result.json missing"
[[ -f idea-stage/idea-creator.manifest.result.json ]] || fail_en "idea-creator.manifest.result.json missing"

if ! grep -q '^## Executive Summary' "$REPORT"; then
  fail_en "$REPORT has no line starting with: ## Executive Summary"
fi
if ! grep -q '^## 执行摘要' "$ZH_REPORT"; then
  if grep -q '^## Executive Summary' "$ZH_REPORT"; then
    fail_zh "$ZH_REPORT 请用契约 H2「## 执行摘要」，不要仅用英文「## Executive Summary」"
  else
    fail_zh "$ZH_REPORT 缺少以「## 执行摘要」开头的行（§1 契约 H2）"
  fi
fi

python3 <<'PY'
import json, os, sys

def fail_en(msg):
    print(f"STEP5 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP5 失败: {msg}", file=sys.stderr)
    sys.exit(1)

try:
    ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
except Exception as e:
    fail_en(f"idea-creator.manifest.result.json invalid: {e}")

try:
    rf = json.load(open("idea-stage/research-refine-pipeline.manifest.result.json"))
except Exception as e:
    fail_en(f"research-refine-pipeline.manifest.result.json invalid: {e}")

if rf.get("result") != "success":
    fail_en(f'manifest result={rf.get("result")!r}, need "success"')

ideas = ic.get("ideas") or []
r1_list = [x for x in ideas if x.get("rank") == 1]
if not r1_list:
    fail_en("idea-creator manifest has no idea with rank 1")
r1 = r1_list[0]

if rf.get("idea_title") != r1["title"]:
    fail_en(
        f"idea_title mismatch:\n"
        f"  idea-creator rank 1: {r1['title']!r}\n"
        f"  refine manifest:       {rf.get('idea_title')!r}"
    )

if not (rf.get("executive_summary") or "").strip():
    fail_en("executive_summary empty")

d = rf.get("deliverables") or {}
for k in ("final_proposal", "experiment_plan", "experiment_tracker"):
    p = d.get(k)
    if not p:
        fail_en(f"deliverables.{k} path missing")
    if not os.path.isfile(p):
        fail_en(f"deliverables.{k} not a file: {p!r}")
    if os.path.getsize(p) <= 0:
        fail_en(f"deliverables.{k} empty file: {p!r}")

body = open("idea-stage/IDEA_REPORT.en.md").read()
body_zh = open("idea-stage/IDEA_REPORT.zh.md").read()

def scaffold_leftovers(path, text):
    bad = []
    prefixes = (
        "[Step 5:",
        "[From RESEARCH_BRIEF.md]",
        "[Venue]",
        "[YYYY-MM-DD]",
        "[generated]",
        "[第 5 步：",
        "[来自 RESEARCH_BRIEF.md]",
    )
    for line in text.splitlines():
        s = line.strip()
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: scaffold line left: {s[:80]}")
    return bad

for msg in scaffold_leftovers("IDEA_REPORT.en.md", body):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md", body_zh):
    fail_zh(f"§1 仍留模版占位行: {msg}")

print("STEP5 OK")
PY
