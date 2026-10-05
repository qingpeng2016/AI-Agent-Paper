ALTER TABLE `vip_domain_config`
  DROP KEY `idx_vip_domain_config_pool`,
  DROP COLUMN `assigned_user_id`;
