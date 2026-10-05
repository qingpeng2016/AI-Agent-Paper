-- 用户可提现/累计佣金余额（元）；流水 type=commission|withdraw 时与本字段配合。
ALTER TABLE `users`
  ADD COLUMN `commission_balance` DECIMAL(16,2) NOT NULL DEFAULT 0 COMMENT '佣金余额（元）' AFTER `wallet_balance`;
