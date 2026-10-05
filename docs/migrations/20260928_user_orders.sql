-- 订单 / 支付 / 回调 / 订阅（若已执行 ai-platform-schema 交易段可跳过）

CREATE TABLE IF NOT EXISTS `user_orders` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_no`        VARCHAR(64)  NOT NULL COMMENT '业务订单号',
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `product_id`      BIGINT UNSIGNED NOT NULL,
  `order_type`      VARCHAR(32)  NOT NULL DEFAULT 'purchase' COMMENT 'purchase|renewal|upgrade|quota_addon',
  `user_subscription_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `quantity`        INT          NOT NULL DEFAULT 1,
  `unit_price`      DECIMAL(16,2) NOT NULL COMMENT '下单单价（元）',
  `status`          VARCHAR(32)  NOT NULL,
  `total_amount`    DECIMAL(16,2) NOT NULL COMMENT '应付总额（元）',
  `currency`        CHAR(3)      NOT NULL DEFAULT 'CNY',
  `enterprise_invoice` TINYINT(1) NOT NULL DEFAULT 0,
  `pay_channel`     VARCHAR(32)  DEFAULT NULL,
  `out_trade_no`    VARCHAR(64)  DEFAULT NULL,
  `third_trade_no`  VARCHAR(128) DEFAULT NULL,
  `raw_request_json`  JSON       DEFAULT NULL,
  `raw_notify_json`   JSON       DEFAULT NULL,
  `paid_at`         DATETIME     DEFAULT NULL,
  `expire_at`       DATETIME     NOT NULL,
  `closed_at`       DATETIME     DEFAULT NULL,
  `fail_reason`     VARCHAR(512) DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_orders_order_no` (`order_no`),
  UNIQUE KEY `uk_user_orders_out_trade_no` (`out_trade_no`),
  KEY `idx_user_orders_user_status` (`user_id`, `status`),
  KEY `idx_user_orders_expire` (`expire_at`),
  KEY `idx_user_orders_product` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `payment_callbacks` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `channel`         VARCHAR(32)  NOT NULL,
  `idempotency_key` VARCHAR(128) NOT NULL,
  `payload_json`    JSON         NOT NULL,
  `signature_ok`    TINYINT(1)   NOT NULL DEFAULT 0,
  `process_result`  VARCHAR(32)  NOT NULL,
  `processed_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_callbacks_idem` (`channel`, `idempotency_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_wallet_flows` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`             BIGINT UNSIGNED NOT NULL,
  `type`                VARCHAR(32)  NOT NULL,
  `amount`              DECIMAL(16,2) NOT NULL COMMENT '正入负出（元）',
  `balance_after`       DECIMAL(16,2) DEFAULT NULL COMMENT '变动后余额（元）',
  `currency`            CHAR(3)      NOT NULL DEFAULT 'CNY',
  `ref_type`            VARCHAR(32)  DEFAULT NULL,
  `ref_id`              BIGINT UNSIGNED DEFAULT NULL,
  `remark`              VARCHAR(512) DEFAULT NULL,
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_wallet_flows_ref` (`type`, `ref_type`, `ref_id`),
  KEY `idx_user_wallet_flows_user_time` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_subscriptions` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `product_id`        BIGINT UNSIGNED NOT NULL,
  `orders`            JSON         NOT NULL DEFAULT (JSON_ARRAY()),
  `products_category_name` VARCHAR(32)  NOT NULL,
  `sku_product_name`  VARCHAR(128) NOT NULL,
  `base_limit_tokens` BIGINT       NOT NULL,
  `limit_tokens`      BIGINT       NOT NULL,
  `used_tokens`       BIGINT       NOT NULL DEFAULT 0,
  `started_at`        DATETIME     NOT NULL,
  `expires_at`        DATETIME     NOT NULL,
  `period_start`      DATETIME     NOT NULL,
  `period_end`        DATETIME     NOT NULL,
  `status`            VARCHAR(32)  NOT NULL DEFAULT 'active',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_subscriptions_user` (`user_id`),
  KEY `idx_user_subscriptions_status_expires` (`status`, `expires_at`),
  KEY `idx_user_subscriptions_period_end` (`period_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_invoices` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id`        BIGINT UNSIGNED NOT NULL,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `invoice_type`    VARCHAR(32)  NOT NULL,
  `title`           VARCHAR(256) NOT NULL,
  `tax_no`          VARCHAR(64)  DEFAULT NULL,
  `amount`          DECIMAL(16,2) NOT NULL,
  `status`          VARCHAR(32)  NOT NULL,
  `file_url`        VARCHAR(512) DEFAULT NULL,
  `issued_at`       DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_invoices_user` (`user_id`),
  KEY `idx_user_invoices_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
