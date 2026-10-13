-- 若曾创建 paper_manuscript_category_rel，重命名为 paper_user_literature_manuscript_rel
SET @has_old := (
  SELECT COUNT(*)
  FROM information_schema.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'paper_manuscript_category_rel'
);
SET @has_new := (
  SELECT COUNT(*)
  FROM information_schema.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'paper_user_literature_manuscript_rel'
);

SET @sql_rename = IF(
  @has_old > 0 AND @has_new = 0,
  'RENAME TABLE `paper_manuscript_category_rel` TO `paper_user_literature_manuscript_rel`',
  'SELECT 1'
);
PREPARE stmt_rename FROM @sql_rename;
EXECUTE stmt_rename;
DEALLOCATE PREPARE stmt_rename;
