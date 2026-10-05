-- products 表：与 upstream_info 等路由字段区分命名
ALTER TABLE `products`
  CHANGE COLUMN `upstream_name` `sku_upstream_name` VARCHAR(32) NOT NULL,
  CHANGE COLUMN `upstream_product` `sku_product_name` VARCHAR(128) NOT NULL COMMENT 'SKU 绑定上游档位名，如 GPT PRO 5X';

ALTER TABLE `products`
  DROP INDEX `idx_products_upstream_name`,
  DROP INDEX `idx_products_upstream`;

ALTER TABLE `products`
  ADD KEY `idx_products_sku_upstream_name` (`sku_upstream_name`),
  ADD KEY `idx_products_sku_upstream` (`sku_upstream_name`, `sku_product_name`);
