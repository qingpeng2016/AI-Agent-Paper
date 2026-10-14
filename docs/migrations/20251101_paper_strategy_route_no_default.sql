-- 去掉 topic.stage.default；capability_id 必须是 ARIS skill 名（非 topic.stage.*）。

DELETE FROM `paper_llm_strategy_route`
WHERE `behavior_key` = 'topic.stage.default';

UPDATE `paper_llm_strategy_route` SET `capability_id` = 'idea-creator'
WHERE `behavior_key` = 'topic.stage.topic_generate';

UPDATE `paper_llm_strategy_route` SET `capability_id` = 'research-review'
WHERE `behavior_key` = 'topic.stage.topic_audit';

UPDATE `paper_llm_strategy_route` SET `capability_id` = 'experiment-bridge'
WHERE `behavior_key` = 'topic.stage.experiment_plan';

UPDATE `paper_llm_strategy_route` SET `capability_id` = 'idea-discovery'
WHERE `behavior_key` = 'topic.discovery.run';
