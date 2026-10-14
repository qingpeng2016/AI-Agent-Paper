-- 我的论文：draft_version（若缺）+ 定稿状态 + 投递状态
SET @db := DATABASE();

-- 1) draft_version（未跑 20251091 或库无 draft_updated_at 时在此补齐）
SET @dv_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'paper_manuscript'
    AND COLUMN_NAME = 'draft_version'
);
SET @sql := IF(
  @dv_exists = 0,
  'ALTER TABLE `paper_manuscript`
     ADD COLUMN `draft_version` INT UNSIGNED NOT NULL DEFAULT 0
       COMMENT ''论文写作稿版本（每次成功生成整稿 +1；0=尚未生成）''
       AFTER `is_current`',
  'SELECT ''skip: paper_manuscript.draft_version exists'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2) 定稿 / 投递汇总字段
SET @ds_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'paper_manuscript'
    AND COLUMN_NAME = 'draft_status'
);
SET @sql := IF(
  @ds_exists = 0,
  'ALTER TABLE `paper_manuscript`
     ADD COLUMN `draft_status` VARCHAR(16) NOT NULL DEFAULT ''writing''
       COMMENT ''定稿：writing|ready|frozen''
       AFTER `draft_version`,
     ADD COLUMN `submission_status` VARCHAR(24) NOT NULL DEFAULT ''not_submitted''
       COMMENT ''投递汇总：not_submitted|submitted|under_review|revision|rebuttal|resubmitted|accepted|rejected''
       AFTER `draft_status`,
     ADD COLUMN `submission_label` VARCHAR(128) DEFAULT NULL
       COMMENT ''投递展示，如 NeurIPS 2026 · 审稿中''
       AFTER `submission_status`,
     ADD COLUMN `submission_deadline` DATETIME(3) DEFAULT NULL
       COMMENT ''最近投递相关截止（修回/rebuttal 等）''
       AFTER `submission_label`',
  'SELECT ''skip: paper_manuscript draft/submission status exists'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
