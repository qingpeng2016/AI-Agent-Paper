-- 分页 UI 测试：为 user_id=2 插入 100 条已完成购单（order_no 前缀 TST，便于清理）
-- 当前前端 page_size=2 时约 50 页（若已有 4 条则共 52 页）
-- 清理：DELETE FROM user_orders WHERE order_no LIKE 'TST%' AND user_id = 2;

SET NAMES utf8mb4;

INSERT INTO `user_orders` (
  `order_no`,
  `user_id`,
  `product_id`,
  `order_type`,
  `user_subscription_id`,
  `quantity`,
  `unit_price`,
  `status`,
  `total_amount`,
  `currency`,
  `enterprise_invoice`,
  `pay_channel`,
  `out_trade_no`,
  `third_trade_no`,
  `paid_at`,
  `expire_at`,
  `created_at`,
  `updated_at`
)
WITH RECURSIVE `seq` AS (
  SELECT 1 AS `n`
  UNION ALL
  SELECT `n` + 1 FROM `seq` WHERE `n` < 100
)
SELECT
  CONCAT('TST', LPAD(`n`, 10, '0')) AS `order_no`,
  2 AS `user_id`,
  IF(`n` % 2 = 0, 2, 5) AS `product_id`,
  'purchase' AS `order_type`,
  0 AS `user_subscription_id`,
  1 AS `quantity`,
  IF(`n` % 2 = 0, 899.00, 89.00) AS `unit_price`,
  'completed' AS `status`,
  IF(`n` % 2 = 0, 899.00, 89.00) AS `total_amount`,
  'CNY' AS `currency`,
  0 AS `enterprise_invoice`,
  'alipay' AS `pay_channel`,
  CONCAT('PAYTST', LPAD(`n`, 10, '0')) AS `out_trade_no`,
  CONCAT('MOCK_alipay_seed_', `n`) AS `third_trade_no`,
  DATE_SUB(NOW(), INTERVAL `n` MINUTE) AS `paid_at`,
  DATE_ADD(DATE_SUB(NOW(), INTERVAL `n` MINUTE), INTERVAL 30 MINUTE) AS `expire_at`,
  DATE_SUB(NOW(), INTERVAL `n` MINUTE) AS `created_at`,
  DATE_SUB(NOW(), INTERVAL `n` MINUTE) AS `updated_at`
FROM `seq`;
