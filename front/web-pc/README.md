# AI Research Studio / AIRS（主前端）

本目录是 **唯一 Web 入口**：科研工作台 + 登录 / 注册。

| 内容 | 路径 |
|------|------|
| 工作台 UI | `PaperWorkbenchView.vue` 及同目录各 `Paper*.vue` |
| 模块类型、导航 | `types.ts` |
| 登录 / 注册 | `src/views/LoginView.vue`、`RegisterView.vue` |
| API 客户端 | `shared/`（`@ai-agent-paper/shared`）+ `src/api/` |
| 路由 | `/` → `/workbench`；`/login`、`/register`；`/workbench` 需登录 |

## 开发

```bash
cd front/web-pc
pnpm install
pnpm dev
```

默认页面 `http://127.0.0.1:5173`，浏览器请求 `/api/...` 由 Vite **代理**到 `http://127.0.0.1:8886`（与是否写 IP 无关，走相对路径即可）。

### 指定后端地址

任选一种（不要末尾 `/`）：

```bash
# 临时：连远程 API
VITE_DEV_API_TARGET=http://216.195.208.199:8886 pnpm dev

# 或写入 .env.development.local（git 可忽略）
# VITE_DEV_API_TARGET=http://216.195.208.199:8886
```

`pnpm dev` 后面 **`--` 之后** 的参数会传给 Vite，例如：

```bash
pnpm dev -- --host 0.0.0.0          # 局域网可访问 dev 页面
pnpm dev -- --port 5174             # 改前端端口
VITE_DEV_API_TARGET=http://216.195.208.199:8886 pnpm dev -- --host 0.0.0.0
```

说明：工作台里大量 `fetch('/api/v1/...')` 走代理；登录等 shared 客户端在 **`VITE_API_BASE_URL` 为空** 时同样走 `/api` 代理。若设置 `VITE_API_BASE_URL=http://远程:8886` 会直连远程，需后端开 CORS，一般本地开发不必设。

## 构建

```bash
pnpm build
# 产物：front/web-pc/dist
```
