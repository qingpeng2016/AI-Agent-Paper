ALTER TABLE `paper_llm_workflow_binding`
  DROP INDEX `idx_paper_llm_bind_lookup`,
  DROP INDEX `idx_paper_llm_bind_ms`;

ALTER TABLE `paper_llm_workflow_binding`
  DROP COLUMN `user_id`,
  DROP COLUMN `manuscript_id`,
  DROP COLUMN `intensity_code`,
  DROP COLUMN `audit_level_code`,
  DROP COLUMN `note`;

ALTER TABLE `paper_llm_workflow_binding`
  ADD KEY `idx_paper_llm_bind_lookup` (`module_code`, `stage_code`, `llm_role`, `status`, `priority`);
