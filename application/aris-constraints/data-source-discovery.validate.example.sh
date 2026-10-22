#!/usr/bin/env bash
# /data-source-discovery — from project root: bash aris-constraints/data-source-discovery.validate.example.sh [root]
# contract-id: data-source-discovery-validate-2026-10-21-data-source-dir
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
CONTRACT_TAG="data-source-discovery-validate-2026-10-21-data-source-dir"

fail() { echo "DATA-SOURCE-DISCOVERY FAIL [$CONTRACT_TAG]: $*" >&2; exit 1; }

[[ -d data-source ]] || fail "data-source/ missing (run /data-source-discovery first)"
MANIFEST="data-source/data-source-discovery.manifest.result.json"
[[ -f "$MANIFEST" ]] || fail "$MANIFEST missing"
if [[ -f idea-stage/data-source-discovery.manifest.result.json ]]; then
  fail "manifest is under idea-stage/; use data-source/data-source-discovery.manifest.result.json"
fi
if [[ -f refine-logs/data-source-discovery.manifest.result.json ]]; then
  fail "manifest is under refine-logs/; use data-source/data-source-discovery.manifest.result.json"
fi

python3 <<'PY'
import json, os, sys

CONTRACT = "data-source-discovery-validate-2026-10-21-data-source-dir"
ALLOW_TIER = {"open_free_direct"}
ALLOW_LICENSE_SUBSTR = (
    "cc0", "cc-by", "public domain", "pddl", "odc-by", "odbl",
)

REQUIRED_H2_EN = [
    "## Research data needs",
    "## Candidate datasets",
    "## Primary recommendation",
]
REQUIRED_H2_ZH = [
    "## 研究数据需求",
    "## 候选数据集",
    "## 主推荐数据集",
]
REQUIRED_H4_EN = [
    "#### Cohort and design",
    "#### Associated publication",
    "#### Directory and file structure",
    "#### Access and license",
    "#### Download procedure",
]
REQUIRED_H4_ZH = [
    "#### 队列与设计",
    "#### 关联发表",
    "#### 目录与文件结构",
    "#### 访问和许可",
    "#### 下载步骤",
]
FORBIDDEN_TOP_H2_EN = [
    "## Directory and file structure",
    "## Access and license",
    "## Download procedure",
    "## Gaps and mitigations",
    "## Downstream hooks",
]
FORBIDDEN_TOP_H2_ZH = [
    "## 目录与文件结构",
    "## 访问和许可",
    "## 下载步骤",
    "## 缺口与应对",
    "## 下游衔接",
]

def fail(msg):
    print(f"DATA-SOURCE-DISCOVERY FAIL [{CONTRACT}]: {msg}", file=sys.stderr)
    sys.exit(1)

def norm_h2(line: str) -> str:
    line = line.replace("\r", "").strip().lstrip("\ufeff")
    if not line.startswith("##"):
        return line
    rest = line[2:].strip()
    return "## " + rest if rest else "##"

def has_heading(text: str, headings) -> bool:
    want = set(headings)
    for line in text.splitlines():
        s = line.replace("\r", "").strip().lstrip("\ufeff")
        if s in want:
            return True
    return False

def has_forbidden_top_h2(text: str, forbidden):
    for line in text.splitlines():
        s = line.replace("\r", "").strip().lstrip("\ufeff")
        if s.startswith("###"):
            continue
        if s.startswith("## ") and s in forbidden:
            return s
    return None

def check_file(path, h2_list, h4_list, forbidden_h2, label):
    if not os.path.isfile(path):
        fail(f"{label} missing: {path}")
    body = open(path, encoding="utf-8").read().strip()
    if len(body) < 200:
        fail(f"{label} too short or empty: {path}")
    bad = has_forbidden_top_h2(body, forbidden_h2)
    if bad:
        fail(f"{path} must nest {bad!r} under ### 1./2. (use ####), not document-level H2")
    for h in h2_list:
        if not has_heading(body, (h,)):
            fail(f"{path} missing H2 {h!r}")
    for h in h4_list:
        if not has_heading(body, (h,)):
            fail(f"{path} missing H4 {h!r} (under ### 1. primary block)")

if not os.path.isdir("data-source"):
    fail("data-source/ directory missing")

m = json.load(open("data-source/data-source-discovery.manifest.result.json"))
if m.get("aris_phase") != "data-source-discovery":
    fail(f"aris_phase must be data-source-discovery, got {m.get('aris_phase')!r}")
if m.get("result") != "success":
    fail('manifest result must be "success" for validate pass (or run agent to fix gates)')

summary = (m.get("executive_summary") or "").strip()
if not summary:
    fail("executive_summary empty")

pid = m.get("primary_recommendation_id")
datasets = m.get("datasets") or []
if not pid:
    fail("primary_recommendation_id missing")
by_id = {d.get("id"): d for d in datasets if d.get("id")}
primary = by_id.get(pid)
if not primary:
    fail(f"primary_recommendation_id {pid!r} not in datasets[]")

tier = primary.get("access_tier")
if tier not in ALLOW_TIER:
    fail(f"primary access_tier must be open_free_direct, got {tier!r}")

lic = (primary.get("license") or "").lower()
if not any(s in lic for s in ALLOW_LICENSE_SUBSTR):
    fail(f"primary license not in allow-list: {primary.get('license')!r}")

ver = primary.get("verification") or {}
if not ver.get("structure_verified"):
    fail("primary verification.structure_verified must be true")

deliverables = m.get("deliverables") or {}
pen = deliverables.get("profile_en", "data-source/DATA_SOURCE_PROFILE.en.md")
pzh = deliverables.get("profile_zh", "data-source/DATA_SOURCE_PROFILE.zh.md")
for p in (pen, pzh):
    if not p.startswith("data-source/"):
        fail(f"deliverable must be under data-source/: {p}")
check_file(pen, REQUIRED_H2_EN, REQUIRED_H4_EN, FORBIDDEN_TOP_H2_EN, "profile EN")
check_file(pzh, REQUIRED_H2_ZH, REQUIRED_H4_ZH, FORBIDDEN_TOP_H2_ZH, "profile ZH")

print(f"DATA-SOURCE-DISCOVERY OK [{CONTRACT}] primary={pid} profiles={pen},{pzh}")
PY
