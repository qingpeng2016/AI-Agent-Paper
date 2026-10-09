ALTER TABLE `paper_output_literature_review`
  ADD COLUMN `paper_output_experiment_plan_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联 paper_output_experiment_plan.id' AFTER `meta`,
  ADD KEY `idx_paper_output_lit_review_exp_plan` (`paper_output_experiment_plan_id`);
