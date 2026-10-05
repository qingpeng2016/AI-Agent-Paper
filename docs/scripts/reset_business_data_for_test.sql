-- 测试环境：清空订单/订阅/履约/流水/日志等业务数据，保留 users 与平台配置（商品、支付通道、Bot 调度、教程等）
-- 用法：mysql -u... -p... your_db < docs/scripts/reset_business_data_for_test.sql

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 访问与通知（依赖 subscription / api_key）
TRUNCATE TABLE `user_access_logs`;
TRUNCATE TABLE `user_notifications`;

-- 履约
TRUNCATE TABLE `user_api_keys`;
TRUNCATE TABLE `user_subscriptions`;

-- 交易与资金
TRUNCATE TABLE `user_invoices`;
TRUNCATE TABLE `user_refunds`;
TRUNCATE TABLE `user_wallet_flows`;
TRUNCATE TABLE `payment_callbacks`;
TRUNCATE TABLE `user_orders`;

-- 用户侧配置（非 users 主表）
TRUNCATE TABLE `user_invoice_config`;

-- 企业咨询与成员（若未建表可注释）
TRUNCATE TABLE `enterprise_users`;
TRUNCATE TABLE `enterprise_inquiry`;

-- 可选：上游容量计数归零（路由/容量测试时用）
-- UPDATE `upstream_info` SET `used_tokens` = 0;

SET FOREIGN_KEY_CHECKS = 1;
