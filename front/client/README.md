# 原生客户端（占位）

Android / iOS **不在** `front/apps` 的 pnpm workspace 内，与 Web 分目录维护。

若需与后端字段一致，可参考 `front/apps/shared/src/types` 自行实现 HTTP。Web 的 `pnpm install` 只在 **`front/apps`** 执行，本目录不会有 `node_modules`。
