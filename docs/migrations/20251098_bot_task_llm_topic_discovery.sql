-- bot_schedule_config：topic_discovery_llm → llm_topic_discovery（与调度器 task_name 一致）

UPDATE `bot_schedule_config`
SET
  `task_name` = 'llm_topic_discovery',
  `updated_at` = CURRENT_TIMESTAMP
WHERE `module` = 'ai_agent_paper'
  AND `task_name` = 'topic_discovery_llm';

INSERT INTO `bot_schedule_config` (
  `module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`
)
SELECT
  'ai_agent_paper',
  'llm_topic_discovery',
  2,
  '选题 topic_discovery=running：生成包↔三门禁审计迭代，写入综述与实验方案',
  1,
  1
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM `bot_schedule_config`
  WHERE `module` = 'ai_agent_paper' AND `task_name` = 'llm_topic_discovery'
);
