#!/usr/bin/env bash
# Step 3 — from project root: bash aris-constraints/novelty-check.validate.example.sh [root]
# contract-id: novelty-check-validate-2026-03-20
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"
MANIFEST="idea-stage/novelty-check.manifest.result.json"
CONTRACT_TAG="novelty-check-validate-2026-03-20"

fail_en() { echo "STEP3 FAIL [$CONTRACT_TAG]: $*" >&2; exit 1; }
fail_zh() { echo "STEP3 失败 [$CONTRACT_TAG]: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f idea-stage/novelty-check.manifest.result.json ]] || fail_en "novelty-check.manifest.result.json missing"
[[ -f idea-stage/idea-creator.manifest.result.json ]] || fail_en "idea-creator.manifest.result.json missing"

python3 <<'PY'
import json, sys

CONTRACT = "novelty-check-validate-2026-03-20"

def fail_en(msg):
    print(f"STEP3 FAIL [{CONTRACT}]: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP3 失败 [{CONTRACT}]: {msg}", file=sys.stderr)
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

if not has_h2(body, "## Novelty Verification"):
    fail_en("IDEA_REPORT.en.md 缺少契约 H2「## Novelty Verification」")

if not has_h2(body_zh, "## 新颖性评估", "## Novelty Verification"):
    fail_zh("IDEA_REPORT.zh.md 缺少 §4 契约 H2：需要「## 新颖性评估」或「## Novelty Verification」")

ic = json.load(open("idea-stage/idea-creator.manifest.result.json", encoding="utf-8"))
nc = json.load(open("idea-stage/novelty-check.manifest.result.json", encoding="utf-8"))

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
        s = norm_line(line)
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: {s[:80]}")
    return bad

for msg in scaffold_leftovers("IDEA_REPORT.en.md", body):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md", body_zh):
    fail_zh(f"§4 仍留模版占位行: {msg}")

print("STEP3 OK")
PY
python3 aris-constraints/data-anchor.validate.example.py "STEP3" "$MANIFEST" "$REPORT" "$ZH_REPORT"
