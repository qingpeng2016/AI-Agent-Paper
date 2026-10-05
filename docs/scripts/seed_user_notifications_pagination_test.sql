-- 站内消息测试：user_id=2 插入 100 条 in_app（顶栏弹窗取最近 20 条）
-- 清理：DELETE FROM user_notifications WHERE user_id = 2 AND sent_at >= '2099-01-01 00:00:00' AND sent_at < '2099-01-02 00:00:00';

SET NAMES utf8mb4;

INSERT INTO `user_notifications` (
  `user_id`,
  `user_subscription_id`,
  `api_key_id`,
  `channel`,
  `template_code`,
  `status`,
  `sent_at`,
  `created_at`
)
WITH RECURSIVE `seq` AS (
  SELECT 1 AS `n`
  UNION ALL
  SELECT `n` + 1 FROM `seq` WHERE `n` < 100
)
SELECT
  2,
  NULL,
  NULL,
  'in_app',
  ELT(
    ((`n` - 1) % 7) + 1,
    'subscription_activated',
    'subscription_renewed',
    'subscription_upgraded',
    'subscription_quota_added',
    'subscription_expired',
    'renew_reminder',
    'key_issued'
  ),
  IF(`n` % 4 = 0, 'read', 'sent'),
  DATE_ADD('2099-01-01 00:00:00', INTERVAL `n` MINUTE),
  DATE_SUB(NOW(), INTERVAL `n` HOUR)
FROM `seq`;
