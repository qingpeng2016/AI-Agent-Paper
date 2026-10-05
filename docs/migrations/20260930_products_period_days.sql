-- 商品计费周期天数（订阅时间计算统一依据）
ALTER TABLE `products`
  ADD COLUMN `period_days` INT NOT NULL DEFAULT 30 COMMENT '每个计费周期天数' AFTER `billing_period`;

UPDATE `products` SET `period_days` = CASE `billing_period`
  WHEN 'year' THEN 365
  WHEN 'once' THEN 30
  ELSE 30
END;
