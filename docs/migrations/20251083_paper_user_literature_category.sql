-- 用户文献分类（文献管理模块）
CREATE TABLE IF NOT EXISTS `paper_user_literature_category` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类 ID',
  `name`       VARCHAR(128)    NOT NULL COMMENT '分类名称',
  `sort`       INT             NOT NULL DEFAULT 100 COMMENT '展示顺序，越小越靠前',
  `status`     VARCHAR(16)     NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_paper_user_lit_cat_status_sort` (`status`, `sort`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文献分类';
