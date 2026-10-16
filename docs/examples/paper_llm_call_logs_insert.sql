-- 单次 ARIS/LLM 调用明细示例（列名 use_in_tokens / use_out_tokens）
-- 将 @manuscript_id、@model_config_id 换成真实 paper_manuscript.id、paper_llm_model_config.id

INSERT INTO `paper_llm_call_logs` (
  `manuscript_id`,
  `stage_code`,
  `behavior_key`,
  `strategy_package_id`,
  `capability_id`,
  `external_trace_id`,
  `model_config_id`,
  `model_name`,
  `request_json`,
  `response_json`,
  `use_in_tokens`,
  `use_out_tokens`,
  `latency_ms`,
  `status`
) VALUES (
  1,
  'idea_discovery',
  'topic.discovery.run',
  (SELECT `id` FROM `paper_llm_strategy_package` WHERE `code` = 'aris_local' LIMIT 1),
  'idea-discovery',
  'idea-discovery',
  (SELECT `id` FROM `paper_llm_model_config` WHERE `status` = 'active' ORDER BY `id` LIMIT 1),
  'aris:claude-sonnet-5',
  JSON_OBJECT(
    'behavior_key', 'topic.discovery.run',
    'capability_id', 'idea-discovery',
    'platform_user_prompt', '/idea-discovery …'
  ),
  JSON_OBJECT(
    'result_json', JSON_OBJECT('raw', 'checkpoint'),
    'external_trace_id', 'idea-discovery',
    'raw_text', '…'
  ),
  89,
  9,
  1200,
  200
);
