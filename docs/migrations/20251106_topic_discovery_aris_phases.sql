-- ARIS idea-discovery 五 phase 路由（替代 topic_generate/topic_audit 双 invoke 循环）
INSERT INTO `paper_llm_strategy_route`
  (`behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`)
SELECT * FROM (
  SELECT 'topic.stage.research_lit' AS behavior_key, 'topic_discovery' AS module, 'research_lit' AS action,
         sp.id AS strategy_package_id, 'research-lit' AS capability_id, COALESCE(m.id, 0) AS default_model_config_id, 'active' AS status
  FROM `paper_llm_strategy_package` sp
  LEFT JOIN `paper_llm_model_config` m ON m.model_name LIKE '%claude%' AND m.status = 'active'
  WHERE sp.code = 'aris_local' LIMIT 1
) t
WHERE NOT EXISTS (SELECT 1 FROM `paper_llm_strategy_route` r WHERE r.behavior_key = 'topic.stage.research_lit');

INSERT INTO `paper_llm_strategy_route`
  (`behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`)
SELECT * FROM (
  SELECT 'topic.stage.idea_creator', 'topic_discovery', 'idea_creator', sp.id, 'idea-creator', COALESCE(m.id, 0), 'active'
  FROM `paper_llm_strategy_package` sp
  LEFT JOIN `paper_llm_model_config` m ON m.model_name LIKE '%claude%' AND m.status = 'active'
  WHERE sp.code = 'aris_local' LIMIT 1
) t
WHERE NOT EXISTS (SELECT 1 FROM `paper_llm_strategy_route` r WHERE r.behavior_key = 'topic.stage.idea_creator');

INSERT INTO `paper_llm_strategy_route`
  (`behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`)
SELECT * FROM (
  SELECT 'topic.stage.novelty_check', 'topic_discovery', 'novelty_check', sp.id, 'novelty-check', COALESCE(m.id, 0), 'active'
  FROM `paper_llm_strategy_package` sp
  LEFT JOIN `paper_llm_model_config` m ON m.model_name LIKE '%codex%' AND m.status = 'active'
  WHERE sp.code = 'aris_local' LIMIT 1
) t
WHERE NOT EXISTS (SELECT 1 FROM `paper_llm_strategy_route` r WHERE r.behavior_key = 'topic.stage.novelty_check');

INSERT INTO `paper_llm_strategy_route`
  (`behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`)
SELECT * FROM (
  SELECT 'topic.stage.research_review', 'topic_discovery', 'research_review', sp.id, 'research-review', COALESCE(m.id, 0), 'active'
  FROM `paper_llm_strategy_package` sp
  LEFT JOIN `paper_llm_model_config` m ON m.model_name LIKE '%codex%' AND m.status = 'active'
  WHERE sp.code = 'aris_local' LIMIT 1
) t
WHERE NOT EXISTS (SELECT 1 FROM `paper_llm_strategy_route` r WHERE r.behavior_key = 'topic.stage.research_review');

INSERT INTO `paper_llm_strategy_route`
  (`behavior_key`, `module`, `action`, `strategy_package_id`, `capability_id`, `default_model_config_id`, `status`)
SELECT * FROM (
  SELECT 'topic.stage.research_refine', 'topic_discovery', 'research_refine', sp.id, 'research-refine-pipeline', COALESCE(m.id, 0), 'active'
  FROM `paper_llm_strategy_package` sp
  LEFT JOIN `paper_llm_model_config` m ON m.model_name LIKE '%claude%' AND m.status = 'active'
  WHERE sp.code = 'aris_local' LIMIT 1
) t
WHERE NOT EXISTS (SELECT 1 FROM `paper_llm_strategy_route` r WHERE r.behavior_key = 'topic.stage.research_refine');
