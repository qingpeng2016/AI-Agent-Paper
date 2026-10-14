-- 选题发现：合并 generate_ideas + audit + experiment_plan 为 topic_discovery 单步 + 单 bot

UPDATE `paper_llm_template_topic`
SET `stage_code` = 'topic_generate', `stage_name` = '选题生成包', `updated_at` = CURRENT_TIMESTAMP
WHERE `stage_code` = 'generate_ideas';

UPDATE `paper_llm_template_topic`
SET `stage_code` = 'topic_audit', `stage_name` = '选题审计', `updated_at` = CURRENT_TIMESTAMP
WHERE `stage_code` = 'audit';

UPDATE `paper_llm_template_topic`
SET `status` = 'inactive', `updated_at` = CURRENT_TIMESTAMP
WHERE `stage_code` = 'experiment_plan';

UPDATE `paper_exec_intensity_topic`
SET `exec_rounds` = CASE `code`
  WHEN 'fast' THEN 1
  WHEN 'balanced' THEN 2
  WHEN 'deep' THEN 3
  ELSE `exec_rounds`
END;

ALTER TABLE `paper_exec_intensity_topic`
  MODIFY COLUMN `exec_rounds` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '选题生成↔审计迭代轮数（Corpus 冻结）';

DELETE FROM `bot_schedule_config`
WHERE `module` = 'ai_agent_paper'
  AND `task_name` IN ('topic_generate_ideas_llm', 'topic_audit_llm', 'experiment_plan_llm');

INSERT INTO `bot_schedule_config` (`module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`)
VALUES (
  'ai_agent_paper',
  'llm_topic_discovery',
  2,
  '选题 topic_discovery=running：生成包↔三门禁审计迭代，写入综述与实验方案',
  1,
  1
);
