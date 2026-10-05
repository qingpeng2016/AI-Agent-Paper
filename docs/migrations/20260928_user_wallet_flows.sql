-- 用户资金流水（充值 / 消费 / 退款 / 佣金 / 提现等）

CREATE TABLE IF NOT EXISTS `user_wallet_flows` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`             BIGINT UNSIGNED NOT NULL,
  `type`                VARCHAR(32)  NOT NULL COMMENT 'recharge|pay|refund|commission|withdraw',
  `amount`              DECIMAL(16,2) NOT NULL COMMENT '正入负出（元）',
  `balance_after`       DECIMAL(16,2) DEFAULT NULL COMMENT '变动后账户余额（元），渠道直付可为 NULL',
  `currency`            CHAR(3)      NOT NULL DEFAULT 'CNY',
  `ref_type`            VARCHAR(32)  DEFAULT NULL COMMENT 'order|refund|withdraw|recharge',
  `ref_id`              BIGINT UNSIGNED DEFAULT NULL,
  `remark`              VARCHAR(512) DEFAULT NULL,
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_wallet_flows_ref` (`type`, `ref_type`, `ref_id`),
  KEY `idx_user_wallet_flows_user_time` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户资金流水';
