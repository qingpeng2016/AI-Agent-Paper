-- 文献综述 → 生成实验方案：LLM 环节 binding + prompt（stage_code=experiment_plan）

SET @paper_llm_model_id = (
  SELECT c.`id` FROM `paper_llm_model_config` c
  WHERE c.`provider_code` = 'anthropic' AND c.`status` = 'active'
  ORDER BY c.`id` DESC LIMIT 1
);

INSERT INTO `paper_llm_workflow_binding` (`stage_code`, `stage_name`, `model_config_id`, `status`) VALUES
  ('experiment_plan', '生成实验方案', @paper_llm_model_id, 'active')
ON DUPLICATE KEY UPDATE
  `stage_name` = VALUES(`stage_name`),
  `model_config_id` = VALUES(`model_config_id`),
  `status` = VALUES(`status`),
  `updated_at` = CURRENT_TIMESTAMP;

INSERT INTO `paper_llm_prompt_template` (`stage_code`, `stage_name`, `template_body`, `status`) VALUES
  (
    'experiment_plan',
    '生成实验方案',
    '环节：生成实验方案（experiment planning）。方向：{{direction}}。目标期刊/会议：{{venue}}。须以用户消息中的 LiteratureReview（文献综述全文/摘要）与选题 Corpus 为依据，不得编造未给出的基线或数据。输出可审稿的实验计划：核心假设、主实验与对照基线、评价指标、消融设计、算力/数据资源与时间线（写入 paper_output_experiment_plan：title、summary、content_medium、meta）。严格 JSON，字段以用户消息 schema 为准（含 experiment_plan）。',
    'active'
  )
ON DUPLICATE KEY UPDATE
  `stage_name` = VALUES(`stage_name`),
  `template_body` = VALUES(`template_body`),
  `status` = VALUES(`status`),
  `updated_at` = CURRENT_TIMESTAMP;
