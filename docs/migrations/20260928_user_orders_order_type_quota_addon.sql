ALTER TABLE `user_orders`
  MODIFY COLUMN `order_type` VARCHAR(32) NOT NULL DEFAULT 'purchase'
    COMMENT 'purchase=新购 renewal=续费 upgrade=升档 quota_addon=加购额度';
