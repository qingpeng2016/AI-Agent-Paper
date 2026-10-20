#!/usr/bin/env bash
# Step 1 — from project root: bash aris-constraints/research-lit.validate.example.sh [root]
# contract-id: research-lit-validate-2026-03-20 (EN §2=Literature Landscape; ZH §2=文献综述 or Literature Landscape)
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
REPORT="idea-stage/IDEA_REPORT.en.md"
ZH_REPORT="idea-stage/IDEA_REPORT.zh.md"
MANIFEST="idea-stage/research-lit.manifest.result.json"
CONTRACT_TAG="research-lit-validate-2026-03-20"

fail_en() { echo "STEP1 FAIL [$CONTRACT_TAG]: $*" >&2; exit 1; }
fail_zh() { echo "STEP1 失败 [$CONTRACT_TAG]: $*" >&2; exit 1; }

[[ -f "$REPORT" ]] || fail_en "$REPORT missing"
[[ -f "$ZH_REPORT" ]] || fail_zh "缺少 $ZH_REPORT"
[[ -f "$MANIFEST" ]] || fail_en "$MANIFEST missing"

python3 <<'PY'
import json, sys

CONTRACT = "research-lit-validate-2026-03-20"

def fail_en(msg):
    print(f"STEP1 FAIL [{CONTRACT}]: {msg}", file=sys.stderr)
    sys.exit(1)

def fail_zh(msg):
    print(f"STEP1 失败 [{CONTRACT}]: {msg}", file=sys.stderr)
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

if not has_h2(body, "## Literature Landscape"):
    fail_en("IDEA_REPORT.en.md 缺少契约 H2「## Literature Landscape」")

if not has_h2(body_zh, "## 文献综述", "## Literature Landscape"):
    fail_zh("IDEA_REPORT.zh.md 缺少 §2 契约 H2：需要「## 文献综述」或「## Literature Landscape」")

try:
    m = json.load(open("idea-stage/research-lit.manifest.result.json"))
except Exception as e:
    fail_en(f"research-lit.manifest.result.json invalid: {e}")

if m.get("result") != "success":
    fail_en(f'manifest result={m.get("result")!r}, need "success"')

papers = m.get("papers") or []
if not papers:
    fail_en("manifest papers[] empty")

def scaffold_leftovers(path, text):
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
        s = norm_line(line)
        if any(s.startswith(p) for p in prefixes):
            bad.append(f"{path}: {s[:80]}")
    return bad

for msg in scaffold_leftovers("IDEA_REPORT.en.md", body):
    fail_en(f"scaffold line left: {msg}")
for msg in scaffold_leftovers("IDEA_REPORT.zh.md", body_zh):
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
