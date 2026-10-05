-- 商城分类表 + products 关联（胶囊筛选来自 products_category.sort，展示名用 name）

CREATE TABLE IF NOT EXISTS `products_category` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(64)  NOT NULL COMMENT '展示名，如 ChatGPT',
  `dot_color`  VARCHAR(16)  DEFAULT NULL COMMENT '胶囊圆点色',
  `active_bg`  VARCHAR(16)  DEFAULT NULL COMMENT '选中胶囊背景色',
  `sort`       INT          NOT NULL DEFAULT 0 COMMENT '分类排序（越小越靠前）',
  `status`     VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_products_category_name` (`name`),
  KEY `idx_products_category_sort` (`status`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商城前台分类（全部套餐胶囊）';

INSERT INTO `products_category` (`name`, `dot_color`, `active_bg`, `sort`, `status`) VALUES
  ('ChatGPT',    '#10a37f', '#10a37f', 10, 'active'),
  ('Claude',     '#d97757', '#d97757', 20, 'active'),
  ('Grok',       '#0f172a', '#334155', 30, 'active'),
  ('Gemini',     '#4285f4', '#4285f4', 40, 'active'),
  ('Cursor',     '#7c3aed', '#7c3aed', 50, 'active'),
  ('Perplexity', '#0d9488', '#0d9488', 60, 'active')
ON DUPLICATE KEY UPDATE
  `dot_color` = VALUES(`dot_color`),
  `active_bg` = VALUES(`active_bg`),
  `sort` = VALUES(`sort`),
  `status` = VALUES(`status`);

ALTER TABLE `products`
  CHANGE COLUMN `sku_upstream_name` `products_category_name` VARCHAR(32) NOT NULL COMMENT '上游/导航分组键 openai|cursor 等（非分类展示名）',
  ADD COLUMN `products_category_id` BIGINT UNSIGNED NULL COMMENT '分类 ID → products_category.id' AFTER `share_seats`,
  ADD KEY `idx_products_category_id` (`products_category_id`);

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
SET p.`products_category_id` = c.`id`;

UPDATE `products` p
INNER JOIN `products_category` c ON c.`name` = 'Cursor'
SET p.`products_category_id` = c.`id`
WHERE p.`sku_code` LIKE 'CUR-%';

ALTER TABLE `products`
  MODIFY COLUMN `sort_order` INT NOT NULL DEFAULT 0 COMMENT '分类内排序（越小越靠前）';
