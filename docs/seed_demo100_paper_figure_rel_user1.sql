-- Demo: 图表分类 / 论文关联（figure id 1–100, user1）
-- 依赖 seed_demo100_paper_user_figure_user1.sql、seed_demo100_paper_user_figure_category.sql

DELETE FROM `paper_user_figure_category_rel`
WHERE `figure_id` BETWEEN 1 AND 100;

INSERT IGNORE INTO `paper_user_figure_category_rel` (`figure_id`, `category_id`)
SELECT `id`, ((`id` - 1) % 100) + 1
FROM `paper_user_figure`
WHERE `user_id` = 1 AND `id` BETWEEN 1 AND 100 AND `status` = 'active';

INSERT IGNORE INTO `paper_user_figure_category_rel` (`figure_id`, `category_id`)
SELECT `id`, ((`id`) % 100) + 1
FROM `paper_user_figure`
WHERE `user_id` = 1 AND `id` BETWEEN 1 AND 100 AND `status` = 'active'
  AND `id` % 6 = 0
  AND ((`id`) % 100) + 1 <> ((`id` - 1) % 100) + 1;

DELETE FROM `paper_user_figure_manuscript_rel`
WHERE `figure_id` BETWEEN 1 AND 100;

INSERT IGNORE INTO `paper_user_figure_manuscript_rel` (`figure_id`, `manuscript_id`)
SELECT `id`, `manuscript_id`
FROM `paper_user_figure`
WHERE `user_id` = 1 AND `id` BETWEEN 1 AND 100 AND `status` = 'active';

-- 部分图表额外关联一篇论文（多论文）
INSERT IGNORE INTO `paper_user_figure_manuscript_rel` (`figure_id`, `manuscript_id`)
SELECT `id`, ((`id` % 100) + 1)
FROM `paper_user_figure`
WHERE `user_id` = 1 AND `id` BETWEEN 1 AND 100 AND `status` = 'active'
  AND `id` % 7 = 0
  AND ((`id` % 100) + 1) <> `manuscript_id`;
