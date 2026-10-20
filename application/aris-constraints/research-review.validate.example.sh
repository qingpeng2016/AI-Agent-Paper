#!/usr/bin/env bash
# Step 4 — from project root: bash aris-constraints/research-review.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"

fail_en() { echo "STEP4 FAIL: $*" >&2; exit 1; }
fail_zh() { echo "STEP4 失败: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f idea-stage/research-review.manifest.result.json ]] || fail_en "research-review.manifest.result.json missing"
[[ -f idea-stage/idea-creator.manifest.result.json ]] || fail_en "idea-creator.manifest.result.json missing"

if ! grep -q '^## External Critical Review' "$REPORT"; then
  fail_en "$REPORT has no line starting with: ## External Critical Review"
fi
if ! grep -q '^## 外部审稿意见' "$ZH_REPORT"; then
  if grep -q '^## External Critical Review' "$ZH_REPORT"; then
    fail_zh "$ZH_REPORT 请用契约 H2「## 外部审稿意见」，不要仅用英文「## External Critical Review」"
  else
    fail_zh "$ZH_REPORT 缺少以「## 外部审稿意见」开头的行（§5 契约 H2）"
  fi
fi

python3 <<'PY'
import json, sys

def fail_en(msg):
    print(f"STEP4 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP4 失败: {msg}", file=sys.stderr)
    sys.exit(1)

try:
    ic = json.load(open("idea-stage/idea-creator.manifest.result.json"))
except Exception as e:
    fail_en(f"idea-creator.manifest.result.json invalid: {e}")

try:
    rr = json.load(open("idea-stage/research-review.manifest.result.json"))
except Exception as e:
    fail_en(f"research-review.manifest.result.json invalid: {e}")

if rr.get("result") != "success":
    fail_en(f'manifest result={rr.get("result")!r}, need "success"')

ideas = ic.get("ideas") or []
r1_list = [x for x in ideas if x.get("rank") == 1]
if not r1_list:
    fail_en("idea-creator manifest has no idea with rank 1")
r1 = r1_list[0]

rev = (rr.get("reviews") or [None])[0]
if not rev:
    fail_en("reviews[0] missing")

if rev.get("idea_title") != r1["title"]:
    fail_en(
        f"reviews[0].idea_title mismatch:\n"
        f"  idea-creator rank 1: {r1['title']!r}\n"
        f"  research-review:     {rev.get('idea_title')!r}"
    )

if not (rev.get("verdict_summary") or "").strip():
    fail_en("reviews[0].verdict_summary empty")

body = open("idea-stage/IDEA_REPORT.en.md").read()
body_zh = open("idea-stage/IDEA_REPORT.zh.md").read()

def scaffold_leftovers(path, text):
    bad = []
    prefixes = (
        "[Step 4:",
        "[Summary verdict]",
        "[Strength]",
        "[Weakness]",
        "[Concern]",
        "[Venue fit",
        "[Concrete follow-ups]",
        "[第 4 步：",
        "[总体结论]",
        "[优点]",
        "[问题]",
        "[关切]",
        "[期刊匹配",
        "[可操作的修改建议]",
    )
    for line in text.splitlines():
        s = line.strip()
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: scaffold line left: {s[:80]}")
    return bad

for msg in scaffold_leftovers("IDEA_REPORT.en.md", body):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md", body_zh):
    fail_zh(f"§5 仍留模版占位行: {msg}")

print("STEP4 OK")
PY
