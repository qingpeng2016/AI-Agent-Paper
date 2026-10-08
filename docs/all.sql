-- AI Agent Paper 全量 schema（唯一 migration；新库执行本文件即可）
-- 含：平台账户/返利/优惠券/Bot/埋点（与 domain/persistent/entity 对齐）+ Paper 工作流（paper_*）
-- 引擎：MySQL 8.0+，utf8mb4；paper_* 与 users.id 逻辑关联，不设 DB 外键
-- 已移除旧商城表：products、user_orders、user_subscriptions、user_api_keys、user_notifications 等
--
-- 表名规范：paper_{类型}[_{子实体}]
--   · 域前缀 paper_ = Paper Agent 全家桶
--   · {类型}       = 业务归类（见下表「类型」列），同类一眼可扫
--   · {子实体}     = 归属某类型的从表
--
-- | 类型       | 表名 |
-- |------------|------|
-- | ref        | paper_ref_* |
-- | manuscript | paper_manuscript（我的论文）, paper_manuscript_runtime, |
-- |            | paper_manuscript_citation_gate, paper_manuscript_literature_hit, paper_manuscript_review |
-- | output     | paper_output_*（各模块业务产出，均含 manuscript_id） |
-- | user       | paper_user_preference, paper_user_literature |
-- | model      | paper_llm_model_config, paper_llm_workflow_binding, paper_llm_prompt_template |
-- | llm        | paper_llm_call_logs |
-- | audit      | paper_operation_log |
-- 业务表均含 manuscript_id → paper_manuscript.id（逻辑外键，不设 DB FK）

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- 0. 平台：用户、返利、优惠券、Bot 调度、行为埋点（Go entity）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `users` (
  `id`                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户 ID',
  `email`                VARCHAR(255) DEFAULT NULL COMMENT '邮箱（登录）',
  `phone`                VARCHAR(32)  DEFAULT NULL COMMENT '手机号（登录）',
  `password_hash`        VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `password_plain`       VARCHAR(255) NOT NULL COMMENT '密码明文（仅限受控环境）',
  `nickname`             VARCHAR(64)  DEFAULT NULL COMMENT '昵称',
  `parent_user_id`       BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上级用户 ID，0 无上级',
  `vip_config_id`        BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'VIP 档位 vip_config.id',
  `vip_domain`           VARCHAR(255) DEFAULT NULL COMMENT '专属推广独立域名',
  `status`               VARCHAR(32)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled|banned',
  `wallet_balance`       DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '钱包可用余额（元）',
  `commission_balance`   DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '佣金余额（元）',
  `last_login_at`        DATETIME     DEFAULT NULL,
  `created_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_email` (`email`),
  UNIQUE KEY `uk_users_phone` (`phone`),
  UNIQUE KEY `uk_users_vip_domain` (`vip_domain`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_parent_user_id` (`parent_user_id`),
  KEY `idx_users_vip_config_id` (`vip_config_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='注册用户';

CREATE TABLE IF NOT EXISTS `vip_config` (
  `id`                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `level_label`             VARCHAR(64)  NOT NULL COMMENT '等级名称',
  `min_valid_invites`       INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '有效邀请人数下限（≥）',
  `min_invitee_paid_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '直属下级累计消费（元，≥）；当前无订单链路时常为 0',
  `rate_percent`            DECIMAL(5,2) NOT NULL COMMENT '返佣比例（%）',
  `sort_order`              INT          NOT NULL DEFAULT 0 COMMENT '门槛排序，越大越高',
  `enabled`                 TINYINT(1)   NOT NULL DEFAULT 1,
  `is_default`              TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '新用户默认档位，仅一条应为 1',
  `created_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_vip_config_enabled_sort` (`enabled`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请返利 VIP 档位';

CREATE TABLE IF NOT EXISTS `vip_domain_config` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `domain`      VARCHAR(255) NOT NULL COMMENT '完整域名',
  `is_official` TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否官网域名',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_vip_domain_config_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='可选推广域名池';

CREATE TABLE IF NOT EXISTS `user_wallet_flows` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`       BIGINT UNSIGNED NOT NULL,
  `type`          VARCHAR(32)  NOT NULL COMMENT 'recharge|pay|refund|commission|withdraw',
  `amount`        DECIMAL(16,2) NOT NULL COMMENT '正入负出（元）',
  `balance_after` DECIMAL(16,2) DEFAULT NULL COMMENT '变动后余额（元）',
  `currency`      CHAR(3)      NOT NULL DEFAULT 'CNY',
  `ref_type`      VARCHAR(32)  DEFAULT NULL,
  `ref_id`        BIGINT UNSIGNED DEFAULT NULL,
  `remark`        VARCHAR(512) DEFAULT NULL,
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_wallet_flows_ref` (`type`, `ref_type`, `ref_id`),
  KEY `idx_user_wallet_flows_user_time` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户资金流水';

CREATE TABLE IF NOT EXISTS `user_commission_payout_config` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL,
  `channel`    VARCHAR(16)  NOT NULL COMMENT 'alipay|wechat',
  `qr_mime`    VARCHAR(64)  DEFAULT NULL COMMENT 'image/png|image/jpeg|image/webp',
  `qr_image`   MEDIUMBLOB   DEFAULT NULL COMMENT '收款码图片',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_commission_payout_user_channel` (`user_id`, `channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='佣金提现收款配置';

CREATE TABLE IF NOT EXISTS `user_commission_records` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `inviter_user_id` BIGINT UNSIGNED NOT NULL,
  `invitee_user_id` BIGINT UNSIGNED NOT NULL,
  `order_id`        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '历史订单 id；无订单链路时为 0',
  `order_no`        VARCHAR(64)  NOT NULL DEFAULT '',
  `product_name`    VARCHAR(128) NOT NULL DEFAULT '',
  `order_amount`    DECIMAL(16,2) NOT NULL DEFAULT 0.00,
  `rate_percent`    DECIMAL(5,2) NOT NULL,
  `rebate_amount`   DECIMAL(16,2) NOT NULL,
  `status`          VARCHAR(32)  NOT NULL DEFAULT 'settled' COMMENT 'settled|reversed',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_commission_records_order` (`order_id`),
  KEY `idx_user_commission_records_inviter_time` (`inviter_user_id`, `created_at`),
  KEY `idx_user_commission_records_invitee` (`invitee_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请返利明细';

CREATE TABLE IF NOT EXISTS `user_commission_withdrawals` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`       BIGINT UNSIGNED NOT NULL,
  `amount`        DECIMAL(16,2) NOT NULL,
  `channel`       VARCHAR(16)  NOT NULL COMMENT 'alipay|wechat',
  `payout_qr_url` VARCHAR(512) NOT NULL DEFAULT '',
  `status`        VARCHAR(32)  NOT NULL DEFAULT 'pending' COMMENT 'pending|completed|failed',
  `fail_reason`   VARCHAR(512) DEFAULT NULL,
  `processed_at`  DATETIME     DEFAULT NULL,
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_commission_withdrawals_user_time` (`user_id`, `created_at`),
  KEY `idx_user_commission_withdrawals_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='佣金提现申请';

CREATE TABLE IF NOT EXISTS `coupon_campaigns` (
  `id`                     BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`                   VARCHAR(32)  NOT NULL,
  `name`                   VARCHAR(128) NOT NULL,
  `title`                  VARCHAR(128) NOT NULL,
  `subtitle`               VARCHAR(512) DEFAULT NULL,
  `discount_type`          VARCHAR(16)  NOT NULL COMMENT 'fixed_amount|percent',
  `discount_value`         DECIMAL(16,2) NOT NULL,
  `min_order_amount`       DECIMAL(16,2) NOT NULL DEFAULT 0.00,
  `valid_days`             INT UNSIGNED NOT NULL DEFAULT 30,
  `auto_grant_on_register` TINYINT(1) NOT NULL DEFAULT 0,
  `enabled`                TINYINT(1) NOT NULL DEFAULT 1,
  `created_at`             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_coupon_campaigns_code` (`code`),
  KEY `idx_coupon_campaigns_register` (`auto_grant_on_register`, `enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券活动';

CREATE TABLE IF NOT EXISTS `user_coupons` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`          BIGINT UNSIGNED NOT NULL,
  `campaign_id`      BIGINT UNSIGNED NOT NULL,
  `coupon_code`      VARCHAR(40)  NOT NULL,
  `discount_type`    VARCHAR(16)  NOT NULL,
  `discount_value`   DECIMAL(16,2) NOT NULL,
  `min_order_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00,
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

CREATE TABLE IF NOT EXISTS `bot_schedule_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `module` varchar(50) NOT NULL,
  `task_name` varchar(100) NOT NULL,
  `interval_seconds` float NOT NULL DEFAULT 60,
  `exe_sort` int DEFAULT NULL,
  `concurrency` int DEFAULT NULL,
  `description` varchar(500) DEFAULT NULL,
  `is_enabled` tinyint(1) NOT NULL DEFAULT 1,
  `is_strategy_enabled` tinyint(1) DEFAULT 1,
  `last_live_time` varchar(255) DEFAULT NULL,
  `last_live_time_by_machine` json DEFAULT NULL,
  `is_primary_machine_run` int NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_module_task` (`module`,`task_name`),
  KEY `idx_module` (`module`),
  KEY `idx_bot_schedule_module_enabled` (`module`,`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Bot 脚本调度';

-- ---------------------------------------------------------------------------
-- 1. ref 字典：学科、venue、文献源、强度与审计档位（七模块 code 见前端 front/web-pc/types.ts）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_ref_discipline` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`        VARCHAR(64)  NOT NULL COMMENT '学科键，如 cs_ai|medicine|law|economics',
  `name`        VARCHAR(128) NOT NULL COMMENT '展示名',
  `name_en`     VARCHAR(128) DEFAULT NULL,
  `sort`        INT          NOT NULL DEFAULT 0,
  `literature_source_codes` JSON         NOT NULL DEFAULT (JSON_ARRAY()) COMMENT '该学科推荐文献源 paper_ref_literature_source.code 列表，顺序即 UI 默认',
  `status`      VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_discipline_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学科';

CREATE TABLE IF NOT EXISTS `paper_ref_venue` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`            VARCHAR(64)  NOT NULL COMMENT 'venue 键，如 neurips|iclr|nature|ssci_q1',
  `name`            VARCHAR(256) NOT NULL COMMENT 'NeurIPS 2026 / Nature 等',
  `venue_type`      VARCHAR(16)  NOT NULL COMMENT 'conference|journal|workshop|other',
  `discipline_id`   BIGINT UNSIGNED DEFAULT NULL COMMENT 'NULL=跨学科通用',
  `publisher`       VARCHAR(128) DEFAULT NULL,
  `website_url`     VARCHAR(512) DEFAULT NULL,
  `rubric`          JSON         NOT NULL COMMENT 'venue 规则：contribution_types|experiment_bar|writing_style 等 prompt/结构化约束',
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_venue_code` (`code`),
  KEY `idx_paper_ref_venue_discipline` (`discipline_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='目标会议/期刊';

CREATE TABLE IF NOT EXISTS `paper_ref_literature_source` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`            VARCHAR(64)  NOT NULL COMMENT 'arxiv|openalex|semantic_scholar|pubmed|crossref|cnki|…',
  `name`            VARCHAR(128) NOT NULL,
  `api_kind`        VARCHAR(32)  NOT NULL COMMENT 'rest|oai|custom',
  `base_url`        VARCHAR(512) NOT NULL,
  `auth_type`       VARCHAR(32)  NOT NULL DEFAULT 'none' COMMENT 'none|api_key|oauth',
  `config_schema`   JSON         DEFAULT NULL COMMENT '连接参数 JSON Schema（Key 名、必填项）',
  `default_config`  JSON         DEFAULT NULL COMMENT '默认非密钥配置',
  `rate_limit_hint` VARCHAR(256) DEFAULT NULL,
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_literature_source_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文献/API 数据源';

CREATE TABLE IF NOT EXISTS `paper_ref_execution_intensity` (
  `code`         VARCHAR(16)  NOT NULL COMMENT 'fast|balanced|deep',
  `name`         VARCHAR(64)  NOT NULL,
  `multiplier`   DECIMAL(4,2) NOT NULL DEFAULT 1.00 COMMENT '相对检索量/迭代轮数系数',
  `max_papers`   INT UNSIGNED NOT NULL DEFAULT 80 COMMENT '检索文献上限（选题/综述等）',
  `max_ideas`    INT UNSIGNED NOT NULL DEFAULT 12 COMMENT '选题候选 idea 上限',
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='执行强度档位';

CREATE TABLE IF NOT EXISTS `paper_ref_audit_level` (
  `code`                   VARCHAR(16)  NOT NULL COMMENT 'standard|polished|strict',
  `name`                   VARCHAR(64)  NOT NULL,
  `citation_strength`      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '0=关,1-3=引用审计强度',
  `claim_strength`         TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '0=关,1-3=论断审计强度',
  `kill_argument_strength` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=关,1-3=驳论审计强度',
  `audit_rounds`           TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '审计迭代轮数',
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计等级';

-- ---------------------------------------------------------------------------
-- 2. 稿件线与用户默认（工作台「当前论文」= paper_manuscript）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_manuscript` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT 'users.id',
  `title`             VARCHAR(256) NOT NULL COMMENT '工作标题 / 暂定篇名',
  `description`       TEXT         DEFAULT NULL,
  `manuscript_kind`   VARCHAR(32)  DEFAULT NULL COMMENT 'conference|journal|thesis|course|report',
  `deadline_at`       DATETIME     DEFAULT NULL,
  `target_words`      INT UNSIGNED DEFAULT NULL,
  `citation_style`    VARCHAR(64)  DEFAULT NULL COMMENT 'apa|ieee|acm|…',
  `discipline_id`     BIGINT UNSIGNED DEFAULT NULL,
  `default_venue_id`  BIGINT UNSIGNED DEFAULT NULL,
  `reference_gate_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '参考文献门禁',
  `workspace_uri`     VARCHAR(512) DEFAULT NULL COMMENT '平台 provision 的工作区根路径',
  `forked_from_manuscript_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '从另一篇论文 fork 时链接',
  `draft_format`      VARCHAR(16)  DEFAULT NULL COMMENT 'md|tex|pdf；论文写作当前稿',
  `draft_content_medium` MEDIUMTEXT DEFAULT NULL COMMENT '当前正文（小稿内联）',
  `draft_storage_uri` VARCHAR(1024) DEFAULT NULL COMMENT '当前正文文件路径',
  `draft_bibtex_uri`  VARCHAR(1024) DEFAULT NULL,
  `draft_word_count`  INT UNSIGNED DEFAULT NULL,
  `draft_updated_at`  DATETIME     DEFAULT NULL COMMENT '最近一次写作/保存',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|archived',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_manuscript_user` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='我的论文（元数据 + 当前论文正文草稿）';

CREATE TABLE IF NOT EXISTS `paper_user_preference` (
  `user_id`                   BIGINT UNSIGNED NOT NULL,
  `default_discipline_id`     BIGINT UNSIGNED DEFAULT NULL,
  `default_venue_id`          BIGINT UNSIGNED DEFAULT NULL,
  `default_intensity_code`    VARCHAR(16)  NOT NULL DEFAULT 'balanced',
  `default_audit_level_code`  VARCHAR(16)  NOT NULL DEFAULT 'polished',
  `default_human_checkpoint`  TINYINT(1)   NOT NULL DEFAULT 1,
  `default_literature_source_ids` JSON     DEFAULT NULL COMMENT '默认文献源 id 数组',
  `updated_at`                DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户全局默认（个人中心·默认配置）';

-- ---------------------------------------------------------------------------
-- 3. 平台运行时（模型、算力、文献源凭证 — 不对用户暴露）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_llm_model_config` (
  `id`                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `label`                VARCHAR(128) NOT NULL COMMENT '展示名',
  `provider_code`        VARCHAR(32)  NOT NULL COMMENT 'openai|anthropic|azure_openai|openai_compatible|ollama|gateway|other',
  `model_name`           VARCHAR(128) NOT NULL COMMENT '上游 model 参数',
  `api_base_url`         VARCHAR(512) NOT NULL COMMENT 'API Host / Base URL（具体 path 由 provider_code 在代码中决定）',
  `api_key`              VARCHAR(512) DEFAULT NULL COMMENT 'LLM API Key 明文（库内存储；生产建议库权限+TLS）',
  `timeout_ms`           INT UNSIGNED NOT NULL DEFAULT 120000,
  `max_retries`          TINYINT UNSIGNED NOT NULL DEFAULT 2,
  `supports_vision`      TINYINT(1)   NOT NULL DEFAULT 0,
  `context_window_hint`  INT UNSIGNED DEFAULT NULL,
  `extra`                JSON         DEFAULT NULL,
  `status`               VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_model_cfg_status` (`status`),
  KEY `idx_paper_llm_model_cfg_provider` (`provider_code`, `model_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台 LLM 端点配置（Key/Host/模型名）';

CREATE TABLE IF NOT EXISTS `paper_llm_workflow_binding` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `stage_code`        VARCHAR(64)  NOT NULL DEFAULT 'default' COMMENT '环节键：default|retrieve|…（与 prompt 表同键；由代码约定）',
  `model_config_id`   BIGINT UNSIGNED NOT NULL COMMENT 'paper_llm_model_config.id',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_llm_bind_stage` (`stage_code`),
  KEY `idx_paper_llm_bind_model` (`model_config_id`),
  KEY `idx_paper_llm_bind_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台：stage_code → model_config';

CREATE TABLE IF NOT EXISTS `paper_llm_prompt_template` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `stage_code`        VARCHAR(64)  NOT NULL DEFAULT 'default' COMMENT '环节键：default|retrieve|generate_ideas|…（由代码约定）',
  `template_body`     MEDIUMTEXT   NOT NULL COMMENT '话术正文；占位符 {{var_name}}，由代码渲染',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_llm_prompt_stage` (`stage_code`),
  KEY `idx_paper_llm_prompt_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台话术（按 stage_code）';

CREATE TABLE IF NOT EXISTS `paper_manuscript_runtime` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`         BIGINT UNSIGNED NOT NULL,
  `label`                 VARCHAR(128) NOT NULL DEFAULT 'default',
  `literature_credentials` JSON        DEFAULT NULL COMMENT 'source_id -> 密钥密文/引用 id',
  `literature_source_ids` JSON        NOT NULL COMMENT '本稿件启用的 source id 列表',
  `gpu_profile`           JSON         DEFAULT NULL COMMENT 'SSH/集群/本地',
  `wiki_uri`              VARCHAR(512) DEFAULT NULL COMMENT 'Research Wiki 路径',
  `is_active`             TINYINT(1)   NOT NULL DEFAULT 1,
  `created_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_manuscript_runtime_ms` (`manuscript_id`, `is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='稿件平台运行时绑定（内部）';

-- ---------------------------------------------------------------------------
-- 4. 本篇检索文献与引用门禁
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_manuscript_literature_hit` (
  `manuscript_id`       BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `source_id`           BIGINT UNSIGNED NOT NULL COMMENT 'paper_ref_literature_source.id',
  `external_key`        VARCHAR(256) NOT NULL COMMENT 'arxiv:2401.12345 / doi:… / s2:…',
  `stage_code`          VARCHAR(64)  DEFAULT NULL COMMENT '写入时的环节键，如 retrieve',
  `relevance_score`     DECIMAL(6,4) DEFAULT NULL,
  `query_text`          VARCHAR(512) DEFAULT NULL,
  `meta`                JSON         DEFAULT NULL COMMENT 'title、authors、doi 等快照',
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`manuscript_id`, `source_id`, `external_key`),
  KEY `idx_paper_ms_lit_hit_doi` (`external_key`(64))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='本篇检索命中的文献（选题等）';

CREATE TABLE IF NOT EXISTS `paper_manuscript_citation_gate` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`       BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `stage_code`          VARCHAR(64)  DEFAULT NULL COMMENT '校验发生时的环节键',
  `target_id`           BIGINT UNSIGNED DEFAULT NULL COMMENT '如 paper_manuscript.id 或其它 paper_output_*',
  `source_id`           BIGINT UNSIGNED DEFAULT NULL COMMENT '门禁关联文献源',
  `external_key`        VARCHAR(256) DEFAULT NULL COMMENT '与 literature_hit 同源键',
  `cited_key`           VARCHAR(256) DEFAULT NULL COMMENT '正文中的 cite key',
  `gate_type`           VARCHAR(32)  NOT NULL COMMENT 'reference|citation_audit|claim_audit|kill_argument',
  `status`              VARCHAR(16)  NOT NULL COMMENT 'verified|rejected|pending|waived',
  `detail`              JSON         DEFAULT NULL,
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_ms_citation_gate_ms` (`manuscript_id`, `gate_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='本篇引用文献与审计校验记录（参考文献门禁台账）';

-- ---------------------------------------------------------------------------
-- 5. 各模块业务产出 paper_output_*（均关联 paper_manuscript.id）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_output_topic` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`     BIGINT UNSIGNED NOT NULL,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `version`           INT          NOT NULL DEFAULT 1,
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '本篇当前生效版本',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed',
  `direction`         TEXT         DEFAULT NULL,
  `venue_label`       VARCHAR(256) DEFAULT NULL,
  `literature_hit_count` INT UNSIGNED NOT NULL DEFAULT 0,
  `verified_hit_count`   INT UNSIGNED NOT NULL DEFAULT 0,
  `novelty_report`    MEDIUMTEXT   DEFAULT NULL,
  `audit_summary`     TEXT         DEFAULT NULL,
  `input_params`      JSON         DEFAULT NULL,
  `meta`              JSON         DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_topic_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='选题发现 · 本轮汇总';

CREATE TABLE IF NOT EXISTS `paper_output_topic_idea` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`     BIGINT UNSIGNED NOT NULL,
  `topic_id`          BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_output_topic.id',
  `rank_no`           INT          NOT NULL DEFAULT 0,
  `title`             VARCHAR(512) NOT NULL,
  `problem`           TEXT         DEFAULT NULL,
  `approach`          TEXT         DEFAULT NULL,
  `contribution`      TEXT         DEFAULT NULL,
  `novelty_summary`   TEXT         DEFAULT NULL,
  `novelty_risk`      VARCHAR(16)  DEFAULT NULL COMMENT 'low|medium|high',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'candidate' COMMENT 'candidate|selected|rejected',
  `meta`              JSON         DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_topic_idea_ms` (`manuscript_id`, `rank_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='选题发现 · 结构化 idea';

CREATE TABLE IF NOT EXISTS `paper_output_literature_review` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`     BIGINT UNSIGNED NOT NULL,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `version`           INT          NOT NULL DEFAULT 1,
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 1,
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed' COMMENT 'draft|completed|superseded',
  `structure`         VARCHAR(32)  DEFAULT NULL COMMENT 'thematic|chronological|method',
  `title`             VARCHAR(256) DEFAULT NULL,
  `summary`           TEXT         DEFAULT NULL,
  `content_medium`    MEDIUMTEXT   DEFAULT NULL,
  `storage_uri`       VARCHAR(1024) DEFAULT NULL,
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md',
  `citations`         JSON         DEFAULT NULL COMMENT '引用表 / cite key 列表',
  `input_params`      JSON         DEFAULT NULL,
  `meta`              JSON         DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_lit_review_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文献综述';

CREATE TABLE IF NOT EXISTS `paper_output_experiment_plan` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`     BIGINT UNSIGNED NOT NULL,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `version`           INT          NOT NULL DEFAULT 1,
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 1,
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed',
  `title`             VARCHAR(256) DEFAULT NULL,
  `summary`           TEXT         DEFAULT NULL,
  `content_medium`    MEDIUMTEXT   DEFAULT NULL,
  `storage_uri`       VARCHAR(1024) DEFAULT NULL,
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md',
  `input_params`      JSON         DEFAULT NULL,
  `meta`              JSON         DEFAULT NULL COMMENT 'baseline、ablation、资源估算等',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_exp_plan_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实验规划';

CREATE TABLE IF NOT EXISTS `paper_output_experiment_review` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`     BIGINT UNSIGNED NOT NULL,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `version`           INT          NOT NULL DEFAULT 1,
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 1,
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed',
  `title`             VARCHAR(256) DEFAULT NULL,
  `summary`           TEXT         DEFAULT NULL,
  `content_medium`    MEDIUMTEXT   DEFAULT NULL COMMENT '结果审查报告正文',
  `storage_uri`       VARCHAR(1024) DEFAULT NULL,
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md',
  `findings`          JSON         DEFAULT NULL COMMENT '结构化问题与建议',
  `input_params`      JSON         DEFAULT NULL,
  `meta`              JSON         DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_exp_review_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实验审查（结果审查 auto_review）';

CREATE TABLE IF NOT EXISTS `paper_manuscript_review` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`     BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `version`           INT          NOT NULL DEFAULT 1,
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 1,
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed',
  `title`             VARCHAR(256) DEFAULT NULL,
  `summary`           TEXT         DEFAULT NULL,
  `content_medium`    MEDIUMTEXT   DEFAULT NULL COMMENT '论文审查报告',
  `storage_uri`       VARCHAR(1024) DEFAULT NULL,
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md',
  `findings`          JSON         DEFAULT NULL COMMENT '分维度审稿意见',
  `input_params`      JSON         DEFAULT NULL,
  `meta`              JSON         DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_ms_review_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文审查（manuscript_analysis 模块产出）';

CREATE TABLE IF NOT EXISTS `paper_output_figure` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`     BIGINT UNSIGNED NOT NULL,
  `user_id`           BIGINT UNSIGNED NOT NULL,
  `origin`            VARCHAR(16)  NOT NULL COMMENT 'upload|generated',
  `file_kind`         VARCHAR(16)  NOT NULL COMMENT 'image|pdf|vector|other',
  `file_name`         VARCHAR(512) NOT NULL,
  `title`             VARCHAR(512) NOT NULL,
  `note`              TEXT         DEFAULT NULL,
  `size_bytes`        BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `storage_uri`       VARCHAR(1024) DEFAULT NULL,
  `content_hash`      CHAR(64)     DEFAULT NULL,
  `prompt_snapshot`   TEXT         DEFAULT NULL COMMENT '生成图时的提示词摘要',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_figure_ms` (`manuscript_id`, `origin`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表记录（上传 + 一键生成）';

-- ---------------------------------------------------------------------------
-- 7. LLM 调用日志（计费、排错；不存完整 prompt 时可只存 hash）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_llm_call_logs` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `manuscript_id`   BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `stage_code`      VARCHAR(64)  DEFAULT NULL COMMENT '环节键（与 binding/prompt 同 stage_code）',
  `model_config_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_model_config.id',
  `workflow_binding_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_workflow_binding.id',
  `prompt_template_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_prompt_template.id',
  `model_name`      VARCHAR(128) NOT NULL,
  `prompt_tokens`   INT          DEFAULT NULL,
  `completion_tokens` INT        DEFAULT NULL,
  `latency_ms`      INT          NOT NULL DEFAULT 0,
  `status`          SMALLINT     NOT NULL COMMENT 'HTTP 或业务码',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_call_logs_ms` (`manuscript_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='LLM 调用明细日志';

-- ---------------------------------------------------------------------------
-- 8. 上传文献、操作日志（阶段进度由 paper_output_* / draft_* 是否存在推导，不落单独 progress 表）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_user_literature` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`       BIGINT UNSIGNED NOT NULL COMMENT 'users.id',
  `manuscript_id` BIGINT UNSIGNED NOT NULL COMMENT '本篇「上传文献」；选题 user_library 语料',
  `file_kind`     VARCHAR(16)  NOT NULL COMMENT 'pdf|bib|txt|other',
  `file_name`     VARCHAR(512) NOT NULL,
  `title`         VARCHAR(512) NOT NULL,
  `note`          TEXT         DEFAULT NULL,
  `size_bytes`    BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `storage_uri`   VARCHAR(1024) DEFAULT NULL COMMENT '对象存储或工作区相对路径',
  `content_hash`  CHAR(64)     DEFAULT NULL COMMENT 'SHA-256',
  `status`        VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_user_lit_ms` (`manuscript_id`, `status`),
  KEY `idx_paper_user_lit_user` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户上传文献（上传文献模块 / 参考文献门禁语料）';

CREATE TABLE IF NOT EXISTS `paper_operation_log` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`            BIGINT UNSIGNED NOT NULL,
  `manuscript_id`      BIGINT UNSIGNED DEFAULT NULL,
  `manuscript_title`   VARCHAR(256) DEFAULT NULL,
  `stage_code`         VARCHAR(64)  NOT NULL DEFAULT 'default' COMMENT '环节键（展示名见 action / note）',
  `action`             VARCHAR(512) NOT NULL,
  `tokens_prompt`      INT UNSIGNED NOT NULL DEFAULT 0,
  `tokens_completion`  INT UNSIGNED NOT NULL DEFAULT 0,
  `tokens_total`       INT UNSIGNED NOT NULL DEFAULT 0,
  `status`             VARCHAR(16)  NOT NULL COMMENT 'success|failed|cancelled',
  `note`               TEXT         DEFAULT NULL,
  `occurred_at`        DATETIME     NOT NULL,
  `created_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_oplog_user` (`user_id`, `occurred_at`),
  KEY `idx_paper_oplog_ms` (`manuscript_id`, `occurred_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Paper Agent 操作与 Token 消耗日志';

-- ---------------------------------------------------------------------------
-- 9. 种子数据 — 平台（VIP、优惠券、Bot）
-- ---------------------------------------------------------------------------

INSERT INTO `vip_config` (
  `id`, `level_label`, `min_valid_invites`, `min_invitee_paid_amount`, `rate_percent`, `sort_order`, `enabled`, `is_default`
) VALUES
  (1, '入门推广', 0,  0.00,    3.00, 10, 1, 1),
  (2, '标准推广', 3,  0.00,    5.00, 20, 1, 0),
  (3, '高级推广', 5,  500.00,  8.00, 30, 1, 0),
  (4, '合伙人',   20, 2000.00, 12.00, 40, 1, 0)
ON DUPLICATE KEY UPDATE
  `level_label` = VALUES(`level_label`),
  `min_valid_invites` = VALUES(`min_valid_invites`),
  `min_invitee_paid_amount` = VALUES(`min_invitee_paid_amount`),
  `rate_percent` = VALUES(`rate_percent`),
  `sort_order` = VALUES(`sort_order`),
  `enabled` = VALUES(`enabled`),
  `is_default` = VALUES(`is_default`);

INSERT INTO `coupon_campaigns` (
  `code`, `name`, `title`, `subtitle`, `discount_type`, `discount_value`,
  `min_order_amount`, `valid_days`, `auto_grant_on_register`, `enabled`
) VALUES (
  'WELCOME',
  '新用户注册礼',
  '新人专享优惠券',
  '注册即领',
  'fixed_amount',
  20.00,
  0.00,
  30,
  1,
  1
)
ON DUPLICATE KEY UPDATE
  `title` = VALUES(`title`),
  `subtitle` = VALUES(`subtitle`),
  `enabled` = VALUES(`enabled`);

INSERT INTO `bot_schedule_config` (
  `module`, `task_name`, `interval_seconds`, `description`, `is_enabled`, `is_strategy_enabled`
) VALUES
  (
    'ai_agent_paper',
    'vip_level_sync',
    3600,
    '按邀请人数匹配 vip_config，仅升不降',
    1,
    1
  ),
  (
    'ai_agent_paper',
    'coupon_expire',
    3600,
    '过期 available 且 valid_until 已过的 user_coupons',
    1,
    1
  )
ON DUPLICATE KEY UPDATE
  `interval_seconds` = VALUES(`interval_seconds`),
  `description` = VALUES(`description`),
  `is_enabled` = VALUES(`is_enabled`),
  `is_strategy_enabled` = VALUES(`is_strategy_enabled`),
  `updated_at` = CURRENT_TIMESTAMP;

-- ---------------------------------------------------------------------------
-- 10. 种子数据 — Paper（强度、审计、学科、venue、文献源）
-- ---------------------------------------------------------------------------

INSERT INTO `paper_ref_execution_intensity` (`code`, `name`, `multiplier`, `max_papers`, `max_ideas`) VALUES
  ('fast',     '更快', 0.60, 30,  6),
  ('balanced', 'Balanced（平衡）', 1.00, 80,  12),
  ('deep',     '更深', 1.80, 200, 20)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `multiplier` = VALUES(`multiplier`),
  `max_papers` = VALUES(`max_papers`),
  `max_ideas` = VALUES(`max_ideas`);

INSERT INTO `paper_ref_audit_level` (
  `code`, `name`, `citation_strength`, `claim_strength`, `kill_argument_strength`, `audit_rounds`
) VALUES
  ('standard', 'Standard', 1, 1, 0, 1),
  ('polished', 'Polished（精修）', 2, 2, 1, 2),
  ('strict',   'Strict', 3, 3, 2, 3)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `citation_strength` = VALUES(`citation_strength`),
  `claim_strength` = VALUES(`claim_strength`),
  `kill_argument_strength` = VALUES(`kill_argument_strength`),
  `audit_rounds` = VALUES(`audit_rounds`);

INSERT INTO `paper_ref_discipline` (`code`, `name`, `name_en`, `sort`, `literature_source_codes`) VALUES
  ('cs_ai', '计算机/人工智能', 'Computer Science & AI', 10,
   JSON_ARRAY('arxiv', 'openalex', 'semantic_scholar')),
  ('general', '跨学科通用', 'General', 0, JSON_ARRAY('openalex', 'crossref'))
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `literature_source_codes` = VALUES(`literature_source_codes`);

INSERT INTO `paper_ref_venue` (`code`, `name`, `venue_type`, `discipline_id`, `rubric`) 
SELECT 'ml_top3', 'NeurIPS/ICLR/ICML', 'conference', d.id,
  JSON_OBJECT(
    'contribution_types', JSON_ARRAY('method', 'theory', 'empirical'),
    'experiment_bar', 'strong_baselines_ablations',
    'writing_style', 'ml_conference'
  )
FROM `paper_ref_discipline` d WHERE d.code = 'cs_ai' LIMIT 1
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

INSERT INTO `paper_ref_literature_source` (`code`, `name`, `api_kind`, `base_url`, `auth_type`, `default_config`, `config_schema`) VALUES
  ('arxiv', 'arXiv', 'rest', 'https://export.arxiv.org/api/query', 'none', NULL, NULL),
  ('openalex', 'OpenAlex', 'rest', 'https://api.openalex.org/works', 'none',
   JSON_OBJECT('mailto', ''),
   JSON_OBJECT('type', 'object', 'properties', JSON_OBJECT('mailto', JSON_OBJECT('type', 'string', 'description', 'User-Agent 礼貌池')))),
  ('semantic_scholar', 'Semantic Scholar', 'rest', 'https://api.semanticscholar.org/graph/v1/paper/search', 'api_key',
   NULL,
   JSON_OBJECT('type', 'object', 'properties', JSON_OBJECT('api_key', JSON_OBJECT('type', 'string', 'description', 'Header x-api-key；可选'))))
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `base_url` = VALUES(`base_url`),
  `auth_type` = VALUES(`auth_type`),
  `default_config` = VALUES(`default_config`),
  `config_schema` = VALUES(`config_schema`);

-- ---------------------------------------------------------------------------
-- 11. 种子 — LLM（model 首次插入后请 UPDATE api_key；重复跑请跳过 11.1 或自行去重）
-- ---------------------------------------------------------------------------

-- 11.1 Claude 端点（勿重复 INSERT 多条相同 label，除非手动清理）
INSERT INTO `paper_llm_model_config` (
  `label`, `provider_code`, `model_name`, `api_base_url`, `api_key`,
  `timeout_ms`, `max_retries`, `supports_vision`, `context_window_hint`, `status`
) VALUES (
  'Claude Sonnet（平台默认）',
  'anthropic',
  'claude-sonnet-4-20250514',
  'https://api.anthropic.com',
  'PASTE_YOUR_ANTHROPIC_API_KEY_HERE',
  120000,
  2,
  1,
  200000,
  'active'
);

-- 11.2 各环节 → 同一 Claude（@model_id 取自最新 active anthropic 配置）
SET @paper_llm_model_id = (
  SELECT c.`id` FROM `paper_llm_model_config` c
  WHERE c.`provider_code` = 'anthropic' AND c.`status` = 'active'
  ORDER BY c.`id` DESC LIMIT 1
);

INSERT INTO `paper_llm_workflow_binding` (`stage_code`, `model_config_id`, `status`) VALUES
  ('default', @paper_llm_model_id, 'active'),
  ('retrieve', @paper_llm_model_id, 'active'),
  ('generate_ideas', @paper_llm_model_id, 'active'),
  ('novelty', @paper_llm_model_id, 'active'),
  ('audit', @paper_llm_model_id, 'active'),
  ('literature_review', @paper_llm_model_id, 'active'),
  ('experiment_planning', @paper_llm_model_id, 'active'),
  ('auto_review', @paper_llm_model_id, 'active'),
  ('paper_writing', @paper_llm_model_id, 'active'),
  ('figure_generation', @paper_llm_model_id, 'active'),
  ('citation_audit', @paper_llm_model_id, 'active'),
  ('claim_audit', @paper_llm_model_id, 'active'),
  ('kill_argument', @paper_llm_model_id, 'active')
ON DUPLICATE KEY UPDATE
  `model_config_id` = VALUES(`model_config_id`),
  `status` = VALUES(`status`);

-- 11.3 环节话术（与 front/web-pc TOPIC_DISCOVERY_FLOW_STEPS + 各 PaperModule 对齐）
INSERT INTO `paper_llm_prompt_template` (`stage_code`, `template_body`, `status`) VALUES
  ('default',
   'You are an expert academic research assistant for the Paper Agent platform. Be precise, cite evidence when provided, and state uncertainty clearly. Default to clear scholarly English.',
   'active'),
  ('retrieve',
   'Stage: multi-source literature retrieve and validate for direction {{direction}}. Dedupe by external_key; list coverage gaps.',
   'active'),
  ('generate_ideas',
   'Stage: generate_ideas for {{direction}}. Propose up to {{max_ideas}} testable ideas with title, one-line claim, and reference keys.',
   'active'),
  ('novelty',
   'Stage: novelty check for {{direction}}. Compare each idea to ingested literature; flag overlap and suggest differentiation.',
   'active'),
  ('audit',
   'Stage: topic audit. Review claims like a strict reviewer: unsupported claims, missing baselines, vague contributions. Output blocker/major/minor issues.',
   'active'),
  ('literature_review',
   'Stage: literature_review. Synthesize ingested corpus into a structured review for {{direction}}; gap analysis and related-work outline for {{venue}}.',
   'active'),
  ('experiment_planning',
   'Stage: experiment_planning for {{direction}}. Hypotheses, datasets, strong baselines, ablations, metrics, reproducible steps (intensity {{intensity}}).',
   'active'),
  ('auto_review',
   'Stage: auto_review before writing. Simulate a reviewer on plan and evidence: controls, baselines, overclaiming.',
   'active'),
  ('paper_writing',
   'Stage: paper_writing for venue {{venue}}. Conference/journal sections, consistent notation; citation keys from bibliography gate.',
   'active'),
  ('figure_generation',
   'Stage: figure_generation. Publication-ready figure spec: chart type, axes, statistics, caption.',
   'active'),
  ('citation_audit',
   'Stage: citation_audit on full draft. Missing keys, mismatched claims, over-citation, uncited facts; section anchors.',
   'active'),
  ('claim_audit',
   'Stage: claim_audit. Match claims to evidence; separate contributions, limitations, speculation.',
   'active'),
  ('kill_argument',
   'Stage: kill_argument for venue {{venue}}. Strongest rejection reasons; stay constructive.',
   'active')
ON DUPLICATE KEY UPDATE
  `template_body` = VALUES(`template_body`),
  `status` = VALUES(`status`);

SET FOREIGN_KEY_CHECKS = 1;
