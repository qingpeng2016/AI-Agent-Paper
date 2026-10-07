# Research 工作台（Paper Agent）

| 内容 | 路径 |
|------|------|
| **UI 源码** | `agent/paper/PaperWorkbenchView.vue`（与 `front/.../src/views/paper/PaperWorkbenchView.vue` 保持同步；后者 import 用 `@paper/types`） |
| 模块类型、导航定义 | `agent/paper/types.ts` |
| 路由 | `/workbench` |
| 入口 | 会员中心 Hero **进入工作台**（原「选购套餐」位） |

功能对齐 ARIS Desktop；色系对齐 AIPlan 会员中心（紫渐变侧栏、白卡片、主色按钮）。
