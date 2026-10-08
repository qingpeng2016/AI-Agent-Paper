-- 已有库：密文列改为明文 api_key（执行前请自行备份）
-- 若 api_key_ciphertext 里是 UNHEX 密文，需先用 Python open 或重新 INSERT 明文 Key

ALTER TABLE `paper_llm_model_config`
  ADD COLUMN `api_key` VARCHAR(512) NULL COMMENT 'LLM API Key 明文' AFTER `api_base_url`;

ALTER TABLE `paper_llm_model_config`
  DROP COLUMN `api_key_ciphertext`;

-- 已有行：写入明文 Key（把 sk-ant-... 换成你的）
-- UPDATE `paper_llm_model_config` SET `api_key` = 'sk-ant-api03-...' WHERE `id` = 1;
-- 删 user_id: docs/migrations/20261008_paper_llm_model_config_drop_user_id.sql
