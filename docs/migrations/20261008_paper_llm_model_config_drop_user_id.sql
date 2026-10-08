-- paper_llm_model_config 仅平台端点；按用户选模型见 paper_llm_workflow_binding.user_id

ALTER TABLE `paper_llm_model_config`
  DROP INDEX `idx_paper_llm_model_cfg_user`;

ALTER TABLE `paper_llm_model_config`
  DROP COLUMN `user_id`;

ALTER TABLE `paper_llm_model_config`
  ADD KEY `idx_paper_llm_model_cfg_status` (`status`);
