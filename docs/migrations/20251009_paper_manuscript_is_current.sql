ALTER TABLE `paper_manuscript`
  ADD COLUMN `is_current` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否工作台当前论文（每用户 active 至多一条为 1）' AFTER `status`;

UPDATE `paper_manuscript` m
INNER JOIN (
  SELECT `user_id`, MAX(`id`) AS `latest_id`
  FROM `paper_manuscript`
  WHERE `status` = 'active'
  GROUP BY `user_id`
) t ON m.`id` = t.`latest_id`
SET m.`is_current` = 1;
