-- MySQL 8：连接若仍用 utf8(utf8mb3) 写 JSON 会报 Error 3988。
-- 应用侧已改 DSN charset=utf8mb4；旧库可执行本脚本统一表字符集。

ALTER TABLE `paper_output_topic_step`
  CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

ALTER TABLE `paper_llm_call_logs`
  CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
