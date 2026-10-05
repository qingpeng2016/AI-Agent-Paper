-- 邀请返利测试数据：inviter user_id=2，下级 3/4/5/6
-- 100 条返利记录 + 100 条提现记录
-- 清理见文末 DELETE

SET NAMES utf8mb4;

UPDATE `users`
SET `commission_balance` = 1288.50
WHERE `id` = 2;

DELETE FROM `user_commission_records`
WHERE `order_no` LIKE 'TST-CRB-%' AND `inviter_user_id` = 2;

DELETE FROM `user_commission_withdrawals`
WHERE `user_id` = 2 AND `payout_qr_url` = 'seed:test';

INSERT INTO `user_commission_records` (
  `inviter_user_id`,
  `invitee_user_id`,
  `order_id`,
  `order_no`,
  `product_name`,
  `order_amount`,
  `rate_percent`,
  `rebate_amount`,
  `status`,
  `created_at`,
  `updated_at`
)
WITH RECURSIVE `seq` AS (
  SELECT 1 AS `n`
  UNION ALL
  SELECT `n` + 1 FROM `seq` WHERE `n` < 100
)
SELECT
  2 AS `inviter_user_id`,
  3 + ((`n` - 1) % 4) AS `invitee_user_id`,
  900000 + `n` AS `order_id`,
  CONCAT('TST-CRB-', LPAD(`n`, 5, '0')) AS `order_no`,
  ELT(1 + (`n` % 4), 'ChatGPT Plus 月卡', 'GPT Go 月卡', 'Claude Pro 月卡', 'Gemini Advanced 月卡') AS `product_name`,
  ROUND(89.00 + (`n` % 7) * 30.00, 2) AS `order_amount`,
  5.00 AS `rate_percent`,
  ROUND((89.00 + (`n` % 7) * 30.00) * 0.05, 2) AS `rebate_amount`,
  IF(`n` % 17 = 0, 'reversed', 'settled') AS `status`,
  DATE_SUB('2026-09-30 12:00:00', INTERVAL `n` HOUR) AS `created_at`,
  DATE_SUB('2026-09-30 12:00:00', INTERVAL `n` HOUR) AS `updated_at`
FROM `seq`;

INSERT INTO `user_commission_withdrawals` (
  `user_id`,
  `amount`,
  `channel`,
  `payout_qr_url`,
  `status`,
  `fail_reason`,
  `processed_at`,
  `created_at`,
  `updated_at`
)
WITH RECURSIVE `seq` AS (
  SELECT 1 AS `n`
  UNION ALL
  SELECT `n` + 1 FROM `seq` WHERE `n` < 100
)
SELECT
  2 AS `user_id`,
  ROUND(10.00 + (`n` % 20) * 5.00, 2) AS `amount`,
  IF(`n` % 2 = 0, 'alipay', 'wechat') AS `channel`,
  'seed:test' AS `payout_qr_url`,
  ELT(1 + (`n` % 3), 'pending', 'completed', 'failed') AS `status`,
  IF(`n` % 3 = 0, '演示失败', NULL) AS `fail_reason`,
  IF(`n` % 3 = 1, NULL, DATE_SUB('2026-09-30 12:00:00', INTERVAL `n` HOUR)) AS `processed_at`,
  DATE_SUB('2026-09-30 12:00:00', INTERVAL `n` HOUR) AS `created_at`,
  DATE_SUB('2026-09-30 12:00:00', INTERVAL `n` HOUR) AS `updated_at`
FROM `seq`;

-- 清理：
-- DELETE FROM user_commission_records WHERE order_no LIKE 'TST-CRB-%' AND inviter_user_id = 2;
-- DELETE FROM user_commission_withdrawals WHERE user_id = 2 AND payout_qr_url = 'seed:test';
