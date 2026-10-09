INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`)
VALUES (
  'ai_agent_paper',
  'experiment_plan_llm',
  15,
  '文献综述 generating_experiment_plan：异步 LLM 生成实验方案并写入 paper_output_experiment_plan',
  1,
  1
)
ON DUPLICATE KEY UPDATE
  `interval_seconds` = VALUES(`interval_seconds`),
  `description` = VALUES(`description`),
  `is_enabled` = VALUES(`is_enabled`),
  `is_strategy_enabled` = VALUES(`is_strategy_enabled`),
  `updated_at` = CURRENT_TIMESTAMP;
