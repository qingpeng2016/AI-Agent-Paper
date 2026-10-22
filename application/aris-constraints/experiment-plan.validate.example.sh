#!/usr/bin/env bash
# /experiment-plan — from project root: bash aris-constraints/experiment-plan.validate.example.sh [root]
# contract-id: experiment-plan-validate-2026-10-21-data-acq
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
CONTRACT_TAG="experiment-plan-validate-2026-10-21-data-acq"

fail_en() { echo "EXPERIMENT-PLAN FAIL [$CONTRACT_TAG]: $*" >&2; exit 1; }

MANIFEST="idea-stage/experiment-plan.manifest.result.json"
[[ -f "$MANIFEST" ]] || fail_en "$MANIFEST missing"
if [[ -f refine-logs/experiment-plan.manifest.result.json ]]; then
  fail_en "manifest is under refine-logs/; move or rewrite to idea-stage/experiment-plan.manifest.result.json"
fi

python3 <<'PY'
import json, os, re, sys

CONTRACT = "experiment-plan-validate-2026-10-21-data-acq"
REQUIRED_H2_EN = [
    "## Claim Map",
    "## Paper Storyline",
    "## Experiment Data Acquisition",
    "## Experiment Blocks",
    "## Run Order and Milestones",
    "## Compute and Data Budget",
    "## Risks and Mitigations",
    "## Final Checklist",
]
REQUIRED_H3_EN = [
    "### Experimental acquisition",
    "#### Subjects or units of study",
    "#### Methods",
    "#### Data products and layout",
    "### Data download",
    "#### Download URLs",
    "#### Access procedure",
    "#### Data products and layout",
]
REQUIRED_H3_ZH = [
    "### 方法一：实验获取",
    "#### 实验对象",
    "#### 实验方法",
    "### 方法二：数据下载",
    "#### 下载地址",
    "#### 如何获取",
]
REQUIRED_H2_ZH = [
    "## 论点与实验对照",
    "## 论文叙事",
    "## 实验数据获取",
    "## 实验模块",
    "## 运行顺序与里程碑",
    "## 算力与数据预算",
    "## 风险与缓解",
    "## 终检清单",
]
DELIVERABLE_KEYS = (
    "experiment_plan_en",
    "experiment_plan_zh",
    "experiment_tracker_en",
    "experiment_tracker_zh",
)
DEFAULT_PATHS = {
    "experiment_plan_en": "refine-logs/EXPERIMENT_PLAN.en.md",
    "experiment_plan_zh": "refine-logs/EXPERIMENT_PLAN.zh.md",
    "experiment_tracker_en": "refine-logs/EXPERIMENT_TRACKER.en.md",
    "experiment_tracker_zh": "refine-logs/EXPERIMENT_TRACKER.zh.md",
}

def fail_en(msg):
    print(f"EXPERIMENT-PLAN FAIL [{CONTRACT}]: {msg}", file=sys.stderr)
    sys.exit(1)

def norm_h2(line: str) -> str:
    line = line.replace("\r", "").strip().lstrip("\ufeff")
    if not line.startswith("##"):
        return line
    rest = line[2:].strip()
    return "## " + rest if rest else "##"

def has_h2(text: str, heading: str) -> bool:
    want = norm_h2(heading)
    for line in text.splitlines():
        h = norm_h2(line)
        if h == want or h.startswith(want + " ") or h.startswith(want + "（"):
            return True
    return False

def check_plan_h2(path: str, headings: list[str]) -> None:
    text = open(path, encoding="utf-8").read()
    for h2 in headings:
        if not has_h2(text, h2):
            fail_en(f"{path} missing required H2: {h2!r}")

def line_starts(text: str, prefix: str) -> int:
    return sum(1 for line in text.splitlines() if line.strip().startswith(prefix))

def check_data_acquisition_headings(path: str, h3_h4: list[str], dup_h4: str) -> None:
    text = open(path, encoding="utf-8").read()
    for h in h3_h4:
        if line_starts(text, h) < 1:
            fail_en(f"{path} missing required heading: {h!r}")
    if line_starts(text, dup_h4) < 2:
        fail_en(f"{path} need two {dup_h4!r} headings (方法一 + 方法二 各一节 #### 数据结果)")

def check_tracker_table(path: str) -> None:
    tracker = open(path, encoding="utf-8").read()
    if "|" not in tracker:
        fail_en(f"{path} must contain a markdown table")
    if not re.search(r"^\|[^|\n]+\|", tracker, re.MULTILINE):
        fail_en(f"{path} table has no data rows")

try:
    mf = json.load(open("idea-stage/experiment-plan.manifest.result.json", encoding="utf-8"))
except Exception as e:
    fail_en(f"manifest invalid JSON: {e}")

if mf.get("result") != "success":
    fail_en(f'manifest result={mf.get("result")!r}, need "success"')

if mf.get("aris_phase") != "experiment-plan":
    fail_en(f'aris_phase={mf.get("aris_phase")!r}, need "experiment-plan"')

summary = mf.get("summary") or {}
for key in ("must_run_blocks", "highest_risk", "first_runs"):
    val = summary.get(key)
    if key == "must_run_blocks":
        if not isinstance(val, list) or not any(str(x).strip() for x in val):
            fail_en(f"summary.{key} must be non-empty array")
    elif not (val and str(val).strip()):
        fail_en(f"summary.{key} empty")

d = mf.get("deliverables") or {}
for k in DELIVERABLE_KEYS:
    p = d.get(k) or DEFAULT_PATHS[k]
    if d.get(k) and p != DEFAULT_PATHS[k]:
        fail_en(f"deliverables.{k} must be exactly {DEFAULT_PATHS[k]!r}, got {p!r}")
    if not os.path.isfile(p):
        fail_en(f"deliverables.{k} not a file: {p!r}")
    if os.path.getsize(p) <= 0:
        fail_en(f"deliverables.{k} empty file: {p!r}")

plan_en = d.get("experiment_plan_en") or DEFAULT_PATHS["experiment_plan_en"]
plan_zh = d.get("experiment_plan_zh") or DEFAULT_PATHS["experiment_plan_zh"]
check_plan_h2(plan_en, REQUIRED_H2_EN)
check_plan_h2(plan_zh, REQUIRED_H2_ZH)
check_data_acquisition_headings(
    plan_en,
    [h for h in REQUIRED_H3_EN if h != "#### Data products and layout"],
    "#### Data products and layout",
)
check_data_acquisition_headings(plan_zh, REQUIRED_H3_ZH, "#### 数据结果")

tracker_en = d.get("experiment_tracker_en") or DEFAULT_PATHS["experiment_tracker_en"]
tracker_zh = d.get("experiment_tracker_zh") or DEFAULT_PATHS["experiment_tracker_zh"]
check_tracker_table(tracker_en)
check_tracker_table(tracker_zh)

ic_path = "idea-stage/idea-creator.manifest.result.json"
if os.path.isfile(ic_path):
    try:
        ic = json.load(open(ic_path, encoding="utf-8"))
    except Exception as e:
        fail_en(f"idea-creator manifest invalid: {e}")
    ideas = ic.get("ideas") or []
    r1 = next((x for x in ideas if x.get("rank") == 1), None)
    if not r1:
        fail_en("idea-creator manifest has no rank 1")
    if mf.get("idea_title") != r1.get("title"):
        fail_en(
            "idea_title mismatch vs idea-creator rank 1:\n"
            f"  idea-creator: {r1.get('title')!r}\n"
            f"  experiment-plan: {mf.get('idea_title')!r}"
        )

print(f"EXPERIMENT-PLAN OK [{CONTRACT}]")
PY
