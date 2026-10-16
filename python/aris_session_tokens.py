#!/usr/bin/env python3
"""List per-round token usage from ARIS / Claude session JSON under a project root."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any


def find_usage(obj: Any) -> dict[str, Any] | None:
    if isinstance(obj, dict):
        keys = set(obj.keys())
        has_in = "input_tokens" in keys or "prompt_tokens" in keys
        has_out = "output_tokens" in keys or "completion_tokens" in keys
        if has_in and has_out:
            return obj
        for v in obj.values():
            found = find_usage(v)
            if found is not None:
                return found
    elif isinstance(obj, list):
        for item in obj:
            found = find_usage(item)
            if found is not None:
                return found
    return None


def token_counts(usage: dict[str, Any]) -> tuple[int, int]:
    inp = int(usage.get("input_tokens") or usage.get("prompt_tokens") or 0)
    out = int(usage.get("output_tokens") or usage.get("completion_tokens") or 0)
    return inp, out


def summarize_tools(blocks: list[Any] | None) -> str:
    if not blocks:
        return "(text)"
    parts: list[str] = []
    for b in blocks:
        if not isinstance(b, dict) or b.get("type") != "tool_use":
            continue
        name = str(b.get("name") or "?")
        raw_in = b.get("input")
        inp: Any = raw_in
        if isinstance(raw_in, str):
            try:
                inp = json.loads(raw_in)
            except json.JSONDecodeError:
                inp = {}
        if name == "Skill" and isinstance(inp, dict):
            parts.append(f"Skill:{inp.get('skill', '')}")
        elif name == "bash" and isinstance(inp, dict):
            cmd = str(inp.get("command") or "")[:60].replace("\n", " ")
            parts.append(f"bash:{cmd}")
        else:
            parts.append(name)
    return ", ".join(parts) if parts else "(text)"


def parse_session(path: Path) -> list[dict[str, Any]]:
    with path.open(encoding="utf-8") as f:
        data = json.load(f)
    rows: list[dict[str, Any]] = []
    for m in data.get("messages") or []:
        if m.get("role") != "assistant":
            continue
        usage = find_usage(m)
        if usage is None:
            continue
        inp, out = token_counts(usage)
        rows.append(
            {
                "turn": len(rows) + 1,
                "input_tokens": inp,
                "output_tokens": out,
                "total_tokens": inp + out,
                "action": summarize_tools(m.get("blocks")),
            }
        )
    return rows


def print_session_table(path: Path, rows: list[dict[str, Any]]) -> None:
    print(f"\n=== {path.name} ({len(rows)} API rounds) ===")
    if not rows:
        print("(no usage in this session file)")
        return
    print(f"{'#':>3}  {'input':>10}  {'output':>8}  {'total':>10}  action")
    print("-" * 90)
    tin = tout = 0
    for r in rows:
        tin += r["input_tokens"]
        tout += r["output_tokens"]
        act = str(r["action"])[:70]
        print(
            f"{r['turn']:3d}  {r['input_tokens']:10,}  {r['output_tokens']:8,}  "
            f"{r['total_tokens']:10,}  {act}"
        )
    print("-" * 90)
    print(f"{'Σ':>3}  {tin:10,}  {tout:8,}  {tin + tout:10,}")


def _session_text(path: Path) -> str:
    try:
        return path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return ""


def _matches_project(path: Path, root: Path) -> bool:
    text = _session_text(path)
    if not text:
        return False
    root_s = str(root)
    return root_s in text or f"/{root.name}/" in text or f"/{root.name}\"" in text


def collect_session_candidates(root: Path, include_global: bool) -> list[Path]:
    seen: set[Path] = set()
    out: list[Path] = []
    for base in (root / ".claude" / "sessions",):
        if base.is_dir():
            for p in base.glob("session-*.json"):
                rp = p.resolve()
                if rp not in seen:
                    seen.add(rp)
                    out.append(p)
    if include_global:
        global_dir = Path.home() / ".claude" / "sessions"
        if global_dir.is_dir():
            for p in global_dir.glob("session-*.json"):
                rp = p.resolve()
                if rp not in seen:
                    seen.add(rp)
                    out.append(p)
    return out


def session_paths(
    root: Path,
    all_sessions: bool,
    include_global: bool,
) -> list[Path]:
    candidates = collect_session_candidates(root, include_global)
    if not candidates:
        return []

    if all_sessions:
        proj = _project_sessions_dir(root)
        project_only = sorted(
            [p for p in candidates if p.parent.resolve() == proj],
            key=lambda p: p.stat().st_mtime,
        )
        if project_only:
            return project_only
        matched = sorted(
            [p for p in candidates if _matches_project(p, root)],
            key=lambda p: p.stat().st_mtime,
        )
        return matched or sorted(candidates, key=lambda p: p.stat().st_mtime)

    def rank(p: Path) -> tuple[int, int, int, float]:
        rows = parse_session(p)
        usage = 1 if rows else 0
        match = 1 if _matches_project(p, root) else 0
        in_project = 1 if p.parent.resolve() == _project_sessions_dir(root) else 0
        return (usage, match, in_project, p.stat().st_mtime)

    best = max(candidates, key=rank)
    if parse_session(best):
        return [best]
    # Newest under project even if empty (caller may print hint).
    proj_dir = root / ".claude" / "sessions"
    if proj_dir.is_dir():
        under = sorted(proj_dir.glob("session-*.json"), key=lambda p: p.stat().st_mtime, reverse=True)
        if under:
            return [under[0]]
    return [best]


def _project_sessions_dir(root: Path) -> Path:
    return (root / ".claude" / "sessions").resolve()


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Per-round token usage from .claude/sessions/*.json (ARIS REPL)."
    )
    parser.add_argument(
        "project_dir",
        type=Path,
        help="ARIS project root (contains .claude/sessions/)",
    )
    parser.add_argument(
        "--all-sessions",
        action="store_true",
        help="Print every session file (default: newest only)",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Emit JSON instead of a text table",
    )
    parser.add_argument(
        "--include-global",
        action="store_true",
        help="Also search ~/.claude/sessions (if you started aris outside project dir)",
    )
    args = parser.parse_args()
    root = args.project_dir.expanduser().resolve()
    if not root.is_dir():
        print(f"error: not a directory: {root}", file=sys.stderr)
        return 1

    include_global = args.include_global
    paths = session_paths(root, args.all_sessions, include_global)
    if not paths:
        print(
            f"error: no session logs under {root / '.claude' / 'sessions'}",
            file=sys.stderr,
        )
        print("Run idea-discovery in this directory first, then retry.", file=sys.stderr)
        return 1

    report: dict[str, Any] = {"project_dir": str(root), "sessions": []}
    any_usage = False
    for path in paths:
        rows = parse_session(path)
        if rows:
            any_usage = True
        if not args.json and len(paths) == 1:
            print(f"(session file: {path})")
        if args.json:
            tin = sum(r["input_tokens"] for r in rows)
            tout = sum(r["output_tokens"] for r in rows)
            report["sessions"].append(
                {
                    "file": path.name,
                    "path": str(path),
                    "rounds": rows,
                    "totals": {
                        "input_tokens": tin,
                        "output_tokens": tout,
                        "total_tokens": tin + tout,
                    },
                }
            )
        else:
            print_session_table(path, rows)

    if not any_usage and not include_global:
        retry = session_paths(root, args.all_sessions, include_global=True)
        for path in retry:
            if parse_session(path) and _matches_project(path, root):
                if not args.json:
                    print(
                        "\nhint: project session has no usage yet; "
                        f"found tokens in {path}",
                        file=sys.stderr,
                    )
                    print_session_table(path, parse_session(path))
                any_usage = True
                break

    if not any_usage and not args.json:
        print(
            "\nhint: 0 rounds usually means: (1) REPL still running — use /cost; "
            "(2) start aris after `cd` project dir; (3) retry with --include-global",
            file=sys.stderr,
        )

    if args.json:
        print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
