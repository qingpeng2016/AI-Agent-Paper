-- 若已按旧版建表，执行本脚本删除 label、enabled
ALTER TABLE `vip_domain_config`
  DROP COLUMN `label`,
  DROP COLUMN `enabled`;

ALTER TABLE `vip_domain_config`
  DROP KEY `idx_vip_domain_config_pool`;
