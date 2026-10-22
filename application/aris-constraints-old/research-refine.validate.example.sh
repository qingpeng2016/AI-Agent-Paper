#!/usr/bin/env bash
# /research-refine — from project root: bash aris-constraints/research-refine.validate.example.sh [root]
# contract-id: research-refine-validate-2026-10-24
set -euo pipefail
ROOT="${1:-.}"
cd "$ROOT"
CONTRACT_TAG="research-refine-validate-2026-10-24"

fail_en() { echo "RESEARCH-REFINE FAIL [$CONTRACT_TAG]: $*" >&2; exit 1; }

MANIFEST="idea-stage/research-refine.manifest.result.json"
[[ -f "$MANIFEST" ]] || fail_en "$MANIFEST missing"
if [[ -f refine-logs/research-refine.manifest.result.json ]]; then
  fail_en "manifest is under refine-logs/; move or rewrite to idea-stage/research-refine.manifest.result.json"
fi

python3 <<'PY'
import glob, json, os, re, sys

CONTRACT = "research-refine-validate-2026-10-24"
MAX_REVIEW_ROUNDS = 5
REQUIRED_H2_PROPOSAL = [
    "## Problem Anchor",
    "## Technical Gap",
    "## Method Thesis",
    "## Contribution Focus",
    "## Proposed Method",
    "## Claim-Driven Validation Sketch",
    "## Experiment Handoff Inputs",
    "## Compute & Timeline Estimate",
]
DELIVERABLE_KEYS = (
    "final_proposal",
    "review_summary",
    "refinement_report",
    "refine_state",
)
DEFAULT_PATHS = {
    "final_proposal": "refine-logs/FINAL_PROPOSAL.md",
    "review_summary": "refine-logs/REVIEW_SUMMARY.md",
    "refinement_report": "refine-logs/REFINEMENT_REPORT.md",
    "refine_state": "refine-logs/REFINE_STATE.json",
}
ROUND0 = "refine-logs/round-0-initial-proposal.md"

def fail_en(msg):
    print(f"RESEARCH-REFINE FAIL [{CONTRACT}]: {msg}", file=sys.stderr)
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

def require_nonempty_file(path: str, label: str) -> None:
    if not os.path.isfile(path):
        fail_en(f"{label} missing: {path!r}")
    if os.path.getsize(path) <= 0:
        fail_en(f"{label} empty: {path!r}")

def check_proposal(path: str) -> None:
    text = open(path, encoding="utf-8").read()
    for h2 in REQUIRED_H2_PROPOSAL:
        if not has_h2(text, h2):
            fail_en(f"{path} missing required H2: {h2!r}")

def check_review_summary(path: str) -> None:
    text = open(path, encoding="utf-8").read()
    if not has_h2(text, "## Problem Anchor"):
        fail_en(f"{path} missing required H2: '## Problem Anchor'")
    if not (has_h2(text, "## Final Status") or "Final Verdict" in text):
        fail_en(f"{path} must include '## Final Status' or 'Final Verdict' metadata")

try:
    mf = json.load(open("idea-stage/research-refine.manifest.result.json", encoding="utf-8"))
except Exception as e:
    fail_en(f"manifest invalid JSON: {e}")

if mf.get("result") != "success":
    fail_en(f'manifest result={mf.get("result")!r}, need "success"')

if mf.get("aris_phase") != "research-refine":
    fail_en(f'aris_phase={mf.get("aris_phase")!r}, need "research-refine"')

phases = mf.get("pipeline_phases_completed") or []
if "5" not in phases and "5" not in [str(x) for x in phases]:
    fail_en('pipeline_phases_completed must include "5"')

summary = mf.get("summary") or {}
for key in ("final_verdict", "problem_anchor_one_liner"):
    if not (summary.get(key) and str(summary.get(key)).strip()):
        fail_en(f"summary.{key} empty")
if summary.get("final_score") is None:
    fail_en("summary.final_score missing")
R = summary.get("rounds_completed")
if not isinstance(R, int) or R < 1 or R > MAX_REVIEW_ROUNDS:
    fail_en(
        f"summary.rounds_completed must be integer 1..{MAX_REVIEW_ROUNDS}, got {R!r}"
    )
if summary.get("max_review_rounds") != MAX_REVIEW_ROUNDS:
    fail_en(
        f"summary.max_review_rounds must be {MAX_REVIEW_ROUNDS}, "
        f"got {summary.get('max_review_rounds')!r}"
    )
exited_after = summary.get("exited_after_review")
if not isinstance(exited_after, bool):
    fail_en("summary.exited_after_review must be boolean (upstream STOP / max-rounds exit)")

for path in glob.glob("refine-logs/round-*-review.md"):
    m = re.search(r"round-(\d+)-review\.md$", os.path.basename(path))
    if m and int(m.group(1)) > MAX_REVIEW_ROUNDS:
        fail_en(f"review round exceeds cap {MAX_REVIEW_ROUNDS}: {path!r}")
for path in glob.glob("refine-logs/round-*-refinement.md"):
    m = re.search(r"round-(\d+)-refinement\.md$", os.path.basename(path))
    if m and int(m.group(1)) > MAX_REVIEW_ROUNDS:
        fail_en(f"refinement round exceeds cap {MAX_REVIEW_ROUNDS}: {path!r}")

require_nonempty_file(ROUND0, "round-0-initial-proposal")
for n in range(1, R + 1):
    require_nonempty_file(f"refine-logs/round-{n}-review.md", f"round-{n}-review")
for n in range(1, R):
    require_nonempty_file(f"refine-logs/round-{n}-refinement.md", f"round-{n}-refinement")
refinement_final = f"refine-logs/round-{R}-refinement.md"
if not exited_after:
    require_nonempty_file(refinement_final, f"round-{R}-refinement")
elif os.path.isfile(refinement_final) and os.path.getsize(refinement_final) <= 0:
    fail_en(f"{refinement_final} exists but empty")

d = mf.get("deliverables") or {}
for k in DELIVERABLE_KEYS:
    p = d.get(k) or DEFAULT_PATHS[k]
    if d.get(k) and p != DEFAULT_PATHS[k]:
        fail_en(f"deliverables.{k} must be exactly {DEFAULT_PATHS[k]!r}, got {p!r}")
    require_nonempty_file(p, f"deliverables.{k}")

check_proposal(d.get("final_proposal") or DEFAULT_PATHS["final_proposal"])
check_review_summary(d.get("review_summary") or DEFAULT_PATHS["review_summary"])

state_path = d.get("refine_state") or DEFAULT_PATHS["refine_state"]
try:
    state = json.load(open(state_path, encoding="utf-8"))
except Exception as e:
    fail_en(f"{state_path} invalid JSON: {e}")
if state.get("status") != "completed":
    fail_en(f'{state_path} status={state.get("status")!r}, need "completed"')
if state.get("phase") != "done":
    fail_en(f'{state_path} phase={state.get("phase")!r}, need "done"')
state_round = state.get("round")
if not isinstance(state_round, int) or state_round != R:
    fail_en(f"REFINE_STATE.round={state_round!r} must equal summary.rounds_completed={R}")
if state_round > MAX_REVIEW_ROUNDS:
    fail_en(f"REFINE_STATE.round exceeds cap {MAX_REVIEW_ROUNDS}")

ic_path = "idea-stage/idea-creator.manifest.result.json"
if os.path.isfile(ic_path):
    ic = json.load(open(ic_path, encoding="utf-8"))
    ideas = ic.get("ideas") or []
    r1 = next((x for x in ideas if x.get("rank") == 1), None)
    if not r1:
        fail_en("idea-creator manifest has no rank 1")
    if mf.get("idea_title") != r1.get("title"):
        fail_en("idea_title mismatch vs idea-creator rank 1")

print(f"RESEARCH-REFINE OK [{CONTRACT}]")
PY
