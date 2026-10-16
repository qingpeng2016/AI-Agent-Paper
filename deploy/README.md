# AI-Agent-Paper 部署

本机只做 **SSH + 上传 `conf/local.conf.yaml`**（`run_conf=local` 时）。  
**代码拉取、`go run`、启停** 全部在服务器 `MALL_APP_DIR` 执行。

## 两个进程（可同时跑）

| 脚本 | 服务器命令 | PID/日志 |
|------|------------|----------|
| `./deploy-service.sh` | `go run main.go --run_conf=local --run_bot=true` | `logs/ai-agent-paper-bot-*` |
| `./deploy-api.sh` | `go run main.go --run_conf=local --run_bot=false` | `logs/ai-agent-paper-user-*` |

流程相同：`git checkout .` → 拉分支 → 清理同模式旧进程 → `go run` → 等待就绪。

```bash
cp .deploy.env.example .deploy.env   # SSH_PASS、MALL_USER_PORT 等
chmod +x deploy-service.sh deploy-api.sh deploy-backend.sh

./deploy-service.sh master           # bot
./deploy-api.sh master               # HTTP API（端口见 local.conf / MALL_USER_PORT）

./deploy-service.sh master status
./deploy-api.sh master restart
```

`.deploy.env` 中 **`MALL_BOT_RUN_CONF`** / **`MALL_USER_RUN_CONF`** 默认 `local`；**`MALL_USER_PORT`** 需与 `conf/local.conf.yaml` 的 `server.port` 一致。

### 服务器前置

1. **git**（`GIT_REPO_URL` 为 git@ 时，服务器需能读 GitHub）
2. **Go 1.22+**（`PAPER_GO_ROOT=/usr/local/go` 可选）
3. `conf/local.conf.yaml` 由本机上传
4. 原生 ARIS 需 **`libs/aris/tools`** 或环境变量 **`PAPER_ARIS_REPO`**（`libs/aris` 不在 git 里时要在服务器单独准备）

### 底层入口

`./deploy-backend.sh <分支> [deploy|restart|stop|status] [user|bot]`
