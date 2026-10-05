-- 已建 user_ledger_entries 时重命名为 user_wallet_flows

RENAME TABLE `user_ledger_entries` TO `user_wallet_flows`;

ALTER TABLE `user_wallet_flows`
  RENAME INDEX `uk_user_ledger_ref` TO `uk_user_wallet_flows_ref`,
  RENAME INDEX `idx_user_ledger_user_time` TO `idx_user_wallet_flows_user_time`;
