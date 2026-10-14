-- 图表分类（图表管理模块）
CREATE TABLE IF NOT EXISTS `paper_user_figure_category` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类 ID',
  `name`       VARCHAR(128)    NOT NULL COMMENT '分类名称',
  `status`     VARCHAR(16)     NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_user_fig_cat_status_id` (`status`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表分类';

CREATE TABLE IF NOT EXISTS `paper_user_figure_category_rel` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `figure_id`   BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_figure.id',
  `category_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_figure_category.id',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_fig_cat_rel` (`figure_id`, `category_id`),
  KEY `idx_paper_user_fig_cat_rel_cat` (`category_id`, `figure_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表与分类多对多';

CREATE TABLE IF NOT EXISTS `paper_user_figure_manuscript_rel` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关联 ID',
  `figure_id`     BIGINT UNSIGNED NOT NULL COMMENT 'paper_user_figure.id',
  `manuscript_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user_fig_ms_rel` (`figure_id`, `manuscript_id`),
  KEY `idx_paper_user_fig_ms_rel_ms` (`manuscript_id`, `figure_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图表与论文章多对多';

INSERT IGNORE INTO `paper_user_figure_manuscript_rel` (`figure_id`, `manuscript_id`)
SELECT `id`, `manuscript_id`
FROM `paper_user_figure`
WHERE `status` = 'active';
