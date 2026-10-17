-- 累计记录每次本地 aris/claude CLI 调用的完整命令（含参数）
ALTER TABLE `paper_output_topic_step`
  ADD COLUMN `exec_command` mediumtext NULL COMMENT '累计 ARIS CLI 执行命令（每次 invoke 追加）' AFTER `error_msg`;
