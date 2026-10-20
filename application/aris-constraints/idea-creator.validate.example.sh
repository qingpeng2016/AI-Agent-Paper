#!/usr/bin/env bash
# Step 2 — from project root: bash aris-constraints/idea-creator.validate.example.sh [root]
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"
MANIFEST="idea-stage/idea-creator.manifest.result.json"

fail_en() { echo "STEP2 FAIL: $*" >&2; exit 1; }
fail_zh() { echo "STEP2 失败: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f "$MANIFEST" ]] || fail_en "$MANIFEST missing"

if ! grep -q '^## Recommended Ideas (ranked)' "$REPORT"; then
  fail_en "$REPORT has no line starting with: ## Recommended Ideas (ranked)"
fi
if ! grep -q '^## 推荐研究方案' "$ZH_REPORT"; then
  if grep -q '^## Recommended Ideas (ranked)' "$ZH_REPORT"; then
    fail_zh "$ZH_REPORT 请用契约 H2「## 推荐研究方案」，不要仅用英文「## Recommended Ideas (ranked)」"
  else
    fail_zh "$ZH_REPORT 缺少以「## 推荐研究方案」开头的行（§3 契约 H2）"
  fi
fi

python3 <<'PY'
import json, sys

def fail_en(msg):
    print(f"STEP2 FAIL: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP2 失败: {msg}", file=sys.stderr)
    sys.exit(1)

try:
    m = json.load(open("idea-stage/idea-creator.manifest.result.json"))
except Exception as e:
    fail_en(f"idea-creator.manifest.result.json invalid: {e}")

if m.get("result") != "success":
    fail_en(f'manifest result={m.get("result")!r}, need "success"')

ideas = m.get("ideas") or []
if not ideas:
    fail_en("manifest ideas[] empty, need at least 1")

ev = m.get("ideas_evaluated") or {}
rec = ev.get("recommended")
if rec is not None and rec != len(ideas):
    fail_en(f"ideas_evaluated.recommended={rec!r} but ideas.length={len(ideas)} (must match if set)")

ranks = [i.get("rank") for i in ideas]
want = list(range(1, len(ideas) + 1))
if sorted(ranks) != want:
    fail_en(f"ideas[].rank must be exactly {want}, got {ranks}")

body = open("idea-stage/IDEA_REPORT.en.md").read()
body_zh = open("idea-stage/IDEA_REPORT.zh.md").read()

def scaffold_leftovers(path, text):
    bad = []
    prefixes = (
        "[Optional:",
        "[Same 3.1.1",
        "[Same structure",
        "[结构同",
        "### Idea 1: [Title",
        "### Idea 2: [Title",
        "### Idea 3: [Title",
    )
    for line in text.splitlines():
        s = line.strip()
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: scaffold line left: {s[:80]}")
    return bad

for msg in scaffold_leftovers("IDEA_REPORT.en.md", body):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md", body_zh):
    fail_zh(f"§3 仍留模版占位行: {msg}")

for i in ideas:
    line = f"### Idea {i['rank']}: {i['title']}"
    if line not in body:
        fail_en(f"IDEA_REPORT.en.md missing exact line: {line!r}")
    if line not in body_zh:
        fail_zh(f"IDEA_REPORT.zh.md 缺少与 manifest 一致的标题行: {line!r}")

print("STEP2 OK")
PY
