-- 优惠券：活动模板 + 用户持券

CREATE TABLE IF NOT EXISTS `coupon_campaigns` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`                  VARCHAR(32)  NOT NULL COMMENT '活动码，如 WELCOME',
  `name`                  VARCHAR(128) NOT NULL COMMENT '内部名称',
  `title`                 VARCHAR(128) NOT NULL COMMENT 'C 端标题',
  `subtitle`              VARCHAR(512) DEFAULT NULL COMMENT 'C 端副文案',
  `discount_type`         VARCHAR(16)  NOT NULL COMMENT 'fixed_amount|percent',
  `discount_value`        DECIMAL(16,2) NOT NULL COMMENT '减免金额(元)或折扣百分比(0-100)',
  `min_order_amount`      DECIMAL(16,2) NOT NULL DEFAULT 0 COMMENT '最低订单金额(元)',
  `valid_days`            INT UNSIGNED NOT NULL DEFAULT 30 COMMENT '领取后有效天数',
  `auto_grant_on_register` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '新用户注册自动发放',
  `enabled`               TINYINT(1) NOT NULL DEFAULT 1,
  `created_at`            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_coupon_campaigns_code` (`code`),
  KEY `idx_coupon_campaigns_register` (`auto_grant_on_register`, `enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券活动';

CREATE TABLE IF NOT EXISTS `user_coupons` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`          BIGINT UNSIGNED NOT NULL,
  `campaign_id`      BIGINT UNSIGNED NOT NULL,
  `coupon_code`      VARCHAR(40)  NOT NULL COMMENT '用户券码 UC-{id} 等',
  `discount_type`    VARCHAR(16)  NOT NULL,
  `discount_value`   DECIMAL(16,2) NOT NULL,
  `min_order_amount` DECIMAL(16,2) NOT NULL DEFAULT 0,
  `status`           VARCHAR(16)  NOT NULL DEFAULT 'available' COMMENT 'available|used|expired',
  `valid_from`       DATETIME     NOT NULL,
  `valid_until`      DATETIME     NOT NULL,
  `used_at`          DATETIME     DEFAULT NULL,
  `order_id`         BIGINT UNSIGNED DEFAULT NULL,
  `created_at`       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_coupons_code` (`coupon_code`),
  UNIQUE KEY `uk_user_coupons_user_campaign` (`user_id`, `campaign_id`),
  KEY `idx_user_coupons_user_status` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户优惠券';

INSERT INTO `coupon_campaigns` (
  `code`, `name`, `title`, `subtitle`, `discount_type`, `discount_value`,
  `min_order_amount`, `valid_days`, `auto_grant_on_register`, `enabled`
) VALUES (
  'WELCOME',
  '新用户注册礼',
  '新人专享优惠券',
  '注册即领，首单立减，畅享 AI 套餐',
  'fixed_amount',
  20.00,
  0.00,
  30,
  1,
  1
);
