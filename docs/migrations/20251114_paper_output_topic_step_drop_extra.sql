-- 步骤扩展字段并入 result JSON；表上仅保留 input_params / result / error_msg（及 status、files 等）
UPDATE `paper_output_topic_step`
SET `result` = JSON_MERGE_PATCH(COALESCE(`result`, JSON_OBJECT()), `extra`)
WHERE `extra` IS NOT NULL AND JSON_TYPE(`extra`) = 'OBJECT';

ALTER TABLE `paper_output_topic_step`
  DROP COLUMN `extra`;
