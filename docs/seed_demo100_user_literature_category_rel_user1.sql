-- Demo: 文献分类关联（user1, id 1–100）；主分类 + 部分文献多分类
DELETE FROM `paper_user_literature_category_rel`
WHERE `user_literature_id` BETWEEN 1 AND 100;

INSERT IGNORE INTO `paper_user_literature_category_rel` (`user_literature_id`, `category_id`)
SELECT `id`, ((`id` - 1) % 5) + 1
FROM `paper_user_literature`
WHERE `user_id` = 1 AND `id` BETWEEN 1 AND 100 AND `status` = 'active';

INSERT IGNORE INTO `paper_user_literature_category_rel` (`user_literature_id`, `category_id`)
SELECT `id`, ((`id`) % 5) + 1
FROM `paper_user_literature`
WHERE `user_id` = 1 AND `id` BETWEEN 1 AND 100 AND `status` = 'active'
  AND `id` % 6 = 0
  AND ((`id`) % 5) + 1 <> ((`id` - 1) % 5) + 1;
