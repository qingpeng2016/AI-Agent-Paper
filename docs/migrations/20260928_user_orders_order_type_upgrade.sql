-- 已有 order_type 列时补充升档类型说明（可选）

ALTER TABLE `user_orders`
  MODIFY COLUMN `order_type` VARCHAR(32) NOT NULL DEFAULT 'purchase'
    COMMENT 'purchase=新购 renewal=续费 upgrade=升档';
