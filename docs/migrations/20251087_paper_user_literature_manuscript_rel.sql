-- 用户文献与论文（paper_manuscript）多对多
CREATE TABLE IF NOT EXISTS `paper_user_literature_manuscript_rel` (
  `id`                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `user_literature_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_literature.id',
  `manuscript_id`      BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `created_at`         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_lit_ms_rel` (`user_literature_id`, `manuscript_id`),
  KEY `idx_paper_user_lit_ms_rel_ms` (`manuscript_id`, `user_literature_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文献与论文章多对多';
