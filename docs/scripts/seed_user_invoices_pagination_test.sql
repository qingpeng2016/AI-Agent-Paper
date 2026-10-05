-- 发票 Tab 分页测试：user_id=2 插入 100 条（title 含 TSTI 便于清理）
-- 需该用户至少有一条 user_orders；order_id 在现有订单间轮询
-- 清理：DELETE FROM user_invoices WHERE user_id = 2 AND title LIKE 'TSTI%';

SET NAMES utf8mb4;

INSERT INTO `user_invoices` (
  `order_id`,
  `user_id`,
  `invoice_type`,
  `title`,
  `tax_no`,
  `amount`,
  `status`,
  `issued_at`,
  `created_at`,
  `updated_at`
)
WITH RECURSIVE `seq` AS (
  SELECT 1 AS `n`
  UNION ALL
  SELECT `n` + 1 FROM `seq` WHERE `n` < 100
),
`ord` AS (
  SELECT
    `id`,
    ROW_NUMBER() OVER (ORDER BY `id` DESC) AS `rn`,
    COUNT(*) OVER () AS `cnt`
  FROM `user_orders`
  WHERE `user_id` = 2
)
SELECT
  COALESCE(
    (SELECT `id` FROM `ord` WHERE `rn` = 1 + ((`seq`.`n` - 1) % GREATEST((SELECT MAX(`cnt`) FROM `ord`), 1))),
    1
  ) AS `order_id`,
  2 AS `user_id`,
  'enterprise_vat' AS `invoice_type`,
  CONCAT('TSTI 测试抬头 #', LPAD(`seq`.`n`, 3, '0')) AS `title`,
  CONCAT('91310000TSTI', LPAD(`seq`.`n`, 6, '0')) AS `tax_no`,
  IF(`seq`.`n` % 2 = 0, 899.00, 89.00) AS `amount`,
  ELT(1 + (`seq`.`n` % 3), 'pending', 'issued', 'failed') AS `status`,
  IF(`seq`.`n` % 3 = 2, DATE_SUB(NOW(), INTERVAL `seq`.`n` MINUTE), NULL) AS `issued_at`,
  DATE_SUB(NOW(), INTERVAL `seq`.`n` MINUTE) AS `created_at`,
  DATE_SUB(NOW(), INTERVAL `seq`.`n` MINUTE) AS `updated_at`
FROM `seq`;
