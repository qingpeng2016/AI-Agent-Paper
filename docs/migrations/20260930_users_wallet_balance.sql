-- 用户钱包余额（元），与 user_wallet_flows 流水配合；余额以本字段为准。
ALTER TABLE `users`
  ADD COLUMN `wallet_balance` DECIMAL(16,2) NOT NULL DEFAULT 0 COMMENT '钱包可用余额（元）' AFTER `status`;
