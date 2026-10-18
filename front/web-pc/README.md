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

- 页面：`http://localhost:5173/workbench`（本机 Vite）
- 接口：浏览器仍显示 `localhost:5173/api/...`，由 Vite **转发**到 **`https://216.195.208.199/api/...`**（需宝塔 443 反代到 Go `8886`）

指定后端（`/api` 代理目标），任选其一：

```bash
pnpm dev 127.0.0.1:8886
pnpm dev http://216.195.208.199:8886
VITE_DEV_API_TARGET=http://127.0.0.1:8886 pnpm dev   # 不写参数时仍可用环境变量
```

不写 IP/端口时，默认仍为 `http://216.195.208.199:8886`。

Vite 自身参数用 **`--` 之后** 传递，例如：

```bash
pnpm dev 127.0.0.1:8886 -- --host 0.0.0.0
pnpm dev -- --port 5174             # 仅改前端端口，后端仍用默认
```

说明：工作台里大量 `fetch('/api/v1/...')` 走代理；登录等 shared 客户端在 **`VITE_API_BASE_URL` 为空** 时同样走 `/api` 代理。若设置 `VITE_API_BASE_URL=http://远程:8886` 会直连远程，需后端开 CORS，一般本地开发不必设。

## 构建

```bash
pnpm build
# 产物：front/web-pc/dist
```
