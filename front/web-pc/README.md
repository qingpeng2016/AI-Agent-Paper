# ARIS Research Desktop（主前端）

本目录是 **唯一 Web 入口**：科研工作台 + 登录 / 注册。

| 内容 | 路径 |
|------|------|
| 工作台 UI | `PaperWorkbenchView.vue` 及同目录各 `Paper*.vue` |
| 模块类型、导航 | `types.ts` |
| 登录 / 注册 | `src/views/LoginView.vue`、`RegisterView.vue` |
| API 客户端 | `shared/`（`@ai-token-mall/shared`）+ `src/api/` |
| 路由 | `/` → `/workbench`；`/login`、`/register`；`/workbench` 需登录 |

## 开发

```bash
cd front/web-pc
pnpm install
pnpm dev
```

默认 `http://127.0.0.1:5173`，API 代理到 `127.0.0.1:8886`。

## 构建

```bash
pnpm build
# 产物：front/web-pc/dist
```
