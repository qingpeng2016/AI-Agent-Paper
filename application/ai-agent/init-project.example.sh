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

mkdir -p "$ROOT/idea-stage" "$ROOT/refine-logs" "$ROOT/literature" "$ROOT/aris-constraints" "$ROOT/research-wiki" "$ROOT/data-source"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEST="$ROOT/aris-constraints"
COPIED=0
for src in "$SCRIPT_DIR"/all-init-project.example.sh; do
  [[ -f "$src" ]] || continue
  cp -f "$src" "$DEST/"
  COPIED=$((COPIED + 1))
done
if [[ -d "$SCRIPT_DIR/md_template" ]]; then
  mkdir -p "$DEST/md_template"
  # Copy all md_template files (including /experiment-plan scaffolds at top level)
  cp -f "$SCRIPT_DIR"/md_template/* "$DEST/md_template/" 2>/dev/null || true
fi
if [[ -d "$SCRIPT_DIR/skills" ]]; then
  mkdir -p "$DEST/skills"
  cp -R "$SCRIPT_DIR/skills/"* "$DEST/skills/" 2>/dev/null || true
  # Mirror for ARIS slash-command discovery (canonical copy: aris-constraints/skills/)
  mkdir -p "$ROOT/skills"
  cp -R "$SCRIPT_DIR/skills/"* "$ROOT/skills/" 2>/dev/null || true
  find "$DEST/skills" -name '*.sh' -exec chmod +x {} \; 2>/dev/null || true
fi
EN_REPORT_TEMPLATE="$DEST/md_template/IDEA_REPORT.en.md"
ZH_REPORT_TEMPLATE="$DEST/md_template/IDEA_REPORT.zh.md"
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
基于 Sentinel-2 与 Landsat 地表温度产品，获取武汉市夏季典型日期的热环境空间分布；结合土地利用/本地气候分区（LCZ），采用 GIS 分区统计与空间相关分析，比较不同下垫面类型间地表温度差异，并探讨植被指数、不透水面比例、水体距离等因素对热岛格局的影响，为城市降温规划提供空间依据。

## Keywords
urban heat island, land surface temperature, Sentinel-2, GIS, remote sensing, Wuhan

## Target venue
Applied Geography
EOF

echo "Initialized: $ROOT"
echo "  dirs: idea-stage refine-logs literature aris-constraints research-wiki data-source skills"
echo "  file: RESEARCH_BRIEF.md"
echo "  report templates: aris-constraints/md_template/IDEA_REPORT.en.md, IDEA_REPORT.zh.md"
echo "  experiment-plan templates: md_template/EXPERIMENT_PLAN.en|zh.md, EXPERIMENT_TRACKER.en|zh.md"
echo "  skill bundles: aris-constraints/skills/<name>/ (*.aris.sh, *.desc, *.validate.sh; mirror to project/skills/ for ARIS)"
echo "  data profile templates: md_template/DATA_SOURCE_PROFILE.en|zh.md"
echo "  idea-stage: IDEA_REPORT.en.md, IDEA_REPORT.zh.md (working copies from templates)"
echo "  copied: all-init + md_template + skills/ ($COPIED top-level script(s) from $SCRIPT_DIR)"
