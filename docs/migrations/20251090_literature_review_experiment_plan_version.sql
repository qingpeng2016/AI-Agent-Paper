-- 文献综述 / 实验方案：版本号 version（按 manuscript_id 递增，前端展示 vN）
-- 新库见 docs/all.sql；本脚本用于旧库补列 + 历史数据回填（可重复执行）

SET @db := DATABASE();

-- ---------------------------------------------------------------------------
-- 1. paper_output_literature_review.version
-- ---------------------------------------------------------------------------
SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'paper_output_literature_review'
    AND COLUMN_NAME = 'version'
);
SET @sql := IF(
  @col_exists = 0,
  'ALTER TABLE `paper_output_literature_review`
     ADD COLUMN `version` INT NOT NULL DEFAULT 1 COMMENT ''版本号（同一 manuscript 内递增）'' AFTER `user_id`',
  'SELECT ''skip: paper_output_literature_review.version exists'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `paper_output_literature_review` r
INNER JOIN (
  SELECT
    `id`,
    ROW_NUMBER() OVER (
      PARTITION BY `manuscript_id`
      ORDER BY `created_at`, `id`
    ) AS `vn`
  FROM `paper_output_literature_review`
) t ON t.`id` = r.`id`
SET r.`version` = t.`vn`
WHERE r.`version` IS NULL OR r.`version` < 1;

-- 同一论文下若曾全部为 DEFAULT 1，按创建顺序重排为 1..N
UPDATE `paper_output_literature_review` r
INNER JOIN (
  SELECT
    `id`,
    ROW_NUMBER() OVER (
      PARTITION BY `manuscript_id`
      ORDER BY `created_at`, `id`
    ) AS `vn`
  FROM `paper_output_literature_review`
  WHERE `status` <> 'deleted'
) t ON t.`id` = r.`id`
SET r.`version` = t.`vn`;

-- ---------------------------------------------------------------------------
-- 2. paper_output_experiment_plan.version + is_current（实验方案原有设计）
-- ---------------------------------------------------------------------------
SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'paper_output_experiment_plan'
    AND COLUMN_NAME = 'version'
);
SET @sql := IF(
  @col_exists = 0,
  'ALTER TABLE `paper_output_experiment_plan`
     ADD COLUMN `version` INT NOT NULL DEFAULT 1 COMMENT ''版本号（同一 manuscript 内递增）'' AFTER `user_id`',
  'SELECT ''skip: paper_output_experiment_plan.version exists'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'paper_output_experiment_plan'
    AND COLUMN_NAME = 'is_current'
);
SET @sql := IF(
  @col_exists = 0,
  'ALTER TABLE `paper_output_experiment_plan`
     ADD COLUMN `is_current` TINYINT(1) NOT NULL DEFAULT 1 COMMENT ''是否当前版本'' AFTER `version`',
  'SELECT ''skip: paper_output_experiment_plan.is_current exists'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `paper_output_experiment_plan` p
INNER JOIN (
  SELECT
    `id`,
    ROW_NUMBER() OVER (
      PARTITION BY `manuscript_id`
      ORDER BY `created_at`, `id`
    ) AS `vn`
  FROM `paper_output_experiment_plan`
  WHERE `status` <> 'deleted'
) t ON t.`id` = p.`id`
SET p.`version` = t.`vn`;

-- 每篇论文仅最新一条 is_current = 1（按 created_at, id）
UPDATE `paper_output_experiment_plan`
SET `is_current` = 0
WHERE `status` <> 'deleted';

UPDATE `paper_output_experiment_plan` p
INNER JOIN (
  SELECT `id`
  FROM (
    SELECT
      `id`,
      ROW_NUMBER() OVER (
        PARTITION BY `manuscript_id`
        ORDER BY `created_at` DESC, `id` DESC
      ) AS `rn`
    FROM `paper_output_experiment_plan`
    WHERE `status` <> 'deleted'
  ) x
  WHERE x.`rn` = 1
) cur ON cur.`id` = p.`id`
SET p.`is_current` = 1;

-- 可选索引（不存在则加；若已存在会报错，可忽略或手工跳过）
-- ALTER TABLE `paper_output_literature_review`
--   ADD KEY `idx_paper_lit_review_ms_version` (`manuscript_id`, `version`);
-- ALTER TABLE `paper_output_experiment_plan`
--   ADD KEY `idx_paper_exp_plan_ms_version` (`manuscript_id`, `version`);
