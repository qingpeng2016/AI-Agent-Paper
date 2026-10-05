-- 聚合 AI 平台 — 表结构（依据 docs/ai-platform-design.md）
-- 引擎：MySQL 8.0+，字符集 utf8mb4
-- 说明：Key / 上游 Key 仅存哈希或密文；金额以「分」为单位存储

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- 会员与账户
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `users` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户 ID',
  `email`           VARCHAR(255) DEFAULT NULL COMMENT '邮箱（登录）',
  `phone`           VARCHAR(32)  DEFAULT NULL COMMENT '手机号（登录）',
  `password_hash`   VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `password_plain`  VARCHAR(255) NOT NULL COMMENT '密码明文（业务要求留存，仅限受控环境）',
  `nickname`        VARCHAR(64)  DEFAULT NULL COMMENT '昵称',
  `parent_user_id`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上级用户 ID，0 无上级',
  `vip_config_id`   BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'VIP 档位 vip_config.id',
  `vip_domain`      VARCHAR(255) DEFAULT NULL COMMENT '专属推广独立域名',
  `status`          VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled|banned',
  `wallet_balance`     DECIMAL(16,2) NOT NULL DEFAULT 0 COMMENT '钱包可用余额（元）',
  `commission_balance` DECIMAL(16,2) NOT NULL DEFAULT 0 COMMENT '佣金余额（元）',
  `last_login_at`      DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_email` (`email`),
  UNIQUE KEY `uk_users_phone` (`phone`),
  KEY `idx_users_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='注册用户';

CREATE TABLE IF NOT EXISTS `vip_domain_config` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `domain`           VARCHAR(255) NOT NULL COMMENT '完整域名',
  `is_official`      TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否官网域名',
  `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_vip_domain_config_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户可选专属推广域名池';

CREATE TABLE IF NOT EXISTS `user_invoice_config` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`               BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `enterprise_inquiry_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'enterprise_inquiry.id，0 表示无',
  `profile_type`          VARCHAR(16)  NOT NULL DEFAULT 'enterprise' COMMENT 'enterprise|personal',
  `title`           VARCHAR(256) NOT NULL COMMENT '发票抬头',
  `tax_no`          VARCHAR(64)  DEFAULT NULL COMMENT '税号',
  `bank_name`       VARCHAR(128) DEFAULT NULL,
  `bank_account`    VARCHAR(64)  DEFAULT NULL,
  `address`         VARCHAR(512) DEFAULT NULL,
  `phone`           VARCHAR(32)  DEFAULT NULL,
  `is_default`      TINYINT(1)   NOT NULL DEFAULT 0,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_invoice_config_user` (`user_id`),
  KEY `idx_user_invoice_config_enterprise_inquiry` (`enterprise_inquiry_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户发票抬头配置（含企业 VAT）';

-- ---------------------------------------------------------------------------
-- 目录与 SKU（quota / 池路由引用）
-- ---------------------------------------------------------------------------

-- 上游信息：逻辑总池 + 单条上游账号/凭证；容量颗粒度到每一条记录
CREATE TABLE IF NOT EXISTS `upstream_info` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `products_category_name` VARCHAR(32)  NOT NULL COMMENT '与 products 分组键一致：openai|anthropic|xai|gemini|perplexity',
  `sku_product_name`      VARCHAR(128) NOT NULL COMMENT '上游采购档位，如 GPT PRO 5X',
  `account_label`         VARCHAR(128) NOT NULL COMMENT '上游账号/条目运营标签',
  `api_key_ciphertext`    VARBINARY(1024) NOT NULL COMMENT '官方 API Key 密文',
  `api_secret_ciphertext` VARBINARY(1024) DEFAULT NULL COMMENT '部分厂商 Secret 密文',
  `procurement_cost_note` VARCHAR(256) DEFAULT NULL COMMENT '采购成本备注',
  `expires_at`            DATETIME     DEFAULT NULL COMMENT '凭证到期日',
  -- 容量快照：remain=cap_tokens-used_tokens；used>=cap 时停选路由（status=over_cap）
  `cap_tokens`            BIGINT       DEFAULT NULL COMMENT 'token 配额上限，NULL=不限或未录入',
  `used_tokens`           BIGINT       NOT NULL DEFAULT 0 COMMENT '当前周期已消耗 token',
  `quota_reset_at`        DATETIME     DEFAULT NULL COMMENT '下次清零 used_tokens（自然月/采购周期由业务写入）',
  `weight`                INT          NOT NULL DEFAULT 100 COMMENT '同 products_category_name+sku_product_name 组内路由权重',
  `status`                VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active=参与路由；disabled|unhealthy|over_cap=不参与',
  `fail_rate`             DECIMAL(5,4) DEFAULT NULL COMMENT '近期失败率',
  `created_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_upstream_info_product` (`products_category_name`, `sku_product_name`),
  KEY `idx_upstream_info_name_status` (`products_category_name`, `status`),
  KEY `idx_upstream_info_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上游信息（厂商+档位 + 凭证 + 单条容量）';

CREATE TABLE IF NOT EXISTS `products_category` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(64)  NOT NULL COMMENT '展示名，如 ChatGPT',
  `dot_color`  VARCHAR(16)  DEFAULT NULL COMMENT '胶囊圆点色',
  `active_bg`  VARCHAR(16)  DEFAULT NULL COMMENT '选中胶囊背景色',
  `sort`       INT          NOT NULL DEFAULT 0 COMMENT '分类排序（越小越靠前）',
  `status`     VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `hot_tag_name` VARCHAR(32) DEFAULT NULL COMMENT '顶栏分类标签，空则不显示',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_products_category_name` (`name`),
  KEY `idx_products_category_sort` (`status`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商城前台分类（全部套餐胶囊）';

CREATE TABLE IF NOT EXISTS `products` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `sku_code`          VARCHAR(64)  NOT NULL COMMENT 'SKU 编码',
  `card_title`        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '卡片标题',
  `card_subtitle`     VARCHAR(512) NOT NULL DEFAULT '' COMMENT '卡片副标题',
  `card_features`     JSON         NOT NULL COMMENT '卡片卖点条目；可用 {limit_tokens}、{rpm_limit} 占位',
  `share_seats`       INT          NOT NULL DEFAULT 1 COMMENT '子 Key 可共用人数（展示）',
  `products_category_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '分类 ID',
  `products_category_name` VARCHAR(32) NOT NULL COMMENT '上游/导航分组键 openai|cursor 等',
  `sku_product_name`  VARCHAR(128) NOT NULL COMMENT 'SKU 绑定上游档位，如 GPT PRO 5X',
  `limit_tokens`      BIGINT       NOT NULL COMMENT '每计费周期 token 额度（售卖给用户）',
  `rpm_limit`         INT          NOT NULL DEFAULT 0 COMMENT '用户 RPM，0=不限',
  `tpm_limit`         INT          DEFAULT NULL COMMENT '用户 TPM（可选）',
  `allowed_models`    JSON         NOT NULL COMMENT '允许 model 列表',
  `billing_period`    VARCHAR(16)  NOT NULL DEFAULT 'month' COMMENT 'month|year|once（展示用）',
  `period_days`       INT          NOT NULL DEFAULT 30 COMMENT '每个计费周期天数；订阅到期/续费/升档周期计算均据此',
  `price`             DECIMAL(16,2) NOT NULL COMMENT '售价（元）',
  `currency`          CHAR(3)      NOT NULL DEFAULT 'CNY',
  `is_hot`            TINYINT(1)   NOT NULL DEFAULT 0,
  `hot_tag_name`      VARCHAR(32)  DEFAULT NULL COMMENT '卡片右上角标签，空则不显示',
  `sort_order`        INT          NOT NULL DEFAULT 0,
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'on_sale' COMMENT 'on_sale|off_sale',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_products_sku` (`sku_code`),
  KEY `idx_products_category_id` (`products_category_id`),
  KEY `idx_products_category_name` (`products_category_name`),
  KEY `idx_products_category_product` (`products_category_name`, `sku_product_name`),
  KEY `idx_products_status_sort` (`status`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商城 SKU（分类：products_category）';

-- ---------------------------------------------------------------------------
-- 支付通道配置（平台级，非用户支付流水）
-- ---------------------------------------------------------------------------

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

-- ---------------------------------------------------------------------------
-- 交易与订单
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `user_orders` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_no`        VARCHAR(64)  NOT NULL COMMENT '业务订单号',
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `product_id`      BIGINT UNSIGNED NOT NULL COMMENT '购买的 SKU（MVP 一单一件商品）',
  `order_type`      VARCHAR(32)  NOT NULL DEFAULT 'purchase' COMMENT 'purchase=新购 renewal=续费 upgrade=升档 quota_addon=加购额度',
  `user_subscription_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '续费/升档关联的原订阅；新购为 0',
  `quantity`        INT          NOT NULL DEFAULT 1 COMMENT '购买数量',
  `unit_price`      DECIMAL(16,2) NOT NULL COMMENT '下单时单价快照（元）',
  `status`          VARCHAR(32)  NOT NULL COMMENT 'pending_payment|paid|fulfilling|completed|failed|cancelled|refunding',
  `total_amount`    DECIMAL(16,2) NOT NULL COMMENT '应付总额（元），一般=unit_price*quantity',
  `currency`        CHAR(3)      NOT NULL DEFAULT 'CNY',
  `enterprise_invoice` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否企业开票',
  `pay_channel`     VARCHAR(32)  DEFAULT NULL COMMENT 'alipay|wechat|stripe|paypal',
  `out_trade_no`    VARCHAR(64)  DEFAULT NULL COMMENT '平台支付单号',
  `third_trade_no`  VARCHAR(128) DEFAULT NULL COMMENT '渠道流水号',
  `raw_request_json`  JSON       DEFAULT NULL,
  `raw_notify_json`   JSON       DEFAULT NULL,
  `paid_at`         DATETIME     DEFAULT NULL,
  `expire_at`       DATETIME     NOT NULL COMMENT '待支付关单时间',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户订单（MVP：一单对应一个 product）';

CREATE TABLE IF NOT EXISTS `payment_callbacks` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `channel`         VARCHAR(32)  NOT NULL,
  `idempotency_key` VARCHAR(128) NOT NULL COMMENT '幂等键',
  `payload_json`    JSON         NOT NULL,
  `signature_ok`    TINYINT(1)   NOT NULL DEFAULT 0,
  `process_result`  VARCHAR(32)  NOT NULL COMMENT 'success|ignored|failed',
  `processed_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_callbacks_idem` (`channel`, `idempotency_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付回调幂等与对账';

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

CREATE TABLE IF NOT EXISTS `user_refunds` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id`        BIGINT UNSIGNED NOT NULL,
  `refund_no`       VARCHAR(64)  NOT NULL,
  `amount`          DECIMAL(16,2) NOT NULL,
  `reason`          VARCHAR(512) DEFAULT NULL,
  `status`          VARCHAR(32)  NOT NULL COMMENT 'pending|success|failed',
  `refunded_at`     DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_refunds_refund_no` (`refund_no`),
  KEY `idx_user_refunds_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户退款';

CREATE TABLE IF NOT EXISTS `user_invoices` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id`              BIGINT UNSIGNED NOT NULL,
  `enterprise_inquiry_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'enterprise_inquiry.id，0 表示无',
  `user_id`               BIGINT UNSIGNED NOT NULL,
  `invoice_type`    VARCHAR(32)  NOT NULL COMMENT 'personal|electronic|enterprise_vat',
  `title`           VARCHAR(256) NOT NULL,
  `tax_no`          VARCHAR(64)  DEFAULT NULL,
  `amount`          DECIMAL(16,2) NOT NULL,
  `status`          VARCHAR(32)  NOT NULL COMMENT 'pending|issued|failed',
  `file_url`        VARCHAR(512) DEFAULT NULL,
  `issued_at`       DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_invoices_user` (`user_id`),
  KEY `idx_user_invoices_order` (`order_id`),
  KEY `idx_user_invoices_enterprise_inquiry` (`enterprise_inquiry_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户发票';

-- ---------------------------------------------------------------------------
-- 履约：用户订阅 + API Key
-- 履约约定：started_at/expires_at=整段服务有效期（续费延长 expires_at）；
-- period_start/period_end=当前 token 计费周期（滚动时 used=0, limit=base）；新开 base=limit=products.limit_tokens；
-- 永久升档改 base_limit_tokens；本周期加购 limit_tokens+=N，user_orders 追加订单 id
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `user_subscriptions` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `product_id`        BIGINT UNSIGNED NOT NULL,
  `orders`            JSON         NOT NULL DEFAULT (JSON_ARRAY()) COMMENT '关联 user_orders.id 数组（同一 product 的开通、续费等）',
  `products_category_name` VARCHAR(32)  NOT NULL COMMENT '开通时 SKU 分组键快照',
  `sku_product_name`  VARCHAR(128) NOT NULL COMMENT '开通时 SKU 档位名快照',
  `base_limit_tokens` BIGINT       NOT NULL COMMENT '套餐基准上限；续费升档改此值；周期重置时 limit_tokens 回到此值',
  `limit_tokens`      BIGINT       NOT NULL COMMENT '当前周期有效上限（=base+本周期临时加购，可>base）',
  `used_tokens`       BIGINT       NOT NULL DEFAULT 0 COMMENT '当前周期已用；剩余=limit_tokens-used_tokens',
  `started_at`        DATETIME     NOT NULL COMMENT '服务整体开始时间（首开）',
  `expires_at`        DATETIME     NOT NULL COMMENT '服务整体到期（续费延长；网关校验是否仍可调用）',
  `period_start`      DATETIME     NOT NULL COMMENT '当前 token 计费周期起（如自然月）',
  `period_end`        DATETIME     NOT NULL COMMENT '当前 token 计费周期止（到期滚动，与 expires_at 可不同）',
  `status`            VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active|expired|suspended|cancelled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_subscriptions_user` (`user_id`),
  KEY `idx_user_subscriptions_status_expires` (`status`, `expires_at`),
  KEY `idx_user_subscriptions_period_end` (`period_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户订阅（额度+周期+新开/续费/升档/加购）';

CREATE TABLE IF NOT EXISTS `user_api_keys` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `owner_user_id`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '企业主账号 users.id；个人订阅与 user_id 相同',
  `enterprise_inquiry_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'enterprise_inquiry.id，0 表示未关联',
  `user_subscription_id` BIGINT UNSIGNED NOT NULL COMMENT '归属的用户订阅；同订阅下可有多条 Key',
  `key_type`        VARCHAR(16)  NOT NULL DEFAULT 'main' COMMENT 'main=主Key，sub=子Key',
  `key_hash`        CHAR(64)     NOT NULL COMMENT '平台 Key 的 SHA-256（明文仅创建时展示一次，不入库）',
  `products_category_name` VARCHAR(32)  NOT NULL COMMENT '网关 Path 隔离，如 openai',
  `limit_tokens`      BIGINT       NOT NULL COMMENT '该 Key 本周期 token 上限（多条 Key 分配之和不超过 user_subscriptions.limit_tokens）',
  `used_tokens`       BIGINT       NOT NULL DEFAULT 0 COMMENT '该 Key 本周期已用；剩余=limit_tokens-used_tokens',
  `status`          VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled|rotated',
  `rotated_at`      DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_api_keys_hash` (`key_hash`),
  KEY `idx_user_api_keys_subscription` (`user_subscription_id`),
  KEY `idx_user_api_keys_user` (`user_id`),
  KEY `idx_user_api_keys_owner` (`owner_user_id`),
  KEY `idx_user_api_keys_enterprise_inquiry` (`enterprise_inquiry_id`),
  KEY `idx_user_api_keys_sub_key_type` (`user_subscription_id`, `key_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户 API Key（同 user_subscriptions 下多条平级）';

CREATE TABLE IF NOT EXISTS `user_notifications` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`         BIGINT UNSIGNED NOT NULL,
  `user_subscription_id` BIGINT UNSIGNED DEFAULT NULL,
  `api_key_id`      BIGINT UNSIGNED DEFAULT NULL,
  `channel`         VARCHAR(32)  NOT NULL COMMENT 'email|in_app|sms',
  `template_code`   VARCHAR(64)  NOT NULL COMMENT 'subscription_activated|subscription_renewed|subscription_upgraded|subscription_quota_added|key_issued|renew_reminder 等',
  `status`          VARCHAR(32)  NOT NULL COMMENT 'email/sms: pending|sent|failed；in_app 收件: sent=未读 read=已读',
  `sent_at`         DATETIME     DEFAULT NULL,
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_notifications_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户通知（站内/邮件等，开通 Key、续费提醒）';

-- ---------------------------------------------------------------------------
-- 用户 API 网关访问日志
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `user_access_logs` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `request_id`      VARCHAR(64)  NOT NULL,
  `user_id`         BIGINT UNSIGNED DEFAULT NULL,
  `user_subscription_id` BIGINT UNSIGNED DEFAULT NULL,
  `api_key_id`      BIGINT UNSIGNED DEFAULT NULL,
  `products_category_name` VARCHAR(32)  NOT NULL,
  `sku_product_name` VARCHAR(128) DEFAULT NULL COMMENT '实际路由档位（冗余便于检索）',
  `upstream_info_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '实际选用的上游条目',
  `model`           VARCHAR(128) DEFAULT NULL,
  `http_method`     VARCHAR(16)  NOT NULL,
  `path`            VARCHAR(512) NOT NULL,
  `client_ip`       VARCHAR(64)  DEFAULT NULL,
  `gateway_status`  SMALLINT     NOT NULL COMMENT '网关响应码',
  `upstream_status` SMALLINT     DEFAULT NULL,
  `latency_ms`      INT          NOT NULL DEFAULT 0,
  `tokens_prompt`   INT          DEFAULT NULL,
  `tokens_completion` INT        DEFAULT NULL,
  `tokens_total`    INT          DEFAULT NULL,
  `is_stream`       TINYINT(1)   NOT NULL DEFAULT 0,
  `error_code`      VARCHAR(32)  DEFAULT NULL COMMENT '401|402|429|502 等',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_access_logs_request_id` (`request_id`),
  KEY `idx_user_access_logs_user_time` (`user_id`, `created_at`),
  KEY `idx_user_access_logs_key_time` (`api_key_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户 API 网关访问日志（默认不存完整 prompt）';

-- ---------------------------------------------------------------------------
-- 企业采购与成员
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `enterprise_inquiry` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `owner_user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '公司管理人员 users.id',
  `company_name`  VARCHAR(256) NOT NULL COMMENT '公司名称',
  `contact_name`  VARCHAR(128) NOT NULL COMMENT '联系人',
  `phone`         VARCHAR(32)  NOT NULL COMMENT '手机',
  `email`         VARCHAR(255) DEFAULT NULL COMMENT '邮箱',
  `status`        VARCHAR(32)  NOT NULL DEFAULT 'pending' COMMENT 'pending|contacted|closed',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_enterprise_inquiry_owner_user` (`owner_user_id`),
  KEY `idx_enterprise_inquiry_status` (`status`),
  KEY `idx_enterprise_inquiry_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业采购需求';

CREATE TABLE IF NOT EXISTS `enterprise_users` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `owner_user_id`         BIGINT UNSIGNED NOT NULL COMMENT '企业主账号 users.id',
  `enterprise_inquiry_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'enterprise_inquiry.id，0 表示未关联',
  `user_id`               BIGINT UNSIGNED DEFAULT NULL COMMENT '成员商城账号 users.id',
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

-- ---------------------------------------------------------------------------
-- 脚本调度配置
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `bot_schedule_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `module` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `interval_seconds` float NOT NULL DEFAULT '60',
  `exe_sort` int DEFAULT NULL,
  `concurrency` int DEFAULT NULL,
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_enabled` tinyint(1) NOT NULL DEFAULT '1',
  `is_strategy_enabled` tinyint(1) DEFAULT '1',
  `last_live_time` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `last_live_time_by_machine` json DEFAULT NULL,
  `is_primary_machine_run` int NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_module_task` (`module`,`task_name`),
  KEY `idx_module` (`module`),
  KEY `idx_bot_schedule_module_enabled` (`module`,`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='脚本调度配置';

CREATE TABLE IF NOT EXISTS `tutorial_category` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`       VARCHAR(32)  NOT NULL COMMENT '筛选键，如 payment、pro',
  `name`       VARCHAR(64)  NOT NULL COMMENT '展示名',
  `sort`       INT          NOT NULL DEFAULT 0,
  `status`     VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tutorial_category_code` (`code`),
  KEY `idx_tutorial_category_sort` (`status`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教程分类';

CREATE TABLE IF NOT EXISTS `tutorial_article` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `category_id`  BIGINT UNSIGNED NOT NULL COMMENT 'tutorial_category.id',
  `slug`         VARCHAR(128) NOT NULL COMMENT 'URL 唯一标识',
  `title`        VARCHAR(256) NOT NULL,
  `excerpt`      VARCHAR(512) NOT NULL DEFAULT '',
  `body`         JSON         NOT NULL COMMENT '正文段落数组',
  `published_at` DATE         NOT NULL COMMENT '展示日期',
  `sort`         INT          NOT NULL DEFAULT 0,
  `status`       VARCHAR(16)  NOT NULL DEFAULT 'published' COMMENT 'draft|published',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tutorial_article_slug` (`slug`),
  KEY `idx_tutorial_article_category` (`category_id`),
  KEY `idx_tutorial_article_list` (`status`, `published_at`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教程文章';

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

CREATE TABLE IF NOT EXISTS `user_track_events` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `event_type`    VARCHAR(16)  NOT NULL COMMENT 'page_view|click',
  `action`        VARCHAR(16)  NOT NULL COMMENT 'enter|click',
  `user_id`       BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `visitor_id`    VARCHAR(64)  NOT NULL DEFAULT '',
  `session_id`    VARCHAR(64)  NOT NULL DEFAULT '',
  `channel`       VARCHAR(16)  NOT NULL DEFAULT '',
  `app_version`   VARCHAR(32)  NOT NULL DEFAULT '',
  `page_id`       VARCHAR(64)  NOT NULL DEFAULT '',
  `page_path`     VARCHAR(256) NOT NULL DEFAULT '',
  `page_title`    VARCHAR(128) NOT NULL DEFAULT '',
  `element_id`    VARCHAR(128) NOT NULL DEFAULT '',
  `element_name`  VARCHAR(128) NOT NULL DEFAULT '',
  `target_url`    VARCHAR(512) NOT NULL DEFAULT '',
  `api_method`    VARCHAR(16)  NOT NULL DEFAULT '',
  `api_path`      VARCHAR(256) NOT NULL DEFAULT '',
  `api_params`    JSON DEFAULT NULL,
  `locale`        VARCHAR(16)  NOT NULL DEFAULT 'zh-Hans',
  `ip`            VARCHAR(64)  NOT NULL DEFAULT '',
  `user_agent`    VARCHAR(512) NOT NULL DEFAULT '',
  `device_type`   VARCHAR(32)  NOT NULL DEFAULT '',
  `device_model`  VARCHAR(128) NOT NULL DEFAULT '',
  `os_name`       VARCHAR(32)  NOT NULL DEFAULT '',
  `os_version`    VARCHAR(32)  NOT NULL DEFAULT '',
  `screen_width`  INT NOT NULL DEFAULT 0,
  `screen_height` INT NOT NULL DEFAULT 0,
  `referrer`      VARCHAR(512) NOT NULL DEFAULT '',
  `extra_json`    JSON DEFAULT NULL,
  `event_at`      DATETIME NOT NULL,
  `created_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_track_events_user_time` (`user_id`, `event_at`),
  KEY `idx_user_track_events_page_time` (`page_id`, `event_at`),
  KEY `idx_user_track_events_visitor_time` (`visitor_id`, `event_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户端行为埋点';

SET FOREIGN_KEY_CHECKS = 1;
