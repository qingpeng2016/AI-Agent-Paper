-- 档位展示统一用 sku_product_name，删除冗余 marketing_tier
ALTER TABLE `products` DROP COLUMN `marketing_tier`;
