-- 策略包仅本地 CLI：去掉 http/能力清单等列；路由默认指向 aris 包。

ALTER TABLE `paper_llm_strategy_package`
  DROP COLUMN `provider_kind`,
  DROP COLUMN `base_url`,
  DROP COLUMN `auth_type`,
  DROP COLUMN `auth_secret_ref`,
  DROP COLUMN `manifest_version`,
  DROP COLUMN `capabilities`;

INSERT INTO `paper_llm_strategy_package` (`code`, `name`, `status`)
VALUES ('aris', 'ARIS 本地 CLI', 'active')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `status` = VALUES(`status`);

UPDATE `paper_llm_strategy_route` r
INNER JOIN `paper_llm_strategy_package` p ON p.`code` = 'aris' AND p.`status` = 'active'
SET r.`strategy_package_id` = p.`id`
WHERE r.`status` = 'active';
