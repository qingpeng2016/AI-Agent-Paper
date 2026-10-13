-- 文献可属于多个分类：多对多关联表。
-- 若曾执行 20251084（paper_user_literature.category_id），会自动迁入 rel 并删旧列/索引；否则只建 rel 表。

CREATE TABLE IF NOT EXISTS `paper_user_literature_category_rel` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `user_literature_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_literature.id',
  `category_id`        BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_literature_category.id',
  `created_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_lit_cat_rel` (`user_literature_id`, `category_id`),
  KEY `idx_paper_user_lit_cat_rel_cat` (`category_id`, `user_literature_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文献与分类多对多';

SET @has_category_col := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'paper_user_literature'
    AND COLUMN_NAME = 'category_id'
);

SET @sql_migrate = IF(
  @has_category_col > 0,
  'INSERT IGNORE INTO `paper_user_literature_category_rel` (`user_literature_id`, `category_id`)
   SELECT `id`, `category_id` FROM `paper_user_literature`
   WHERE `category_id` IS NOT NULL AND `status` = ''active''',
  'SELECT 1'
);
PREPARE stmt_migrate FROM @sql_migrate;
EXECUTE stmt_migrate;
DEALLOCATE PREPARE stmt_migrate;

SET @has_category_idx := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'paper_user_literature'
    AND INDEX_NAME = 'idx_paper_user_lit_category'
);

SET @sql_drop_idx = IF(
  @has_category_idx > 0,
  'ALTER TABLE `paper_user_literature` DROP KEY `idx_paper_user_lit_category`',
  'SELECT 1'
);
PREPARE stmt_drop_idx FROM @sql_drop_idx;
EXECUTE stmt_drop_idx;
DEALLOCATE PREPARE stmt_drop_idx;

SET @sql_drop_col = IF(
  @has_category_col > 0,
  'ALTER TABLE `paper_user_literature` DROP COLUMN `category_id`',
  'SELECT 1'
);
PREPARE stmt_drop_col FROM @sql_drop_col;
EXECUTE stmt_drop_col;
DEALLOCATE PREPARE stmt_drop_col;
