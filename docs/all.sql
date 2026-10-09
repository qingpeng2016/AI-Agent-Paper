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
-- |            | paper_manuscript_citation_gate, paper_manuscript_review |
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
  `last_login_at`        DATETIME     DEFAULT NULL COMMENT '最近登录时间',
  `created_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  `updated_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '资料更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_email` (`email`),
  UNIQUE KEY `uk_users_phone` (`phone`),
  UNIQUE KEY `uk_users_vip_domain` (`vip_domain`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_parent_user_id` (`parent_user_id`),
  KEY `idx_users_vip_config_id` (`vip_config_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='注册用户';

CREATE TABLE IF NOT EXISTS `vip_config` (
  `id`                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '档位 ID',
  `level_label`             VARCHAR(64)  NOT NULL COMMENT '等级名称',
  `min_valid_invites`       INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '有效邀请人数下限（≥）',
  `min_invitee_paid_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '直属下级累计消费（元，≥）；当前无订单链路时常为 0',
  `rate_percent`            DECIMAL(5,2) NOT NULL COMMENT '返佣比例（%）',
  `sort_order`              INT          NOT NULL DEFAULT 0 COMMENT '门槛排序，越大越高',
  `enabled`                 TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否启用',
  `is_default`              TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '新用户默认档位，仅一条应为 1',
  `created_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_vip_config_enabled_sort` (`enabled`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请返利 VIP 档位';

CREATE TABLE IF NOT EXISTS `vip_domain_config` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录 ID',
  `domain`      VARCHAR(255) NOT NULL COMMENT '完整域名',
  `is_official` TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否官网域名',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_vip_domain_config_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='可选推广域名池';

CREATE TABLE IF NOT EXISTS `user_wallet_flows` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水 ID',
  `user_id`       BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `type`          VARCHAR(32)  NOT NULL COMMENT 'recharge|pay|refund|commission|withdraw',
  `amount`        DECIMAL(16,2) NOT NULL COMMENT '正入负出（元）',
  `balance_after` DECIMAL(16,2) DEFAULT NULL COMMENT '变动后余额（元）',
  `currency`      CHAR(3)      NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `ref_type`      VARCHAR(32)  DEFAULT NULL COMMENT '关联业务类型',
  `ref_id`        BIGINT UNSIGNED DEFAULT NULL COMMENT '关联业务 ID',
  `remark`        VARCHAR(512) DEFAULT NULL COMMENT '备注',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发生时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_wallet_flows_ref` (`type`, `ref_type`, `ref_id`),
  KEY `idx_user_wallet_flows_user_time` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户资金流水';

CREATE TABLE IF NOT EXISTS `user_commission_payout_config` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录 ID',
  `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `channel`    VARCHAR(16)  NOT NULL COMMENT 'alipay|wechat',
  `qr_mime`    VARCHAR(64)  DEFAULT NULL COMMENT 'image/png|image/jpeg|image/webp',
  `qr_image`   MEDIUMBLOB   DEFAULT NULL COMMENT '收款码图片',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_commission_payout_user_channel` (`user_id`, `channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='佣金提现收款配置';

CREATE TABLE IF NOT EXISTS `user_commission_records` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录 ID',
  `inviter_user_id` BIGINT UNSIGNED NOT NULL COMMENT '邀请人用户 ID',
  `invitee_user_id` BIGINT UNSIGNED NOT NULL COMMENT '被邀请人用户 ID',
  `order_id`        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '历史订单 id；无订单链路时为 0',
  `order_no`        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '订单号',
  `product_name`    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '商品名称',
  `order_amount`    DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '订单实付（元）',
  `rate_percent`    DECIMAL(5,2) NOT NULL COMMENT '返佣比例（%）',
  `rebate_amount`   DECIMAL(16,2) NOT NULL COMMENT '返利金额（元）',
  `status`          VARCHAR(32)  NOT NULL DEFAULT 'settled' COMMENT 'settled|reversed',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_commission_records_order` (`order_id`),
  KEY `idx_user_commission_records_inviter_time` (`inviter_user_id`, `created_at`),
  KEY `idx_user_commission_records_invitee` (`invitee_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邀请返利明细';

CREATE TABLE IF NOT EXISTS `user_commission_withdrawals` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '申请 ID',
  `user_id`       BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `amount`        DECIMAL(16,2) NOT NULL COMMENT '提现金额（元）',
  `channel`       VARCHAR(16)  NOT NULL COMMENT 'alipay|wechat',
  `payout_qr_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '收款码 URL 快照',
  `status`        VARCHAR(32)  NOT NULL DEFAULT 'pending' COMMENT 'pending|completed|failed',
  `fail_reason`   VARCHAR(512) DEFAULT NULL COMMENT '失败原因',
  `processed_at`  DATETIME     DEFAULT NULL COMMENT '处理完成时间',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_commission_withdrawals_user_time` (`user_id`, `created_at`),
  KEY `idx_user_commission_withdrawals_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='佣金提现申请';

CREATE TABLE IF NOT EXISTS `coupon_campaigns` (
  `id`                     BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '活动 ID',
  `code`                   VARCHAR(32)  NOT NULL COMMENT '活动编码',
  `name`                   VARCHAR(128) NOT NULL COMMENT '内部名称',
  `title`                  VARCHAR(128) NOT NULL COMMENT '展示标题',
  `subtitle`               VARCHAR(512) DEFAULT NULL COMMENT '副标题',
  `discount_type`          VARCHAR(16)  NOT NULL COMMENT 'fixed_amount|percent',
  `discount_value`         DECIMAL(16,2) NOT NULL COMMENT '面额或折扣值',
  `min_order_amount`       DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '最低消费（元）',
  `valid_days`             INT UNSIGNED NOT NULL DEFAULT 30 COMMENT '领取后有效天数',
  `auto_grant_on_register` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '注册是否自动发放',
  `enabled`                TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `created_at`             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_coupon_campaigns_code` (`code`),
  KEY `idx_coupon_campaigns_register` (`auto_grant_on_register`, `enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券活动';

CREATE TABLE IF NOT EXISTS `user_coupons` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '券 ID',
  `user_id`          BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `campaign_id`      BIGINT UNSIGNED NOT NULL COMMENT '活动 ID',
  `coupon_code`      VARCHAR(40)  NOT NULL COMMENT '券码',
  `discount_type`    VARCHAR(16)  NOT NULL COMMENT 'fixed_amount|percent',
  `discount_value`   DECIMAL(16,2) NOT NULL COMMENT '面额或折扣值',
  `min_order_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00 COMMENT '最低消费（元）',
  `status`           VARCHAR(16)  NOT NULL DEFAULT 'available' COMMENT 'available|used|expired',
  `valid_from`       DATETIME     NOT NULL COMMENT '生效时间',
  `valid_until`      DATETIME     NOT NULL COMMENT '过期时间',
  `used_at`          DATETIME     DEFAULT NULL COMMENT '使用时间',
  `order_id`         BIGINT UNSIGNED DEFAULT NULL COMMENT '核销订单 ID',
  `created_at`       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
  `updated_at`       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_coupons_code` (`coupon_code`),
  UNIQUE KEY `uk_user_coupons_user_campaign` (`user_id`, `campaign_id`),
  KEY `idx_user_coupons_user_status` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户优惠券';

CREATE TABLE IF NOT EXISTS `user_track_events` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '事件 ID',
  `event_type`    VARCHAR(16)  NOT NULL COMMENT 'page_view|click',
  `action`        VARCHAR(16)  NOT NULL COMMENT 'enter|click',
  `user_id`       BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户 ID，0 为游客',
  `visitor_id`    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '访客标识',
  `session_id`    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '会话 ID',
  `channel`       VARCHAR(16)  NOT NULL DEFAULT '' COMMENT '渠道',
  `app_version`   VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '客户端版本',
  `page_id`       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '页面 ID',
  `page_path`     VARCHAR(256) NOT NULL DEFAULT '' COMMENT '页面路径',
  `page_title`    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '页面标题',
  `element_id`    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '元素 ID',
  `element_name`  VARCHAR(128) NOT NULL DEFAULT '' COMMENT '元素名称',
  `target_url`    VARCHAR(512) NOT NULL DEFAULT '' COMMENT '跳转 URL',
  `api_method`    VARCHAR(16)  NOT NULL DEFAULT '' COMMENT 'API 方法',
  `api_path`      VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'API 路径',
  `api_params`    JSON DEFAULT NULL COMMENT 'API 参数快照',
  `locale`        VARCHAR(16)  NOT NULL DEFAULT 'zh-Hans' COMMENT '语言',
  `ip`            VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '客户端 IP',
  `user_agent`    VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'UA',
  `device_type`   VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '设备类型',
  `device_model`  VARCHAR(128) NOT NULL DEFAULT '' COMMENT '设备型号',
  `os_name`       VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '操作系统',
  `os_version`    VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '系统版本',
  `screen_width`  INT NOT NULL DEFAULT 0 COMMENT '屏宽 px',
  `screen_height` INT NOT NULL DEFAULT 0 COMMENT '屏高 px',
  `referrer`      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '来源页',
  `extra_json`    JSON DEFAULT NULL COMMENT '扩展字段',
  `event_at`      DATETIME NOT NULL COMMENT '事件发生时间',
  `created_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '入库时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_track_events_user_time` (`user_id`, `event_at`),
  KEY `idx_user_track_events_page_time` (`page_id`, `event_at`),
  KEY `idx_user_track_events_visitor_time` (`visitor_id`, `event_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户端行为埋点';

CREATE TABLE IF NOT EXISTS `bot_schedule_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '任务 ID',
  `module` varchar(50) NOT NULL COMMENT '模块名',
  `task_name` varchar(100) NOT NULL COMMENT '任务名',
  `interval_seconds` float NOT NULL DEFAULT 60 COMMENT '执行间隔（秒）',
  `exe_sort` int DEFAULT NULL COMMENT '同模块内执行顺序',
  `concurrency` int DEFAULT NULL COMMENT '并发度',
  `description` varchar(500) DEFAULT NULL COMMENT '说明',
  `is_enabled` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `is_strategy_enabled` tinyint(1) DEFAULT 1 COMMENT '策略是否启用',
  `last_live_time` varchar(255) DEFAULT NULL COMMENT '上次存活时间',
  `last_live_time_by_machine` json DEFAULT NULL COMMENT '各机器上次存活',
  `is_primary_machine_run` int NOT NULL DEFAULT 0 COMMENT '是否主机器执行',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_module_task` (`module`,`task_name`),
  KEY `idx_module` (`module`),
  KEY `idx_bot_schedule_module_enabled` (`module`,`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Bot 脚本调度';

-- ---------------------------------------------------------------------------
-- 1. ref 字典：学科、venue、文献源、强度与审计档位（七模块 code 见前端 front/web-pc/types.ts）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_ref_discipline` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '学科 ID',
  `code`        VARCHAR(64)  NOT NULL COMMENT '学科键，如 cs_ai|medicine|law|economics',
  `name`        VARCHAR(128) NOT NULL COMMENT '展示名',
  `name_en`     VARCHAR(128) DEFAULT NULL COMMENT '英文名',
  `sort`        INT          NOT NULL DEFAULT 0 COMMENT '排序权重',
  `literature_source_codes` JSON         NOT NULL DEFAULT (JSON_ARRAY()) COMMENT '该学科推荐文献源 paper_ref_literature_source.code 列表，顺序即 UI 默认',
  `status`      VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_discipline_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学科';

CREATE TABLE IF NOT EXISTS `paper_ref_venue` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'venue ID',
  `code`            VARCHAR(64)  NOT NULL COMMENT 'venue 键，如 neurips|iclr|nature|ssci_q1',
  `name`            VARCHAR(256) NOT NULL COMMENT 'NeurIPS 2026 / Nature 等',
  `venue_type`      VARCHAR(16)  NOT NULL COMMENT 'conference|journal|workshop|other',
  `discipline_id`   BIGINT UNSIGNED DEFAULT NULL COMMENT 'NULL=跨学科通用',
  `publisher`       VARCHAR(128) DEFAULT NULL COMMENT '出版方',
  `website_url`     VARCHAR(512) DEFAULT NULL COMMENT '官网',
  `rubric`          JSON         NOT NULL COMMENT 'venue 规则：contribution_types|experiment_bar|writing_style 等 prompt/结构化约束',
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_venue_code` (`code`),
  KEY `idx_paper_ref_venue_discipline` (`discipline_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='目标会议/期刊';

CREATE TABLE IF NOT EXISTS `paper_ref_literature_source` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '数据源 ID',
  `code`            VARCHAR(64)  NOT NULL COMMENT 'arxiv|openalex|semantic_scholar|pubmed|crossref|cnki|…',
  `name`            VARCHAR(128) NOT NULL COMMENT '展示名',
  `api_kind`        VARCHAR(32)  NOT NULL COMMENT 'rest|oai|custom',
  `base_url`        VARCHAR(512) NOT NULL COMMENT 'API 根地址',
  `auth_type`       VARCHAR(32)  NOT NULL DEFAULT 'none' COMMENT 'none|api_key|oauth',
  `config_schema`   JSON         DEFAULT NULL COMMENT '连接参数 JSON Schema（Key 名、必填项）',
  `default_config`  JSON         DEFAULT NULL COMMENT '默认非密钥配置',
  `rate_limit_hint` VARCHAR(256) DEFAULT NULL COMMENT '限流说明',
  `priority`        INT          NOT NULL DEFAULT 100 COMMENT '检索与 UI 展示顺序，越小越优先',
  `default_selected` TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '选题表单默认勾选',
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_literature_source_code` (`code`),
  KEY `idx_paper_ref_literature_source_active_priority` (`status`, `priority`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文献/API 数据源';

CREATE TABLE IF NOT EXISTS `paper_ref_execution_intensity` (
  `code`         VARCHAR(16)  NOT NULL COMMENT '档位键 fast|balanced|deep',
  `name`         VARCHAR(64)  NOT NULL COMMENT '展示名',
  `multiplier`   DECIMAL(4,2) NOT NULL DEFAULT 1.00 COMMENT '相对检索量/迭代轮数系数',
  `max_papers`   INT UNSIGNED NOT NULL DEFAULT 80 COMMENT '检索文献上限（选题/综述等）',
  `max_ideas`    INT UNSIGNED NOT NULL DEFAULT 12 COMMENT '选题候选 idea 上限',
  `is_default`   TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1',
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='执行强度档位';

CREATE TABLE IF NOT EXISTS `paper_ref_audit_level` (
  `code`                   VARCHAR(16)  NOT NULL COMMENT '档位键 standard|polished|strict',
  `name`                   VARCHAR(64)  NOT NULL COMMENT '展示名',
  `citation_strength`      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '0=关,1-3=引用审计强度',
  `claim_strength`         TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '0=关,1-3=论断审计强度',
  `kill_argument_strength` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=关,1-3=驳论审计强度',
  `audit_rounds`           TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '审计迭代轮数',
  `is_default`             TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1',
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计等级';

-- ---------------------------------------------------------------------------
-- 2. 稿件线与用户默认（工作台「当前论文」= paper_manuscript）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_manuscript` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '论文/项目 ID',
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT 'users.id',
  `title`             VARCHAR(256) NOT NULL COMMENT '工作标题 / 暂定篇名',
  `description`       TEXT         DEFAULT NULL COMMENT '简介',
  `manuscript_kind`   VARCHAR(32)  DEFAULT NULL COMMENT 'conference|journal|thesis|course|report',
  `deadline_at`       DATETIME     DEFAULT NULL COMMENT '截止日期',
  `target_words`      INT UNSIGNED DEFAULT NULL COMMENT '目标字数',
  `citation_style`    VARCHAR(64)  DEFAULT NULL COMMENT 'apa|ieee|acm|…',
  `discipline_id`     BIGINT UNSIGNED DEFAULT NULL COMMENT '默认学科 ID',
  `default_venue_id`  BIGINT UNSIGNED DEFAULT NULL COMMENT '默认 venue ID',
  `reference_gate_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '参考文献门禁',
  `workspace_uri`     VARCHAR(512) DEFAULT NULL COMMENT '平台 provision 的工作区根路径',
  `forked_from_manuscript_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '从另一篇论文 fork 时链接',
  `draft_format`      VARCHAR(16)  DEFAULT NULL COMMENT 'md|tex|pdf；论文写作当前稿',
  `draft_content_medium` MEDIUMTEXT DEFAULT NULL COMMENT '当前正文（小稿内联）',
  `draft_storage_uri` VARCHAR(1024) DEFAULT NULL COMMENT '当前正文文件路径',
  `draft_bibtex_uri`  VARCHAR(1024) DEFAULT NULL COMMENT 'BibTeX 文件路径',
  `draft_word_count`  INT UNSIGNED DEFAULT NULL COMMENT '当前稿字数',
  `draft_updated_at`  DATETIME     DEFAULT NULL COMMENT '最近一次写作/保存',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|archived',
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否工作台当前论文（每用户 active 至多一条为 1）',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_manuscript_user` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='我的论文（元数据 + 当前论文正文草稿）';

CREATE TABLE IF NOT EXISTS `paper_user_preference` (
  `user_id`                   BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `default_discipline_id`     BIGINT UNSIGNED DEFAULT NULL COMMENT '默认学科 ID',
  `default_venue_id`          BIGINT UNSIGNED DEFAULT NULL COMMENT '默认 venue ID',
  `default_intensity_code`    VARCHAR(16)  NOT NULL DEFAULT 'balanced' COMMENT '默认执行强度',
  `default_audit_level_code`  VARCHAR(16)  NOT NULL DEFAULT 'polished' COMMENT '默认审计等级',
  `default_human_checkpoint`  TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '默认开启人工检查点',
  `default_literature_source_ids` JSON     DEFAULT NULL COMMENT '默认文献源 id 数组',
  `updated_at`                DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户全局默认（个人中心·默认配置）';

-- ---------------------------------------------------------------------------
-- 3. 平台运行时（模型、算力、文献源凭证 — 不对用户暴露）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_llm_model_config` (
  `id`                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置 ID',
  `label`                VARCHAR(128) NOT NULL COMMENT '展示名',
  `provider_code`        VARCHAR(32)  NOT NULL COMMENT 'openai|anthropic|azure_openai|openai_compatible|ollama|gateway|other',
  `model_name`           VARCHAR(128) NOT NULL COMMENT '上游 model 参数',
  `api_base_url`         VARCHAR(512) NOT NULL COMMENT 'API Origin（仅 scheme://host；HTTP path 由 provider_code 在代码中拼接）',
  `api_key`              VARCHAR(512) DEFAULT NULL COMMENT 'LLM API Key 明文（库内存储；生产建议库权限+TLS）',
  `timeout_ms`           INT UNSIGNED NOT NULL DEFAULT 120000 COMMENT '超时毫秒',
  `max_retries`          TINYINT UNSIGNED NOT NULL DEFAULT 2 COMMENT '最大重试次数',
  `supports_vision`      TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否支持视觉',
  `context_window_hint`  INT UNSIGNED DEFAULT NULL COMMENT '上下文长度提示',
  `tokens_used_total`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计消耗 token（prompt+completion）',
  `extra`                JSON         DEFAULT NULL COMMENT '扩展配置',
  `status`               VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_model_cfg_status` (`status`),
  KEY `idx_paper_llm_model_cfg_provider` (`provider_code`, `model_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台 LLM 端点配置（Key/Host/模型名）';

CREATE TABLE IF NOT EXISTS `paper_llm_workflow_binding` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '绑定 ID',
  `stage_code`        VARCHAR(64)  NOT NULL DEFAULT 'default' COMMENT '环节键：default|retrieve|…（与 prompt 表同键；由代码约定）',
  `stage_name`        VARCHAR(128) NOT NULL COMMENT '环节中文名（运营/排查）',
  `model_config_id`   BIGINT UNSIGNED NOT NULL COMMENT 'paper_llm_model_config.id',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_llm_bind_stage` (`stage_code`),
  KEY `idx_paper_llm_bind_model` (`model_config_id`),
  KEY `idx_paper_llm_bind_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台：stage_code → model_config';

CREATE TABLE IF NOT EXISTS `paper_llm_prompt_template` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板 ID',
  `stage_code`        VARCHAR(64)  NOT NULL DEFAULT 'default' COMMENT '环节键：default|retrieve|generate_ideas|…（由代码约定）',
  `stage_name`        VARCHAR(128) NOT NULL COMMENT '环节中文名（运营/排查）',
  `template_body`     MEDIUMTEXT   NOT NULL COMMENT '话术正文；占位符 {{var_name}}，由代码渲染',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_llm_prompt_stage` (`stage_code`),
  KEY `idx_paper_llm_prompt_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台话术（按 stage_code）';

CREATE TABLE IF NOT EXISTS `paper_manuscript_runtime` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '运行时配置 ID',
  `manuscript_id`         BIGINT UNSIGNED NOT NULL COMMENT '论文 ID',
  `label`                 VARCHAR(128) NOT NULL DEFAULT 'default' COMMENT '配置标签',
  `literature_credentials` JSON        DEFAULT NULL COMMENT 'source_id -> 密钥密文/引用 id',
  `literature_source_ids` JSON        NOT NULL COMMENT '本稿件启用的 source id 列表',
  `gpu_profile`           JSON         DEFAULT NULL COMMENT 'SSH/集群/本地',
  `wiki_uri`              VARCHAR(512) DEFAULT NULL COMMENT 'Research Wiki 路径',
  `is_active`             TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否当前生效',
  `created_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_manuscript_runtime_ms` (`manuscript_id`, `is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='稿件平台运行时绑定（内部）';

-- ---------------------------------------------------------------------------
-- 4. 引用门禁（选题检索命中文献见 paper_output_topic_step retrieve 步 result）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_manuscript_citation_gate` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '门禁记录 ID',
  `manuscript_id`       BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `stage_code`          VARCHAR(64)  DEFAULT NULL COMMENT '校验发生时的环节键',
  `target_id`           BIGINT UNSIGNED DEFAULT NULL COMMENT '如 paper_manuscript.id 或其它 paper_output_*',
  `source_id`           BIGINT UNSIGNED DEFAULT NULL COMMENT '门禁关联文献源',
  `external_key`        VARCHAR(256) DEFAULT NULL COMMENT '文献外部键 arxiv:… / doi:…（与 retrieve 步 result 内 hits 一致）',
  `cited_key`           VARCHAR(256) DEFAULT NULL COMMENT '正文中的 cite key',
  `gate_type`           VARCHAR(32)  NOT NULL COMMENT '门禁类型：引用/引用审计/论断审计/驳论',
  `status`              VARCHAR(16)  NOT NULL COMMENT 'verified|rejected|pending|waived',
  `detail`              JSON         DEFAULT NULL COMMENT '校验详情 JSON',
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_ms_citation_gate_ms` (`manuscript_id`, `gate_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='本篇引用文献与审计校验记录（参考文献门禁台账）';

-- ---------------------------------------------------------------------------
-- 5. 各模块业务产出 paper_output_*（均关联 paper_manuscript.id）
-- ---------------------------------------------------------------------------

-- 选题发现：一轮 run 固定四步，每步一行（stage_code 见 paper_llm_workflow_binding）
CREATE TABLE IF NOT EXISTS `paper_output_topic_step` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录 ID',
  `manuscript_id`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '论文 ID；选题未确认综述前为 0，确认后反填',
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `run_version`       INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '该用户选题第几轮（同轮四步相同）',
  `stage_code`        VARCHAR(32)  NOT NULL COMMENT 'retrieve|generate_ideas|audit（旧 run 可有 novelty）',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'pending' COMMENT 'pending|running|completed|failed|cancelled',
  `result`            JSON         DEFAULT NULL COMMENT '本步产出：retrieve→literature_hits[]；generate_ideas→ideas[]；novelty/audit→报告结构',
  `summary_text`      MEDIUMTEXT   DEFAULT NULL COMMENT '本步可读摘要/报告',
  `input_params`      JSON         DEFAULT NULL COMMENT '本轮表单快照（通常写在 retrieve 步）',
  `extra`             JSON         DEFAULT NULL COMMENT '扩展（token、error、llm_*；retrieve 可存 hit_count、literature_platform_requests[] 等）',
  `files`             JSON         DEFAULT NULL COMMENT '本步关联文件列表（PDF 路径等，JSON 数组只追加）',
  `started_at`        DATETIME     DEFAULT NULL COMMENT '本步开始时间',
  `completed_at`      DATETIME     DEFAULT NULL COMMENT '本步结束时间',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '入库时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_output_topic_step_run` (`user_id`, `run_version`, `stage_code`),
  KEY `idx_paper_output_topic_step_ms` (`manuscript_id`, `stage_code`),
  KEY `idx_paper_output_topic_step_user` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='选题发现 · 四步各一行（含检索文献 hits）';

CREATE TABLE IF NOT EXISTS `paper_output_literature_review` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '综述 ID',
  `manuscript_id`     BIGINT UNSIGNED NOT NULL COMMENT '论文 ID',
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `version`           INT          NOT NULL DEFAULT 1 COMMENT '版本号',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed' COMMENT 'draft|completed|generating_experiment_plan|deleted（软删）',
  `structure`         VARCHAR(32)  DEFAULT NULL COMMENT 'thematic|chronological|method',
  `title`             VARCHAR(256) DEFAULT NULL COMMENT '标题',
  `summary`           TEXT         DEFAULT NULL COMMENT '摘要',
  `content_medium`    MEDIUMTEXT   DEFAULT NULL COMMENT '正文（内联）',
  `storage_uri`       VARCHAR(1024) DEFAULT NULL COMMENT '正文文件路径',
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md' COMMENT 'md|tex 等',
  `citations`         JSON         DEFAULT NULL COMMENT '引用表 / cite key 列表',
  `input_params`      JSON         DEFAULT NULL COMMENT '生成参数快照',
  `meta`              JSON         DEFAULT NULL COMMENT '扩展元数据',
  `paper_output_experiment_plan_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联 paper_output_experiment_plan.id',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_lit_review_ms` (`manuscript_id`, `created_at`),
  KEY `idx_paper_output_lit_review_exp_plan` (`paper_output_experiment_plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文献综述';

CREATE TABLE IF NOT EXISTS `paper_output_experiment_plan` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '实验计划 ID',
  `manuscript_id`     BIGINT UNSIGNED NOT NULL COMMENT '论文 ID',
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `version`           INT          NOT NULL DEFAULT 1 COMMENT '版本号',
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否当前版本',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed' COMMENT 'draft|completed',
  `title`             VARCHAR(256) DEFAULT NULL COMMENT '标题',
  `summary`           TEXT         DEFAULT NULL COMMENT '摘要',
  `content_medium`    MEDIUMTEXT   DEFAULT NULL COMMENT '正文（内联）',
  `storage_uri`       VARCHAR(1024) DEFAULT NULL COMMENT '正文文件路径',
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md' COMMENT 'md|tex 等',
  `input_params`      JSON         DEFAULT NULL COMMENT '生成参数快照',
  `meta`              JSON         DEFAULT NULL COMMENT 'baseline、ablation、资源估算等',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_exp_plan_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实验规划';

CREATE TABLE IF NOT EXISTS `paper_output_experiment_review` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '实验审查 ID',
  `manuscript_id`     BIGINT UNSIGNED NOT NULL COMMENT '论文 ID',
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `version`           INT          NOT NULL DEFAULT 1 COMMENT '版本号',
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否当前版本',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed' COMMENT 'draft|completed',
  `title`             VARCHAR(256) DEFAULT NULL COMMENT '标题',
  `summary`           TEXT         DEFAULT NULL COMMENT '摘要',
  `content_medium`    MEDIUMTEXT   DEFAULT NULL COMMENT '结果审查报告正文',
  `storage_uri`       VARCHAR(1024) DEFAULT NULL COMMENT '报告文件路径',
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md' COMMENT 'md|tex 等',
  `findings`          JSON         DEFAULT NULL COMMENT '结构化问题与建议',
  `input_params`      JSON         DEFAULT NULL COMMENT '生成参数快照',
  `meta`              JSON         DEFAULT NULL COMMENT '扩展元数据',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_exp_review_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实验审查（结果审查 auto_review）';

CREATE TABLE IF NOT EXISTS `paper_manuscript_review` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '论文审查 ID',
  `manuscript_id`     BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `version`           INT          NOT NULL DEFAULT 1 COMMENT '版本号',
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否当前版本',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed' COMMENT 'draft|completed',
  `title`             VARCHAR(256) DEFAULT NULL COMMENT '标题',
  `summary`           TEXT         DEFAULT NULL COMMENT '摘要',
  `content_medium`    MEDIUMTEXT   DEFAULT NULL COMMENT '论文审查报告',
  `storage_uri`       VARCHAR(1024) DEFAULT NULL COMMENT '报告文件路径',
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md' COMMENT 'md|tex 等',
  `findings`          JSON         DEFAULT NULL COMMENT '分维度审稿意见',
  `input_params`      JSON         DEFAULT NULL COMMENT '生成参数快照',
  `meta`              JSON         DEFAULT NULL COMMENT '扩展元数据',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_ms_review_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文审查（manuscript_analysis 模块产出）';

CREATE TABLE IF NOT EXISTS `paper_output_figure` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '图表 ID',
  `manuscript_id`     BIGINT UNSIGNED NOT NULL COMMENT '论文 ID',
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `origin`            VARCHAR(16)  NOT NULL COMMENT 'upload|generated',
  `file_kind`         VARCHAR(16)  NOT NULL COMMENT 'image|pdf|vector|other',
  `file_name`         VARCHAR(512) NOT NULL COMMENT '原始文件名',
  `title`             VARCHAR(512) NOT NULL COMMENT '图表标题',
  `note`              TEXT         DEFAULT NULL COMMENT '备注',
  `size_bytes`        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件大小字节',
  `storage_uri`       VARCHAR(1024) DEFAULT NULL COMMENT '文件路径',
  `content_hash`      CHAR(64)     DEFAULT NULL COMMENT '文件 SHA-256',
  `prompt_snapshot`   TEXT         DEFAULT NULL COMMENT '生成图时的提示词摘要',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_figure_ms` (`manuscript_id`, `origin`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表记录（上传 + 一键生成）';

-- ---------------------------------------------------------------------------
-- 7. LLM 调用日志（计费、排错；不存完整 prompt 时可只存 hash）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_llm_call_logs` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志 ID',
  `manuscript_id`   BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `stage_code`      VARCHAR(64)  DEFAULT NULL COMMENT '环节键（与 binding/prompt 同 stage_code）',
  `model_config_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_model_config.id',
  `workflow_binding_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_workflow_binding.id',
  `prompt_template_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_prompt_template.id',
  `model_name`      VARCHAR(128) NOT NULL COMMENT '模型名快照',
  `prompt_tokens`   INT          DEFAULT NULL COMMENT '输入 token',
  `completion_tokens` INT        DEFAULT NULL COMMENT '输出 token',
  `latency_ms`      INT          NOT NULL DEFAULT 0 COMMENT '耗时毫秒',
  `status`          SMALLINT     NOT NULL COMMENT 'HTTP 或业务码',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '调用时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_call_logs_ms` (`manuscript_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='LLM 调用明细日志';

-- ---------------------------------------------------------------------------
-- 8. 上传文献、操作日志（阶段进度由 paper_output_* / draft_* 是否存在推导，不落单独 progress 表）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_user_literature` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '上传记录 ID',
  `user_id`       BIGINT UNSIGNED NOT NULL COMMENT 'users.id',
  `manuscript_id` BIGINT UNSIGNED NOT NULL COMMENT '本篇「上传文献」；选题 user_library 语料',
  `file_kind`     VARCHAR(16)  NOT NULL COMMENT 'pdf|bib|txt|other',
  `file_name`     VARCHAR(512) NOT NULL COMMENT '原始文件名',
  `title`         VARCHAR(512) NOT NULL COMMENT '文献标题',
  `note`          TEXT         DEFAULT NULL COMMENT '备注',
  `size_bytes`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件大小字节',
  `storage_uri`   VARCHAR(1024) DEFAULT NULL COMMENT '对象存储或工作区相对路径',
  `content_hash`  CHAR(64)     DEFAULT NULL COMMENT 'SHA-256',
  `status`        VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_user_lit_ms` (`manuscript_id`, `status`),
  KEY `idx_paper_user_lit_user` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户上传文献（上传文献模块 / 参考文献门禁语料）';

CREATE TABLE IF NOT EXISTS `paper_operation_log` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志 ID',
  `user_id`            BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `manuscript_id`      BIGINT UNSIGNED DEFAULT NULL COMMENT '论文 ID',
  `manuscript_title`   VARCHAR(256) DEFAULT NULL COMMENT '论文标题快照',
  `stage_code`         VARCHAR(64)  NOT NULL DEFAULT 'default' COMMENT '环节键（展示名见 action / note）',
  `action`             VARCHAR(512) NOT NULL COMMENT '操作描述',
  `tokens_prompt`      INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '输入 token',
  `tokens_completion`  INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '输出 token',
  `tokens_total`       INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总 token',
  `status`             VARCHAR(16)  NOT NULL COMMENT 'success|failed|cancelled',
  `note`               TEXT         DEFAULT NULL COMMENT '备注',
  `occurred_at`        DATETIME     NOT NULL COMMENT '业务发生时间',
  `created_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '入库时间',
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
  ),
  (
    'ai_agent_paper',
    'topic_literature_pdf_download',
    15,
    '选题 retrieve=running：下载 literature_downloads PDF 至 storage/literature',
    1,
    1
  ),
  (
    'ai_agent_paper',
    'topic_generate_ideas_llm',
    15,
    '选题 generate_ideas=running：异步调用 LLM 脑暴+新颖性预填',
    1,
    1
  ),
  (
    'ai_agent_paper',
    'topic_audit_llm',
    15,
    '选题 audit=running：异步调用 LLM 审查结论+文献综述并写入 paper_output_literature_review',
    1,
    1
  ),
  (
    'ai_agent_paper',
    'experiment_plan_llm',
    15,
    '文献综述 generating_experiment_plan：异步 LLM 生成实验方案',
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

INSERT INTO `paper_ref_execution_intensity` (`code`, `name`, `multiplier`, `max_papers`, `max_ideas`, `is_default`) VALUES
  ('fast',     '更快', 0.60, 30,  6, 0),
  ('balanced', 'Balanced（平衡）', 1.00, 80,  12, 1),
  ('deep',     '更深', 1.80, 200, 20, 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `multiplier` = VALUES(`multiplier`),
  `max_papers` = VALUES(`max_papers`),
  `max_ideas` = VALUES(`max_ideas`),
  `is_default` = VALUES(`is_default`);

INSERT INTO `paper_ref_audit_level` (
  `code`, `name`, `citation_strength`, `claim_strength`, `kill_argument_strength`, `audit_rounds`, `is_default`
) VALUES
  ('standard', '标准', 1, 1, 0, 1, 0),
  ('polished', '精修', 2, 2, 1, 2, 1),
  ('strict',   '严格', 3, 3, 2, 3, 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `citation_strength` = VALUES(`citation_strength`),
  `claim_strength` = VALUES(`claim_strength`),
  `kill_argument_strength` = VALUES(`kill_argument_strength`),
  `audit_rounds` = VALUES(`audit_rounds`),
  `is_default` = VALUES(`is_default`);

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

INSERT INTO `paper_ref_literature_source` (`code`, `name`, `api_kind`, `base_url`, `auth_type`, `default_config`, `config_schema`, `priority`, `default_selected`) VALUES
  ('arxiv', 'arXiv', 'rest', 'https://export.arxiv.org/api/query', 'none', NULL, NULL, 1, 1),
  ('openalex', 'OpenAlex', 'rest', 'https://api.openalex.org/works', 'none',
   JSON_OBJECT('mailto', ''),
   JSON_OBJECT('type', 'object', 'properties', JSON_OBJECT('mailto', JSON_OBJECT('type', 'string', 'description', 'User-Agent 礼貌池'))), 2, 1),
  ('semantic_scholar', 'Semantic Scholar', 'rest', 'https://api.semanticscholar.org/graph/v1/paper/search', 'api_key',
   NULL,
   JSON_OBJECT('type', 'object', 'properties', JSON_OBJECT('api_key', JSON_OBJECT('type', 'string', 'description', 'Header x-api-key；可选'))), 3, 1),
  ('crossref', 'Crossref', 'rest', 'https://api.crossref.org/works', 'none', NULL, NULL, 4, 0),
  ('pubmed', 'PubMed', 'rest', 'https://eutils.ncbi.nlm.nih.gov/entrez/eutils', 'none', NULL, NULL, 5, 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `base_url` = VALUES(`base_url`),
  `auth_type` = VALUES(`auth_type`),
  `default_config` = VALUES(`default_config`),
  `config_schema` = VALUES(`config_schema`),
  `priority` = VALUES(`priority`),
  `default_selected` = VALUES(`default_selected`);

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
  'claude-sonnet-5-5',
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

-- 当前产品范围：选题发现（topic-discovery）四步 + default 兜底
INSERT INTO `paper_llm_workflow_binding` (`stage_code`, `stage_name`, `model_config_id`, `status`) VALUES
  ('default', '默认（全局兜底）', @paper_llm_model_id, 'active'),
  ('retrieve', '多源文献检索与校验入库', @paper_llm_model_id, 'active'),
  ('generate_ideas', '脑暴选题', @paper_llm_model_id, 'active'),
  ('novelty', '新颖性检查', @paper_llm_model_id, 'active'),
  ('audit', '审查结论+生成文献综述', @paper_llm_model_id, 'active'),
  ('experiment_plan', '生成实验方案', @paper_llm_model_id, 'active')
ON DUPLICATE KEY UPDATE
  `stage_name` = VALUES(`stage_name`),
  `model_config_id` = VALUES(`model_config_id`),
  `status` = VALUES(`status`);

-- 11.3 选题发现话术（front/web-pc TOPIC_DISCOVERY_FLOW_STEPS）
INSERT INTO `paper_llm_prompt_template` (`stage_code`, `stage_name`, `template_body`, `status`) VALUES
  ('default', '默认（全局兜底）',
   'You are an expert academic research assistant for the Paper Agent platform. Be precise, cite evidence when provided, and state uncertainty clearly. Default to clear scholarly English.',
   'active'),
  ('retrieve', '多源文献检索与校验入库',
   'Stage: multi-source literature retrieve and validate for direction {{direction}}. Dedupe by external_key; list coverage gaps.',
   'active'),
  ('generate_ideas', '脑暴选题',
   '环节：脑暴 + 新颖性（与 novelty 同一次模型调用完成）。最多 {{max_ideas}} 条 idea；仅依据用户消息 Corpus 与 CorpusFiles；输出 JSON 含 ideas 与 novelty（lines/risks/synthesis），以用户消息 schema 为准。',
   'active'),
  ('novelty', '新颖性检查',
   '（通常与 generate_ideas 合并为一次调用；本模板仅用于旧 run 单独补跑 novelty。）方向 {{direction}}：对照 Corpus 比较每条 idea，输出 lines/risks。',
   'active'),
  ('audit', '审查结论+生成文献综述',
   '环节：审查结论 + 生成文献综述。方向：{{direction}}。目标期刊：{{venue}}。须再次审查第二步 generate_ideas（含 novelty）是否贴题；以严格审稿人视角给出 issues/summary/off_topic；并基于 Corpus 生成文献综述（structure/content_medium/citations）。输出 JSON 以用户消息 schema 为准（含 audit 与 literature_review）。',
   'active'),
  ('experiment_plan', '生成实验方案',
   '环节：生成实验方案（experiment planning）。方向：{{direction}}。目标期刊/会议：{{venue}}。须以用户消息中的 LiteratureReview（文献综述全文/摘要）与选题 Corpus 为依据，不得编造未给出的基线或数据。输出可审稿的实验计划：核心假设、主实验与对照基线、评价指标、消融设计、算力/数据资源与时间线（写入 paper_output_experiment_plan：title、summary、content_medium、meta）。严格 JSON，字段以用户消息 schema 为准（含 experiment_plan）。',
   'active')
ON DUPLICATE KEY UPDATE
  `stage_name` = VALUES(`stage_name`),
  `template_body` = VALUES(`template_body`),
  `status` = VALUES(`status`);

-- ---------------------------------------------------------------------------
-- 升级脚本（已有库执行一次即可）
-- ---------------------------------------------------------------------------
-- ALTER TABLE `paper_ref_execution_intensity`
--   ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1' AFTER `max_ideas`;
-- ALTER TABLE `paper_ref_audit_level`
--   ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1' AFTER `audit_rounds`;
-- UPDATE `paper_ref_execution_intensity` SET `is_default` = 0;
-- UPDATE `paper_ref_execution_intensity` SET `is_default` = 1 WHERE `code` = 'balanced';
-- UPDATE `paper_ref_audit_level` SET `is_default` = 0;
-- UPDATE `paper_ref_audit_level` SET `is_default` = 1 WHERE `code` = 'polished';
-- ALTER TABLE `paper_output_topic_step`
--   MODIFY COLUMN `manuscript_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '论文 ID；选题未确认综述前为 0，确认后反填';
-- ALTER TABLE `paper_output_topic_step` DROP INDEX `uk_paper_output_topic_step_run`;
-- ALTER TABLE `paper_output_topic_step` ADD UNIQUE KEY `uk_paper_output_topic_step_run` (`user_id`, `run_version`, `stage_code`);
-- ALTER TABLE `paper_output_topic_step` DROP COLUMN `is_current_run`;
-- ALTER TABLE `paper_output_topic_step` DROP INDEX `idx_paper_output_topic_step_ms`;
-- ALTER TABLE `paper_output_topic_step` ADD KEY `idx_paper_output_topic_step_ms` (`manuscript_id`, `stage_code`);
-- ALTER TABLE `paper_output_topic_step`
--   CHANGE COLUMN `meta` `extra` JSON DEFAULT NULL COMMENT '扩展（token、error、llm_*；retrieve 可存 hit_count 等）';

SET FOREIGN_KEY_CHECKS = 1;
