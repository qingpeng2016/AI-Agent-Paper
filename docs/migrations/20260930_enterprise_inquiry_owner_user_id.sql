-- enterprise_inquiry.user_id 更名为 owner_user_id（公司管理人员 users.id）

ALTER TABLE `enterprise_inquiry`
  DROP INDEX `idx_enterprise_inquiry_user`;

ALTER TABLE `enterprise_inquiry`
  CHANGE COLUMN `user_id` `owner_user_id` BIGINT UNSIGNED DEFAULT NULL
    COMMENT '公司管理人员 users.id' AFTER `id`;

ALTER TABLE `enterprise_inquiry`
  ADD KEY `idx_enterprise_inquiry_owner_user` (`owner_user_id`);
