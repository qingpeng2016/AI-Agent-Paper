#!/usr/bin/env bash
# Initialize idea-discovery workspace — run: bash aris-constraints/all-init-project.example.sh [project_root]
# Default project root: ~/my-paper-new
set -euo pipefail

ROOT="${1:-$HOME/my-paper-new}"
mkdir -p "$ROOT/idea-stage" "$ROOT/refine-logs" "$ROOT/literature" "$ROOT/aris-constraints" "$ROOT/research-wiki"

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
