-- 已有库：paper_output_experiment_plan_id → experiment_plan_id
ALTER TABLE `paper_output_literature_review`
  CHANGE COLUMN `paper_output_experiment_plan_id` `experiment_plan_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联 paper_output_experiment_plan.id';

ALTER TABLE `paper_output_literature_review`
  DROP INDEX `idx_paper_output_lit_review_exp_plan`,
  ADD KEY `idx_paper_lit_review_experiment_plan` (`experiment_plan_id`);
