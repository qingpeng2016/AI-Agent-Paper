#!/usr/bin/env bash
# 本机仅 SSH；服务器：git checkout . → 拉分支 → go run 或 go build → 启停
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
PAPER_USE_GO_RUN="${PAPER_USE_GO_RUN}"
PAPER_GO_ROOT="${PAPER_GO_ROOT}"
PAPER_GO_VERSION="${PAPER_GO_VERSION:-1.22.5}"
GIT_FETCH_TIMEOUT="${GIT_FETCH_TIMEOUT}"
GIT_REPO_URL="${GIT_REPO_URL}"
GIT_CLEAN_BIN="1"

LOG_DIR="\${APP_DIR}/logs"
BIN_DIR="\${APP_DIR}/bin"
PID_NAME="ai-agent-paper-\${MODE}"
PID_FILE="\${LOG_DIR}/\${PID_NAME}.pid"
LOG_FILE="\${LOG_DIR}/\${PID_NAME}-nohup.log"

mkdir -p "\${LOG_DIR}" 2>/dev/null || true

${REMOTE_LIB}

proc_cwd() {
  local pid="\$1"
  readlink -f "/proc/\${pid}/cwd" 2>/dev/null || true
}

read_proc_cmdline() {
  local pid="\$1"
  [[ -r "/proc/\${pid}/cmdline" ]] || return 1
  tr '\\0' ' ' <"/proc/\${pid}/cmdline" 2>/dev/null || true
}

is_mall_go_cmd() {
  local cmd="\$1"
  [[ "\${cmd}" == *"main.go"* || "\${cmd}" == *"/bin/\${MALL_BIN_NAME}"* || "\${cmd}" == *"\${MALL_BIN_NAME}"* || "\${cmd}" == *"ai-agent-paper"* || "\${cmd}" == *"--run_conf="* ]]
}

# go run 编译后的子进程 cmdline 常无 --run_bot；用 cwd=APP_DIR + go-build/main 识别
cmdline_matches_mode() {
  local pid="\$1" cmd cwd
  cmd=\$(read_proc_cmdline "\${pid}" 2>/dev/null || true)
  cwd=\$(proc_cwd "\${pid}")
  if [[ -n "\${cmd}" ]] && is_mall_go_cmd "\${cmd}"; then
    case "\${MODE}" in
      user|api)
        [[ "\${cmd}" == *"--run_bot=false"* || "\${cmd}" == *"--run_bot=0"* ]] || return 1
        [[ "\${cmd}" == *"--run_conf=\${RUN_CONF}"* ]] || return 1
        ;;
      bot|robot)
        if [[ "\${cmd}" == *"--run_bot=true"* || "\${cmd}" == *"--run_bot=1"* ]]; then
          [[ "\${cmd}" == *"--run_conf=\${RUN_CONF}"* ]] || return 1
          return 0
        fi
        return 1
        ;;
    esac
    return 0
  fi
  if [[ "\${RUN_BOT}" == "true" && "\${cwd}" == "\${APP_DIR}" ]]; then
    if [[ "\${cmd}" == *"\${MALL_BIN_NAME}"* || "\${cmd}" == *go-build* || "\${cmd}" == *"/exe/main"* ]]; then
      return 0
    fi
  fi
  return 1
}

mall_bot_ready() {
  local pid
  pid="\$(find_matching_pid 2>/dev/null || true)"
  if [[ -n "\${pid}" ]] && kill -0 "\${pid}" 2>/dev/null; then
    return 0
  fi
  if [[ -f "\${PID_FILE}" ]]; then
    pid=\$(tr -d '[:space:]' <"\${PID_FILE}" 2>/dev/null || true)
    if [[ -n "\${pid}" ]] && kill -0 "\${pid}" 2>/dev/null; then
      return 0
    fi
  fi
  if [[ -f "\${LOG_FILE}" ]] && grep -qE 'bot scheduler-Started|bot Start|AI-Agent-Paper bot' "\${LOG_FILE}" 2>/dev/null; then
    if [[ -f "\${PID_FILE}" ]]; then
      pid=\$(tr -d '[:space:]' <"\${PID_FILE}" 2>/dev/null || true)
      [[ -n "\${pid}" ]] && kill -0 "\${pid}" 2>/dev/null && return 0
    fi
    pid="\$(find_matching_pid 2>/dev/null || true)"
    [[ -n "\${pid}" ]] && return 0
  fi
  return 1
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
  ensure_go
  cd "\${APP_DIR}"
  mkdir -p "\${BIN_DIR}"
  local bin_path="\${BIN_DIR}/\${MALL_BIN_NAME}"
  export GO111MODULE=on
  export GOPROXY="\${GOPROXY:-https://goproxy.cn,direct}"
  echo ">>> [服务器] \${GO_BIN} build -o \${bin_path} ."
  "\${GO_BIN}" build -o "\${bin_path}" .
}

find_matching_pid() {
  local pid
  for pid_path in /proc/[0-9]*; do
    [[ -e "\${pid_path}" ]] || continue
    pid=\${pid_path#/proc/}
    [[ "\${pid}" =~ ^[0-9]+\$ ]] || continue
    cmdline_matches_mode "\${pid}" || continue
    echo "\${pid}"
    return 0
  done
  return 1
}

kill_matching() {
  local sig="\${1:-TERM}" pid
  for pid_path in /proc/[0-9]*; do
    [[ -e "\${pid_path}" ]] || continue
    pid=\${pid_path#/proc/}
    [[ "\${pid}" =~ ^[0-9]+\$ ]] || continue
    cmdline_matches_mode "\${pid}" || continue
    kill "-\${sig}" "\${pid}" 2>/dev/null || true
  done
}

# 启动前务必清掉所有 run_bot=true（含 go run 子进程；不限 run_conf）
proc_is_app_run_bot() {
  local pid="\$1" cmd cwd
  cmd=\$(read_proc_cmdline "\${pid}" 2>/dev/null || true)
  cwd=\$(proc_cwd "\${pid}")
  [[ -n "\${cmd}" ]] || return 1
  if [[ "\${cmd}" == *"--run_bot=false"* || "\${cmd}" == *"--run_bot=0"* ]]; then
    return 1
  fi
  if [[ "\${cmd}" == *"--run_bot=true"* || "\${cmd}" == *"--run_bot=1"* ]]; then
    if [[ "\${cwd}" == "\${APP_DIR}" ]]; then
      return 0
    fi
    if [[ "\${cmd}" == *"main.go"* || "\${cmd}" == *"\${MALL_BIN_NAME}"* || "\${cmd}" == *"ai-agent-paper"* ]]; then
      return 0
    fi
  fi
  if [[ "\${cwd}" == "\${APP_DIR}" ]]; then
    if [[ "\${cmd}" == *go-build* || "\${cmd}" == *"/exe/main"* ]]; then
      return 0
    fi
    if [[ "\${cmd}" == *"/bin/\${MALL_BIN_NAME}"* || "\${cmd}" == *"\${MALL_BIN_NAME}"* ]]; then
      [[ "\${cmd}" == *"--run_bot=true"* || "\${cmd}" == *"--run_bot=1"* ]] && return 0
    fi
  fi
  return 1
}

kill_all_run_bot_processes() {
  local sig="\${1:-TERM}" pid cmd
  echo ">>> [服务器] 清理 run_bot=true 进程（\${APP_DIR}）"
  if [[ -f "\${PID_FILE}" ]]; then
    pid=\$(tr -d '[:space:]' <"\${PID_FILE}" 2>/dev/null || true)
    if [[ -n "\${pid}" ]]; then
      kill "-\${sig}" "\${pid}" 2>/dev/null || true
    fi
    rm -f "\${PID_FILE}"
  fi
  for pid_path in /proc/[0-9]*; do
    [[ -e "\${pid_path}" ]] || continue
    pid=\${pid_path#/proc/}
    [[ "\${pid}" =~ ^[0-9]+\$ ]] || continue
    proc_is_app_run_bot "\${pid}" || continue
    cmd=\$(read_proc_cmdline "\${pid}" 2>/dev/null || true)
    echo ">>> [服务器] kill -\${sig} pid=\${pid} \${cmd}"
    kill "-\${sig}" "\${pid}" 2>/dev/null || true
  done
}

stop_mall() {
  echo ">>> [服务器] 停止 ai-agent-paper/\${MODE}"
  if [[ "\${RUN_BOT}" == "true" ]]; then
    kill_all_run_bot_processes TERM
  else
    if [[ -f "\${PID_FILE}" ]]; then
      old_pid=\$(tr -d '[:space:]' <"\${PID_FILE}" 2>/dev/null || true)
      kill -TERM "\${old_pid}" 2>/dev/null || true
      rm -f "\${PID_FILE}"
    fi
    kill_matching TERM
  fi
  if [[ "\${RUN_BOT}" == "false" && -n "\${LISTEN_PORT}" ]]; then
    kill_port_listeners "\${LISTEN_PORT}" TERM
  fi
  sleep 2
  if [[ "\${RUN_BOT}" == "true" ]]; then
    kill_all_run_bot_processes KILL
  else
    kill_matching KILL
  fi
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
    [[ -e "\${pid_path}" ]] || continue
    pid=\${pid_path#/proc/}
    [[ "\${pid}" =~ ^[0-9]+\$ ]] || continue
    cmdline_matches_mode "\${pid}" || continue
    found=1
    cmd=\$(read_proc_cmdline "\${pid}" 2>/dev/null || true)
    cwd=\$(proc_cwd "\${pid}")
    echo ">>> RUNNING pid=\${pid} cwd=\${cwd}"
    echo ">>>   \${cmd}"
  done
  if [[ -f "\${LOG_FILE}" ]]; then
    echo ">>> 日志: \${LOG_FILE}"
  fi
  if [[ "\${found}" -eq 0 ]]; then
    echo ">>> 无匹配进程"
  fi
  return 0
}

wait_for_mall() {
  local i pid wait_sec=30
  if [[ "\${PAPER_USE_GO_RUN}" == "1" ]]; then
    wait_sec=120
  fi
  echo ">>> [服务器] 等待进程就绪（最多 \${wait_sec}s）..."
  for i in \$(seq 1 "\${wait_sec}"); do
    if [[ "\${RUN_BOT}" == "true" ]]; then
      if mall_bot_ready; then
        show_mall_status
        return 0
      fi
    else
      pid="\$(find_matching_pid 2>/dev/null || true)"
      if [[ -n "\${pid}" ]] && kill -0 "\${pid}" 2>/dev/null && port_in_use "\${LISTEN_PORT}"; then
        show_mall_status
        return 0
      fi
    fi
    if [[ -f "\${LOG_FILE}" ]] && grep -qE 'panic:|FATAL|run_conf parameter is required' "\${LOG_FILE}"; then
      echo "ERROR: [服务器] 启动失败，日志:" >&2
      tail -n 50 "\${LOG_FILE}" >&2
      exit 1
    fi
    sleep 1
  done
  if [[ "\${RUN_BOT}" == "true" ]]; then
    echo "ERROR: [服务器] bot 启动超时，请查看 \${LOG_FILE}（go 未安装或 go run 编译失败？）" >&2
  else
    echo "ERROR: [服务器] 启动超时（端口 \${LISTEN_PORT} 未监听？检查 conf 的 server.port）" >&2
  fi
  tail -n 40 "\${LOG_FILE}" 2>/dev/null >&2 || true
  exit 1
}

start_mall() {
  cd "\${APP_DIR}"
  mkdir -p "\${BIN_DIR}" "\${LOG_DIR}"
  local bin_path="\${BIN_DIR}/\${MALL_BIN_NAME}"
  if [[ "\${RUN_BOT}" == "true" ]]; then
    kill_all_run_bot_processes TERM
    sleep 2
    kill_all_run_bot_processes KILL
    sleep 1
  fi
  : >"\${LOG_FILE}"
  export GO111MODULE=on
  export GOPROXY="\${GOPROXY:-https://goproxy.cn,direct}"
  if [[ "\${PAPER_USE_GO_RUN}" == "1" ]]; then
    ensure_go
    echo ">>> [服务器] 启动 \${GO_BIN} run main.go --run_conf=\${RUN_CONF} --run_bot=\${RUN_BOT}"
    nohup "\${GO_BIN}" run main.go --run_conf="\${RUN_CONF}" --run_bot="\${RUN_BOT}" >>"\${LOG_FILE}" 2>&1 &
  else
    [[ -x "\${bin_path}" ]] || build_mall_binary
    echo ">>> [服务器] 启动 \${bin_path} --run_conf=\${RUN_CONF} --run_bot=\${RUN_BOT}"
    nohup "\${bin_path}" --run_conf="\${RUN_CONF}" --run_bot="\${RUN_BOT}" >>"\${LOG_FILE}" 2>&1 &
  fi
  echo \$! >"\${PID_FILE}"
  wait_for_mall
}

case "\${ACTION}" in
  deploy)
    git_sync_branch
    if [[ "\${PAPER_USE_GO_RUN}" != "1" ]]; then
      build_mall_binary
    fi
    stop_mall
    start_mall
    ;;
  restart)
    if [[ "\${PAPER_USE_GO_RUN}" != "1" ]]; then
      build_mall_binary
    fi
    stop_mall
    start_mall
    ;;
  stop) stop_mall ;;
  status) show_mall_status ;;
  *) exit 1 ;;
esac
EOF

echo ">>> 后端 分支=${BRANCH} 动作=${ACTION} 模式=${MODE} run_conf=${RUN_CONF} go_run=${PAPER_USE_GO_RUN}"
run_ssh "mkdir -p '${APP_DIR}'"
deploy_sync_local_conf "$RUN_CONF"
if ! run_remote_bash "$REMOTE_SCRIPT_BODY"; then
  echo ">>> 远程脚本返回非 0" >&2
  echo ">>> 服务器日志: ${APP_DIR}/logs/ai-agent-paper-${MODE}-nohup.log" >&2
  echo ">>> 若上方已有 RUNNING，可: ./deploy-service.sh master status" >&2
  exit 1
fi
echo ">>> 后端完成"
