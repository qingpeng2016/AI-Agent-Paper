-- Demo: 5 个常用图表分类（快速联调）
-- 分页压测请用 seed_demo100_paper_user_figure_category.sql（100 条）
INSERT INTO `paper_user_figure_category` (`id`, `name`, `status`, `created_at`, `updated_at`)
VALUES
(1, '主文 Figure', 'active', NOW(), NOW()),
(2, '补充材料', 'active', NOW(), NOW()),
(3, '实验结果', 'active', NOW(), NOW()),
(4, '方法示意图', 'active', NOW(), NOW()),
(5, '待替换草稿', 'active', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `status` = VALUES(`status`),
  `updated_at` = VALUES(`updated_at`);
