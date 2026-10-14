-- ARIS 策略包：本地 inline（子进程 paper_strategy_invoke.py），非 http OpenAPI

UPDATE `paper_llm_strategy_package`
SET
  `name` = 'ARIS 本地（inline CLI）',
  `provider_kind` = 'inline',
  `base_url` = NULL,
  `auth_type` = 'none',
  `manifest_version` = '1',
  `capabilities` = JSON_ARRAY('idea-creator', 'research-review', 'idea-discovery', 'research-lit'),
  `status` = 'active',
  `updated_at` = CURRENT_TIMESTAMP
WHERE `code` = 'aris';

INSERT INTO `paper_llm_strategy_package` (
  `code`, `name`, `provider_kind`, `auth_type`, `manifest_version`, `capabilities`, `status`
)
SELECT
  'aris',
  'ARIS 本地（inline CLI）',
  'inline',
  'none',
  '1',
  JSON_ARRAY('idea-creator', 'research-review', 'idea-discovery', 'research-lit'),
  'active'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `paper_llm_strategy_package` WHERE `code` = 'aris');

-- 可选：选题发现改走 ARIS（capability_id = ARIS skill 名）
-- UPDATE `paper_llm_strategy_route` r
-- JOIN `paper_llm_strategy_package` p ON p.`code` = 'aris'
-- SET r.`strategy_package_id` = p.`id`,
--     r.`capability_id` = CASE r.`behavior_key`
--       WHEN 'topic.stage.topic_generate' THEN 'idea-creator'
--       WHEN 'topic.stage.topic_audit' THEN 'research-review'
--       ELSE r.`capability_id`
--     END,
--     r.`default_model_config_id` = 0,
--     r.`updated_at` = CURRENT_TIMESTAMP
-- WHERE r.`behavior_key` IN ('topic.stage.topic_generate', 'topic.stage.topic_audit');
