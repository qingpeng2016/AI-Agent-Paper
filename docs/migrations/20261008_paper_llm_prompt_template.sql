-- MVP：一条全局 system 即可；各模块具体任务句建议在代码里拼 user message
-- 重插前: TRUNCATE TABLE `paper_llm_prompt_template`;

SET NAMES utf8mb4;

INSERT INTO `paper_llm_prompt_template` (
  `module_code`, `stage_code`, `llm_role`, `message_role`,
  `template_body`, `priority`, `status`
) VALUES (
  '*', NULL, 'system', 'system',
  'You are an expert academic research assistant for the Paper Agent platform. Help with topic discovery, literature review, experiment planning, writing, and review. Be precise, cite evidence when provided, and state uncertainty clearly. Default to clear scholarly English.',
  10, 'active'
);
