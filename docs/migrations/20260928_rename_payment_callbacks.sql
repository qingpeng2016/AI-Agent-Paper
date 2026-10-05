-- 已有 user_payment_callbacks 时重命名

RENAME TABLE `user_payment_callbacks` TO `payment_callbacks`;

ALTER TABLE `payment_callbacks`
  RENAME INDEX `uk_user_payment_callbacks_idem` TO `uk_payment_callbacks_idem`;
