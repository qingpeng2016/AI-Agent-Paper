# 聚合 AI 平台 — 产品与技术设计

**模式**：商城卖档位 → 发 **API Key** → 网关 **校验权限与剩余 token** → **透明转发**上游池。用户自选客户端（Aider、Claude Code、Cursor 自定义模型等），配置 **Base URL + Key** 即可；改代码在用户本机，平台只处理模型 HTTP。

---

## 1. 平台功能总表

| 模块 | 功能 | 说明 | 优先级 |
|------|------|------|--------|
| **商城前台** | 多品牌目录 | ChatGPT / Claude / Grok / Gemini / Cursor / Perplexity 等品牌页 +「全部套餐」 | P0 |
| | SKU 展示 | 档位名、周期、价格、卖点、允许模型 | P0 |
| | 筛选排序 | 按品牌、价格、热销；搜索 SKU | P1 |
| | 营销展示 | Banner、热销标签、与官方价对比（可选） | P1 |
| | 购买弹窗 | 数量 1/5/10/20、企业开票勾选 | P1 |
| **交易与订单** | 支付 | 支付宝、微信；PayPal/Stripe（可选） | P0 |
| | 订单状态机 | 待支付 → 已支付 → 履约中 → 完成 / 失败退款 | P0 |
| | 支付回调 | 幂等、签名校验、对账 | P0 |
| | 超时关单 | 未支付自动关闭 | P0 |
| | 退款 | 规则可配置；原路退或余额 | P0 |
| | 发票 | 个人电子票；企业 VAT（会员中心填抬头） | P1 |
| | 优惠券 / 分销 | 渠道价、分佣结算 | P2 |
| **履约 · API 通道** | 订阅开通 | 支付成功 → 创建 subscription、绑定 SKU | P0 |
| | 发放主 Key | 支付成功生成 **主 Key**；**子 Key** 由账号在会员中心自行开通（见 **用量与计费**） | P0 |
| | 开通通知 | 邮件 / 站内信：Key、Base URL、文档链接 | P0 |
| | Token 加购包 | 单独 SKU 增加 `remain_tokens` | P1 |
| | 续费 | 延长周期或叠加额度 | P0 |
| **API 网关** | 鉴权 | 平台 Key → **主 Key** 或 **子 Key** → 同一主订阅 / SKU | P0 |
| | 订阅校验 | 未过期、状态 active（子 Key 继承主订阅） | P0 |
| | 模型白名单 | 请求 model ∈ SKU 允许列表 | P0 |
| | 用量校验 | 读取 **用量与计费** 侧主账 `remain_tokens`、子 Key 金额 cap 后判定 | P0 |
| | 限流 | 用户 RPM/TPM/并发（Redis） | P1 |
| | 池路由 | 按 `pool_route` 选 upstream Key（轮询/权重/主备） | P0 |
| | 透明转发 | 同 path/body；替换 Authorization；SSE 原样透传 | P0 |
| | 访问日志 | request_id、latency、tokens（默认不存完整 prompt） | P0 |
| | 错误响应 | 401/403/402/429/502 等与文档一致 | P0 |
| | Path 隔离 | 按 upstream_line 分前缀（openai/anthropic/…） | P0 |
| **用量与计费** | 主 Key 凭证 | 主 Key 创建（履约触发）、禁用、泄露轮换；哈希存储 | P0 |
| | 调用权限 | Key 绑定订阅/SKU：`allowed_models`、`upstream_line`（子 Key 继承主订阅） | P0 |
| | 用户额度账 | `quota_profile`：`quota_tokens_month`、`used`、`remain` | P0 |
| | 周期重置 | 按 SKU 计费周期滚动或自然月重置用户额度 | P0 |
| | 扣减规则 | 上游 2xx 且带 usage → 扣 `remain`；流式结束后记账；5xx/超时/网关拒单不扣 | P0 |
| | 扣费幂等 | `Idempotency-Key` 防重复扣费（可选） | P2 |
| | 加购入账 | Token 加购包、补偿 token 写入用户账 | P1 |
| | 用量明细 | 按主/子 Key、订阅、时间查询 token 与折算金额 | P1 |
| | 定价与 SKU 额度 | 档位 token/RPM/售价（不含上游池成本） | P0 |
| | 子 Key 开通 | 主 Key 下创建 **子 Key**（独立密钥，共享主订阅额度） | P1 |
| | 子 Key 管理 | 命名/备注、禁用、轮换、删除 | P1 |
| | 子 Key 分账统计 | 按子 Key：调用、token、**折算金额** | P1 |
| | 子 Key 费用 cap | 单 Key **金额预算**；达上限 402 | P1 |
| | 拆分用量总览 | 主订阅 remain + 各子 Key 消耗排行 | P1 |
| **上游资源池** | 总池定义 | 按 **upstream_line**（如 OpenAI、Anthropic）建「该 AI 总池」；一个总池下挂多个官方账号/API Key | P0 |
| | 账号 Key 管理 | 单 Key 加密存储、标签、到期日、归属总池、采购成本 | P0 |
| | 池用量统计 | 单 Key 与 **总池汇总** 的真实消耗（token/次数等，依上游回传或探测） | P0 |
| | 池额度与硬顶 | 总池或单 Key 的官方限额、剩余、超额策略（告警/停选该 Key） | P1 |
| | 池路由配置 | 总池内选 Key：轮询/权重/主备；供网关 `pool_route` 引用 | P0 |
| | 健康检查 | 失败率剔除、手动禁用 Key、临时移出路由 | P1 |
| | 池成本与毛利 | 池侧消耗 × 成本价；与 SKU 售价对账（运营报表，与用户账独立） | P1 |
| | 池级告警 | 总池或单 Key 余量、错误率、即将到期 | P1 |
| **会员中心** | 注册登录 | 邮箱/手机；可选微信；改密、2FA | P0 |
| | API 密钥 | 主/子 Key 的 UI（生命周期与 cap 归属 **用量与计费**） | P0 |
| | 用量仪表盘 | 主订阅 used/remain；各 **子 Key** 分账用量与金额 | P0 |
| | 调用文档 | 各 line 的 Base URL、兼容工具说明 | P0 |
| | 订单与发票 | 历史订单、退款入口 | P0 |
| | 续费提醒 | 到期前通知、一键再买 | P1 |
| | 配置教程 | Cursor BYOK、Claude Code、Aider 等 | P1 |
| **管理后台** | 商品 | SKU、quota_profile、pool_route CRUD | P0 |
| | 订单与用户 | 查询、手动延期、补偿 token | P0 |
| | 用户用量与订单 | 用户额度、补偿、订阅延期 | P0 |
| | 上游池与路由 | 总池、多账号 Key、池用量、路由与健康 | P0 |
| | 调用日志检索 | 按 user/key/request_id | P0 |
| | 报表 | GMV、用户 SKU 用量、**池消耗与毛利**、429/退款率 | P1 |
| | 风控操作 | 封 Key、黑名单 | P1 |
| **客服与售后** | 在线客服 | 售前 / 履约异常 | P0 |
| | FAQ / 帮助 | API 配置、常见 402/429 | P0 |
| | 工单 | 关联订单、SLA | P1 |
| **企业采购** | 企业专页 | 批量、对公说明 | P1 |
| | 对公订单 | 汇款确认后开通 | P1 |
| | 发票抬头 | 企业 VAT（可与会员中心共用） | P1 |
| **运营与增长** | 数据埋点 | 支付漏斗 | P2 |
| | 组合包 / 捆绑 | 多 SKU 打包 | P2 |
| **合规与风控** | 第三方声明 | 非官方授权；nominative 商标 | P0 |
| | 条款隐私 | 退款、禁止转售 Key、合法使用 | P0 |
| | 支付防刷 | 频控、欺诈规则 | P1 |
| | Key 泄露检测 | 单 Key 异常 QPS | P1 |
| **基础设施** | 网关高可用 | 无状态多实例 | P0 |
| | 安全 | TLS、Key/upstream KMS、哈希存 Key | P0 |
| | 监控告警 | 支付、上游失败率、池余量 | P1 |
| | WAF / 限流 | 登录、支付与网关入口 | P0 |
| | 移动端 | H5 购买与会员中心 | P0 |

---

## 2. 系统总览

```
商城 → 订单/订阅 → API Key
                        ↓
用户工具 ──HTTP──► API 网关（鉴权·剩余 token·限流）──► 上游池（多 Key/多协议线）
```

**履约**：支付成功 → 订阅 + 主 Key（**用量与计费** 写凭证与额度）。  
**调用**：网关只 **读** 用量与计费中的 Key 状态、额度与子 Key cap，通过后转发上游（与上游池 Key 无关）。

---

## 3. SKU 与目录

| 字段 | 含义 |
|------|------|
| `sku_product_name` | SKU 档位名（如 GPT PRO 5X），详情页档位文案同此字段 |
| `upstream_line` | openai / anthropic / xai / gemini / perplexity |
| （见 `user_subscriptions`） | 用户使用容量在履约表；SKU 上定义默认 quota/RPM/models |
| `upstream_name` / `upstream_product` | 网关按组选 `upstream_info`（组内 `weight` 加权） |

| 对外 SKU（例） | upstream_line | API 通道 |
|----------------|---------------|----------|
| GPT Go/Plus/Pro、Codex、Images | openai | ✅ |
| Claude Pro/Max、Claude Code | anthropic | ✅ |
| Grok 系列 | xai | ✅ |
| Gemini / Veo 等 | gemini | ✅ |
| Perplexity Pro | perplexity | ✅ |
| Cursor 档位名 | — | ⚠️ 仅 BYOK；非 Cursor 会员 |
| Cloud Agents | — | ❌ 不售 |

**档位与池分离**：对外卖 20X 等档位 → **用量与计费**只记用户 `quota_profile`；真实官方账号在 **上游资源池** 聚成该 AI 总池（可多个 5X 账号拼池）。网关转发只换 Authorization，不改响应 body。

---

## 4. 用户旅程

**买**：选 SKU → 支付 → subscription + **主 Key** → 会员中心看用量与文档。  
**拆**：任意用户可在 **用量与计费** 下开 **子 Key**（分给同事/项目/环境），可设单 Key **金额 cap**；调用仍扣 **主订阅**，分 Key 单独统计。  
**用**：工具填主 Key 或子 Key + Base URL → 网关校验主额度与子 Key cap → 转发。  
**尽**：主 token/RPM 用尽、子 Key 达金额 cap 或过期 → 402/429；续费或加购包。

---

## 5. API 网关（单次请求）

1. 解析 Key → **主 Key** 或 **子 Key**（映射同一主订阅）/ SKU  
2. 校验过期、model 白名单、主账 `remain_tokens`、RPM；子 Key 再校验 **费用 cap**（按 SKU/token 单价折算累计金额）  
3. 按 `pool_route` 从该 line **总池** 选 upstream Key → 透传 HTTP/SSE → 成功则 **扣主订阅额度**、**累计子 Key 分账**（用量与计费）并 **记池侧消耗** → 写访问日志（`key_id`、是否子 Key）  

**必须**：流式、大 body、tool_calls 透传。**第一版不做**：自研 chat、替用户改文件、未披露 model 映射。

---

## 6. 用量规则摘要（用户计费）

| 场景 | 扣用户 token |
|------|----------------|
| 上游 2xx + usage | 扣 |
| 5xx / 超时 | 不扣 |
| 网关 402/429 | 不转发、不扣 |
| 子 Key 达金额 cap | 402（主账仍有余额也不转发） |

主订阅 token 与 **子 Key 金额 cap** 同时满足才放行。池侧见 **上游资源池**；子 Key 分账与 cap 配置见 **用量与计费**（个人与企业购买均可用）。

---

## 7. 错误码（建议）

401 Key 无效｜403 禁用/model 不允许｜402 过期、主额度不足或 **子 Key 费用达 cap**｜429 限流｜502/503 上游不可用

---

## 8. 数据实体（参考）

`users`, `user_orders`, `products`, `user_subscriptions`（订阅+用户额度）, `user_api_keys`（同 `user_subscriptions` 下多条平级 Key）, `upstream_info`（上游凭证与容量）, `user_access_logs`

---

## 9. 设计原则

1. 无脑转发 — 门禁 + 记账 + HTTP 代理  
2. 工具无关 — 能配 API 即可  
3. 档位差价 — 用户 quota（用量与计费）vs 总池采购成本（上游资源池），两模块分账，不伪造 body  
4. 多协议线 — 转发按 line 分  
5. 闭源 SKU 不硬塞 API 模式  

---

## 10. 分期交付

**MVP**：1 条 upstream_line + 池；3~5 SKU；网关鉴权+token+SSE；支付宝/微信→Key；会员用量+文档；后台 SKU/池/日志；FAQ。  

**V1**：五条转发线；RPM/多 Key；池告警与报表；发票与加购；**子 Key + 分统计 + 金额 cap**；对公与教程。  

**V2**：分销/优惠券；子 Key 数量上限策略、IP 白名单（可选）。

---

*实现以各上游官方 API 文档为准。*
