#!/usr/bin/env bash
# Initialize idea-discovery workspace — run: bash aris-constraints/all-init-project.example.sh <project_name>
# Example: bash aris-constraints/all-init-project.example.sh my-paper-new  →  ~/my-paper-new
# Absolute path also allowed: bash aris-constraints/all-init-project.example.sh /data/my-paper-new
set -euo pipefail

if [[ $# -lt 1 || -z "${1// /}" ]]; then
  echo "usage: $0 <project_name>" >&2
  echo "  project_name: directory name under HOME (e.g. my-paper-new) or absolute path" >&2
  exit 1
fi

NAME="$1"
if [[ "$NAME" == /* ]]; then
  ROOT="$NAME"
else
  ROOT="$HOME/$NAME"
fi

mkdir -p "$ROOT/idea-stage" "$ROOT/refine-logs" "$ROOT/literature" "$ROOT/aris-constraints" "$ROOT/research-wiki"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEST="$ROOT/aris-constraints"
COPIED=0
for pattern in \
  '*.aris.example.sh' \
  '*.manifest.example.desc' \
  '*.explain.example.desc' \
  '*.validate.example.sh' \
  'validate-idea-discovery.example.sh' \
  'all-init-project.example.sh'
do
  for src in "$SCRIPT_DIR"/$pattern; do
    [[ -f "$src" ]] || continue
    cp -f "$src" "$DEST/"
    COPIED=$((COPIED + 1))
  done
done
if [[ -d "$SCRIPT_DIR/md_template" ]]; then
  mkdir -p "$DEST/md_template"
  cp -f "$SCRIPT_DIR"/md_template/* "$DEST/md_template/" 2>/dev/null || true
fi
EN_REPORT_TEMPLATE="$DEST/md_template/IDEA_REPORT.en.example.md"
ZH_REPORT_TEMPLATE="$DEST/md_template/IDEA_REPORT.zh.example.md"
if [[ -f "$EN_REPORT_TEMPLATE" ]]; then
  cp -f "$EN_REPORT_TEMPLATE" "$ROOT/idea-stage/IDEA_REPORT.en.md"
fi
if [[ -f "$ZH_REPORT_TEMPLATE" ]]; then
  cp -f "$ZH_REPORT_TEMPLATE" "$ROOT/idea-stage/IDEA_REPORT.zh.md"
fi
chmod +x "$DEST"/*.sh 2>/dev/null || true

cat > "$ROOT/RESEARCH_BRIEF.md" <<'EOF'
# Research Brief

## 研究内容
研究重度抑郁症（MDD）患者脑功能网络拓扑是否相对健康对照存在稳定、可重复的异常。基于静息态功能磁共振（rs-fMRI）采集 BOLD 信号，经头动校正、配准、标准化与滤波等预处理，以 AAL/功能分区或 data-driven parcellation 定义节点，构建复杂脑网络（Pearson/偏相关等功能连接），采用图论分析全局与局部拓扑（全局/局部效率、聚类系数、特征路径长度、小世界属性、模块化、枢纽与 rich-club 等），并检验与 HAMD 等临床严重程度及认知/情绪加工的相关或组间差异；关注默认模式网络、突显网络、边缘系统等子网络的拓扑改变及可解释性。

## Keywords
major depressive disorder, resting-state fMRI, functional connectivity, graph theory

## Target venue
NeuroImage
EOF

echo "Initialized: $ROOT"
echo "  dirs: idea-stage refine-logs literature aris-constraints research-wiki"
echo "  file: RESEARCH_BRIEF.md"
echo "  report templates: aris-constraints/md_template/IDEA_REPORT.en.example.md, IDEA_REPORT.zh.example.md"
echo "  idea-stage: IDEA_REPORT.en.md, IDEA_REPORT.zh.md (working copies from templates)"
echo "  templates: copied $COPIED file(s) from $SCRIPT_DIR -> $DEST"
