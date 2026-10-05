-- 资金流水 Tab 分页测试：user_id=2 插入 100 条（remark 含 TSTW 便于清理）
-- page_size=9 时约 12 页
-- 清理：DELETE FROM user_wallet_flows WHERE user_id = 2 AND remark LIKE 'TSTW%';

SET NAMES utf8mb4;

INSERT INTO `user_wallet_flows` (
  `user_id`,
  `type`,
  `amount`,
  `balance_after`,
  `currency`,
  `ref_type`,
  `ref_id`,
  `remark`,
  `created_at`
)
WITH RECURSIVE `seq` AS (
  SELECT 1 AS `n`
  UNION ALL
  SELECT `n` + 1 FROM `seq` WHERE `n` < 100
)
SELECT
  2 AS `user_id`,
  ELT(1 + (`n` % 5), 'recharge', 'pay', 'refund', 'commission', 'withdraw') AS `type`,
  CASE ELT(1 + (`n` % 5), 'recharge', 'pay', 'refund', 'commission', 'withdraw')
    WHEN 'pay' THEN -89.00
    WHEN 'withdraw' THEN -50.00
    WHEN 'recharge' THEN 500.00
    WHEN 'refund' THEN 89.00
    WHEN 'commission' THEN -12.50
    ELSE 12.50
  END AS `amount`,
  9000.00 + (`n` * 0.01) AS `balance_after`,
  'CNY' AS `currency`,
  'seed' AS `ref_type`,
  `n` AS `ref_id`,
  CONCAT('TSTW 分页测试流水 #', LPAD(`n`, 3, '0')) AS `remark`,
  DATE_SUB(NOW(), INTERVAL `n` MINUTE) AS `created_at`
FROM `seq`;
