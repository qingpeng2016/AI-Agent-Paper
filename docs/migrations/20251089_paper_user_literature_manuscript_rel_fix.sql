-- 修正 paper_user_literature_manuscript_rel：文献与论文多对多（非分类）
SET @has_wrong_col := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'paper_user_literature_manuscript_rel'
    AND COLUMN_NAME = 'category_id'
);

SET @drop_sql := IF(
  @has_wrong_col > 0,
  'DROP TABLE `paper_user_literature_manuscript_rel`',
  'SELECT 1'
);
PREPARE stmt_drop FROM @drop_sql;
EXECUTE stmt_drop;
DEALLOCATE PREPARE stmt_drop;

CREATE TABLE IF NOT EXISTS `paper_user_literature_manuscript_rel` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `user_literature_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_literature.id',
  `manuscript_id`      BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `created_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_lit_ms_rel` (`user_literature_id`, `manuscript_id`),
  KEY `idx_paper_user_lit_ms_rel_ms` (`manuscript_id`, `user_literature_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文献与论文章多对多';
