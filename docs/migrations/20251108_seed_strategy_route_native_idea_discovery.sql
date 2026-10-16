-- 策略路由为空时一次性执行（原生 /idea-discovery 选题 + 可选旧路由/实验计划）
-- 不需要 20251106（五 phase 拆分）；当前代码只用 topic.discovery.run

INSERT INTO `paper_llm_strategy_package` (`code`, `name`, `status`)
VALUES ('aris', 'ARIS 本地 CLI', 'active')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `status` = VALUES(`status`);

SET @pkg_aris := (SELECT `id` FROM `paper_llm_strategy_package` WHERE `code` = 'aris' LIMIT 1);

SET @paper_llm_any := (
  SELECT `id` FROM `paper_llm_model_config`
  WHERE `status` = 'active'
  ORDER BY `id` DESC
  LIMIT 1
);

INSERT INTO `paper_llm_strategy_route` (
  `behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`
) VALUES
  (
    'topic.discovery.run',
    'topic_discovery',
    'run',
    @pkg_aris,
    'idea-discovery',
    IFNULL(@paper_llm_any, 0),
    'active'
  ),
  (
    'topic.stage.experiment_plan',
    'experiment_plan',
    'experiment_plan',
    @pkg_aris,
    'experiment-bridge',
    IFNULL(@paper_llm_any, 0),
    'active'
  )
ON DUPLICATE KEY UPDATE
  `strategy_package_id` = VALUES(`strategy_package_id`),
  `capability_id` = VALUES(`capability_id`),
  `default_model_config_id` = VALUES(`default_model_config_id`),
  `status` = VALUES(`status`);
