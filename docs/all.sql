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
-- | strategy   | paper_llm_strategy_package, paper_llm_strategy_route |
-- |            | paper_ref_venue（strategy_profile_key 由策略包解释） |
-- | manuscript | paper_manuscript（我的论文） |
-- | output     | paper_output_*（各模块业务产出，均含 manuscript_id） |
-- | user       | paper_user_preference, paper_user_literature |
-- | model      | paper_llm_model_config, paper_llm_constraint_gate, paper_exec_intensity_topic, paper_exec_intensity_write_paper |
-- | llm        | paper_llm_call_logs |
-- 多数 output 表含 manuscript_id → paper_manuscript.id（逻辑外键，不设 DB FK）；paper_user_literature 按 user_id

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
  `status`      VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_discipline_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学科';

CREATE TABLE IF NOT EXISTS `paper_ref_venue` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'venue ID',
  `code`            VARCHAR(64)  NOT NULL COMMENT 'venue 键，如 neurips|iclr|nature',
  `name`            VARCHAR(256) NOT NULL COMMENT 'NeurIPS / Nature 等',
  `sort`            INT          NOT NULL DEFAULT 100 COMMENT '选题下拉展示顺序，越小越靠前',
  `venue_type`      VARCHAR(16)  NOT NULL COMMENT 'conference|journal|workshop|other',
  `strategy_profile_key` VARCHAR(64) NOT NULL DEFAULT 'venue/generic' COMMENT '策略包 venue profile，如 venue/ml_top3',
  `publisher`       VARCHAR(128) DEFAULT NULL COMMENT '出版方',
  `website_url`     VARCHAR(512) DEFAULT NULL COMMENT '官网',
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_ref_venue_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='目标会议/期刊（具体可选）';

CREATE TABLE IF NOT EXISTS `paper_literature_platform` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '平台 ID',
  `code`            VARCHAR(64)  NOT NULL COMMENT 'arxiv|openalex|semantic_scholar|pubmed|crossref|cnki|…',
  `name`            VARCHAR(128) NOT NULL COMMENT '展示名',
  `api_kind`        VARCHAR(32)  NOT NULL COMMENT 'rest|oai|custom',
  `base_url`        VARCHAR(512) NOT NULL COMMENT 'API 根地址',
  `auth_type`       VARCHAR(32)  NOT NULL DEFAULT 'none' COMMENT 'none|api_key|oauth',
  `config_schema`   JSON         DEFAULT NULL COMMENT '连接参数 JSON Schema（Key 名、必填项）',
  `default_config`  JSON         DEFAULT NULL COMMENT '默认非密钥配置',
  `rate_limit_hint` VARCHAR(256) DEFAULT NULL COMMENT '限流说明',
  `sort`            INT          NOT NULL DEFAULT 100 COMMENT '检索与 UI 展示顺序，越小越靠前',
  `default_selected` TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '选题表单默认勾选',
  `status`          VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_literature_platform_code` (`code`),
  KEY `idx_paper_literature_platform_active_sort` (`status`, `sort`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文献检索平台（API 配置）';

CREATE TABLE IF NOT EXISTS `paper_exec_intensity_topic` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `code`         VARCHAR(16)  NOT NULL COMMENT '档位键 fast|balanced|deep',
  `name`         VARCHAR(64)  NOT NULL COMMENT '展示名',
  `exec_rounds`  TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '选题生成↔审计迭代轮数（Corpus 冻结）',
  `max_papers`   INT UNSIGNED NOT NULL DEFAULT 80 COMMENT '检索文献上限（选题发现）',
  `max_ideas`    INT UNSIGNED NOT NULL DEFAULT 12 COMMENT '选题候选 idea 上限',
  `is_default`   TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_exec_intensity_topic_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='选题发现·执行强度档位';

CREATE TABLE IF NOT EXISTS `paper_exec_intensity_write_paper` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `code`         VARCHAR(16)  NOT NULL COMMENT '档位键 fast|balanced|deep',
  `name`         VARCHAR(64)  NOT NULL COMMENT '展示名',
  `exec_rounds`  TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '选题生成↔审计迭代轮数（Corpus 冻结）',
  `is_default`   TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_exec_intensity_write_paper_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文写作·执行强度档位';

CREATE TABLE IF NOT EXISTS `paper_llm_constraint_gate` (
  `prompt_scene`       VARCHAR(32)  NOT NULL COMMENT 'topic|write_paper',
  `code`               VARCHAR(16)  NOT NULL COMMENT '档位键 standard|polished|strict',
  `name`               VARCHAR(64)  NOT NULL COMMENT '展示名',
  `is_default`         TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '前端默认选中项，每 prompt_scene 仅一条应为 1',
  PRIMARY KEY (`prompt_scene`, `code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计等级（语义由策略包解释）';

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
  `discipline_id`     BIGINT UNSIGNED DEFAULT NULL COMMENT '学科 ID',
  `content_language`  VARCHAR(8)   NOT NULL DEFAULT 'en' COMMENT 'zh|en：与 AI 沟通及产出文本语言',
  `reference_gate_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '参考文献门禁',
  `workspace_uri`     VARCHAR(512) DEFAULT NULL COMMENT '平台 provision 的工作区根路径',
  `forked_from_manuscript_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '从另一篇论文 fork 时链接',
  `draft_format`      VARCHAR(16)  DEFAULT NULL COMMENT 'md|tex|pdf；论文写作当前稿',
  `draft_content_medium` MEDIUMTEXT DEFAULT NULL COMMENT '当前正文（小稿内联）',
  `draft_storage_uri` VARCHAR(1024) DEFAULT NULL COMMENT '当前正文文件路径',
  `draft_bibtex_uri`  VARCHAR(1024) DEFAULT NULL COMMENT 'BibTeX 文件路径',
  `draft_word_count`  INT UNSIGNED DEFAULT NULL COMMENT '当前稿字数',
  `draft_updated_at`  DATETIME     DEFAULT NULL COMMENT '最近一次写作/保存',
  `draft_version`     INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '论文写作稿版本（每次成功生成整稿 +1；0=尚未生成）',
  `draft_status`        VARCHAR(16)  NOT NULL DEFAULT 'writing' COMMENT '定稿：writing|ready|frozen',
  `submission_status` VARCHAR(24)  NOT NULL DEFAULT 'not_submitted' COMMENT '投递汇总：not_submitted|submitted|under_review|revision|rebuttal|resubmitted|accepted|rejected',
  `submission_label`  VARCHAR(128) DEFAULT NULL COMMENT '投递展示，如 NeurIPS 2026 · 审稿中',
  `submission_deadline` DATETIME   DEFAULT NULL COMMENT '最近投递相关截止',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|archived',
  `is_current`        TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否工作台当前论文（每用户 active 至多一条为 1）',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_manuscript_user` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='我的论文（元数据 + 当前论文正文草稿）';

CREATE TABLE IF NOT EXISTS `paper_manuscript_submission` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '投递记录 ID',
  `manuscript_id`       BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `user_id`             BIGINT UNSIGNED NOT NULL COMMENT 'users.id',
  `venue_code`          VARCHAR(64)  NOT NULL COMMENT '目标 venue 编码',
  `venue_label`         VARCHAR(128) NOT NULL COMMENT '展示名',
  `track`               VARCHAR(16)  NOT NULL DEFAULT 'main' COMMENT 'main|short',
  `round`               INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '修回轮次',
  `status`              VARCHAR(24)  NOT NULL DEFAULT 'submitted' COMMENT '投递状态',
  `draft_version`       INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定写作稿本版本',
  `anonymous`           TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '双盲投稿',
  `portal_url`          VARCHAR(1024) DEFAULT NULL,
  `external_id`         VARCHAR(128)  DEFAULT NULL,
  `title`               VARCHAR(512)  DEFAULT NULL,
  `abstract`            TEXT          DEFAULT NULL,
  `main_pdf_uri`        VARCHAR(1024) DEFAULT NULL,
  `main_pdf_filename`   VARCHAR(256)  DEFAULT NULL,
  `decision`            VARCHAR(128)  DEFAULT NULL,
  `comments_summary`    TEXT          DEFAULT NULL,
  `revision_deadline`   DATETIME(3)   DEFAULT NULL,
  `submitted_at`        DATETIME(3)   DEFAULT NULL,
  `checklist`           JSON          DEFAULT NULL,
  `record_status`       VARCHAR(16)   NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`          DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`          DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_pms_manuscript` (`manuscript_id`, `record_status`, `updated_at`),
  KEY `idx_pms_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文对外投递记录';

CREATE TABLE IF NOT EXISTS `paper_manuscript_submission_event` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `submission_id` BIGINT UNSIGNED NOT NULL,
  `event_type`    VARCHAR(32)  NOT NULL,
  `message`       VARCHAR(512) NOT NULL DEFAULT '',
  `created_at`    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_pmse_submission` (`submission_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文投递操作日志';

CREATE TABLE IF NOT EXISTS `paper_user_preference` (
  `user_id`                   BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `default_discipline_id`     BIGINT UNSIGNED DEFAULT NULL COMMENT '默认学科 ID',
  `default_venue_id`          BIGINT UNSIGNED DEFAULT NULL COMMENT '默认 venue ID',
  `default_intensity_code`    VARCHAR(16)  NOT NULL DEFAULT 'balanced' COMMENT '默认执行强度',
  `default_audit_level_code`  VARCHAR(16)  NOT NULL DEFAULT 'standard' COMMENT '默认审计等级',
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
  `scope`                VARCHAR(32)  NOT NULL DEFAULT 'platform_delegate' COMMENT 'platform_delegate|strategy_only',
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

CREATE TABLE IF NOT EXISTS `paper_llm_strategy_package` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '策略包 ID',
  `code`              VARCHAR(64)  NOT NULL COMMENT '包键：aris|…（本地 CLI）',
  `name`              VARCHAR(128) NOT NULL COMMENT '展示名',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_llm_strategy_package_code` (`code`),
  KEY `idx_paper_llm_strategy_package_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 策略包';

CREATE TABLE IF NOT EXISTS `paper_llm_strategy_route` (
  `id`                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `behavior_key`            VARCHAR(128) NOT NULL COMMENT '如 topic.stage.topic_generate',
  `module`                  VARCHAR(64)  NOT NULL DEFAULT '',
  `action`                  VARCHAR(64)  NOT NULL DEFAULT '',
  `strategy_package_id`     BIGINT UNSIGNED NOT NULL COMMENT 'paper_llm_strategy_package.id',
  `capability_id`           VARCHAR(128) NOT NULL DEFAULT '',
  `default_model_config_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `fallback_route_id`       BIGINT UNSIGNED DEFAULT NULL,
  `config_json`             JSON         DEFAULT NULL,
  `priority`                INT          NOT NULL DEFAULT 0,
  `status`                  VARCHAR(16)  NOT NULL DEFAULT 'active',
  `created_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_llm_strategy_route_behavior` (`behavior_key`),
  KEY `idx_paper_llm_strategy_route_pkg` (`strategy_package_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='需求行为 → 策略包';

-- ---------------------------------------------------------------------------
-- 4. 各模块业务产出 paper_output_*（均关联 paper_manuscript.id；retrieve 步 result 承载文献命中）
-- ---------------------------------------------------------------------------

-- 选题发现：一轮 run 固定两步 retrieve + topic_discovery
CREATE TABLE IF NOT EXISTS `paper_output_topic_step` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录 ID',
  `manuscript_id`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '论文 ID；选题未确认综述前为 0，确认后反填',
  `user_id`           BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `run_version`       INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '该用户选题第几轮（同轮四步相同）',
  `stage_code`        VARCHAR(32)  NOT NULL COMMENT 'retrieve|topic_discovery（旧 run 可有 generate_ideas/audit/novelty）',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'pending' COMMENT 'pending|running|completed|failed|cancelled',
  `exec_progress_pct` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '执行进度 0-100（topic_discovery 等多轮 LLM 步）',
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
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'completed' COMMENT 'draft|completed|gen_exp_plan|gen_exp_fail|deleted（软删）',
  `from`              VARCHAR(16)  NOT NULL DEFAULT 'generated' COMMENT 'upload|generated',
  `title`             VARCHAR(256) DEFAULT NULL COMMENT '标题',
  `content_medium`    MEDIUMTEXT   DEFAULT NULL COMMENT '正文（内联）',
  `pdf_uri`           VARCHAR(1024) DEFAULT NULL COMMENT 'PDF 相对 storage 路径',
  `word_uri`          VARCHAR(1024) DEFAULT NULL COMMENT 'Word 相对 storage 路径',
  `ppt_uri`           VARCHAR(1024) DEFAULT NULL COMMENT '汇报 PPT 相对 storage 路径',
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md' COMMENT 'md|tex 等',
  `experiment_plan_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联 paper_output_experiment_plan.id',
  `experiment_plan_llm_request` JSON DEFAULT NULL COMMENT '生成实验方案 LLM 请求',
  `experiment_plan_llm_response` JSON DEFAULT NULL COMMENT '生成实验方案 LLM 返回',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_lit_review_ms` (`manuscript_id`, `created_at`),
  KEY `idx_paper_lit_review_experiment_plan` (`experiment_plan_id`)
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
  `storage_uri`       VARCHAR(1024) DEFAULT NULL COMMENT '正文文件路径（兼容：默认同 pdf_uri）',
  `pdf_uri`           VARCHAR(1024) DEFAULT NULL COMMENT 'PDF 相对 storage 路径',
  `word_uri`          VARCHAR(1024) DEFAULT NULL COMMENT 'Word 相对 storage 路径',
  `experiment_data_uri` VARCHAR(1024) DEFAULT NULL COMMENT '实验数据文件相对 storage/ 路径（experiment-data/…）',
  `format`            VARCHAR(16)  NOT NULL DEFAULT 'md' COMMENT 'md|tex 等',
  `input_params`      JSON         DEFAULT NULL COMMENT '生成参数快照',
  `meta`              JSON         DEFAULT NULL COMMENT 'baseline、ablation、资源估算等',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_output_exp_plan_ms` (`manuscript_id`, `is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实验规划';

CREATE TABLE IF NOT EXISTS `paper_user_figure` (
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
  KEY `idx_paper_user_fig_ms` (`manuscript_id`, `origin`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户图表资产（本地上传 + 一键生成）';

-- ---------------------------------------------------------------------------
-- 7. LLM 调用日志（计费、排错；每次调用存 request/response JSON）
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `paper_llm_call_logs` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志 ID',
  `manuscript_id`   BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `stage_code`      VARCHAR(64)  DEFAULT NULL COMMENT '环节键',
  `behavior_key`    VARCHAR(128) DEFAULT NULL COMMENT 'paper_llm_strategy_route.behavior_key',
  `strategy_package_id` BIGINT UNSIGNED DEFAULT NULL,
  `capability_id`   VARCHAR(128) DEFAULT NULL,
  `external_trace_id` VARCHAR(128) DEFAULT NULL,
  `model_config_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_model_config.id',
  `model_name`      VARCHAR(128) NOT NULL COMMENT '模型名快照',
  `request_json`    JSON         NOT NULL COMMENT '请求快照：stage_code、system_prompt、user_prompt',
  `response_json`   JSON         NOT NULL COMMENT '响应快照：text；失败时含 error',
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
  `from`          VARCHAR(64)  NOT NULL DEFAULT 'user_upload' COMMENT '来源：user_upload|arxiv|openalex|semantic_scholar|crossref|pubmed',
  `file_ext`      VARCHAR(16)  NOT NULL COMMENT 'pdf|bib|txt|other',
  `file_name`         VARCHAR(512) NOT NULL COMMENT '原始文件名',
  `literature_title`  VARCHAR(512) NOT NULL COMMENT '文献标题',
  `note`              TEXT         DEFAULT NULL COMMENT '备注',
  `size_bytes`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '文件大小字节',
  `storage_uri`   VARCHAR(1024) DEFAULT NULL COMMENT '对象存储或工作区相对路径',
  `content_hash`  CHAR(64)     DEFAULT NULL COMMENT 'SHA-256',
  `status`        VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_user_lit_user` (`user_id`, `status`, `created_at`),
  KEY `idx_paper_user_lit_from` (`user_id`, `from`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文献语料（用户上传或第三方检索下载）';

CREATE TABLE IF NOT EXISTS `paper_user_literature_category_rel` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `user_literature_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_literature.id',
  `category_id`        BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_literature_category.id',
  `created_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_lit_cat_rel` (`user_literature_id`, `category_id`),
  KEY `idx_paper_user_lit_cat_rel_cat` (`category_id`, `user_literature_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文献与分类多对多';

CREATE TABLE IF NOT EXISTS `paper_user_literature_manuscript_rel` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `user_literature_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_literature.id',
  `manuscript_id`      BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `created_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_lit_ms_rel` (`user_literature_id`, `manuscript_id`),
  KEY `idx_paper_user_lit_ms_rel_ms` (`manuscript_id`, `user_literature_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文献与论文章多对多';

CREATE TABLE IF NOT EXISTS `paper_user_literature_category` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类 ID',
  `name`       VARCHAR(128)    NOT NULL COMMENT '分类名称',
  `status`     VARCHAR(16)     NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_user_lit_cat_status_id` (`status`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文献分类';

CREATE TABLE IF NOT EXISTS `paper_user_figure_category_rel` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `figure_id`   BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_figure.id',
  `category_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_figure_category.id',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_fig_cat_rel` (`figure_id`, `category_id`),
  KEY `idx_paper_user_fig_cat_rel_cat` (`category_id`, `figure_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表与分类多对多';

CREATE TABLE IF NOT EXISTS `paper_user_figure_manuscript_rel` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `figure_id`     BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_figure.id',
  `manuscript_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_fig_ms_rel` (`figure_id`, `manuscript_id`),
  KEY `idx_paper_user_fig_ms_rel_ms` (`manuscript_id`, `figure_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表与论文章多对多';

CREATE TABLE IF NOT EXISTS `paper_user_figure_category` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类 ID',
  `name`       VARCHAR(128)    NOT NULL COMMENT '分类名称',
  `status`     VARCHAR(16)     NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_user_fig_cat_status_id` (`status`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表分类';

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
    'llm_topic_discovery',
    2,
    '选题 topic_discovery=running：生成包↔三门禁审计迭代，写入综述与实验方案',
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

INSERT INTO `paper_exec_intensity_topic` (`code`, `name`, `exec_rounds`, `max_papers`, `max_ideas`, `is_default`) VALUES
  ('fast',     '更快', 1, 4, 1, 0),
  ('balanced', '平衡', 2, 2, 1, 1),
  ('deep',     '更深', 3, 3, 1, 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `exec_rounds` = VALUES(`exec_rounds`),
  `max_papers` = VALUES(`max_papers`),
  `max_ideas` = VALUES(`max_ideas`),
  `is_default` = VALUES(`is_default`);

INSERT INTO `paper_exec_intensity_write_paper` (`code`, `name`, `exec_rounds`, `is_default`) VALUES
  ('fast',     '更快', 1, 0),
  ('balanced', '平衡', 2, 1),
  ('deep',     '更深', 3, 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `exec_rounds` = VALUES(`exec_rounds`),
  `is_default` = VALUES(`is_default`);

INSERT INTO `paper_llm_constraint_gate` (`prompt_scene`, `code`, `name`, `is_default`) VALUES
  ('topic', 'polished', '基础', 0),
  ('topic', 'standard', '标准', 1),
  ('topic', 'strict',   '严格', 0),
  ('write_paper', 'polished', '基础', 0),
  ('write_paper', 'standard', '标准', 1),
  ('write_paper', 'strict',   '严格', 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `is_default` = VALUES(`is_default`);

INSERT INTO `paper_ref_discipline` (`code`, `name`, `name_en`, `sort`) VALUES
  ('general', '跨学科通用', 'General & Interdisciplinary', 0),
  ('cs_ai', '人工智能', 'Artificial Intelligence', 10),
  ('cs_systems', '计算机', 'Computer Science', 11),
  ('cs_hci_robotics', '人机交互与机器人', 'HCI & Robotics', 12),
  ('engineering', '工程学', 'Engineering', 20),
  ('mathematics', '数学', 'Mathematics', 30),
  ('statistics', '统计学', 'Statistics', 35),
  ('physics', '物理学', 'Physics', 40),
  ('astronomy', '天文学', 'Astronomy & Astrophysics', 45),
  ('chemistry', '化学', 'Chemistry', 50),
  ('materials', '材料科学', 'Materials Science', 55),
  ('earth_environment', '地球与环境科学', 'Earth & Environmental Sciences', 60),
  ('atmospheric_science', '大气科学', 'Atmospheric Science', 61),
  ('ocean_science', '海洋科学', 'Oceanography', 62),
  ('geology', '地质学', 'Geology', 63),
  ('geophysics', '地球物理学', 'Geophysics', 64),
  ('agriculture', '农学', 'Agricultural Sciences', 65),
  ('biology', '生物学', 'Biological Sciences', 70),
  ('ecology', '生态学', 'Ecology', 71),
  ('basic_medicine', '基础医学', 'Basic Medical Sciences', 79),
  ('medicine', '临床医学', 'Clinical Medicine', 80),
  ('tcm', '中医学', 'Traditional Chinese Medicine', 81),
  ('integrative_medicine', '中西医结合', 'Integrative Chinese & Western Medicine', 82),
  ('stomatology', '口腔医学', 'Stomatology', 83),
  ('public_health', '公共卫生与预防医学', 'Public Health & Preventive Medicine', 85),
  ('nursing', '护理学', 'Nursing', 86),
  ('pharmacy', '药学', 'Pharmacy & Pharmacology', 88),
  ('chinese_pharmacy', '中药学', 'Chinese Pharmacology', 89),
  ('forensic_medicine', '法医学', 'Forensic Medicine', 90),
  ('neuroscience', '神经科学', 'Neuroscience', 91),
  ('psychology', '心理学', 'Psychology', 100),
  ('economics', '经济学', 'Economics', 110),
  ('finance', '金融学', 'Finance', 115),
  ('business', '工商管理', 'Business Administration', 120),
  ('management', '管理学', 'Management', 125),
  ('public_administration', '公共管理', 'Public Administration', 126),
  ('agro_forestry_economics', '农林经济管理', 'Agricultural & Forestry Economics', 127),
  ('information_resource_mgmt', '信息资源管理', 'Information Resource Management', 128),
  ('sociology', '社会学', 'Sociology', 130),
  ('ethnology', '民族学', 'Ethnology', 132),
  ('marxism_theory', '马克思主义理论', 'Marxist Theory', 133),
  ('party_history_governance', '中共党史党建学', 'Party History & Governance', 134),
  ('political_science', '政治学', 'Political Science', 135),
  ('law', '法学', 'Law', 140),
  ('education', '教育学', 'Education', 150),
  ('sports_science', '体育学', 'Sport Science', 151),
  ('geography', '地理学', 'Geography', 155),
  ('anthropology', '人类学', 'Anthropology', 160),
  ('communication', '新闻传播学', 'Journalism & Communication', 165),
  ('urban_architecture', '城乡规划与建筑学', 'Urban Planning & Architecture', 170),
  ('philosophy', '哲学', 'Philosophy', 180),
  ('archaeology', '考古学', 'Archaeology', 184),
  ('history', '历史学', 'History', 185),
  ('linguistics', '外国语言文学', 'Foreign Languages & Literature', 190),
  ('literature', '中国语言文学', 'Chinese Language & Literature', 195),
  ('arts_design', '艺术学', 'Arts', 200),
  ('religious_studies', '宗教研究', 'Religious Studies', 205),
  ('area_studies', '区域国别学', 'Area & Country Studies', 210),
  ('cn_stem', '中国理工', 'Chinese STEM', 300),
  ('cn_social_sci', '中国人文社科', 'Chinese Humanities & Social Sciences', 310),
  ('cn_medicine', '中国医学与生命科学', 'Chinese Medicine & Life Sciences', 320)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `name_en` = VALUES(`name_en`),
  `sort` = VALUES(`sort`);

INSERT INTO `paper_ref_venue` (`code`, `name`, `venue_type`, `strategy_profile_key`, `publisher`, `website_url`, `status`)
SELECT v.`code`, v.`name`, v.`venue_type`,
       CASE WHEN v.`profile_code` = 'ml_top3' THEN 'venue/ml_top3' ELSE 'venue/generic' END,
       v.`publisher`, v.`website_url`, 'active'
FROM (
  /* --- ML 顶会 --- */
  SELECT 'neurips' AS `code`, 'NeurIPS' AS `name`, 'conference' AS `venue_type`, 'ml_top3' AS `profile_code`, 'NeurIPS Foundation' AS `publisher`, 'https://neurips.cc' AS `website_url`
  UNION ALL SELECT 'iclr', 'ICLR', 'conference', 'ml_top3', 'OpenReview', 'https://iclr.cc'
  UNION ALL SELECT 'icml', 'ICML', 'conference', 'ml_top3', 'PMLR / ICML', 'https://icml.cc'
  UNION ALL SELECT 'colt', 'COLT', 'conference', 'ml_top3', 'PMLR', 'https://www.learningtheory.org'
  UNION ALL SELECT 'aistats', 'AISTATS', 'conference', 'ml_top3', 'PMLR', 'https://aistats.org'
  /* --- 视觉 --- */
  UNION ALL SELECT 'cvpr', 'CVPR', 'conference', 'vision_top', 'IEEE / CVF', 'https://cvpr.thecvf.com'
  UNION ALL SELECT 'iccv', 'ICCV', 'conference', 'vision_top', 'IEEE / CVF', 'https://iccv.thecvf.com'
  UNION ALL SELECT 'eccv', 'ECCV', 'conference', 'vision_top', 'Springer / ECCV', 'https://eccv.ecva.net'
  UNION ALL SELECT 'acm_mm', 'ACM Multimedia', 'conference', 'vision_top', 'ACM', 'https://www.acmmm.org'
  UNION ALL SELECT 'bmvc', 'BMVC', 'conference', 'vision_top', 'BMVA', 'https://bmvc-conference.com'
  UNION ALL SELECT 'wacv', 'WACV', 'conference', 'vision_top', 'IEEE / CVF', 'https://wacv.thecvf.com'
  UNION ALL SELECT 'miccai', 'MICCAI', 'conference', 'vision_top', 'MICCAI Society', 'https://miccai.org'
  /* --- NLP --- */
  UNION ALL SELECT 'acl', 'ACL', 'conference', 'nlp_top', 'ACL', 'https://aclanthology.org'
  UNION ALL SELECT 'emnlp', 'EMNLP', 'conference', 'nlp_top', 'ACL', 'https://aclanthology.org'
  UNION ALL SELECT 'naacl', 'NAACL', 'conference', 'nlp_top', 'ACL', 'https://aclanthology.org'
  UNION ALL SELECT 'coling', 'COLING', 'conference', 'nlp_top', 'ICCL', 'https://coling2024.org'
  UNION ALL SELECT 'eacl', 'EACL', 'conference', 'nlp_top', 'ACL', 'https://aclanthology.org'
  /* --- AI / 数据 --- */
  UNION ALL SELECT 'aaai', 'AAAI', 'conference', 'ai_data_top', 'AAAI', 'https://aaai.org'
  UNION ALL SELECT 'ijcai', 'IJCAI', 'conference', 'ai_data_top', 'IJCAI', 'https://www.ijcai.org'
  UNION ALL SELECT 'kdd', 'KDD', 'conference', 'ai_data_top', 'ACM', 'https://kdd.org'
  UNION ALL SELECT 'www', 'The Web Conference (WWW)', 'conference', 'ai_data_top', 'ACM / IW3C2', 'https://www2024.thewebconf.org'
  UNION ALL SELECT 'wsdm', 'WSDM', 'conference', 'ai_data_top', 'ACM', 'https://www.wsdm-conference.org'
  UNION ALL SELECT 'cikm', 'CIKM', 'conference', 'ai_data_top', 'ACM', 'https://cikm2024.org'
  UNION ALL SELECT 'sigir', 'SIGIR', 'conference', 'ai_data_top', 'ACM', 'https://sigir.org'
  UNION ALL SELECT 'recsys', 'RecSys', 'conference', 'ai_data_top', 'ACM', 'https://recsys.acm.org'
  UNION ALL SELECT 'icde', 'ICDE', 'conference', 'ai_data_top', 'IEEE', 'https://icde2024.github.io'
  UNION ALL SELECT 'vldb', 'VLDB', 'conference', 'ai_data_top', 'VLDB Endowment', 'https://vldb.org'
  /* --- 系统 --- */
  UNION ALL SELECT 'osdi', 'OSDI', 'conference', 'systems_top', 'USENIX', 'https://www.usenix.org/osdi'
  UNION ALL SELECT 'sosp', 'SOSP', 'conference', 'systems_top', 'ACM', 'https://sosp.org'
  UNION ALL SELECT 'nsdi', 'NSDI', 'conference', 'systems_top', 'USENIX', 'https://www.usenix.org/nsdi'
  UNION ALL SELECT 'eurosys', 'EuroSys', 'conference', 'systems_top', 'ACM', 'https://www.eurosys.org'
  UNION ALL SELECT 'asplos', 'ASPLOS', 'conference', 'systems_top', 'ACM', 'https://asplos-conference.org'
  UNION ALL SELECT 'sigcomm', 'SIGCOMM', 'conference', 'systems_top', 'ACM', 'https://sigcomm.org'
  UNION ALL SELECT 'mobicom', 'MobiCom', 'conference', 'systems_top', 'ACM', 'https://www.sigmobile.org/mobicom'
  UNION ALL SELECT 'usenix_atc', 'USENIX ATC', 'conference', 'systems_top', 'USENIX', 'https://www.usenix.org/atc'
  UNION ALL SELECT 'isca', 'ISCA', 'conference', 'systems_top', 'ACM / IEEE', 'https://iscaconf.org'
  UNION ALL SELECT 'micro', 'MICRO', 'conference', 'systems_top', 'ACM / IEEE', 'https://microarch.org'
  UNION ALL SELECT 'hpca', 'HPCA', 'conference', 'systems_top', 'IEEE', 'https://hpca-conf.org'
  /* --- 安全 --- */
  UNION ALL SELECT 'ieee_sp', 'IEEE S&P (Oakland)', 'conference', 'security_top', 'IEEE', 'https://www.ieee-security.org'
  UNION ALL SELECT 'usenix_security', 'USENIX Security', 'conference', 'security_top', 'USENIX', 'https://www.usenix.org/conference/usenixsecurity'
  UNION ALL SELECT 'ccs', 'ACM CCS', 'conference', 'security_top', 'ACM', 'https://www.sigsac.org/ccs.html'
  UNION ALL SELECT 'ndss', 'NDSS', 'conference', 'security_top', 'ISOC', 'https://www.ndss-symposium.org'
  /* --- HCI / 机器人 --- */
  UNION ALL SELECT 'chi', 'CHI', 'conference', 'hci_robotics_top', 'ACM', 'https://chi.acm.org'
  UNION ALL SELECT 'uist', 'UIST', 'conference', 'hci_robotics_top', 'ACM', 'https://uist.acm.org'
  UNION ALL SELECT 'icra', 'ICRA', 'conference', 'hci_robotics_top', 'IEEE', 'https://www.ieee-ras.org/conferences-workshops/fully-sponsored/icra'
  UNION ALL SELECT 'iros', 'IROS', 'conference', 'hci_robotics_top', 'IEEE / RSJ', 'https://ieee-ras.org/conferences-workshops/fully-sponsored/iros'
  UNION ALL SELECT 'corl', 'CoRL', 'conference', 'hci_robotics_top', 'CoRL', 'https://www.corl.org'
  UNION ALL SELECT 'rss', 'RSS', 'conference', 'hci_robotics_top', 'RSS Foundation', 'https://roboticsconference.org'
  /* --- 理论 --- */
  UNION ALL SELECT 'stoc', 'STOC', 'conference', 'theory_top', 'ACM', 'https://acm-stoc.org'
  UNION ALL SELECT 'focs', 'FOCS', 'conference', 'theory_top', 'IEEE', 'https://ieee-focs.org'
  UNION ALL SELECT 'soda', 'SODA', 'conference', 'theory_top', 'SIAM / ACM', 'https://www.siam.org/conferences/cm/conference/soda'
  /* --- CS 顶刊 --- */
  UNION ALL SELECT 'jmlr', 'JMLR', 'journal', 'cs_journal_top', 'JMLR', 'https://www.jmlr.org'
  UNION ALL SELECT 'tpami', 'IEEE TPAMI', 'journal', 'cs_journal_top', 'IEEE', 'https://ieeexplore.ieee.org/xpl/RecentIssue.jsp?punumber=34'
  UNION ALL SELECT 'ijcv', 'IJCV', 'journal', 'cs_journal_top', 'Springer', 'https://link.springer.com/journal/11263'
  UNION ALL SELECT 'tacl', 'TACL', 'journal', 'cs_journal_top', 'ACL', 'https://transacl.org'
  UNION ALL SELECT 'tmlr', 'TMLR', 'journal', 'cs_journal_top', 'OpenReview', 'https://jmlr.org/tmlr'
  UNION ALL SELECT 'cacm', 'Communications of the ACM', 'journal', 'cs_journal_top', 'ACM', 'https://cacm.acm.org'
  UNION ALL SELECT 'jacm', 'Journal of the ACM', 'journal', 'cs_journal_top', 'ACM', 'https://dl.acm.org/journal/jacm'
  UNION ALL SELECT 'tois', 'ACM TOIS', 'journal', 'cs_journal_top', 'ACM', 'https://dl.acm.org/journal/tois'
  /* --- 综合顶刊 --- */
  UNION ALL SELECT 'nature', 'Nature', 'journal', 'interdisciplinary_top', 'Nature Portfolio', 'https://www.nature.com/nature'
  UNION ALL SELECT 'science', 'Science', 'journal', 'interdisciplinary_top', 'AAAS', 'https://www.science.org/journal/science'
  UNION ALL SELECT 'pnas', 'PNAS', 'journal', 'interdisciplinary_top', 'NAS', 'https://www.pnas.org'
  UNION ALL SELECT 'cell', 'Cell', 'journal', 'interdisciplinary_top', 'Cell Press', 'https://www.cell.com/cell'
  UNION ALL SELECT 'nat_comm', 'Nature Communications', 'journal', 'interdisciplinary_top', 'Nature Portfolio', 'https://www.nature.com/ncomms'
  UNION ALL SELECT 'sci_adv', 'Science Advances', 'journal', 'interdisciplinary_top', 'AAAS', 'https://www.science.org/journal/sciadvances'
  /* --- 生命 / 医学 --- */
  UNION ALL SELECT 'nat_medicine', 'Nature Medicine', 'journal', 'life_med_top', 'Nature Portfolio', 'https://www.nature.com/nm'
  UNION ALL SELECT 'nat_biotech', 'Nature Biotechnology', 'journal', 'life_med_top', 'Nature Portfolio', 'https://www.nature.com/nbt'
  UNION ALL SELECT 'nat_methods', 'Nature Methods', 'journal', 'life_med_top', 'Nature Portfolio', 'https://www.nature.com/nmeth'
  UNION ALL SELECT 'nejm', 'New England Journal of Medicine', 'journal', 'life_med_top', 'NEJM Group', 'https://www.nejm.org'
  UNION ALL SELECT 'lancet', 'The Lancet', 'journal', 'life_med_top', 'Elsevier', 'https://www.thelancet.com'
  UNION ALL SELECT 'jama', 'JAMA', 'journal', 'life_med_top', 'AMA', 'https://jamanetwork.com/journals/jama'
  UNION ALL SELECT 'bmj', 'The BMJ', 'journal', 'life_med_top', 'BMJ', 'https://www.bmj.com'
  /* --- 经管 --- */
  UNION ALL SELECT 'aer', 'American Economic Review', 'journal', 'econ_mgmt_top', 'AEA', 'https://www.aeaweb.org/journals/aer'
  UNION ALL SELECT 'qje', 'Quarterly Journal of Economics', 'journal', 'econ_mgmt_top', 'Oxford UP', 'https://academic.oup.com/qje'
  UNION ALL SELECT 'jpe', 'Journal of Political Economy', 'journal', 'econ_mgmt_top', 'U Chicago Press', 'https://www.journals.uchicago.edu/journals/jpe'
  UNION ALL SELECT 'econometrica', 'Econometrica', 'journal', 'econ_mgmt_top', 'Econometric Society', 'https://www.econometricsociety.org/publications/econometrica'
  UNION ALL SELECT 'jfinance', 'Journal of Finance', 'journal', 'econ_mgmt_top', 'Wiley', 'https://afajof.org'
  UNION ALL SELECT 'jfe', 'Journal of Financial Economics', 'journal', 'econ_mgmt_top', 'Elsevier', 'https://www.jfinec.com'
  UNION ALL SELECT 'management_science', 'Management Science', 'journal', 'econ_mgmt_top', 'INFORMS', 'https://pubsonline.informs.org/journal/mnsc'
  UNION ALL SELECT 'amj', 'Academy of Management Journal', 'journal', 'econ_mgmt_top', 'AOM', 'https://journals.aom.org/journal/amj'
  /* --- 社会科学 SSCI --- */
  UNION ALL SELECT 'asr', 'American Sociological Review', 'journal', 'social_sci_top', 'SAGE', 'https://journals.sagepub.com/home/asr'
  UNION ALL SELECT 'apsr', 'American Political Science Review', 'journal', 'social_sci_top', 'Cambridge UP', 'https://www.cambridge.org/core/journals/american-political-science-review'
  UNION ALL SELECT 'psychological_science', 'Psychological Science', 'journal', 'social_sci_top', 'SAGE', 'https://journals.sagepub.com/home/pss'
  UNION ALL SELECT 'jpe_soc', 'Journal of Personality and Social Psychology', 'journal', 'social_sci_top', 'APA', 'https://www.apa.org/pubs/journals/psp'
  /* --- 物理 / 化学 --- */
  UNION ALL SELECT 'prl', 'Physical Review Letters', 'journal', 'physical_sci_top', 'APS', 'https://journals.aps.org/prl'
  UNION ALL SELECT 'prx', 'Physical Review X', 'journal', 'physical_sci_top', 'APS', 'https://journals.aps.org/prx'
  UNION ALL SELECT 'jacs', 'Journal of the American Chemical Society', 'journal', 'physical_sci_top', 'ACS', 'https://pubs.acs.org/journal/jacsat'
  UNION ALL SELECT 'angew_chem', 'Angewandte Chemie', 'journal', 'physical_sci_top', 'Wiley', 'https://onlinelibrary.wiley.com/journal/15213773'
  UNION ALL SELECT 'adv_materials', 'Advanced Materials', 'journal', 'physical_sci_top', 'Wiley', 'https://onlinelibrary.wiley.com/journal/15214095'
  /* --- 中文核心 --- */
  UNION ALL SELECT 'cn_sci_china', '中国科学', 'journal', 'cn_journal_top', '科学出版社', 'https://www.scichina.com'
  UNION ALL SELECT 'cn_sci_bulletin', '科学通报', 'journal', 'cn_journal_top', '科学出版社', 'https://www.scbulletin.com'
  UNION ALL SELECT 'cn_j_computers', '计算机学报', 'journal', 'cn_journal_top', '中国计算机学会', 'https://cjc.ict.ac.cn'
  UNION ALL SELECT 'cn_j_software', '软件学报', 'journal', 'cn_journal_top', '中国科学院软件所', 'https://www.jos.org.cn'
  UNION ALL SELECT 'cn_j_electronics', '电子学报', 'journal', 'cn_journal_top', '中国电子学会', 'https://www.ejournal.org.cn'
  UNION ALL SELECT 'cn_j_automation', '自动化学报', 'journal', 'cn_journal_top', '中国自动化学会', 'https://www.aas.net.cn'
  /* --- 通用 / 开放 --- */
  UNION ALL SELECT 'plos_one', 'PLOS ONE', 'journal', '', 'PLOS', 'https://journals.plos.org/plosone'
  UNION ALL SELECT 'ieee_access', 'IEEE Access', 'journal', '', 'IEEE', 'https://ieeeaccess.ieee.org'
  UNION ALL SELECT 'arxiv_only', 'arXiv 预印本（无正式 venue）', 'other', '', NULL, 'https://arxiv.org'
) v
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `venue_type` = VALUES(`venue_type`),
  `strategy_profile_key` = VALUES(`strategy_profile_key`),
  `publisher` = VALUES(`publisher`),
  `website_url` = VALUES(`website_url`),
  `status` = VALUES(`status`);

UPDATE `paper_ref_venue` SET `sort` = `id`;

INSERT INTO `paper_literature_platform` (`code`, `name`, `api_kind`, `base_url`, `auth_type`, `default_config`, `config_schema`, `sort`, `default_selected`) VALUES
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
  `sort` = VALUES(`sort`),
  `default_selected` = VALUES(`default_selected`);

-- ---------------------------------------------------------------------------
-- 11. 种子 — LLM（model 首次插入后请 UPDATE api_key；重复跑请跳过 11.1 或自行去重）
-- ---------------------------------------------------------------------------

-- 11.1 Claude 端点（勿重复 INSERT 多条相同 label，除非手动清理）
INSERT INTO `paper_llm_model_config` (
  `label`, `scope`, `provider_code`, `model_name`, `api_base_url`, `api_key`,
  `timeout_ms`, `max_retries`, `supports_vision`, `context_window_hint`, `status`
) VALUES (
  'Claude Sonnet（平台默认）',
  'platform_delegate',
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

INSERT INTO `paper_llm_model_config` (
  `label`, `scope`, `provider_code`, `model_name`, `api_base_url`, `api_key`,
  `timeout_ms`, `max_retries`, `supports_vision`, `context_window_hint`, `extra`, `status`
) VALUES (
  'Codex GPT-5.4（审计 reviewer）',
  'platform_delegate',
  'openai',
  'gpt-5.4',
  'https://api.openai.com',
  NULL,
  3600000,
  1,
  0,
  200000,
  JSON_OBJECT('executor', 'codex', 'notes', '本地 codex / MCP research-review'),
  'active'
);

INSERT INTO `paper_llm_strategy_package` (`code`, `name`, `status`)
VALUES
  ('aris', 'ARIS 本地 CLI', 'active')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `status` = VALUES(`status`);

SET @pkg_aris := (SELECT `id` FROM `paper_llm_strategy_package` WHERE `code` = 'aris' LIMIT 1);
SET @paper_llm_claude_id := (
  SELECT c.`id` FROM `paper_llm_model_config` c
  WHERE c.`status` = 'active' AND c.`provider_code` = 'anthropic'
  ORDER BY c.`id` DESC
  LIMIT 1
);
SET @paper_llm_codex_id := (
  SELECT c.`id` FROM `paper_llm_model_config` c
  WHERE c.`status` = 'active' AND c.`provider_code` = 'openai' AND c.`model_name` = 'gpt-5.4'
  ORDER BY c.`id` DESC
  LIMIT 1
);

INSERT INTO `paper_llm_strategy_route` (
  `behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`
) VALUES
  ('topic.stage.topic_generate', 'topic_discovery', 'topic_generate', @pkg_aris, 'idea-creator', IFNULL(@paper_llm_claude_id, 0), 'active'),
  ('topic.stage.topic_audit', 'topic_discovery', 'topic_audit', @pkg_aris, 'research-review', IFNULL(@paper_llm_codex_id, IFNULL(@paper_llm_claude_id, 0)), 'active'),
  ('topic.stage.experiment_plan', 'experiment_plan', 'experiment_plan', @pkg_aris, 'experiment-bridge', IFNULL(@paper_llm_model_id, 0), 'active'),
  ('topic.discovery.run', 'topic_discovery', 'run', @pkg_aris, 'idea-discovery', IFNULL(@paper_llm_model_id, 0), 'active')
ON DUPLICATE KEY UPDATE
  `strategy_package_id` = VALUES(`strategy_package_id`),
  `capability_id` = VALUES(`capability_id`),
  `default_model_config_id` = VALUES(`default_model_config_id`),
  `status` = VALUES(`status`);

-- ---------------------------------------------------------------------------
-- 升级脚本（已有库执行一次即可）
-- ---------------------------------------------------------------------------
-- ALTER TABLE `paper_exec_intensity_topic`
--   ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1' AFTER `max_ideas`;
-- ALTER TABLE `paper_llm_constraint_audit`
--   ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1' AFTER `audit_rounds`;
-- UPDATE `paper_exec_intensity_topic` SET `is_default` = 0;
-- UPDATE `paper_exec_intensity_topic` SET `is_default` = 1 WHERE `code` = 'balanced';
-- UPDATE `paper_llm_constraint_audit` SET `is_default` = 0;
-- UPDATE `paper_llm_constraint_audit` SET `is_default` = 1 WHERE `code` = 'polished';
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
