-- 已有旧表：删列 + 换索引（保留 template_body 数据）
-- 或空库：直接用 docs/all.sql 中的 CREATE

ALTER TABLE `paper_llm_prompt_template`
  DROP INDEX `idx_paper_llm_prompt_lookup`,
  DROP INDEX `idx_paper_llm_prompt_ms`;

ALTER TABLE `paper_llm_prompt_template`
  DROP COLUMN `user_id`,
  DROP COLUMN `intensity_code`,
  DROP COLUMN `audit_level_code`,
  DROP COLUMN `label`,
  DROP COLUMN `version`,
  DROP COLUMN `variables`,
  DROP COLUMN `note`;

ALTER TABLE `paper_llm_prompt_template`
  DROP COLUMN `manuscript_id`;

ALTER TABLE `paper_llm_prompt_template`
  ADD KEY `idx_paper_llm_prompt_lookup` (
    `module_code`,
    `stage_code`,
    `llm_role`,
    `message_role`,
    `status`,
    `priority`
  );
