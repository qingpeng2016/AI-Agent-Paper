#!/usr/bin/env bash
# Step 4 — from project root: bash aris-constraints/skills/research-review/research-review.validate.sh [root]
# contract-id: research-review-validate-2026-03-20
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"
CONTRACT_TAG="research-review-validate-2026-03-20"

fail_en() { echo "STEP4 FAIL [$CONTRACT_TAG]: $*" >&2; exit 1; }
fail_zh() { echo "STEP4 失败 [$CONTRACT_TAG]: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f idea-stage/research-review.manifest.result.json ]] || fail_en "research-review.manifest.result.json missing"
[[ -f idea-stage/idea-creator.manifest.result.json ]] || fail_en "idea-creator.manifest.result.json missing"

python3 <<'PY'
import json, sys

CONTRACT = "research-review-validate-2026-03-20"

def fail_en(msg):
    print(f"STEP4 FAIL [{CONTRACT}]: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP4 失败 [{CONTRACT}]: {msg}", file=sys.stderr)
    sys.exit(1)

def norm_line(line: str) -> str:
    return line.replace("\r", "").strip().lstrip("\ufeff")

def norm_h2(line: str) -> str:
    line = norm_line(line)
    if not line.startswith("##"):
        return line
    rest = line[2:].strip()
    return "## " + rest if rest else "##"

def has_h2(text: str, *headings: str) -> bool:
    want = set(headings)
    for line in text.splitlines():
        if norm_h2(line) in want:
            return True
    return False

body = open("idea-stage/IDEA_REPORT.en.md", encoding="utf-8").read()
body_zh = open("idea-stage/IDEA_REPORT.zh.md", encoding="utf-8").read()

if not has_h2(body, "## External Critical Review"):
    fail_en("IDEA_REPORT.en.md 缺少契约 H2「## External Critical Review」")

if not has_h2(body_zh, "## 外部审稿意见", "## External Critical Review"):
    fail_zh("IDEA_REPORT.zh.md 缺少 §5 契约 H2：需要「## 外部审稿意见」或「## External Critical Review」")

try:
    ic = json.load(open("idea-stage/idea-creator.manifest.result.json", encoding="utf-8"))
except Exception as e:
    fail_en(f"idea-creator.manifest.result.json invalid: {e}")

try:
    rr = json.load(open("idea-stage/research-review.manifest.result.json", encoding="utf-8"))
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
        s = norm_line(line)
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: {s[:80]}")
    return bad

for msg in scaffold_leftovers("IDEA_REPORT.en.md", body):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md", body_zh):
    fail_zh(f"§5 仍留模版占位行: {msg}")

print("STEP4 OK")
PY
