#!/usr/bin/env bash
# Optional — from project root: bash aris-constraints/translate-idea-report-full-zh.aris.example.sh [root]
# Translates entire idea-stage/IDEA_REPORT.md → idea-stage/IDEA_REPORT.zh.md
set -euo pipefail
cd "${1:-.}"

exec aris prompt "$(cat <<'EOF'
请阅读 idea-stage/IDEA_REPORT.md，全文译为简体中文，写入 idea-stage/IDEA_REPORT.zh.md。不要修改英文原版。 — language: cn
EOF
)"
