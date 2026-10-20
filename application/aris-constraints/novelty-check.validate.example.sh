#!/usr/bin/env bash
# Step 3 — from project root: bash aris-constraints/novelty-check.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"

fail_en() { echo "STEP3 FAIL: $*" >&2; exit 1; }
fail_zh() { echo "STEP3 失败: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f idea-stage/novelty-check.manifest.result.json ]] || fail_en "novelty-check.manifest.result.json missing"
[[ -f idea-stage/idea-creator.manifest.result.json ]] || fail_en "idea-creator.manifest.result.json missing"

if ! grep -q '^## Novelty Verification' "$REPORT"; then
  fail_en "$REPORT has no line starting with: ## Novelty Verification"
fi
if ! grep -q '^## 新颖性评估' "$ZH_REPORT"; then
  if grep -q '^## Novelty Verification' "$ZH_REPORT"; then
    fail_zh "$ZH_REPORT 请用契约 H2「## 新颖性评估」，不要仅用英文「## Novelty Verification」"
  else
    fail_zh "$ZH_REPORT 缺少以「## 新颖性评估」开头的行（§4 契约 H2）"
  fi
fi

python3 <<'PY'
import json, sys

def fail_en(msg):
    print(f"STEP3 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP3 失败: {msg}", file=sys.stderr)
    sys.exit(1)

ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
nc = json.load(open("idea-stage/novelty-check.manifest.result.json"))

if nc.get("result") != "success":
    fail_en(f'novelty manifest result={nc.get("result")!r}, need "success"')

ideas, checks = ic.get("ideas") or [], nc.get("checks") or []
if len(checks) != len(ideas):
    fail_en(f"checks.length={len(checks)} but idea-creator ideas.length={len(ideas)}")

by_rank = {c.get("rank"): c for c in checks}
for idea in ideas:
    r = idea["rank"]
    if r not in by_rank:
        fail_en(f"checks missing rank {r}")
    got = by_rank[r].get("title")
    want = idea["title"]
    if got != want:
        fail_en(f"rank {r} title mismatch:\n  idea-creator: {want!r}\n  novelty-check:  {got!r}")

body = open("idea-stage/IDEA_REPORT.en.md").read()
body_zh = open("idea-stage/IDEA_REPORT.zh.md").read()

def scaffold_leftovers(path, text):
    bad = []
    prefixes = (
        "[Step 3:",
        "[Same structure",
        "[Whether ideas overlap",
        "[第 3 步：",
        "[结构同上",
        "[三个方案是否冗余",
    )
    for line in text.splitlines():
        s = line.strip()
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: scaffold line left: {s[:80]}")
    return bad

for msg in scaffold_leftovers("IDEA_REPORT.en.md", body):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md", body_zh):
    fail_zh(f"§4 仍留模版占位行: {msg}")

print("STEP3 OK")
PY
