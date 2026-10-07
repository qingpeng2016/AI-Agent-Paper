-- Paper Agent 全量 schema（唯一 migration；新库执行本文件即可）
-- 多学科 / 多 venue / 多文献源；模块见 agent/paper/types.ts
-- 引擎：MySQL 8.0+，utf8mb4；表前缀 paper_
-- user_id 与商城 users.id 对齐，不设 FK 便于独立部署
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
-- 1. ref 字典：学科、venue、文献源、强度与审计档位（七模块 code 见前端 agent/paper/types.ts）
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
  `user_id`              BIGINT UNSIGNED NOT NULL COMMENT 'users.id；0=平台预置模板',
  `label`                VARCHAR(128) NOT NULL COMMENT '展示名',
  `provider_code`        VARCHAR(32)  NOT NULL COMMENT 'openai|anthropic|azure_openai|openai_compatible|ollama|gateway|other',
  `model_name`           VARCHAR(128) NOT NULL COMMENT '上游 model 参数',
  `api_base_url`         VARCHAR(512) NOT NULL COMMENT 'Base URL',
  `api_path_chat`        VARCHAR(128) DEFAULT NULL COMMENT '如 /chat/completions；NULL=Provider 默认',
  `api_key_ciphertext`   VARBINARY(4096) DEFAULT NULL COMMENT '本地 Key 密文；可与 user_api_key_id 并存',
  `api_key_header`       VARCHAR(64)  NOT NULL DEFAULT 'Authorization',
  `api_key_prefix`       VARCHAR(32)  DEFAULT 'Bearer ',
  `user_api_key_id`      BIGINT UNSIGNED DEFAULT NULL COMMENT '商城 user_api_keys.id',
  `default_headers`      JSON         DEFAULT NULL,
  `default_params`       JSON         DEFAULT NULL COMMENT 'temperature、max_tokens 等',
  `timeout_ms`           INT UNSIGNED NOT NULL DEFAULT 120000,
  `max_retries`          TINYINT UNSIGNED NOT NULL DEFAULT 2,
  `supports_vision`      TINYINT(1)   NOT NULL DEFAULT 0,
  `context_window_hint`  INT UNSIGNED DEFAULT NULL,
  `extra`                JSON         DEFAULT NULL,
  `status`               VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_model_cfg_user` (`user_id`, `status`),
  KEY `idx_paper_llm_model_cfg_provider` (`provider_code`, `model_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='LLM 端点配置（Key/Host/模型名）';

CREATE TABLE IF NOT EXISTS `paper_llm_workflow_binding` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=平台默认；否则用户覆盖',
  `manuscript_id`     BIGINT UNSIGNED DEFAULT NULL COMMENT '非空=仅本篇 override',
  `module_code`       VARCHAR(32)  NOT NULL COMMENT
    'topic_discovery|literature_review|experiment_planning|auto_review|paper_writing|figure_generation|manuscript_analysis|*',
  `stage_code`        VARCHAR(64)  DEFAULT NULL COMMENT
    'topic: retrieve|generate_ideas|novelty|audit；NULL=整模块默认',
  `llm_role`          VARCHAR(32)  NOT NULL COMMENT
    'executor|reviewer|ingest|system|citation_audit|claim_audit|kill_argument',
  `intensity_code`    VARCHAR(16)  DEFAULT NULL COMMENT 'fast|balanced|deep；NULL=全部',
  `audit_level_code`  VARCHAR(16)  DEFAULT NULL COMMENT 'standard|polished|strict；NULL=全部',
  `model_config_id`   BIGINT UNSIGNED NOT NULL,
  `priority`          INT          NOT NULL DEFAULT 100 COMMENT '同键多条时越小越优先',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `note`              VARCHAR(256) DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_bind_lookup` (`user_id`, `module_code`, `stage_code`, `llm_role`, `status`, `priority`),
  KEY `idx_paper_llm_bind_ms` (`manuscript_id`, `module_code`),
  KEY `idx_paper_llm_bind_model` (`model_config_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环节/阶段/角色 → LLM 配置';

CREATE TABLE IF NOT EXISTS `paper_llm_prompt_template` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`           BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=平台默认；否则用户覆盖',
  `manuscript_id`     BIGINT UNSIGNED DEFAULT NULL COMMENT '非空=仅本篇 override',
  `module_code`       VARCHAR(32)  NOT NULL COMMENT
    'topic_discovery|literature_review|experiment_planning|auto_review|paper_writing|figure_generation|manuscript_analysis|*',
  `stage_code`        VARCHAR(64)  DEFAULT NULL COMMENT
    'topic: retrieve|generate_ideas|novelty|audit；NULL=整模块默认',
  `llm_role`          VARCHAR(32)  NOT NULL COMMENT
    'executor|reviewer|ingest|system|citation_audit|claim_audit|kill_argument',
  `intensity_code`    VARCHAR(16)  DEFAULT NULL COMMENT 'fast|balanced|deep；NULL=全部',
  `audit_level_code`  VARCHAR(16)  DEFAULT NULL COMMENT 'standard|polished|strict；NULL=全部',
  `message_role`      VARCHAR(16)  NOT NULL DEFAULT 'system' COMMENT 'system|user|assistant（拼 chat messages）',
  `label`             VARCHAR(128) DEFAULT NULL COMMENT '运营展示名',
  `version`           INT          NOT NULL DEFAULT 1 COMMENT '同键多版时取最大 active 或按 priority',
  `template_body`     MEDIUMTEXT   NOT NULL COMMENT '话术正文；占位符 {{var_name}}，由代码渲染后请求模型',
  `variables`         JSON         DEFAULT NULL COMMENT
    '[{"name":"direction","required":true,"description":"…","source_hint":"input_params.direction"}]',
  `priority`          INT          NOT NULL DEFAULT 100 COMMENT '同键多条时越小越优先',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `note`              VARCHAR(256) DEFAULT NULL,
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_paper_llm_prompt_lookup` (
    `user_id`,
    `module_code`,
    `stage_code`,
    `llm_role`,
    `message_role`,
    `status`,
    `priority`
  ),
  KEY `idx_paper_llm_prompt_ms` (`manuscript_id`, `module_code`, `stage_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环节/阶段/角色 → 话术模板（与 binding 同维解析）';

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
  `from_module_code`    VARCHAR(32)  NOT NULL DEFAULT 'topic_discovery' COMMENT '哪次模块检索写入',
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
  `module_code`         VARCHAR(32)  DEFAULT NULL COMMENT '校验发生时的模块',
  `target_module`       VARCHAR(32)  DEFAULT NULL COMMENT '产出所属模块，如 paper_writing',
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
  `module_code`     VARCHAR(32)  NOT NULL COMMENT 'topic_discovery|literature_review|…',
  `stage_code`      VARCHAR(64)  DEFAULT NULL COMMENT '模块内阶段，可选',
  `model_config_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_model_config.id',
  `workflow_binding_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_workflow_binding.id',
  `prompt_template_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'paper_llm_prompt_template.id',
  `role`            VARCHAR(16)  NOT NULL COMMENT 'executor|reviewer',
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
  `module_code`        VARCHAR(32)  NOT NULL COMMENT '工作流 module_code 或 system|environment',
  `module_label`       VARCHAR(64)  NOT NULL,
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
-- 9. 种子数据（强度、审计、示例学科与 venue、文献源）
-- ---------------------------------------------------------------------------
-- paper_llm_workflow_binding / paper_llm_prompt_template 不在此 INSERT；接入后写入 platform default（user_id=0）


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

INSERT INTO `paper_ref_literature_source` (`code`, `name`, `api_kind`, `base_url`, `auth_type`) VALUES
  ('arxiv', 'arXiv', 'rest', 'https://export.arxiv.org/api/query', 'none'),
  ('openalex', 'OpenAlex', 'rest', 'https://api.openalex.org', 'none'),
  ('semantic_scholar', 'Semantic Scholar', 'rest', 'https://api.semanticscholar.org/graph/v1', 'api_key')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

SET FOREIGN_KEY_CHECKS = 1;
