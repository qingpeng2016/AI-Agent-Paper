-- 平台 LLM 配置：累计本次配置下所有成功调用的 token 总量（prompt + completion）
ALTER TABLE `paper_llm_model_config`
  ADD COLUMN `tokens_used_total` BIGINT UNSIGNED NOT NULL DEFAULT 0
    COMMENT '累计消耗 token（每次成功调用按 usage 累加 prompt+completion）'
  AFTER `context_window_hint`;
