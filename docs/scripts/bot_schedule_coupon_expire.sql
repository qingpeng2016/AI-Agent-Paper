-- 优惠券过期 Bot：将 status=available 且 valid_until 已过的 user_coupons 更新为 expired
-- 任务代码：application/bot/scripts/coupon_expire/entry.go

SET NAMES utf8mb4;

INSERT INTO `bot_schedule_config` (
  `module`,
  `task_name`,
  `interval_seconds`,
  `description`,
  `is_enabled`,
  `is_strategy_enabled`
)
VALUES (
  'ai_token_mall',
  'coupon_expire',
  3600,
  '扫描 user_coupons：available 且 valid_until < 当前时间则标记为 expired',
  1,
  1
)
ON DUPLICATE KEY UPDATE
  `interval_seconds` = VALUES(`interval_seconds`),
  `description` = VALUES(`description`),
  `is_enabled` = VALUES(`is_enabled`),
  `is_strategy_enabled` = VALUES(`is_strategy_enabled`),
  `updated_at` = CURRENT_TIMESTAMP;
