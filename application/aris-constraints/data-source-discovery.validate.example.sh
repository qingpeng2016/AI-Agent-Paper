#!/usr/bin/env bash
# /data-source-discovery — from project root: bash aris-constraints/data-source-discovery.validate.example.sh [root]
# contract-id: data-source-discovery-validate-2026-10-22-profile-v2
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
CONTRACT_TAG="data-source-discovery-validate-2026-10-22-profile-v2"

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

CONTRACT = "data-source-discovery-validate-2026-10-22-profile-v2"
ALLOW_TIER = {"open_free_direct"}
ALLOW_LICENSE_SUBSTR = (
    "cc0", "cc-by", "public domain", "pddl", "odc-by", "odbl",
)

REQUIRED_H2_EN = [
    "## Data requirements",
    "## Recommended datasets",
]
REQUIRED_H2_ZH = [
    "## 数据需求",
    "## 推荐数据集",
]
# Primary block 2.1 — #### must include numbering prefix + title
H4_SUFFIX_EN = [
    "1 Associated publication",
    "2 Directory structure",
    "3 Access and license",
    "4 Download procedure",
]
H4_SUFFIX_ZH = [
    "1 关联发表",
    "2 目录结构",
    "3 访问许可",
    "4 下载步骤",
]
MIN_DATASETS = 1
MAX_DATASETS = 3
FORBIDDEN_BODY_ZH = (
    "交付模版",
    "Read 本文件",
    "data-source-discovery.validate.example.sh",
    "只写 1 个时删掉",
    "写入 `data-source/DATA_SOURCE_PROFILE",
)
FORBIDDEN_BODY_EN = (
    "deliverable template",
    "Read this file",
    "data-source-discovery.validate.example.sh",
    "delete unused `### 2.2",
    "write `data-source/DATA_SOURCE_PROFILE",
)
LEGACY_H2_ZH = [
    "## 研究数据需求",
    "## 候选数据集",
    "## 主推荐数据集",
]
LEGACY_H2_EN = [
    "## Research data needs",
    "## Candidate datasets",
    "## Primary recommendation",
]
FORBIDDEN_TOP_H2_EN = [
    "## Directory and file structure",
    "## Access and license",
    "## Download procedure",
    "## Gaps and mitigations",
    "## Downstream hooks",
]
FORBIDDEN_TOP_H2_ZH = [
    "## 目录结构",
    "## 访问许可",
    "## 目录与文件结构",
    "## 目录和文件语义",
    "## 访问和许可",
    "## 访问、许可证和伦理",
    "## 下载步骤",
    "## 下载程序",
    "## 缺口与应对",
    "## 数据缺陷和缓解策略",
    "## 下游衔接",
    "## 总结",
]
FORBIDDEN_TOP_H2_EN_EXTRA = [
    "## Directory and file semantics",
    "## Download program",
    "## Gaps and mitigations",
    "## Downstream integration",
    "## Summary",
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

def has_h4_containing(text: str, needle: str) -> bool:
    for line in text.splitlines():
        s = line.replace("\r", "").strip().lstrip("\ufeff")
        if s.startswith("#### ") and needle in s:
            return True
    return False

def dataset_blocks(body: str):
    import re
    nums = [int(m.group(1)) for m in re.finditer(r"^### 2\.(\d+)\b", body, re.MULTILINE)]
    return nums

def check_file(path, h2_list, h4_suffixes, forbidden_h2, legacy_h2, label, need_toc, forbidden_body=()):
    import re
    if not os.path.isfile(path):
        fail(f"{label} missing: {path}")
    body = open(path, encoding="utf-8").read().strip()
    if len(body) < 200:
        fail(f"{label} too short or empty: {path}")
    for needle in forbidden_body:
        if needle in body:
            fail(f"{path} contains agent-instruction text {needle!r}; remove template preamble from deliverable")
    head = "\n".join(body.splitlines()[1:8])
    if "/data-source-discovery" in head:
        fail(f"{path} must not mention /data-source-discovery near top; profile is research doc only")
    lines = [ln.strip() for ln in body.splitlines() if ln.strip()]
    if len(lines) >= 2 and lines[1].startswith("`") and "data-source" in lines[1]:
        fail(f"{path} line 2 looks like agent preamble, not **日期/**Date** metadata")
    blocks = dataset_blocks(body)
    if not blocks:
        fail(f"{path} missing ### 2.1 dataset block")
    if len(blocks) < 1 or len(blocks) > 3:
        fail(f"{path} must have 1–3 dataset blocks (### 2.1–2.3), found {len(blocks)}")
    if blocks != list(range(1, len(blocks) + 1)):
        fail(f"{path} dataset blocks must be consecutive 2.1..2.{len(blocks)}, got {blocks}")
    bad = has_forbidden_top_h2(body, forbidden_h2)
    if bad:
        fail(f"{path} must nest {bad!r} under ### 2.x (use #### 2.x.y), not document-level H2")
    for leg in legacy_h2:
        if has_heading(body, (leg,)):
            fail(f"{path} uses legacy H2 {leg!r}; use v2 template (数据需求 / 推荐数据集)")
    for line in body.splitlines():
        s = line.strip()
        if s.startswith("## ") and ("详细版" in s or "缺口" in s or "下游" in s):
            fail(f"{path} forbidden section {s!r}")
    if need_toc and not has_heading(body, (need_toc,)):
        fail(f"{path} missing {need_toc!r} (copy from md_template)")
    for h in h2_list:
        if not has_heading(body, (h,)):
            fail(f"{path} missing H2 {h!r}")
    for n in blocks:
        for suffix in h4_suffixes:
            needle = f"2.{n}.{suffix}"
            if not has_h4_containing(body, needle):
                fail(f"{path} missing #### {needle!r} under ### 2.{n}")

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
if len(datasets) < MIN_DATASETS or len(datasets) > MAX_DATASETS:
    fail(f"datasets[] must have {MIN_DATASETS}–{MAX_DATASETS} entries, got {len(datasets)}")
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
forbidden_en = FORBIDDEN_TOP_H2_EN + FORBIDDEN_TOP_H2_EN_EXTRA
check_file(
    pen, REQUIRED_H2_EN, H4_SUFFIX_EN, forbidden_en, LEGACY_H2_EN,
    "profile EN", "## Table of contents", FORBIDDEN_BODY_EN,
)
check_file(
    pzh, REQUIRED_H2_ZH, H4_SUFFIX_ZH, FORBIDDEN_TOP_H2_ZH, LEGACY_H2_ZH,
    "profile ZH", "## 目录", FORBIDDEN_BODY_ZH,
)

blocks_zh = dataset_blocks(open(pzh, encoding="utf-8").read())
if len(blocks_zh) != len(datasets):
    fail(f"profile has {len(blocks_zh)} ### 2.x block(s) but datasets[] has {len(datasets)}; must match (1–3)")

print(f"DATA-SOURCE-DISCOVERY OK [{CONTRACT}] primary={pid} datasets={len(datasets)} profiles={pen},{pzh}")
PY
