#!/usr/bin/env bash
# 本机仅 SSH；服务器：进仓库 → 切分支拉代码 → go build → 启停 API/Bot
#
# 用法: ./deploy-backend.sh <分支> [deploy|restart|stop|status] [user|bot]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=deploy-common.sh
source "${SCRIPT_DIR}/deploy-common.sh"
deploy_common_init

usage_backend() {
  cat <<'EOF'
用法: ./deploy-backend.sh <分支> [动作] [模式]
  动作: deploy(默认) | restart | stop | status
  模式: user | api (HTTP) | bot | robot
EOF
  exit 1
}

BRANCH="${1:-}"
[[ -n "$BRANCH" ]] || usage_backend

parse_action_mode_args "${2:-deploy}" "${3:-user}"
MODE="$(resolve_backend_mode "$MODE")"
backend_run_vars "$MODE"

REMOTE_LIB="$(deploy_remote_lib)"

read -r -d '' REMOTE_SCRIPT_BODY <<EOF || true
set -euo pipefail

APP_DIR="${APP_DIR}"
BRANCH="${BRANCH}"
ACTION="${ACTION}"
MODE="${MODE}"
RUN_CONF="${RUN_CONF}"
RUN_BOT="${RUN_BOT}"
LISTEN_PORT="${LISTEN_PORT}"
MALL_BIN_NAME="${MALL_BIN_NAME}"
GIT_FETCH_TIMEOUT="${GIT_FETCH_TIMEOUT}"
GIT_CLEAN_BIN="1"

LOG_DIR="\${APP_DIR}/logs"
BIN_DIR="\${APP_DIR}/bin"
PID_NAME="ai-token-mall-\${MODE}"
PID_FILE="\${LOG_DIR}/\${PID_NAME}.pid"
LOG_FILE="\${LOG_DIR}/\${PID_NAME}-nohup.log"

mkdir -p "\${LOG_DIR}" 2>/dev/null || true

${REMOTE_LIB}

proc_cwd() {
  local pid="\$1"
  readlink -f "/proc/\${pid}/cwd" 2>/dev/null || true
}

is_mall_go_cmd() {
  local cmd="\$1"
  [[ "\${cmd}" == *"main.go"* || "\${cmd}" == *"/bin/\${MALL_BIN_NAME}"* || "\${cmd}" == *"\${MALL_BIN_NAME}"* || "\${cmd}" == *"ai-token-mall"* || "\${cmd}" == *"--run_conf="* ]]
}

cmdline_matches_mode() {
  local pid="\$1"
  local cmd
  cmd=\$(tr '\\0' ' ' <"/proc/\${pid}/cmdline" 2>/dev/null || true)
  [[ -n "\${cmd}" ]] || return 1
  is_mall_go_cmd "\${cmd}" || return 1
  case "\${MODE}" in
    user|api)
      [[ "\${cmd}" == *"--run_bot=false"* || "\${cmd}" == *"--run_bot=0"* ]] || return 1
      [[ "\${cmd}" == *"--run_conf=\${RUN_CONF}"* ]] || return 1
      ;;
    bot|robot)
      [[ "\${cmd}" == *"--run_bot=true"* || "\${cmd}" == *"--run_bot=1"* ]] || return 1
      [[ "\${cmd}" == *"--run_conf=\${RUN_CONF}"* ]] || return 1
      ;;
  esac
  return 0
}

pids_on_port() {
  local port="\$1"
  local pids=""
  if command -v ss >/dev/null 2>&1; then
    pids=\$(ss -lptn "sport = :\${port}" 2>/dev/null | grep -o 'pid=[0-9]*' | cut -d= -f2 | sort -u || true)
  elif command -v lsof >/dev/null 2>&1; then
    pids=\$(lsof -ti ":\${port}" 2>/dev/null | sort -u || true)
  fi
  echo "\${pids}"
}

kill_port_listeners() {
  local port="\$1" sig="\${2:-TERM}" pid
  [[ -n "\${port}" ]] || return 0
  for pid in \$(pids_on_port "\${port}"); do
    [[ -n "\${pid}" ]] || continue
    kill "-\${sig}" "\${pid}" 2>/dev/null || true
  done
}

port_in_use() {
  local port="\$1"
  [[ -n "\$(pids_on_port "\${port}" | tr -d '[:space:]')" ]]
}

build_mall_binary() {
  cd "\${APP_DIR}"
  mkdir -p "\${BIN_DIR}"
  local bin_path="\${BIN_DIR}/\${MALL_BIN_NAME}"
  export PATH="\${PATH}:/usr/local/go/bin"
  export GO111MODULE=on
  export GOPROXY="\${GOPROXY:-https://goproxy.cn,direct}"
  echo ">>> [服务器] go build -o \${bin_path} ."
  go build -o "\${bin_path}" .
}

find_matching_pid() {
  local pid
  for pid_path in /proc/[0-9]*; do
    pid=\${pid_path#/proc/}
    cmdline_matches_mode "\${pid}" || continue
    echo "\${pid}"
    return 0
  done
  return 1
}

kill_matching() {
  local sig="\${1:-TERM}" pid
  for pid_path in /proc/[0-9]*; do
    pid=\${pid_path#/proc/}
    cmdline_matches_mode "\${pid}" || continue
    kill "-\${sig}" "\${pid}" 2>/dev/null || true
  done
}

stop_mall() {
  echo ">>> [服务器] 停止 ai-token-mall/\${MODE}"
  if [[ -f "\${PID_FILE}" ]]; then
    old_pid=\$(tr -d '[:space:]' <"\${PID_FILE}" 2>/dev/null || true)
    kill -TERM "\${old_pid}" 2>/dev/null || true
    rm -f "\${PID_FILE}"
  fi
  kill_matching TERM
  if [[ "\${RUN_BOT}" == "false" && -n "\${LISTEN_PORT}" ]]; then
    kill_port_listeners "\${LISTEN_PORT}" TERM
  fi
  sleep 2
  kill_matching KILL
  if [[ "\${RUN_BOT}" == "false" && -n "\${LISTEN_PORT}" ]]; then
    kill_port_listeners "\${LISTEN_PORT}" KILL
    sleep 1
    if port_in_use "\${LISTEN_PORT}"; then
      echo "ERROR: [服务器] 端口 \${LISTEN_PORT} 仍被占用，无法启动 API：" >&2
      ss -lptn "sport = :\${LISTEN_PORT}" 2>/dev/null >&2 || lsof -i ":\${LISTEN_PORT}" 2>/dev/null >&2 || true
      echo "ERROR: 请 SSH 登录后执行: kill \$(lsof -ti:\${LISTEN_PORT}) 或改 .deploy.env 的 MALL_USER_PORT" >&2
      exit 1
    fi
    echo ">>> [服务器] 端口 \${LISTEN_PORT} 已释放"
  fi
}

show_mall_status() {
  local found=0 pid cmd cwd
  for pid_path in /proc/[0-9]*; do
    pid=\${pid_path#/proc/}
    cmdline_matches_mode "\${pid}" || continue
    found=1
    cmd=\$(tr '\\0' ' ' <"/proc/\${pid}/cmdline" 2>/dev/null || true)
    cwd=\$(proc_cwd "\${pid}")
    echo ">>> RUNNING pid=\${pid} cwd=\${cwd}"
    echo ">>>   \${cmd}"
  done
  [[ -f "\${LOG_FILE}" ]] && echo ">>> 日志: \${LOG_FILE}"
  [[ "\${found}" -eq 0 ]] && echo ">>> 无匹配进程"
}

wait_for_mall() {
  local i pid
  echo ">>> [服务器] 等待进程就绪（最多 30s）..."
  for i in \$(seq 1 30); do
    pid="\$(find_matching_pid || true)"
    if [[ -n "\${pid}" ]] && kill -0 "\${pid}" 2>/dev/null; then
      if [[ "\${RUN_BOT}" == "true" ]] || port_in_use "\${LISTEN_PORT}"; then
        show_mall_status
        return 0
      fi
    fi
    if [[ -f "\${LOG_FILE}" ]] && grep -qE 'panic:|FATAL' "\${LOG_FILE}"; then
      echo "ERROR: [服务器] 启动失败，日志:" >&2
      tail -n 50 "\${LOG_FILE}" >&2
      exit 1
    fi
    sleep 1
  done
  echo "ERROR: [服务器] 启动超时（端口 \${LISTEN_PORT} 未监听？检查 conf/prod.conf.yaml 的 server.port）" >&2
  tail -n 30 "\${LOG_FILE}" 2>/dev/null >&2 || true
  exit 1
}

start_mall() {
  cd "\${APP_DIR}"
  mkdir -p "\${BIN_DIR}" "\${LOG_DIR}"
  local bin_path="\${BIN_DIR}/\${MALL_BIN_NAME}"
  [[ -x "\${bin_path}" ]] || build_mall_binary
  : >"\${LOG_FILE}"
  export PATH="\${PATH}:/usr/local/go/bin"
  echo ">>> [服务器] 启动 \${bin_path} --run_conf=\${RUN_CONF} --run_bot=\${RUN_BOT}"
  nohup "\${bin_path}" --run_conf="\${RUN_CONF}" --run_bot="\${RUN_BOT}" >>"\${LOG_FILE}" 2>&1 &
  echo \$! >"\${PID_FILE}"
  wait_for_mall
}

case "\${ACTION}" in
  deploy)
    git_sync_branch
    build_mall_binary
    stop_mall
    start_mall
    ;;
  restart)
    build_mall_binary
    stop_mall
    start_mall
    ;;
  stop) stop_mall ;;
  status) show_mall_status ;;
  *) exit 1 ;;
esac
EOF

echo ">>> 后端 分支=${BRANCH} 动作=${ACTION} 模式=${MODE}"
if ! run_remote_bash "$REMOTE_SCRIPT_BODY"; then
  echo ">>> 远程脚本返回非 0（若上方已有 RUNNING，可先: ./deploy-backend.sh master status user）" >&2
  echo ">>> 服务器日志: ${APP_DIR}/logs/ai-token-mall-${MODE}-nohup.log" >&2
  exit 1
fi
echo ">>> 后端完成"
