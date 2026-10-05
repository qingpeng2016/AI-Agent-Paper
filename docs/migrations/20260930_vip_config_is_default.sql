ALTER TABLE `vip_config`
  ADD COLUMN `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '新用户默认 VIP 档位，仅一条应为 1' AFTER `enabled`;

UPDATE `vip_config` SET `is_default` = 1 WHERE `id` = 1;
