-- 平台支付通道配置（与 user_orders 支付字段区分）

CREATE TABLE IF NOT EXISTS `payments` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`            VARCHAR(32)  NOT NULL COMMENT '通道键：alipay|wechat|paypal',
  `display_name`    VARCHAR(64)  NOT NULL COMMENT '前台展示名',
  `driver`          VARCHAR(32)  NOT NULL DEFAULT 'epay' COMMENT 'epay|alipay_official|wechat_official',
  `api_base_url`    VARCHAR(512) NOT NULL DEFAULT '' COMMENT '易支付等平台 API 根地址，官方直连可留空',
  `merchant_id`     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '商户号 / PID / mch_id',
  `merchant_secret` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '商户密钥（生产环境建议加密存储）',
  `app_id`          VARCHAR(64)  DEFAULT NULL COMMENT '微信 AppID 等（可选）',
  `extra_json`      JSON         DEFAULT NULL COMMENT '证书路径、v3 密钥等扩展配置',
  `notify_path`     VARCHAR(128) DEFAULT NULL COMMENT '相对回调路径，如 /api/v1/payments/notify/alipay',
  `is_enabled`      TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否启用',
  `sort_order`      INT          NOT NULL DEFAULT 0,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payments_code` (`code`),
  KEY `idx_payments_enabled_sort` (`is_enabled`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付通道配置';

INSERT INTO `payments` (
  `code`, `display_name`, `driver`, `api_base_url`, `merchant_id`, `merchant_secret`,
  `app_id`, `extra_json`, `notify_path`, `is_enabled`, `sort_order`
) VALUES
(
  'alipay',
  '支付宝',
  'epay',
  'https://pay.example.com/',
  '10001',
  'replace-with-alipay-merchant-key',
  NULL,
  JSON_OBJECT('pay_type', 'alipay'),
  '/api/v1/payments/notify/alipay',
  1,
  10
),
(
  'wechat',
  '微信支付',
  'epay',
  'https://pay.example.com/',
  '10001',
  'replace-with-wechat-merchant-key',
  'wx0000000000000000',
  JSON_OBJECT('pay_type', 'wxpay'),
  '/api/v1/payments/notify/wechat',
  1,
  20
)
ON DUPLICATE KEY UPDATE
  `display_name` = VALUES(`display_name`),
  `driver` = VALUES(`driver`),
  `api_base_url` = VALUES(`api_base_url`),
  `merchant_id` = VALUES(`merchant_id`),
  `merchant_secret` = VALUES(`merchant_secret`),
  `app_id` = VALUES(`app_id`),
  `extra_json` = VALUES(`extra_json`),
  `notify_path` = VALUES(`notify_path`),
  `is_enabled` = VALUES(`is_enabled`),
  `sort_order` = VALUES(`sort_order`);
