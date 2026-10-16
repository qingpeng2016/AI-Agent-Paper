-- 步骤可读摘要改由 result / extra / error_msg 承担，不再单独存 summary_text
ALTER TABLE `paper_output_topic_step`
  DROP COLUMN `summary_text`;
