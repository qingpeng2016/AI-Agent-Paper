ALTER TABLE `paper_literature_platform`
  DROP INDEX `idx_paper_literature_platform_active_priority`,
  CHANGE COLUMN `priority` `sort` INT NOT NULL DEFAULT 100 COMMENT '检索与 UI 展示顺序，越小越靠前',
  ADD KEY `idx_paper_literature_platform_active_sort` (`status`, `sort`, `id`);
