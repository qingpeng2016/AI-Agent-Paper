CREATE TABLE IF NOT EXISTS `paper_exec_intensity_write_paper` (
  `code`         VARCHAR(16)  NOT NULL COMMENT '档位键 fast|balanced|deep',
  `name`         VARCHAR(64)  NOT NULL COMMENT '展示名',
  `exec_rounds`  TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '审计迭代轮数',
  `is_default`   TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '前端默认选中项，仅一条应为 1',
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文写作·执行强度档位';

INSERT INTO `paper_exec_intensity_write_paper` (`code`, `name`, `exec_rounds`, `is_default`) VALUES
  ('fast',     '更快', 1, 0),
  ('balanced', '平衡', 2, 1),
  ('deep',     '更深', 3, 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `exec_rounds` = VALUES(`exec_rounds`),
  `is_default` = VALUES(`is_default`);
