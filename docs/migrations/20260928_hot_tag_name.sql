-- 分类导航 / 商品卡片角标文案（空则不显示）

ALTER TABLE `products_category`
  ADD COLUMN `hot_tag_name` VARCHAR(32) DEFAULT NULL COMMENT '顶栏分类标签，空则不显示' AFTER `status`;

ALTER TABLE `products`
  ADD COLUMN `hot_tag_name` VARCHAR(32) DEFAULT NULL COMMENT '卡片右上角标签，空则不显示' AFTER `is_flagship`;

UPDATE `products`
SET `hot_tag_name` = '热销'
WHERE `is_hot` = 1 AND (`hot_tag_name` IS NULL OR `hot_tag_name` = '');

UPDATE `products`
SET `hot_tag_name` = '旗舰'
WHERE `is_flagship` = 1 AND (`hot_tag_name` IS NULL OR `hot_tag_name` = '');
