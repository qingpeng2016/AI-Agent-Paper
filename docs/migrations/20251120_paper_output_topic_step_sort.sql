-- paper_output_topic_step 步骤排序：工作台与段间确认均按 sort 升序
ALTER TABLE paper_output_topic_step
  ADD COLUMN sort INT NOT NULL DEFAULT 0 COMMENT 'run 内步骤顺序，越小越靠前' AFTER stage_code;

UPDATE paper_output_topic_step SET sort = CASE stage_code
  WHEN 'retrieve' THEN 1
  WHEN 'idea_discovery' THEN 1
  WHEN 'topic_discovery' THEN 1
  WHEN 'research_lit' THEN 1
  WHEN 'idea_creator' THEN 2
  WHEN 'novelty_check' THEN 3
  WHEN 'research_review' THEN 4
  WHEN 'research_refine' THEN 5
  WHEN 'topic_confirm' THEN 6
  WHEN 'generate_ideas' THEN 2
  WHEN 'audit' THEN 3
  WHEN 'novelty' THEN 4
  ELSE sort
END WHERE sort = 0;
