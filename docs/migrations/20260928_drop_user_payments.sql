-- 升级用：库中已有 user_payments 时执行；全新库请直接用更新后的 20260928_user_orders.sql，勿重复 ADD 列

ALTER TABLE `user_orders`
  ADD COLUMN `pay_channel` VARCHAR(32) DEFAULT NULL COMMENT 'alipay|wechat|paypal' AFTER `enterprise_invoice`,
  ADD COLUMN `out_trade_no` VARCHAR(64) DEFAULT NULL COMMENT '平台支付单号（渠道 out_trade_no）' AFTER `pay_channel`,
  ADD COLUMN `third_trade_no` VARCHAR(128) DEFAULT NULL COMMENT '渠道流水号' AFTER `out_trade_no`,
  ADD COLUMN `raw_request_json` JSON DEFAULT NULL COMMENT '发起支付请求快照' AFTER `third_trade_no`,
  ADD COLUMN `raw_notify_json` JSON DEFAULT NULL COMMENT '支付回调原文' AFTER `raw_request_json`;

UPDATE `user_orders` o
INNER JOIN `user_payments` p ON p.order_id = o.id
SET
  o.pay_channel = p.channel,
  o.out_trade_no = p.out_trade_no,
  o.third_trade_no = p.third_trade_no,
  o.raw_request_json = p.raw_request_json,
  o.raw_notify_json = p.raw_notify_json
WHERE o.out_trade_no IS NULL;

ALTER TABLE `user_orders`
  ADD UNIQUE KEY `uk_user_orders_out_trade_no` (`out_trade_no`);

DROP TABLE IF EXISTS `user_payments`;

ALTER TABLE `user_refunds` DROP COLUMN `payment_id`;
