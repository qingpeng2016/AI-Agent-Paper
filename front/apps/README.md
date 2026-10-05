# Web monorepo（`front/apps`）

本目录是 **唯一的 pnpm 工作区根**：`node_modules`、`pnpm-lock.yaml` 都应在这里，不要出现在 `front/` 与 `client/` 同级。

| 文件 | 作用 |
|------|------|
| `package.json` | 统一脚本：`dev:pc`、`build:pc` 等 |
| `pnpm-workspace.yaml` | 子包：`shared`、`web-pc`、`web-h5` |
| `pnpm-lock.yaml` | 锁定依赖版本（提交 Git） |
| `node_modules/` | `pnpm install` 生成，**勿手改、勿提交** |

| 目录 | 包名 | 说明 |
|------|------|------|
| `shared/` | `@ai-token-mall/shared` | API、类型；禁止 UI 库 |
| `web-pc/` | `@ai-token-mall/web-pc` | PC + Element Plus |
| `web-h5/` | `@ai-token-mall/web-h5` | H5 + Vant（待 scaffold） |

```bash
cd front/apps
pnpm install
pnpm dev:pc
```

原生端见 `../client/`；不参与本 workspace。
