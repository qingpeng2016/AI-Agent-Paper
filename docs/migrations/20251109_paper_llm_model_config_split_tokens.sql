-- paper_llm_model_config：输入/输出 token 分开累计
-- paper_llm_call_logs：沿用已有 prompt_tokens/completion_tokens 列名（单次调用明细），未新增列

ALTER TABLE `paper_llm_model_config`
  ADD COLUMN `use_in_tokens` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计输入 token' AFTER `context_window_hint`,
  ADD COLUMN `use_out_tokens` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计输出 token' AFTER `use_in_tokens`;
