-- 企业方案参考（方案卡片）

CREATE TABLE IF NOT EXISTS `enterprise_products` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`          VARCHAR(32)  NOT NULL COMMENT '唯一键，如 starter|business|custom',
  `name`          VARCHAR(64)  NOT NULL COMMENT '方案名称',
  `badge`         VARCHAR(32)  DEFAULT NULL COMMENT '角标，如 常用',
  `price_hint`    VARCHAR(128) NOT NULL COMMENT '价格展示文案',
  `seats`         VARCHAR(128) NOT NULL COMMENT '席位范围说明',
  `features`      JSON         NOT NULL COMMENT '特性条目数组',
  `tagline`       VARCHAR(256) NOT NULL COMMENT '底部说明（适合…）',
  `button_label`  VARCHAR(64)  NOT NULL DEFAULT '获取报价' COMMENT '按钮文案',
  `is_featured`   TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否默认推荐高亮',
  `sort`          INT          NOT NULL DEFAULT 0 COMMENT '排序，越小越靠前',
  `status`        VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|hidden',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_enterprise_products_code` (`code`),
  KEY `idx_enterprise_products_list` (`status`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业采购方案';

INSERT INTO `enterprise_products`
  (`code`, `name`, `badge`, `price_hint`, `seats`, `features`, `tagline`, `button_label`, `is_featured`, `sort`, `status`)
VALUES
  (
    'starter',
    '团队试用',
    NULL,
    '¥5,000 起',
    '3–10 席位',
    JSON_ARRAY('混合 SKU 组合', '在线支付为主', '电子普票', '标准客服响应'),
    '适合小团队试点',
    '获取报价',
    0,
    10,
    'active'
  ),
  (
    'business',
    '标准企业',
    '常用',
    '按 SKU 阶梯价',
    '10–100 席位',
    JSON_ARRAY('对公转账 + 合同', 'Dedicated 顾问', '批量账号清单导入', '续费日历提醒'),
    '适合部门统一采购',
    '获取报价',
    1,
    20,
    'active'
  ),
  (
    'custom',
    '定制方案',
    NULL,
    '面议',
    '100+ 或跨品牌',
    JSON_ARRAY('年度框架价', '多主体开票', '分阶段开通', 'SLA 与异常升级通道'),
    '适合集团 / 多子公司',
    '获取报价',
    0,
    30,
    'active'
  )
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `badge` = VALUES(`badge`),
  `price_hint` = VALUES(`price_hint`),
  `seats` = VALUES(`seats`),
  `features` = VALUES(`features`),
  `tagline` = VALUES(`tagline`),
  `button_label` = VALUES(`button_label`),
  `is_featured` = VALUES(`is_featured`),
  `sort` = VALUES(`sort`),
  `status` = VALUES(`status`);
