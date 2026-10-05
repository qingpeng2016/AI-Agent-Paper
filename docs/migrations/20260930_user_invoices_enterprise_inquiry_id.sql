-- 发票 / 抬头配置关联企业采购需求（无关联时为 0）
ALTER TABLE `user_invoices`
  ADD COLUMN `enterprise_inquiry_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'enterprise_inquiry.id，0 表示无' AFTER `order_id`,
  ADD KEY `idx_user_invoices_enterprise_inquiry` (`enterprise_inquiry_id`);

ALTER TABLE `user_invoice_config`
  ADD COLUMN `enterprise_inquiry_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'enterprise_inquiry.id，0 表示无' AFTER `user_id`,
  ADD KEY `idx_user_invoice_config_enterprise_inquiry` (`enterprise_inquiry_id`);
