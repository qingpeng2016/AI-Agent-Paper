-- 若已执行过带 products_category.products_category_name 的旧版迁移，执行本脚本
ALTER TABLE `products_category`
  DROP INDEX `uk_products_category_name`;

ALTER TABLE `products_category`
  DROP COLUMN `products_category_name`;

ALTER TABLE `products_category`
  ADD UNIQUE KEY `uk_products_category_name` (`name`);

UPDATE `products` p
INNER JOIN `products_category` c ON c.`name` = CASE p.`products_category_name`
  WHEN 'openai' THEN 'ChatGPT'
  WHEN 'anthropic' THEN 'Claude'
  WHEN 'xai' THEN 'Grok'
  WHEN 'gemini' THEN 'Gemini'
  WHEN 'perplexity' THEN 'Perplexity'
  WHEN 'cursor' THEN 'Cursor'
  ELSE p.`products_category_name`
END
SET p.`products_category_id` = c.`id`
WHERE p.`products_category_id` IS NULL;
