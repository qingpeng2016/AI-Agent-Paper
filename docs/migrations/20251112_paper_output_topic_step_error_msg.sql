-- 选题步骤失败原因（与 status=failed 配套，供前端展示）
ALTER TABLE `paper_output_topic_step`
  ADD COLUMN `error_msg` mediumtext NULL COMMENT '失败原因（status=failed 时写入）' AFTER `result`;
