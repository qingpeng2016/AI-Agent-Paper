-- 邀请返利：VIP 档位、用户上下级与推广域、佣金明细/提现/收款配置

CREATE TABLE IF NOT EXISTS `vip_config` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `level_label`       VARCHAR(64)  NOT NULL COMMENT '等级名称，如标准推广',
  `min_valid_invites` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '有效邀请人数下限（≥）',
  `min_invitee_paid_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '直属下级已完成订单实付累计（元，≥）',
  `rate_percent`      DECIMAL(5,2) NOT NULL COMMENT '返佣比例（%）',
  `sort_order`        INT          NOT NULL DEFAULT 0 COMMENT '展示与匹配顺序，越大门槛越高',
  `enabled`           TINYINT(1)   NOT NULL DEFAULT 1,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_vip_config_enabled_sort` (`enabled`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请返利 VIP/返佣档位配置';

INSERT INTO `vip_config` (`id`, `level_label`, `min_valid_invites`, `min_invitee_paid_amount`, `rate_percent`, `sort_order`, `enabled`)
VALUES
  (1, '入门推广', 0,  0.00,    3.00, 10, 1),
  (2, '标准推广', 3,  0.00,    5.00, 20, 1),
  (3, '高级推广', 5,  500.00,  8.00, 30, 1),
  (4, '合伙人',   20, 2000.00, 12.00, 40, 1)
ON DUPLICATE KEY UPDATE
  `level_label` = VALUES(`level_label`),
  `min_valid_invites` = VALUES(`min_valid_invites`),
  `min_invitee_paid_amount` = VALUES(`min_invitee_paid_amount`),
  `rate_percent` = VALUES(`rate_percent`),
  `sort_order` = VALUES(`sort_order`),
  `enabled` = VALUES(`enabled`);

ALTER TABLE `users`
  ADD COLUMN `parent_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上级用户 ID，0 表示无上级' AFTER `nickname`,
  ADD COLUMN `vip_config_id`  BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '当前 VIP 档位 vip_config.id' AFTER `parent_user_id`,
  ADD COLUMN `vip_domain`     VARCHAR(255) DEFAULT NULL COMMENT '专属推广独立域名' AFTER `vip_config_id`,
  ADD KEY `idx_users_parent_user_id` (`parent_user_id`),
  ADD KEY `idx_users_vip_config_id` (`vip_config_id`),
  ADD UNIQUE KEY `uk_users_vip_domain` (`vip_domain`);

CREATE TABLE IF NOT EXISTS `user_commission_payout_config` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL,
  `channel`    VARCHAR(16)  NOT NULL COMMENT 'alipay|wechat',
  `qr_mime`    VARCHAR(64)  DEFAULT NULL COMMENT 'image/png|image/jpeg|image/webp',
  `qr_image`   MEDIUMBLOB   DEFAULT NULL COMMENT '收款码图片二进制',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_commission_payout_user_channel` (`user_id`, `channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户佣金提现收款配置';

CREATE TABLE IF NOT EXISTS `user_commission_records` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `inviter_user_id`  BIGINT UNSIGNED NOT NULL COMMENT '获得返佣的用户',
  `invitee_user_id`  BIGINT UNSIGNED NOT NULL COMMENT '下单的被邀请用户',
  `order_id`         BIGINT UNSIGNED NOT NULL,
  `order_no`         VARCHAR(64)  NOT NULL,
  `product_name`     VARCHAR(128) NOT NULL DEFAULT '' COMMENT '展示用，可冗余',
  `order_amount`     DECIMAL(16,2) NOT NULL COMMENT '订单实付基数（元）',
  `rate_percent`     DECIMAL(5,2) NOT NULL COMMENT '结算时返佣比例（%）',
  `rebate_amount`    DECIMAL(16,2) NOT NULL COMMENT '返利金额（元）',
  `status`           VARCHAR(32)  NOT NULL DEFAULT 'settled' COMMENT 'settled|reversed',
  `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_commission_records_order` (`order_id`),
  KEY `idx_user_commission_records_inviter_time` (`inviter_user_id`, `created_at`),
  KEY `idx_user_commission_records_invitee` (`invitee_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请返利明细（按订单）';

CREATE TABLE IF NOT EXISTS `user_commission_withdrawals` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `amount`          DECIMAL(16,2) NOT NULL COMMENT '提现金额（元）',
  `channel`         VARCHAR(16)  NOT NULL COMMENT 'alipay|wechat',
  `payout_qr_url`   VARCHAR(512) NOT NULL DEFAULT '' COMMENT '申请时收款码快照',
  `status`          VARCHAR(32)  NOT NULL DEFAULT 'pending' COMMENT 'pending|completed|failed',
  `fail_reason`     VARCHAR(512) DEFAULT NULL,
  `processed_at`    DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_commission_withdrawals_user_time` (`user_id`, `created_at`),
  KEY `idx_user_commission_withdrawals_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='佣金提现申请';
