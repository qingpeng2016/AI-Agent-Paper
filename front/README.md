# AI-Token-Mall 前端（总目录）

**Web / 原生** 平级，互不混用 Node 工程：

```
front/
├── apps/          # 浏览器：pnpm monorepo（package.json 在这里）
└── client/        # Android / iOS（不进 pnpm）
```

**开发 Web（PC）** — 进入 `apps/` 再装依赖、启动：

```bash
cd front/apps
pnpm install
pnpm dev:pc
```

说明见 [apps/README.md](./apps/README.md)。
