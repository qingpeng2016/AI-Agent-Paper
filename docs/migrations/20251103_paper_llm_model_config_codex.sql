-- Codex（OpenAI API / 本地 codex CLI 审计）端点 + 选题路由绑定 model_config
-- 生成步仍用 Claude 行；审计步绑定 Codex 行（default_model_config_id）

INSERT INTO `paper_llm_model_config` (
  `label`, `scope`, `provider_code`, `model_name`, `api_base_url`, `api_key`,
  `timeout_ms`, `max_retries`, `supports_vision`, `context_window_hint`, `extra`, `status`
) VALUES (
  'Codex GPT-5.4（审计 reviewer）',
  'platform_delegate',
  'openai',
  'gpt-5.4',
  'https://api.openai.com',
  NULL,
  3600000,
  1,
  0,
  200000,
  JSON_OBJECT('executor', 'codex', 'notes', '本地 codex exec / MCP research-review；Key 由 codex login 或 OPENAI_API_KEY'),
  'active'
);

SET @paper_llm_codex_id := LAST_INSERT_ID();

SET @paper_llm_claude_id := (
  SELECT `id` FROM `paper_llm_model_config`
  WHERE `status` = 'active' AND `provider_code` = 'anthropic'
  ORDER BY `id` DESC
  LIMIT 1
);

UPDATE `paper_llm_strategy_route`
SET `default_model_config_id` = IFNULL(@paper_llm_claude_id, `default_model_config_id`)
WHERE `behavior_key` = 'topic.stage.topic_generate';

UPDATE `paper_llm_strategy_route`
SET `default_model_config_id` = @paper_llm_codex_id
WHERE `behavior_key` = 'topic.stage.topic_audit';
