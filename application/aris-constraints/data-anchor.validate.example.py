#!/usr/bin/env python3
"""Shared data-anchor gate for Paper step validates. Usage:
  python3 aris-constraints/data-anchor.validate.example.py <step_tag> <step_manifest.json> [text_file ...]
If data-source/DATA_SOURCE_PROFILE.en.md missing, exit 0.
"""
from __future__ import annotations

import json
import os
import sys

PROFILE = "data-source/DATA_SOURCE_PROFILE.en.md"
DISCOVERY = "data-source/data-source-discovery.manifest.result.json"


def fail(step: str, msg: str) -> None:
    print(f"{step} DATA-ANCHOR FAIL: {msg}", file=sys.stderr)
    sys.exit(1)


def main() -> None:
    if len(sys.argv) < 3:
        fail("?", "usage: data-anchor.validate.example.py <STEP_TAG> <manifest.json> [deliverable ...]")
    step = sys.argv[1]
    manifest_path = sys.argv[2]
    texts = sys.argv[3:]

    if not os.path.isfile(PROFILE):
        sys.exit(0)

    if not os.path.isfile(DISCOVERY):
        fail(step, f"{DISCOVERY} missing while {PROFILE} exists")

    try:
        dm = json.load(open(DISCOVERY, encoding="utf-8"))
    except Exception as e:
        fail(step, f"invalid {DISCOVERY}: {e}")
    if dm.get("result") != "success":
        fail(step, f"{DISCOVERY} result must be success when profile exists")

    pid = dm.get("primary_recommendation_id") or ""
    primary = next(
        (d for d in (dm.get("datasets") or []) if d.get("id") == pid),
        {},
    )
    needles = {pid, (primary.get("stable_id") or ""), (primary.get("name") or "")}
    needles = {n.strip() for n in needles if n and str(n).strip()}

    try:
        m = json.load(open(manifest_path, encoding="utf-8"))
    except Exception as e:
        fail(step, f"invalid manifest {manifest_path}: {e}")

    raw_anchor = m.get("data_anchor")
    if raw_anchor is None:
        fail(step, "manifest missing data_anchor object (profile exists)")
    if isinstance(raw_anchor, str):
        fail(step, f"data_anchor must be a JSON object, not string {raw_anchor!r}")
    if not isinstance(raw_anchor, dict):
        fail(step, f"data_anchor must be object, got {type(raw_anchor).__name__}")
    anchor = raw_anchor
    if not anchor.get("primary_dataset_id"):
        fail(step, "manifest missing data_anchor.primary_dataset_id (profile exists)")
    if anchor.get("profile_en") and anchor["profile_en"] != PROFILE:
        fail(step, f"data_anchor.profile_en must be {PROFILE!r}")

    label = (anchor.get("primary_label") or "").strip()
    if label:
        needles.add(label)

    if not texts:
        sys.exit(0)

    blob = ""
    for p in texts:
        if os.path.isfile(p):
            blob += open(p, encoding="utf-8", errors="replace").read().lower()

    if not needles:
        fail(step, "could not derive primary dataset tokens from discovery manifest")

    if not any(n.lower() in blob for n in needles if len(n) >= 3):
        fail(
            step,
            f"deliverables must mention primary dataset ({', '.join(sorted(needles)[:4])} …); "
            f"adapt work to {PROFILE}",
        )

    sys.exit(0)


if __name__ == "__main__":
    main()
