ALTER TABLE `paper_llm_prompt_template`
  DROP INDEX `idx_paper_llm_prompt_lookup`;

ALTER TABLE `paper_llm_prompt_template`
  DROP INDEX `idx_paper_llm_prompt_ms`;

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
