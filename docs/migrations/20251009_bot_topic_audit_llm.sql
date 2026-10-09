INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`)
VALUES (
  'ai_agent_paper',
  'topic_audit_llm',
  15,
  '选题 audit=running：异步调用 LLM 审查结论+文献综述并写入 paper_output_literature_review',
  1,
  1
)
ON DUPLICATE KEY UPDATE
  `interval_seconds` = VALUES(`interval_seconds`),
  `description` = VALUES(`description`),
  `is_enabled` = VALUES(`is_enabled`),
  `is_strategy_enabled` = VALUES(`is_strategy_enabled`),
  `updated_at` = CURRENT_TIMESTAMP;
