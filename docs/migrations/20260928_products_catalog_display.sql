-- 商城 SKU 卡片展示字段（对齐前端 CatalogProduct）
ALTER TABLE `products`
  ADD COLUMN `card_title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '卡片标题' AFTER `sku_code`,
  ADD COLUMN `card_subtitle` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '卡片副标题' AFTER `card_title`,
  ADD COLUMN `card_features` JSON NOT NULL COMMENT '卡片卖点条目，可用 {limit_tokens}/{rpm_limit} 占位' AFTER `card_subtitle`,
  ADD COLUMN `share_seats` INT NOT NULL DEFAULT 1 COMMENT '子 Key 可共用人数（展示）' AFTER `card_features`,
  ADD COLUMN `is_flagship` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '旗舰卡片样式' AFTER `is_hot`;

-- 初始 SKU（与 front/apps/web-pc mock 一致；可重复执行前先 DELETE FROM products WHERE sku_code IN (...))
INSERT INTO `products` (
  `sku_code`, `card_title`, `card_subtitle`, `card_features`, `share_seats`,
  `sku_upstream_name`, `sku_product_name`, `limit_tokens`, `rpm_limit`, `allowed_models`,
  `product_type`, `billing_period`, `price`, `currency`, `highlights_json`,
  `is_hot`, `is_flagship`, `is_api_enabled`, `sort_order`, `status`
) VALUES
(
  'OAI-PRO20-M', 'ChatGPT Pro 20X 月卡', '约 20 倍配额，团队满血与重度 Codex。',
  JSON_ARRAY('{limit_tokens} tokens / 月', '{rpm_limit} RPM', 'o1-mini 白名单', '专业 / 团队首选'),
  20, 'openai', 'GPT PRO 20X', 20000000, 120, JSON_ARRAY('gpt-4o', 'o1-mini'),
  'subscription', 'month', 1799.00, 'CNY', JSON_ARRAY(), 1, 1, 1, 5, 'on_sale'
),
(
  'OAI-PRO5-M', 'ChatGPT Pro 5X 月卡', 'Pro 5X，主力旗舰组合，约 5 倍配额。',
  JSON_ARRAY('{limit_tokens} tokens / 月', '旗舰模型可用', 'Codex 重度可用', '约 1 分钟开通'),
  10, 'openai', 'GPT PRO 5X', 12000000, 100, JSON_ARRAY('gpt-4o', 'o1-mini'),
  'subscription', 'month', 899.00, 'CNY', JSON_ARRAY(), 0, 0, 1, 6, 'on_sale'
),
(
  'ANT-PRO-M', 'Claude Pro 月卡', 'Sonnet / Haiku 全系，支持 Claude Code。',
  JSON_ARRAY('{limit_tokens} tokens / 月', '200K 上下文', 'Claude 全系列', 'Claude Code 可用'),
  5, 'anthropic', 'Claude Pro 5X', 8000000, 80, JSON_ARRAY('claude-3-5-sonnet-20241022'),
  'subscription', 'month', 219.00, 'CNY', JSON_ARRAY(), 1, 0, 1, 7, 'on_sale'
),
(
  'OAI-PLUS-M', 'ChatGPT Plus 月卡', '主力模型组合，含 Codex 与 Images 档位白名单。',
  JSON_ARRAY('{limit_tokens} tokens / 月', 'gpt-4o + o1-mini', 'Codex 可用', '热销性价比'),
  5, 'openai', 'GPT PRO 5X', 8000000, 90, JSON_ARRAY('gpt-4o', 'o1-mini'),
  'subscription', 'month', 178.00, 'CNY', JSON_ARRAY(), 1, 0, 1, 8, 'on_sale'
),
(
  'OAI-GO-M', 'ChatGPT Go 月卡', '入门档额度，日常对话与轻量开发够用。',
  JSON_ARRAY('{limit_tokens} tokens / 月', 'gpt-4o-mini 可用', 'ChatGPT 线路', '约 1 分钟开通'),
  3, 'openai', 'GPT GO 5X', 2000000, 60, JSON_ARRAY('gpt-4o-mini', 'gpt-4o'),
  'subscription', 'month', 89.00, 'CNY', JSON_ARRAY(), 0, 0, 1, 10, 'on_sale'
),
(
  'PPX-PRO-M', 'Perplexity Pro 月卡', 'Pro Search + 多模型检索，研究类 Agent 首选。',
  JSON_ARRAY('{limit_tokens} tokens / 月', 'sonar / sonar-pro', 'Pro Search', '低价研究首选'),
  3, 'perplexity', 'Perplexity Pro', 5000000, 50, JSON_ARRAY('sonar', 'sonar-pro'),
  'subscription', 'month', 119.00, 'CNY', JSON_ARRAY(), 0, 0, 1, 12, 'on_sale'
),
(
  'CUR-PRO-M', 'Cursor 配置包 月卡', '含使用额度 + Cursor 配置教程（搭配编辑器自助使用）。',
  JSON_ARRAY('额度 + 教程', 'Cloud Agents 说明', 'Auto 模式说明', '按文档自助配置'),
  5, 'openai', 'Mixed Pool', 6000000, 80, JSON_ARRAY('gpt-4o'),
  'subscription', 'month', 198.00, 'CNY', JSON_ARRAY(), 1, 0, 1, 13, 'on_sale'
),
(
  'GEM-PRO-M', 'Gemini AI Pro 月卡', 'Pro + Flash 组合，多模态与长上下文。',
  JSON_ARRAY('{limit_tokens} tokens / 月', '1.5 Pro / Flash', 'Nano 出图', 'Gemini 全系列'),
  5, 'gemini', 'Gemini Pro Pool', 10000000, 90, JSON_ARRAY('gemini-1.5-pro', 'gemini-1.5-flash'),
  'subscription', 'month', 189.00, 'CNY', JSON_ARRAY(), 0, 0, 1, 14, 'on_sale'
),
(
  'ANT-MAX-M', 'Claude Max 5X 月卡', '约 5 倍 Pro 用量，重度编程友好。',
  JSON_ARRAY('{limit_tokens} tokens / 月', 'Opus 可选', 'Claude Code', 'Max 专线'),
  10, 'anthropic', 'Claude Max 10X', 15000000, 100, JSON_ARRAY('claude-3-opus-20240229'),
  'subscription', 'month', 999.00, 'CNY', JSON_ARRAY(), 0, 0, 1, 16, 'on_sale'
),
(
  'XAI-GROK-M', 'Grok Super 3 个月', 'DeepSearch + 联网 + Imagine 线路。',
  JSON_ARRAY('{limit_tokens} tokens / 周期', 'grok-2 系列', 'Grok 全能力', '3 个月包'),
  5, 'xai', 'Grok 5X', 6000000, 60, JSON_ARRAY('grok-2'),
  'subscription', 'month', 790.00, 'CNY', JSON_ARRAY(), 0, 0, 1, 18, 'on_sale'
);
