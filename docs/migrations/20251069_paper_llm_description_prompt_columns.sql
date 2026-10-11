-- prompt_type → prompt_name；原 prompt_name → prompt_desc
ALTER TABLE `paper_llm_description`
  DROP INDEX `uk_paper_llm_description_scope`;

ALTER TABLE `paper_llm_description`
  CHANGE COLUMN `prompt_name` `prompt_desc` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '运营/后台展示名',
  CHANGE COLUMN `prompt_type` `prompt_name` VARCHAR(128) NOT NULL COMMENT '类型名：门禁约束、期刊会议约束、选题环节话术等';

ALTER TABLE `paper_llm_description`
  ADD UNIQUE KEY `uk_paper_llm_description_scope` (`model_config_id`, `prompt_code`, `prompt_desc`);
