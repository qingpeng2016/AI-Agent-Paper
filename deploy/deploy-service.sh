#!/usr/bin/env bash
# 单服务 bot：SSH 到服务器 → git checkout . → 拉分支 → go run main.go --run_conf=local --run_bot=true
#
# 用法:
#   ./deploy-service.sh <分支>              # deploy
#   ./deploy-service.sh <分支> restart
#   ./deploy-service.sh <分支> stop
#   ./deploy-service.sh <分支> status
#
# 本机只上传 conf/local.conf.yaml（若 run_conf=local）；代码与 go run 均在服务器执行。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
export PAPER_USE_GO_RUN=1
BRANCH="${1:-}"
ACTION="${2:-deploy}"

if [[ -z "$BRANCH" ]]; then
  echo "用法: $0 <分支> [deploy|restart|stop|status]" >&2
  exit 1
fi

case "$ACTION" in
  deploy|restart|stop|status) ;;
  *)
    echo "未知动作: $ACTION" >&2
    exit 1
    ;;
esac

exec "${SCRIPT_DIR}/deploy-backend.sh" "$BRANCH" "$ACTION" bot
