# shellcheck shell=bash
# 本机只做 SSH 跳板；git / build / 启停均在服务器 MALL_APP_DIR 执行。
# SSH：优先 sshpass，否则 expect（与 box 项目 remote-deploy.sh 相同）

deploy_common_init() {
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  ENV_FILE="${SCRIPT_DIR}/.deploy.env"
  REMOTE_LIB_FILE="${SCRIPT_DIR}/remote-lib.sh"

  [[ -f "$ENV_FILE" ]] || {
    echo "缺少 ${ENV_FILE}，请先: cp .deploy.env.example .deploy.env" >&2
    exit 1
  }
  [[ -f "$REMOTE_LIB_FILE" ]] || {
    echo "缺少 ${REMOTE_LIB_FILE}" >&2
    exit 1
  }

  # shellcheck source=/dev/null
  source "$ENV_FILE"

  : "${SSH_HOST:?请在 .deploy.env 设置 SSH_HOST}"
  : "${SSH_USER:?请在 .deploy.env 设置 SSH_USER}"
  : "${SSH_PASS:?请在 .deploy.env 设置 SSH_PASS}"

  SSH_PORT="${SSH_PORT:-22}"
  MALL_APP_DIR="${MALL_APP_DIR:-/www/wwwroot/AI-Token-Mall}"
  MALL_USER_RUN_CONF="${MALL_USER_RUN_CONF:-prod}"
  MALL_BOT_RUN_CONF="${MALL_BOT_RUN_CONF:-prod}"
  MALL_USER_PORT="${MALL_USER_PORT:-8886}"
  MALL_WEB_ROOT="${MALL_WEB_ROOT:-front/apps}"
  MALL_WEB_DIST="${MALL_WEB_DIST:-front/apps/web-pc/dist}"
  WEB_BUILD_CMD="${WEB_BUILD_CMD:-pnpm install && pnpm build:pc}"
  # 生产打包写入 web-pc/.env.production（路径含 /api/v1，base 不要末尾斜杠）
  VITE_API_BASE_URL="${VITE_API_BASE_URL:-}"
  MALL_BIN_NAME="${MALL_BIN_NAME:-ai-token-mall}"
  GIT_FETCH_TIMEOUT="${GIT_FETCH_TIMEOUT:-120}"
  LOCAL_REPO_DIR="${LOCAL_REPO_DIR:-$(cd "$SCRIPT_DIR/.." && pwd)}"
  FRONTEND_DIST_COMMIT_MSG="${FRONTEND_DIST_COMMIT_MSG:-chore(deploy): web-pc dist}"

  APP_DIR="$MALL_APP_DIR"
  SSH_TARGET="${SSH_USER}@${SSH_HOST}"

  SSH_OPTS=(
    -o StrictHostKeyChecking=accept-new
    -o PreferredAuthentications=password
    -o PubkeyAuthentication=no
    -p "$SSH_PORT"
  )
}

deploy_remote_lib() {
  cat "${SCRIPT_DIR}/remote-lib.sh"
}

# 远程执行一条 shell 命令（已拼好、由调用方转义）
_deploy_ssh_run() {
  local remote_cmd="$1"
  local rc=0

  if command -v sshpass >/dev/null 2>&1; then
    sshpass -p "$SSH_PASS" ssh "${SSH_OPTS[@]}" "$SSH_TARGET" "$remote_cmd"
    rc=$?
  elif command -v expect >/dev/null 2>&1; then
    # shellcheck disable=SC2016
    expect <<EOF
set timeout 600
log_user 1
spawn ssh -o StrictHostKeyChecking=accept-new -o PreferredAuthentications=password -o PubkeyAuthentication=no -p ${SSH_PORT} ${SSH_TARGET} "${remote_cmd}"
expect {
  -re "(?i)password:" {
    send "${SSH_PASS}\r"
    exp_continue
  }
  eof
}
set wait_result [wait]
set exit_code [lindex \$wait_result 3]
if {\$exit_code eq ""} { set exit_code 0 }
exit \$exit_code
EOF
    rc=$?
  else
    echo "未找到 sshpass 或 expect。Mac 一般自带 expect；或: brew install hudochenkov/sshpass/sshpass" >&2
    return 1
  fi
  return "$rc"
}

run_ssh() {
  local remote_cmd="$*"
  _deploy_ssh_run "$remote_cmd"
}

run_remote_bash() {
  local script_body="$1"
  echo ">>> SSH ${SSH_TARGET} — 以下在服务器 ${APP_DIR} 执行"
  local b64 one_liner
  b64=$(printf '%s' "$script_body" | base64 | tr -d '\n')
  one_liner="echo ${b64} | base64 -d | bash"
  _deploy_ssh_run "$one_liner"
}

resolve_backend_mode() {
  local mode="$1"
  case "$mode" in
    user|api) echo "user" ;;
    bot|robot) echo "bot" ;;
    *)
      echo "ERROR: 模式只能是 user/api 或 bot/robot，当前: ${mode}" >&2
      exit 1
      ;;
  esac
}

backend_run_vars() {
  local mode="$1"
  case "$mode" in
    user|api)
      RUN_CONF="$MALL_USER_RUN_CONF"
      RUN_BOT="false"
      LISTEN_PORT="$MALL_USER_PORT"
      ;;
    bot|robot)
      RUN_CONF="$MALL_BOT_RUN_CONF"
      RUN_BOT="true"
      LISTEN_PORT=""
      ;;
  esac
}

parse_action_mode_args() {
  ACTION="${1:-deploy}"
  MODE="${2:-user}"
  case "$ACTION" in
    user|api|bot|robot)
      MODE="$ACTION"
      ACTION="deploy"
      ;;
  esac
}

ensure_local_node_pnpm() {
  if ! command -v node >/dev/null 2>&1; then
    echo "ERROR: 本机未安装 node（Mac: brew install node）" >&2
    exit 1
  fi
  echo ">>> [本机] node $(node -v)"
  if command -v pnpm >/dev/null 2>&1; then
    echo ">>> [本机] pnpm $(pnpm -v)"
    return 0
  fi
  if command -v corepack >/dev/null 2>&1; then
    corepack enable 2>/dev/null || true
  fi
  if command -v pnpm >/dev/null 2>&1; then
    echo ">>> [本机] pnpm $(pnpm -v)"
    return 0
  fi
  npm install -g pnpm@9
  echo ">>> [本机] pnpm $(pnpm -v)"
}

local_git_sync() {
  local repo="$1"
  local branch="$2"
  cd "$repo"
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "ERROR: 不是 git 仓库: $repo" >&2
    exit 1
  fi
  echo ">>> [本机] git sync ($branch) @ $repo"
  export GIT_TERMINAL_PROMPT=0
  git fetch origin --prune
  if git show-ref --verify --quiet "refs/heads/${branch}"; then
    git checkout "$branch"
  else
    git checkout -B "$branch" "origin/${branch}"
  fi
  git pull --ff-only origin "$branch" || {
    echo "ERROR: 本机 git pull 失败，请先处理本地分支" >&2
    exit 1
  }
  git log -1 --oneline
}
