-- 策略路由：平台为调用方；prompt/门禁文案归策略包（黑盒）
-- 移除 paper_llm_description、paper_llm_template_topic；瘦身 constraint_gate / ref_venue

-- ---------------------------------------------------------------------------
-- 1. 策略包注册
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `paper_llm_strategy_package` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '策略包 ID',
  `code`              VARCHAR(64)  NOT NULL COMMENT '包键：builtin|aris|…',
  `name`              VARCHAR(128) NOT NULL COMMENT '展示名',
  `provider_kind`     VARCHAR(16)  NOT NULL DEFAULT 'inline' COMMENT 'inline|http|process',
  `base_url`          VARCHAR(512) DEFAULT NULL COMMENT 'http 策略服务 Origin',
  `auth_type`         VARCHAR(32)  NOT NULL DEFAULT 'none' COMMENT 'none|bearer|hmac',
  `auth_secret_ref`   VARCHAR(256) DEFAULT NULL COMMENT '平台密钥引用，非明文',
  `manifest_version`  VARCHAR(32)  DEFAULT NULL COMMENT '能力清单版本',
  `capabilities`      JSON         DEFAULT NULL COMMENT '能力 id 列表与 schema 版本',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_llm_strategy_package_code` (`code`),
  KEY `idx_paper_llm_strategy_package_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 策略包（独立服务/内联）';

-- ---------------------------------------------------------------------------
-- 2. 行为 → 策略包路由
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `paper_llm_strategy_route` (
  `id`                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `behavior_key`            VARCHAR(128) NOT NULL COMMENT '如 topic.stage.topic_generate',
  `module`                  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'topic_discovery|write_paper|…',
  `action`                  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'stage 或子动作',
  `strategy_package_id`     BIGINT UNSIGNED NOT NULL COMMENT 'paper_llm_strategy_package.id',
  `capability_id`           VARCHAR(128) NOT NULL DEFAULT '' COMMENT '包内能力/skill id',
  `default_model_config_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=包内默认；非 0=平台代发 LLM',
  `fallback_route_id`       BIGINT UNSIGNED DEFAULT NULL COMMENT '失败时备用 route.id',
  `config_json`             JSON         DEFAULT NULL COMMENT '静态参数',
  `priority`                INT          NOT NULL DEFAULT 0,
  `status`                  VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`              DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_llm_strategy_route_behavior` (`behavior_key`),
  KEY `idx_paper_llm_strategy_route_pkg` (`strategy_package_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='需求行为 → 策略包';

INSERT INTO `paper_llm_strategy_package` (`code`, `name`, `provider_kind`, `base_url`, `auth_type`, `manifest_version`, `capabilities`, `status`)
VALUES
  ('builtin', '内置（过渡内联）', 'inline', NULL, 'none', '1', JSON_ARRAY('topic.stage.default', 'topic.stage.topic_generate', 'topic.stage.topic_audit', 'topic.stage.experiment_plan'), 'active'),
  ('aris', 'ARIS 策略服务', 'http', NULL, 'bearer', NULL, NULL, 'disabled')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `provider_kind` = VALUES(`provider_kind`),
  `status` = VALUES(`status`);

SET @pkg_builtin := (SELECT `id` FROM `paper_llm_strategy_package` WHERE `code` = 'builtin' LIMIT 1);
SET @model_default := (
  SELECT c.`id` FROM `paper_llm_model_config` c
  WHERE c.`status` = 'active'
  ORDER BY (c.`provider_code` = 'anthropic') DESC, c.`id` DESC
  LIMIT 1
);

INSERT INTO `paper_llm_strategy_route` (
  `behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`
) VALUES
  ('topic.stage.default', 'topic_discovery', 'default', @pkg_builtin, 'topic.stage.default', IFNULL(@model_default, 0), 'active'),
  ('topic.stage.topic_generate', 'topic_discovery', 'topic_generate', @pkg_builtin, 'topic.stage.topic_generate', IFNULL(@model_default, 0), 'active'),
  ('topic.stage.topic_audit', 'topic_discovery', 'topic_audit', @pkg_builtin, 'topic.stage.topic_audit', IFNULL(@model_default, 0), 'active'),
  ('topic.stage.experiment_plan', 'experiment_plan', 'experiment_plan', @pkg_builtin, 'topic.stage.experiment_plan', IFNULL(@model_default, 0), 'active'),
  ('topic.discovery.run', 'topic_discovery', 'run', @pkg_builtin, 'topic.discovery.run', IFNULL(@model_default, 0), 'active')
ON DUPLICATE KEY UPDATE
  `strategy_package_id` = VALUES(`strategy_package_id`),
  `capability_id` = VALUES(`capability_id`),
  `default_model_config_id` = VALUES(`default_model_config_id`),
  `status` = VALUES(`status`);

-- ---------------------------------------------------------------------------
-- 3. paper_ref_venue：策略 profile 键（ opaque，由策略包解释）
-- ---------------------------------------------------------------------------
SET @col_spk := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'paper_ref_venue' AND COLUMN_NAME = 'strategy_profile_key'
);
SET @sql := IF(
  @col_spk = 0,
  'ALTER TABLE `paper_ref_venue`
     ADD COLUMN `strategy_profile_key` VARCHAR(64) NOT NULL DEFAULT ''venue/generic''
       COMMENT ''策略包 venue profile，如 venue/ml_top3''
       AFTER `venue_type`',
  'SELECT ''skip: strategy_profile_key exists'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

UPDATE `paper_ref_venue`
SET `strategy_profile_key` = 'venue/ml_top3'
WHERE `llm_description_codes` IS NOT NULL
  AND JSON_CONTAINS(`llm_description_codes`, CAST(20003 AS JSON), '$');

SET @col_ldc := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'paper_ref_venue' AND COLUMN_NAME = 'llm_description_codes'
);
SET @sql := IF(
  @col_ldc > 0,
  'ALTER TABLE `paper_ref_venue` DROP COLUMN `llm_description_codes`',
  'SELECT ''skip: llm_description_codes dropped'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------------------------------------------------------------------------
-- 4. paper_llm_constraint_gate：仅产品档位，不再指向 description
-- ---------------------------------------------------------------------------
SET @col_cit := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'paper_llm_constraint_gate' AND COLUMN_NAME = 'llm_description_citation_check_code'
);
SET @sql := IF(
  @col_cit > 0,
  'ALTER TABLE `paper_llm_constraint_gate`
     DROP COLUMN `llm_description_citation_check_code`,
     DROP COLUMN `llm_description_claim_check_code`,
     DROP COLUMN `llm_description_kill_check_code`',
  'SELECT ''skip: constraint_gate columns dropped'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------------------------------------------------------------------------
-- 5. call log：策略溯源
-- ---------------------------------------------------------------------------
SET @col_bk := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'paper_llm_call_logs' AND COLUMN_NAME = 'behavior_key'
);
SET @sql := IF(
  @col_bk = 0,
  'ALTER TABLE `paper_llm_call_logs`
     ADD COLUMN `behavior_key` VARCHAR(128) DEFAULT NULL COMMENT ''paper_llm_strategy_route.behavior_key'' AFTER `stage_code`,
     ADD COLUMN `strategy_package_id` BIGINT UNSIGNED DEFAULT NULL AFTER `behavior_key`,
     ADD COLUMN `capability_id` VARCHAR(128) DEFAULT NULL AFTER `strategy_package_id`,
     ADD COLUMN `external_trace_id` VARCHAR(128) DEFAULT NULL AFTER `capability_id`',
  'SELECT ''skip: call log strategy cols exist'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_pti := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'paper_llm_call_logs' AND COLUMN_NAME = 'prompt_template_id'
);
SET @sql := IF(
  @col_pti > 0,
  'ALTER TABLE `paper_llm_call_logs` DROP COLUMN `prompt_template_id`',
  'SELECT ''skip: prompt_template_id dropped'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------------------------------------------------------------------------
-- 6. paper_llm_model_config：可选 scope（平台代发）
-- ---------------------------------------------------------------------------
SET @col_scope := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'paper_llm_model_config' AND COLUMN_NAME = 'scope'
);
SET @sql := IF(
  @col_scope = 0,
  'ALTER TABLE `paper_llm_model_config`
     ADD COLUMN `scope` VARCHAR(32) NOT NULL DEFAULT ''platform_delegate''
       COMMENT ''platform_delegate|strategy_only''
       AFTER `label`',
  'SELECT ''skip: scope exists'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------------------------------------------------------------------------
-- 7. 删除 prompt 库表
-- ---------------------------------------------------------------------------
DROP TABLE IF EXISTS `paper_llm_template_topic`;
DROP TABLE IF EXISTS `paper_llm_description`;
