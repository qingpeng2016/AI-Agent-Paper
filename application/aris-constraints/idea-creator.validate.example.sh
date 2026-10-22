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

def matches_contract_h2(line: str, heading: str) -> bool:
    n = norm_line(line)
    if norm_h2(line) == heading or n == heading:
        return True
    # e.g. "## Novelty Verification## Novelty Verification" from a bad merge
    return n.startswith(heading) and len(n) > len(heading) and n[len(heading)] in "# \t"

def section_between(text: str, start_h2s: tuple[str, ...], end_h2s: tuple[str, ...]) -> str:
    """Lines from first §3 contract H2 until §4 contract H2 (Step 2 scope only)."""
    lines = text.splitlines()
    start = end = None
    for i, line in enumerate(lines):
        if start is None and any(matches_contract_h2(line, h) for h in start_h2s):
            start = i
            continue
        if start is not None and any(matches_contract_h2(line, h) for h in end_h2s):
            end = i
            break
    if start is None:
        return ""
    if end is None:
        end = len(lines)
    return "\n".join(lines[start:end])

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

sec3_en = section_between(
    body,
    ("## Recommended Ideas (ranked)",),
    ("## Novelty Verification",),
)
sec3_zh = section_between(
    body_zh,
    ("## 推荐研究方案", "## Recommended Ideas (ranked)"),
    ("## 新颖性评估", "## Novelty Verification"),
)
if not sec3_en.strip():
    fail_en("§3 区间为空（从 ## Recommended Ideas (ranked) 到 ## Novelty Verification）")
if not sec3_zh.strip():
    fail_zh("§3 区间为空（从 ## 推荐研究方案 到 ## 新颖性评估）")

for msg in scaffold_leftovers("IDEA_REPORT.en.md §3", sec3_en):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md §3", sec3_zh):
    fail_zh(f"§3 仍留模版占位行: {msg}")

for i in ideas:
    line = f"### Idea {i['rank']}: {i['title']}"
    if line not in body:
        fail_en(f"IDEA_REPORT.en.md missing exact line: {line!r}")
    if line not in body_zh:
        fail_zh(f"IDEA_REPORT.zh.md 缺少与 manifest 一致的标题行: {line!r}")

print("STEP2 OK")
PY
python3 aris-constraints/data-anchor.validate.example.py "STEP2" "$MANIFEST" "$REPORT" "$ZH_REPORT"
