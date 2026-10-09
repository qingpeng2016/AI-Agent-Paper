ALTER TABLE `paper_output_literature_review`
  ADD COLUMN `experiment_plan_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联 paper_output_experiment_plan.id' AFTER `meta`,
  ADD KEY `idx_paper_lit_review_experiment_plan` (`experiment_plan_id`);
