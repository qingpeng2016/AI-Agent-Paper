ALTER TABLE `paper_output_topic_step`
  ADD COLUMN `exec_progress_pct` TINYINT UNSIGNED NOT NULL DEFAULT 0
    COMMENT '执行进度 0-100（topic_discovery：每轮 generate+audit 各计一次）'
    AFTER `status`;
