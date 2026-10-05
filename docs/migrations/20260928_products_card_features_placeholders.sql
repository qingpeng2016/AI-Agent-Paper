-- 已入库 SKU：将额度/RPM 文案改为占位符（改 limit_tokens / rpm_limit 后接口会自动换算展示）
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 月', '{rpm_limit} RPM', 'o1-mini 白名单', '专业 / 团队首选') WHERE `sku_code` = 'OAI-PRO20-M';
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 月', '旗舰模型可用', 'Codex 重度可用', '约 1 分钟开通') WHERE `sku_code` = 'OAI-PRO5-M';
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 月', '200K 上下文', 'Claude 全系列', 'Claude Code 可用') WHERE `sku_code` = 'ANT-PRO-M';
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 月', 'gpt-4o + o1-mini', 'Codex 可用', '热销性价比') WHERE `sku_code` = 'OAI-PLUS-M';
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 月', 'gpt-4o-mini 可用', 'ChatGPT 线路', '约 1 分钟开通') WHERE `sku_code` = 'OAI-GO-M';
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 月', 'sonar / sonar-pro', 'Pro Search', '低价研究首选') WHERE `sku_code` = 'PPX-PRO-M';
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 月', '1.5 Pro / Flash', 'Nano 出图', 'Gemini 全系列') WHERE `sku_code` = 'GEM-PRO-M';
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 月', 'Opus 可选', 'Claude Code', 'Max 专线') WHERE `sku_code` = 'ANT-MAX-M';
UPDATE `products` SET `card_features` = JSON_ARRAY('{limit_tokens} tokens / 周期', 'grok-2 系列', 'Grok 全能力', '3 个月包') WHERE `sku_code` = 'XAI-GROK-M';
