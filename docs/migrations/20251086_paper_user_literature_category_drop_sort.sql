-- paper_user_literature_category：移除 sort（若列已不存在则跳过）

SET @has_sort_col := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'paper_user_literature_category'
    AND COLUMN_NAME = 'sort'
);

SET @has_sort_idx := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'paper_user_literature_category'
    AND INDEX_NAME = 'idx_paper_user_lit_cat_status_sort'
);

SET @sql_drop_idx = IF(
  @has_sort_idx > 0,
  'ALTER TABLE `paper_user_literature_category` DROP KEY `idx_paper_user_lit_cat_status_sort`',
  'SELECT 1'
);
PREPARE stmt_drop_idx FROM @sql_drop_idx;
EXECUTE stmt_drop_idx;
DEALLOCATE PREPARE stmt_drop_idx;

SET @sql_drop_col = IF(
  @has_sort_col > 0,
  'ALTER TABLE `paper_user_literature_category` DROP COLUMN `sort`',
  'SELECT 1'
);
PREPARE stmt_drop_col FROM @sql_drop_col;
EXECUTE stmt_drop_col;
DEALLOCATE PREPARE stmt_drop_col;

SET @has_status_id_idx := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'paper_user_literature_category'
    AND INDEX_NAME = 'idx_paper_user_lit_cat_status_id'
);

SET @sql_add_idx = IF(
  @has_status_id_idx = 0,
  'ALTER TABLE `paper_user_literature_category` ADD KEY `idx_paper_user_lit_cat_status_id` (`status`, `id`)',
  'SELECT 1'
);
PREPARE stmt_add_idx FROM @sql_add_idx;
EXECUTE stmt_add_idx;
DEALLOCATE PREPARE stmt_add_idx;
