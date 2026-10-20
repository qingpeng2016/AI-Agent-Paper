#!/usr/bin/env bash
# Optional — from project root: bash aris-constraints/translate-idea-report-literature-zh.aris.example.sh [root]
# Translates ## Literature Landscape from idea-stage/IDEA_REPORT.md → idea-stage/LITERATURE_LANDSCAPE.zh.md
set -euo pipefail
cd "${1:-.}"

exec aris prompt "$(cat <<'EOF'
请阅读 idea-stage/IDEA_REPORT.md 的 ## Literature Landscape 一节，译为简体中文，写入 idea-stage/LITERATURE_LANDSCAPE.zh.md。不要修改 IDEA_REPORT.md。 — language: cn
EOF
)"
