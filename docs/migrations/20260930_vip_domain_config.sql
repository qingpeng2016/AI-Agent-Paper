-- 官方开放给用户选用的专属推广域名池（用户占用见 users.vip_domain）

CREATE TABLE IF NOT EXISTS `vip_domain_config` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `domain`           VARCHAR(255) NOT NULL COMMENT '完整域名，如 promo.example.com',
  `is_official`      TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否官网域名：1 是，0 否',
  `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_vip_domain_config_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户可选专属推广域名配置';
