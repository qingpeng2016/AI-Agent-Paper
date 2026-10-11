-- paper_llm_constraint_exec_intensity 展示名与默认档位（balanced=平衡）

INSERT INTO `paper_llm_constraint_exec_intensity` (`code`, `name`, `exec_rounds`, `max_papers`, `max_ideas`, `is_default`) VALUES
  ('fast',     '更快', 1, 4, 1, 0),
  ('balanced', '平衡', 2, 2, 1, 1),
  ('deep',     '更深', 3, 3, 1, 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `exec_rounds` = VALUES(`exec_rounds`),
  `max_papers` = VALUES(`max_papers`),
  `max_ideas` = VALUES(`max_ideas`),
  `is_default` = VALUES(`is_default`);
