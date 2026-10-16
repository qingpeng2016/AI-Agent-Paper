# 仅在远程服务器上执行（由 deploy-*.sh 通过 SSH 注入，勿在本机 source）

git_prepare_for_pull() {
  local clean_bin="${1:-0}"
  cd "${APP_DIR}"

  echo ">>> [服务器] git checkout .（丢弃已跟踪文件的本地改动）"
  git checkout .

  if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git restore --worktree log/info.log log/error.log 2>/dev/null \
      || git checkout HEAD -- log/info.log log/error.log 2>/dev/null \
      || true
  fi
  if [[ "${clean_bin}" == "1" ]]; then
    git clean -fd logs/ bin/ 2>/dev/null || true
  fi
}

git_fetch_with_timeout() {
  export GIT_TERMINAL_PROMPT=0
  if command -v timeout >/dev/null 2>&1; then
    timeout "${GIT_FETCH_TIMEOUT}" git fetch origin --prune --progress
  else
    git fetch origin --prune --progress
  fi
}

ensure_app_git_repo() {
  mkdir -p "$(dirname "${APP_DIR}")"
  if [[ -d "${APP_DIR}/.git" ]]; then
    return 0
  fi
  local url="${GIT_REPO_URL:-}"
  if [[ -z "${url}" ]]; then
    echo "ERROR: ${APP_DIR} 不是 git 仓库，且未设置 GIT_REPO_URL" >&2
    exit 1
  fi
  echo ">>> [服务器] git clone ${url} → ${APP_DIR}"
  rm -rf "${APP_DIR}"
  git clone "${url}" "${APP_DIR}"
  cd "${APP_DIR}"
}

git_sync_branch() {
  ensure_app_git_repo
  cd "${APP_DIR}"
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "ERROR: ${APP_DIR} 不是 git 仓库" >&2
    exit 1
  fi
  git_prepare_for_pull "${GIT_CLEAN_BIN:-0}"
  echo ">>> [服务器] 目录: ${APP_DIR}"
  echo ">>> [服务器] git fetch + checkout ${BRANCH} + pull"
  git_fetch_with_timeout
  if git show-ref --verify --quiet "refs/heads/${BRANCH}"; then
    git checkout "${BRANCH}"
  else
    git checkout -B "${BRANCH}" "origin/${BRANCH}"
  fi
  if ! git pull --ff-only origin "${BRANCH}"; then
    echo "ERROR: git pull --ff-only 失败（已执行 git checkout .，请检查服务器仓库状态）" >&2
    git status -sb >&2 || true
    exit 1
  fi
  git log -1 --oneline
}

setup_node_path() {
  export PATH="${PATH}:/usr/local/node/bin:/usr/local/bin:/usr/bin"
  if [[ -d /www/server/nvm/versions/node ]]; then
    latest=$(ls -1 /www/server/nvm/versions/node 2>/dev/null | sort -V | tail -1 || true)
    if [[ -n "${latest}" ]]; then
      export PATH="/www/server/nvm/versions/node/${latest}/bin:${PATH}"
    fi
  fi
  if [[ -s "${HOME}/.nvm/nvm.sh" ]]; then
    # shellcheck disable=SC1091
    source "${HOME}/.nvm/nvm.sh"
  fi
}

# 非交互 SSH 不会读 .bashrc 里的 PATH，必须用绝对路径或显式 PATH
setup_go_path() {
  local d
  export PATH="/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin:${PATH}"
  if [[ -n "${PAPER_GO_ROOT:-}" && -d "${PAPER_GO_ROOT}/bin" ]]; then
    export PATH="${PAPER_GO_ROOT}/bin:${PATH}"
  fi
  for d in /usr/local/go/bin /opt/go/bin /www/server/go/bin; do
    if [[ -x "${d}/go" ]]; then
      export PATH="${d}:${PATH}"
    fi
  done
  if [[ -s /etc/profile ]]; then
    # shellcheck disable=SC1091
    source /etc/profile 2>/dev/null || true
  fi
  if [[ -s /etc/environment ]]; then
    set -a
    # shellcheck disable=SC1091
    source /etc/environment 2>/dev/null || true
    set +a
  fi
}

resolve_go_bin() {
  local c found
  setup_go_path
  GO_BIN=""
  for c in \
    "${PAPER_GO_ROOT:+"${PAPER_GO_ROOT}/bin/go"}" \
    /usr/local/go/bin/go \
    /opt/go/bin/go \
    /usr/bin/go; do
    [[ -n "${c}" && -x "${c}" ]] || continue
    GO_BIN="${c}"
    export PATH="$(dirname "${c}"):${PATH}"
    export GO_BIN
    return 0
  done
  found="$(type -P go 2>/dev/null || true)"
  if [[ -n "${found}" && -x "${found}" ]]; then
    GO_BIN="${found}"
    export GO_BIN
    export PATH="$(dirname "${found}"):${PATH}"
    return 0
  fi
  return 1
}

install_go_on_server() {
  if [[ -x /usr/local/go/bin/go ]]; then
    return 0
  fi
  local ver="${PAPER_GO_VERSION:-1.22.5}"
  local tarball="go${ver}.linux-amd64.tar.gz"
  local tmp="/tmp/${tarball}"
  local url
  echo ">>> [服务器] 未找到 go，尝试安装 Go ${ver} → /usr/local/go ..."
  for url in \
    "https://go.dev/dl/${tarball}" \
    "https://golang.google.cn/dl/${tarball}"; do
    if command -v wget >/dev/null 2>&1; then
      wget -qO "${tmp}" "${url}" && break
    elif command -v curl >/dev/null 2>&1; then
      curl -fsSL -o "${tmp}" "${url}" && break
    else
      echo "ERROR: [服务器] 需要 wget 或 curl 才能自动安装 Go" >&2
      return 1
    fi
    url=""
  done
  [[ -s "${tmp}" ]] || return 1
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "${tmp}"
  rm -f "${tmp}"
  setup_go_path
}

ensure_go() {
  if resolve_go_bin; then
    echo ">>> [服务器] ${GO_BIN} $("${GO_BIN}" version)"
    return 0
  fi
  install_go_on_server || true
  if resolve_go_bin; then
    echo ">>> [服务器] ${GO_BIN} $("${GO_BIN}" version)"
    return 0
  fi
  echo "ERROR: [服务器] 未找到 go（非交互 SSH 不会加载 .bashrc 的 PATH）。" >&2
  echo "  请确认: ls -l /usr/local/go/bin/go" >&2
  echo "  或在 deploy/.deploy.env 设置: PAPER_GO_ROOT=/usr/local/go" >&2
  exit 1
}

ensure_remote_pnpm() {
  setup_node_path
  if ! command -v node >/dev/null 2>&1; then
    echo "ERROR: [服务器] 未找到 node，请在宝塔安装 Node 18+" >&2
    exit 1
  fi
  echo ">>> [服务器] node $(node -v)"
  if command -v pnpm >/dev/null 2>&1; then
    echo ">>> [服务器] pnpm $(pnpm -v)"
    return 0
  fi
  if command -v corepack >/dev/null 2>&1; then
    corepack enable 2>/dev/null || true
  fi
  if command -v pnpm >/dev/null 2>&1; then
    echo ">>> [服务器] pnpm $(pnpm -v)"
    return 0
  fi
  echo ">>> [服务器] npm install -g pnpm@9"
  npm install -g pnpm@9
  setup_node_path
  command -v pnpm >/dev/null 2>&1 || {
    echo "ERROR: [服务器] 无法安装 pnpm" >&2
    exit 1
  }
  echo ">>> [服务器] pnpm $(pnpm -v)"
}
