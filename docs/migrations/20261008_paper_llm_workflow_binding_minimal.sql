-- 可选：binding 也收成 1 条（executor 用 Claude，reviewer 等同模型时在代码里复用）
-- TRUNCATE 后按 model_config 实际 id 改 1

INSERT INTO `paper_llm_workflow_binding` (
  `module_code`, `stage_code`, `llm_role`, `model_config_id`, `priority`, `status`
) VALUES (
  '*', NULL, 'executor', 1, 10, 'active'
);
