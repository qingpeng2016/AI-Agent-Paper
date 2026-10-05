-- 已有库增量：users 增加密码明文字段
ALTER TABLE `users`
  ADD COLUMN `password_plain` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码明文（业务要求留存，仅限受控环境）'
  AFTER `password_hash`;

-- 若曾用 DEFAULT '' 过渡，新注册用户由应用写入明文；历史行请按需补录或留空后改 NOT NULL 策略
