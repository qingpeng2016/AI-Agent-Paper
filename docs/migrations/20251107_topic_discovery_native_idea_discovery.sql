-- 原生 /idea-discovery：新 run 使用 stage_code=idea_discovery（retrieve → idea_discovery → topic_confirm）
-- 确保 topic.discovery.run 路由存在且 capability=idea-discovery
UPDATE `paper_llm_strategy_route`
SET `capability_id` = 'idea-discovery', `status` = 'active'
WHERE `behavior_key` = 'topic.discovery.run';
