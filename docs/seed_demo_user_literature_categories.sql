-- Demo: 用户文献分类（status=active）
INSERT INTO `paper_user_literature_category` (`id`, `name`, `status`, `created_at`, `updated_at`)
VALUES
(1, '核心参考文献', 'active', NOW(), NOW()),
(2, '背景阅读', 'active', NOW(), NOW()),
(3, '方法对比', 'active', NOW(), NOW()),
(4, '待精读', 'active', NOW(), NOW()),
(5, '数据集与代码', 'active', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `status` = VALUES(`status`),
  `updated_at` = VALUES(`updated_at`);
