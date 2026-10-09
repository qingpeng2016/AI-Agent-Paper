ALTER TABLE `paper_output_literature_review`
  ADD COLUMN `experiment_plan_llm_request` JSON DEFAULT NULL COMMENT '生成实验方案 LLM 请求（system/user/vars 等）' AFTER `experiment_plan_id`,
  ADD COLUMN `experiment_plan_llm_response` JSON DEFAULT NULL COMMENT '生成实验方案 LLM 返回（原文/解析/usage/error）' AFTER `experiment_plan_llm_request`;
