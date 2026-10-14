-- 历史：启用 aris 策略包（列已在 20251100 删除，新库请用 all.sql）。

INSERT INTO `paper_llm_strategy_package` (`code`, `name`, `status`)
SELECT 'aris', 'ARIS 本地 CLI', 'active'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `paper_llm_strategy_package` WHERE `code` = 'aris');

-- 已有库若仍含 provider_kind 等列，请先执行 20251100_paper_strategy_package_local_only.sql
