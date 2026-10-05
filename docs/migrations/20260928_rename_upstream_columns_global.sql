-- 与 products 表命名对齐：upstream_name → products_category_name，upstream_product → sku_product_name
-- products 表若已是新列名可跳过（见 20260928_products_rename_upstream_columns.sql + 20260928_products_category.sql）
-- 以下表在 ai-platform-schema 中仍使用旧列名时需执行本脚本

-- ---------------------------------------------------------------------------
-- upstream_info（上游凭证池，按 SKU 分组键 + 档位路由）
-- ---------------------------------------------------------------------------
ALTER TABLE `upstream_info`
  CHANGE COLUMN `upstream_name` `products_category_name` VARCHAR(32) NOT NULL COMMENT '与 products 分组键一致：openai|anthropic|xai|…',
  CHANGE COLUMN `upstream_product` `sku_product_name` VARCHAR(128) NOT NULL COMMENT '上游采购档位，如 GPT PRO 5X';

ALTER TABLE `upstream_info`
  DROP INDEX `idx_upstream_info_product`,
  DROP INDEX `idx_upstream_info_name_status`;

ALTER TABLE `upstream_info`
  ADD KEY `idx_upstream_info_product` (`products_category_name`, `sku_product_name`),
  ADD KEY `idx_upstream_info_name_status` (`products_category_name`, `status`);

-- ---------------------------------------------------------------------------
-- user_subscriptions（履约快照，与下单 SKU 一致）
-- ---------------------------------------------------------------------------
ALTER TABLE `user_subscriptions`
  CHANGE COLUMN `upstream_name` `products_category_name` VARCHAR(32) NOT NULL COMMENT '开通时 SKU 分组键快照',
  CHANGE COLUMN `upstream_product` `sku_product_name` VARCHAR(128) NOT NULL COMMENT '开通时 SKU 档位名快照';

-- ---------------------------------------------------------------------------
-- user_api_keys（仅原 upstream_name，网关 Path 隔离）
-- ---------------------------------------------------------------------------
ALTER TABLE `user_api_keys`
  CHANGE COLUMN `upstream_name` `products_category_name` VARCHAR(32) NOT NULL COMMENT '网关 Path 隔离，如 openai';

-- ---------------------------------------------------------------------------
-- user_access_logs（访问日志冗余字段）
-- ---------------------------------------------------------------------------
ALTER TABLE `user_access_logs`
  CHANGE COLUMN `upstream_name` `products_category_name` VARCHAR(32) NOT NULL,
  CHANGE COLUMN `upstream_product` `sku_product_name` VARCHAR(128) DEFAULT NULL COMMENT '实际路由档位（冗余便于检索）';
