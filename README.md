# AI-Token-Mall

AI 聚合平台骨架：用户注册/登录 HTTP 服务、调度 Bot、支付宝账务明细 OpenAPI 客户端。

## 启动

需设置 `RUN_CONF` 或 `--run_conf`（如 `local` / `prod`）。

HTTP 接口：

```bash
go run main.go --run_conf=local --run_bot=false
```

Bot（按 `bot_schedule_config` 调度，默认任务 `ai_token_mall:user_stats` 统计用户数与访问日志数）：

```bash
go run main.go --run_conf=local --run_bot=true
```

## HTTP API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/users/register` | 用户注册（邮箱或手机 + 密码） |
| POST | `/api/v1/users/login` | 用户登录 |

## 目录

```
main.go / boot/              入口与 dig 注入
conf/                        yaml（server / db / redis / alipay / notification）
application/core-service/    用户、支付宝、bot 调度配置
application/bot/             Stats 定时任务与 Scheduler
domain/persistent/           GORM 实体与仓储接口
domain/http/                 第三方 API 实体与仓储接口（repository）
infrastructure/http/         通用 HTTP 客户端 + alipay 等实现
infrastructure/              mysql、redis
interfaces/                  handler + gin 路由
docs/                        ai-platform-design.md、ai-platform-schema.sql
```

## Bot 调度示例 SQL

```sql
INSERT INTO bot_schedule_config (module, task_name, interval_seconds, is_enabled, is_strategy_enabled)
VALUES ('ai_token_mall', 'user_stats', 60, 1, 1);
```

## 配置

`conf/*.yaml` 中 `alipay.app_id`、`alipay.private_key` 用于 `alipay.data.bill.accountlog.query`（账务明细）。
