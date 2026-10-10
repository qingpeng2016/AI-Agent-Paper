RENAME TABLE `paper_llm_prompt_template` TO `paper_template_llm_prompt`;

ALTER TABLE `paper_template_llm_prompt`
  RENAME INDEX `uk_paper_llm_prompt_stage` TO `uk_paper_template_llm_prompt_stage`,
  RENAME INDEX `idx_paper_llm_prompt_status` TO `idx_paper_template_llm_prompt_status`;
