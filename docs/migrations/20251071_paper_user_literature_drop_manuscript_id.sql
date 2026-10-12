ALTER TABLE `paper_user_literature`
  DROP INDEX `idx_paper_user_lit_ms`,
  DROP COLUMN `manuscript_id`;

ALTER TABLE `paper_user_literature`
  DROP INDEX `idx_paper_user_lit_user`,
  ADD KEY `idx_paper_user_lit_user` (`user_id`, `status`, `created_at`);
