ALTER TABLE `vip_domain_config`
  ADD COLUMN `is_official` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否官网域名：1 是，0 否' AFTER `domain`;
