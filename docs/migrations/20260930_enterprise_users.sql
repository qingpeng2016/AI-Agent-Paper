-- 企业成员信息（归属企业主账号，可选绑定商城 users、关联 enterprise_inquiry）

CREATE TABLE IF NOT EXISTS `enterprise_users` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `owner_user_id`         BIGINT UNSIGNED NOT NULL COMMENT '企业主账号 users.id',
  `enterprise_inquiry_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'enterprise_inquiry.id，0 表示未关联',
  `user_id`               BIGINT UNSIGNED DEFAULT NULL COMMENT '成员商城账号 users.id，未绑定则为 NULL',
  `member_name`           VARCHAR(128) NOT NULL COMMENT '成员姓名',
  `email`                 VARCHAR(255) DEFAULT NULL COMMENT '工作邮箱',
  `phone`                 VARCHAR(32)  DEFAULT NULL COMMENT '手机号',
  `department`            VARCHAR(128) DEFAULT NULL COMMENT '部门',
  `job_title`             VARCHAR(128) DEFAULT NULL COMMENT '职位',
  `role`                  VARCHAR(32)  NOT NULL DEFAULT 'member' COMMENT 'owner|admin|member',
  `status`                VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'invited|active|disabled|left',
  `remark`                VARCHAR(512) DEFAULT NULL,
  `created_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_enterprise_users_owner` (`owner_user_id`),
  KEY `idx_enterprise_users_inquiry` (`enterprise_inquiry_id`),
  KEY `idx_enterprise_users_user` (`user_id`),
  KEY `idx_enterprise_users_owner_status` (`owner_user_id`, `status`),
  UNIQUE KEY `uk_enterprise_users_owner_email` (`owner_user_id`, `email`),
  UNIQUE KEY `uk_enterprise_users_owner_user` (`owner_user_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业成员信息';
