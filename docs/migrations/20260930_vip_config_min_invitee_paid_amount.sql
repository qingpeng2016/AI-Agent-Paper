-- VIP 档位：直属下级已完成订单实付累计门槛（与 min_valid_invites 同时满足）

ALTER TABLE `vip_config`
  ADD COLUMN `min_invitee_paid_amount` DECIMAL(16,2) NOT NULL DEFAULT 0.00
    COMMENT '直属下级已完成订单实付累计（元，≥）' AFTER `min_valid_invites`;

UPDATE `vip_config` SET `min_invitee_paid_amount` = 0.00 WHERE `id` = 1;
UPDATE `vip_config` SET `min_invitee_paid_amount` = 0.00 WHERE `id` = 2;
UPDATE `vip_config` SET `min_invitee_paid_amount` = 500.00 WHERE `id` = 3;
UPDATE `vip_config` SET `min_invitee_paid_amount` = 2000.00 WHERE `id` = 4;
