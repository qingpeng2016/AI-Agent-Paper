-- ARIS idea-discovery 五 phase 路由（替代 topic_generate/topic_audit 双 invoke 循环）
-- 策略包 code 必须为 aris（与 docs/all.sql 一致；勿写 aris_local）

INSERT INTO `paper_llm_strategy_package` (`code`, `name`, `status`)
VALUES ('aris', 'ARIS 本地 CLI', 'active')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `status` = VALUES(`status`);

SET @pkg_aris := (SELECT `id` FROM `paper_llm_strategy_package` WHERE `code` = 'aris' LIMIT 1);

SET @paper_llm_claude_id := (
  SELECT c.`id` FROM `paper_llm_model_config` c
  WHERE c.`status` = 'active' AND c.`provider_code` = 'anthropic'
  ORDER BY c.`id` DESC
  LIMIT 1
);

SET @paper_llm_codex_id := (
  SELECT c.`id` FROM `paper_llm_model_config` c
  WHERE c.`status` = 'active' AND c.`provider_code` = 'openai'
  ORDER BY c.`id` DESC
  LIMIT 1
);

SET @paper_llm_claude_id := IFNULL(@paper_llm_claude_id, 0);
SET @paper_llm_codex_id := IFNULL(@paper_llm_codex_id, IFNULL(@paper_llm_claude_id, 0));

INSERT INTO `paper_llm_strategy_route` (
  `behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`
) VALUES
  ('topic.stage.research_lit',    'topic_discovery', 'research_lit',    @pkg_aris, 'research-lit',             @paper_llm_claude_id, 'active'),
  ('topic.stage.idea_creator',    'topic_discovery', 'idea_creator',    @pkg_aris, 'idea-creator',             @paper_llm_claude_id, 'active'),
  ('topic.stage.novelty_check',   'topic_discovery', 'novelty_check',   @pkg_aris, 'novelty-check',            @paper_llm_codex_id,   'active'),
  ('topic.stage.research_review', 'topic_discovery', 'research_review', @pkg_aris, 'research-review',          @paper_llm_codex_id,   'active'),
  ('topic.stage.research_refine', 'topic_discovery', 'research_refine', @pkg_aris, 'research-refine-pipeline', @paper_llm_claude_id, 'active')
ON DUPLICATE KEY UPDATE
  `strategy_package_id` = VALUES(`strategy_package_id`),
  `capability_id` = VALUES(`capability_id`),
  `default_model_config_id` = VALUES(`default_model_config_id`),
  `status` = VALUES(`status`),
  `updated_at` = CURRENT_TIMESTAMP;
