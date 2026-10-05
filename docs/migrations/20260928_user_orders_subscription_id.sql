ALTER TABLE `user_orders`
  ADD COLUMN `user_subscription_id` BIGINT UNSIGNED NOT NULL DEFAULT 0
    COMMENT '关联 user_subscriptions.id；新购为 0，续费/升档为原订阅'
    AFTER `order_type`;
