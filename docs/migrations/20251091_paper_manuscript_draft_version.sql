-- 论文写作稿版本号（每篇 paper_manuscript 独立递增；前端「当前论文 · 标题 版本号 vN」）
SET @db := DATABASE();

SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'paper_manuscript'
    AND COLUMN_NAME = 'draft_version'
);
SET @after_draft_updated := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @db
    AND TABLE_NAME = 'paper_manuscript'
    AND COLUMN_NAME = 'draft_updated_at'
);
SET @sql := IF(
  @col_exists = 0,
  IF(
    @after_draft_updated > 0,
    'ALTER TABLE `paper_manuscript`
       ADD COLUMN `draft_version` INT UNSIGNED NOT NULL DEFAULT 0
         COMMENT ''论文写作稿版本（每次成功生成整稿 +1；0=尚未生成）''
         AFTER `draft_updated_at`',
    'ALTER TABLE `paper_manuscript`
       ADD COLUMN `draft_version` INT UNSIGNED NOT NULL DEFAULT 0
         COMMENT ''论文写作稿版本（每次成功生成整稿 +1；0=尚未生成）''
         AFTER `is_current`'
  ),
  'SELECT ''skip: paper_manuscript.draft_version exists'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
