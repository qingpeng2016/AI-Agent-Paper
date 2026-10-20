#!/usr/bin/env bash
# Step 2 — from project root: bash aris-constraints/idea-creator.validate.example.sh [root]
# contract-id: idea-creator-validate-2026-03-20
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"
MANIFEST="idea-stage/idea-creator.manifest.result.json"
CONTRACT_TAG="idea-creator-validate-2026-03-20"

fail_en() { echo "STEP2 FAIL [$CONTRACT_TAG]: $*" >&2; exit 1; }
fail_zh() { echo "STEP2 失败 [$CONTRACT_TAG]: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f "$MANIFEST" ]] || fail_en "$MANIFEST missing"

python3 <<'PY'
import json, sys

CONTRACT = "idea-creator-validate-2026-03-20"

def fail_en(msg):
    print(f"STEP2 FAIL [{CONTRACT}]: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP2 失败 [{CONTRACT}]: {msg}", file=sys.stderr)
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

if not has_h2(body, "## Recommended Ideas (ranked)"):
    fail_en("IDEA_REPORT.en.md 缺少契约 H2「## Recommended Ideas (ranked)」")

if not has_h2(body_zh, "## 推荐研究方案", "## Recommended Ideas (ranked)"):
    fail_zh("IDEA_REPORT.zh.md 缺少 §3 契约 H2：需要「## 推荐研究方案」或「## Recommended Ideas (ranked)」")

try:
    m = json.load(open("idea-stage/idea-creator.manifest.result.json", encoding="utf-8"))
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
        s = norm_line(line)
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: {s[:80]}")
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
