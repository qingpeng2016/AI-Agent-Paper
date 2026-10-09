-- 文献综述版本以 created_at 排序展示，不再使用 is_current
ALTER TABLE `paper_output_literature_review`
  DROP INDEX `idx_paper_output_lit_review_ms`;

ALTER TABLE `paper_output_literature_review`
  DROP COLUMN `is_current`;

ALTER TABLE `paper_output_literature_review`
  ADD KEY `idx_paper_output_lit_review_ms` (`manuscript_id`, `created_at`);
