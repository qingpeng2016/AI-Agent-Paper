-- 累计用量仅保留 use_in_tokens / use_out_tokens

ALTER TABLE `paper_llm_model_config`
  DROP COLUMN `tokens_used_total`;
