# AI-Agent-Paper 部署

本机只做 **SSH + 上传 `conf/local.conf.yaml`**（`run_conf=local` 时）。  
**代码拉取、`go run`、启停** 全部在服务器 `MALL_APP_DIR` 执行。

## 单服务 bot（与本地一致）

服务器流程：`git checkout .` → `git fetch/pull <分支>` →  
`go run main.go --run_conf=local --run_bot=true`

```bash
cp .deploy.env.example .deploy.env   # 填写 SSH_PASS 等
chmod +x deploy-service.sh deploy-backend.sh
./deploy-service.sh master           # deploy
./deploy-service.sh master status
./deploy-service.sh master restart
./deploy-service.sh master stop
```

> `run_bot=true` 时不启 HTTP。需要 API：`./deploy-backend.sh <分支> deploy user`

### 服务器前置

1. **git**（`GIT_REPO_URL` 为 git@ 时，服务器需能读 GitHub）
2. **Go 1.22+**：未安装时 deploy 会尝试下载到 `/usr/local/go`；也可手动安装或设 `PAPER_GO_ROOT`
3. `conf/local.conf.yaml` 不入库，由本机上传到服务器 `conf/`

### 日志

`{MALL_APP_DIR}/logs/ai-agent-paper-bot-nohup.log`
