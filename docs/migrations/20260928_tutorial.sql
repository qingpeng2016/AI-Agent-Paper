-- 教程分类 + 教程文章

CREATE TABLE IF NOT EXISTS `tutorial_category` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`       VARCHAR(32)  NOT NULL COMMENT '筛选键，如 payment、pro',
  `name`       VARCHAR(64)  NOT NULL COMMENT '展示名',
  `sort`       INT          NOT NULL DEFAULT 0,
  `status`     VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tutorial_category_code` (`code`),
  KEY `idx_tutorial_category_sort` (`status`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教程分类';

CREATE TABLE IF NOT EXISTS `tutorial_article` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `category_id`  BIGINT UNSIGNED NOT NULL COMMENT 'tutorial_category.id',
  `slug`         VARCHAR(128) NOT NULL COMMENT 'URL 唯一标识',
  `title`        VARCHAR(256) NOT NULL,
  `excerpt`      VARCHAR(512) NOT NULL DEFAULT '',
  `body`         JSON         NOT NULL COMMENT '正文段落数组',
  `published_at` DATE         NOT NULL COMMENT '展示日期',
  `sort`         INT          NOT NULL DEFAULT 0,
  `status`       VARCHAR(16)  NOT NULL DEFAULT 'published' COMMENT 'draft|published',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tutorial_article_slug` (`slug`),
  KEY `idx_tutorial_article_category` (`category_id`),
  KEY `idx_tutorial_article_list` (`status`, `published_at`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教程文章';

INSERT INTO `tutorial_category` (`code`, `name`, `sort`, `status`) VALUES
  ('plus', 'Plus 代充', 10, 'active'),
  ('pro', 'Pro 升级', 20, 'active'),
  ('perplexity', 'Perplexity', 30, 'active'),
  ('payment', '充值 / 支付', 40, 'active'),
  ('tutorial', '功能教程', 50, 'active'),
  ('quality', '降智 / 认知', 60, 'active'),
  ('account', '账号 / 注册', 70, 'active'),
  ('pricing', '价格 / 对比', 80, 'active')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `sort` = VALUES(`sort`),
  `status` = VALUES(`status`);
