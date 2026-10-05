-- VIP 档位同步 Bot：按有效邀请人数升级 users.vip_config_id（只升不降）
-- 任务代码：application/bot/scripts/vip_level_sync/entry.go

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
  'vip_level_sync',
  3600,
  '扫描全部用户，按下级人数匹配 vip_config 最高可达档位，仅当高于当前档位时更新 vip_config_id',
  1,
  1
)
ON DUPLICATE KEY UPDATE
  `interval_seconds` = VALUES(`interval_seconds`),
  `description` = VALUES(`description`),
  `is_enabled` = VALUES(`is_enabled`),
  `is_strategy_enabled` = VALUES(`is_strategy_enabled`),
  `updated_at` = CURRENT_TIMESTAMP;
