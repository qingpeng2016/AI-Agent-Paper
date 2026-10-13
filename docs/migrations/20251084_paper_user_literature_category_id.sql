-- paper_user_literature：单分类列（已由 20251085 改为多对多 rel 表并删除本列）
ALTER TABLE `paper_user_literature`
  ADD COLUMN `category_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_user_literature_category.id' AFTER `literature_title`,
  ADD KEY `idx_paper_user_lit_category` (`user_id`, `category_id`, `status`);

-- 可选：为 user1 演示数据填分类（需已执行 seed_demo100_user_literature_user1 与 seed_demo_user_literature_categories）
UPDATE `paper_user_literature`
SET `category_id` = ((`id` - 1) % 5) + 1
WHERE `user_id` = 1 AND `id` BETWEEN 1 AND 100 AND `status` = 'active';
