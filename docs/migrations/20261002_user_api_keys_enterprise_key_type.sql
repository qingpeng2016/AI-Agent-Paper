-- user_api_keys：企业归属 + 主/子 Key 类型

ALTER TABLE `user_api_keys`
  ADD COLUMN `owner_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0
    COMMENT '企业主账号 users.id；个人订阅与 user_id 相同' AFTER `user_id`,
  ADD COLUMN `enterprise_inquiry_id` BIGINT UNSIGNED NOT NULL DEFAULT 0
    COMMENT 'enterprise_inquiry.id，0 表示未关联' AFTER `owner_user_id`,
  ADD COLUMN `key_type` VARCHAR(16) NOT NULL DEFAULT 'main'
    COMMENT 'main=主Key，sub=子Key' AFTER `user_subscription_id`;

UPDATE `user_api_keys` SET `owner_user_id` = `user_id` WHERE `owner_user_id` = 0;

ALTER TABLE `user_api_keys`
  ADD KEY `idx_user_api_keys_owner` (`owner_user_id`),
  ADD KEY `idx_user_api_keys_enterprise_inquiry` (`enterprise_inquiry_id`),
  ADD KEY `idx_user_api_keys_sub_key_type` (`user_subscription_id`, `key_type`);
