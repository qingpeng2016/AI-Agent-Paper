#!/usr/bin/env bash
# 前端（CentOS7 无法在服务器跑 Node）：本机构建 dist → git push → 服务器 git pull
#
# 用法: ./deploy-frontend.sh <分支> [动作]
#   deploy(默认) = 本机 build + 提交 dist + push + 服务器 pull
#   pull         = 仅服务器同步 dist（dist 已在 GitHub 时）
#   status       = 查看服务器 dist/index.html
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=deploy-common.sh
source "${SCRIPT_DIR}/deploy-common.sh"
deploy_common_init

usage_frontend() {
  cat <<'EOF'
用法: ./deploy-frontend.sh <分支> [动作]

动作:
  deploy(默认)  本机 pnpm build → git push dist → 服务器 git pull
  pull          仅服务器 git pull（先 checkout .）
  status        查看服务器 dist/index.html
EOF
  exit 1
}

BRANCH="${1:-}"
[[ -n "$BRANCH" ]] || usage_frontend

ACTION="${2:-deploy}"
case "$ACTION" in
  deploy|pull|status) ;;
  *) usage_frontend ;;
esac

write_vite_production_env() {
  [[ -n "${VITE_API_BASE_URL}" ]] || {
    echo "ERROR: 请在 deploy/.deploy.env 设置 VITE_API_BASE_URL（如 https://api.niceboxs.com）" >&2
    exit 1
  }
  local web_pc="${LOCAL_REPO_DIR}/front/apps/web-pc"
  local env_file="${web_pc}/.env.production"
  echo ">>> [本机] 写入 ${env_file} ← VITE_API_BASE_URL=${VITE_API_BASE_URL}"
  cat >"$env_file" <<ENV
# 由 deploy-frontend.sh 根据 deploy/.deploy.env 生成，打包后生效
VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV
}

local_build_and_push_dist() {
  local repo="$LOCAL_REPO_DIR"
  local web_dir="${repo}/${MALL_WEB_ROOT}"
  local dist_rel="$MALL_WEB_DIST"
  local dist_abs="${repo}/${MALL_WEB_DIST}"

  local_git_sync "$repo" "$BRANCH"

  write_vite_production_env

  [[ -d "$web_dir" ]] || {
    echo "ERROR: 目录不存在: $web_dir" >&2
    exit 1
  }

  cd "$web_dir"
  ensure_local_node_pnpm
  echo ">>> [本机] ${WEB_BUILD_CMD}"
  eval "$WEB_BUILD_CMD"

  [[ -f "${dist_abs}/index.html" ]] || {
    echo "ERROR: 本机构建失败，无 ${dist_abs}/index.html" >&2
    exit 1
  }
  echo ">>> [本机] 构建完成: ${dist_abs}/index.html"

  cd "$repo"
  git add -f "$dist_rel"
  if git diff --cached --quiet; then
    echo ">>> [本机] dist 无变化，跳过 commit"
  else
    git commit -m "${FRONTEND_DIST_COMMIT_MSG} $(date +%Y-%m-%d_%H:%M)"
    echo ">>> [本机] 已 commit dist"
  fi

  echo ">>> [本机] git push origin ${BRANCH}"
  git push origin "$BRANCH"
}

REMOTE_LIB="$(deploy_remote_lib)"

server_git_pull() {
  read -r -d '' REMOTE_SCRIPT_BODY <<EOF || true
set -euo pipefail
APP_DIR="${APP_DIR}"
BRANCH="${BRANCH}"
GIT_FETCH_TIMEOUT="${GIT_FETCH_TIMEOUT}"
GIT_CLEAN_BIN="0"
MALL_WEB_DIST="${MALL_WEB_DIST}"
dist_path="\${APP_DIR}/\${MALL_WEB_DIST}"

${REMOTE_LIB}

git_sync_branch

if [[ ! -f "\${dist_path}/index.html" ]]; then
  echo "ERROR: [服务器] pull 后无 \${dist_path}/index.html" >&2
  exit 1
fi
echo ">>> [服务器] 前端静态资源已就绪"
ls -lh "\${dist_path}/index.html"
EOF
  run_remote_bash "$REMOTE_SCRIPT_BODY"
}

case "$ACTION" in
  deploy)
    local_build_and_push_dist
    server_git_pull
    echo ">>> 完成（Nginx: ${APP_DIR}/${MALL_WEB_DIST}）"
    ;;
  pull)
    server_git_pull
    echo ">>> 服务器 git pull 完成"
    ;;
  status)
    run_ssh "ls -lh '${APP_DIR}/${MALL_WEB_DIST}/index.html' 2>/dev/null || echo '>>> 无 dist，执行: ./deploy-frontend.sh master deploy'"
    ;;
esac
