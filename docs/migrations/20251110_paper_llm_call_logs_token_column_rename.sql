-- 与 model_config 命名对齐（若库中仍是 prompt_tokens / completion_tokens）

ALTER TABLE `paper_llm_call_logs`
  CHANGE COLUMN `prompt_tokens` `use_in_tokens` INT DEFAULT NULL COMMENT '输入 token',
  CHANGE COLUMN `completion_tokens` `use_out_tokens` INT DEFAULT NULL COMMENT '输出 token';

-- 验表 INSERT 见 docs/examples/paper_llm_call_logs_insert.sql
