-- 累计记录每次本地 aris/claude CLI 调用的完整命令（含参数）
ALTER TABLE `paper_output_topic_step`
  ADD COLUMN `exec_command` mediumtext NULL COMMENT 'JSON：{stage_code:{"1":{invoked_at,workspace,command,prompt},...}} 每次 invoke 递增下标' AFTER `error_msg`;
