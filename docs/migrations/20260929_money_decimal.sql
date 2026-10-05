-- 金额：分(BIGINT) → 元 DECIMAL(16,2)
-- 幂等：若已是 `price` / `unit_price` 等 DECIMAL 列则跳过对应表。
-- 新库若已用 ai-platform-schema.sql（已是 DECIMAL），执行本文件应无实质变更。

DELIMITER //

DROP PROCEDURE IF EXISTS atm_migrate_money_cents_to_decimal//

CREATE PROCEDURE atm_migrate_money_cents_to_decimal()
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'products' AND COLUMN_NAME = 'price_cents'
  ) THEN
    ALTER TABLE `products`
      ADD COLUMN `price_new` DECIMAL(16,2) NOT NULL DEFAULT 0 COMMENT '售价（元）' AFTER `billing_period`;
    UPDATE `products` SET `price_new` = ROUND(`price_cents` / 100, 2);
    ALTER TABLE `products` DROP COLUMN `price_cents`;
    ALTER TABLE `products` CHANGE COLUMN `price_new` `price` DECIMAL(16,2) NOT NULL COMMENT '售价（元）';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_orders' AND COLUMN_NAME = 'unit_price_cents'
  ) THEN
    ALTER TABLE `user_orders`
      ADD COLUMN `unit_price_new` DECIMAL(16,2) NOT NULL DEFAULT 0 AFTER `quantity`,
      ADD COLUMN `total_amount_new` DECIMAL(16,2) NOT NULL DEFAULT 0 AFTER `status`;
    UPDATE `user_orders` SET
      `unit_price_new` = ROUND(`unit_price_cents` / 100, 2),
      `total_amount_new` = ROUND(`total_amount_cents` / 100, 2);
    ALTER TABLE `user_orders` DROP COLUMN `unit_price_cents`, DROP COLUMN `total_amount_cents`;
    ALTER TABLE `user_orders`
      CHANGE COLUMN `unit_price_new` `unit_price` DECIMAL(16,2) NOT NULL COMMENT '下单单价（元）',
      CHANGE COLUMN `total_amount_new` `total_amount` DECIMAL(16,2) NOT NULL COMMENT '应付总额（元）';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_wallet_flows' AND COLUMN_NAME = 'amount_cents'
  ) THEN
    ALTER TABLE `user_wallet_flows`
      ADD COLUMN `amount_new` DECIMAL(16,2) NOT NULL DEFAULT 0 AFTER `type`,
      ADD COLUMN `balance_after_new` DECIMAL(16,2) DEFAULT NULL AFTER `amount_new`;
    UPDATE `user_wallet_flows` SET
      `amount_new` = ROUND(`amount_cents` / 100, 2),
      `balance_after_new` = CASE WHEN `balance_after_cents` IS NULL THEN NULL ELSE ROUND(`balance_after_cents` / 100, 2) END;
    ALTER TABLE `user_wallet_flows` DROP COLUMN `amount_cents`, DROP COLUMN `balance_after_cents`;
    ALTER TABLE `user_wallet_flows`
      CHANGE COLUMN `amount_new` `amount` DECIMAL(16,2) NOT NULL COMMENT '正入负出（元）',
      CHANGE COLUMN `balance_after_new` `balance_after` DECIMAL(16,2) DEFAULT NULL COMMENT '变动后余额（元）';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.TABLES
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_refunds'
  ) AND EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_refunds' AND COLUMN_NAME = 'amount_cents'
  ) THEN
    ALTER TABLE `user_refunds`
      ADD COLUMN `amount_new` DECIMAL(16,2) NOT NULL DEFAULT 0 AFTER `refund_no`;
    UPDATE `user_refunds` SET `amount_new` = ROUND(`amount_cents` / 100, 2);
    ALTER TABLE `user_refunds` DROP COLUMN `amount_cents`;
    ALTER TABLE `user_refunds` CHANGE COLUMN `amount_new` `amount` DECIMAL(16,2) NOT NULL;
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_invoices' AND COLUMN_NAME = 'amount_cents'
  ) THEN
    ALTER TABLE `user_invoices`
      ADD COLUMN `amount_new` DECIMAL(16,2) NOT NULL DEFAULT 0 AFTER `tax_no`;
    UPDATE `user_invoices` SET `amount_new` = ROUND(`amount_cents` / 100, 2);
    ALTER TABLE `user_invoices` DROP COLUMN `amount_cents`;
    ALTER TABLE `user_invoices` CHANGE COLUMN `amount_new` `amount` DECIMAL(16,2) NOT NULL;
  END IF;
END//

DELIMITER ;

CALL atm_migrate_money_cents_to_decimal();
DROP PROCEDURE IF EXISTS atm_migrate_money_cents_to_decimal;
