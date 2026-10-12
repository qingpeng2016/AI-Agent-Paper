ALTER TABLE `paper_llm_call_logs`
  ADD COLUMN `request_json` JSON NOT NULL DEFAULT (JSON_OBJECT()) COMMENT '请求快照：stage_code、system_prompt、user_prompt' AFTER `model_name`,
  ADD COLUMN `response_json` JSON NOT NULL DEFAULT (JSON_OBJECT()) COMMENT '响应快照：text；失败时含 error' AFTER `request_json`;
