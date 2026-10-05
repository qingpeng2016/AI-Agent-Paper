-- 收款码改为库内存储（二进制），不再使用 qr_url
ALTER TABLE `user_commission_payout_config`
  DROP COLUMN `qr_url`,
  ADD COLUMN `qr_mime`  VARCHAR(64) DEFAULT NULL COMMENT 'image/png|image/jpeg|image/webp' AFTER `channel`,
  ADD COLUMN `qr_image` MEDIUMBLOB DEFAULT NULL COMMENT '收款码图片二进制' AFTER `qr_mime`;
