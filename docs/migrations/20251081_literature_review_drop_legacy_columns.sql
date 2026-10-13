-- 移除文献综述表冗余字段（正文与附件 URI 保留在 content_medium / pdf|word|ppt_uri）
ALTER TABLE `paper_output_literature_review`
  DROP COLUMN `summary`,
  DROP COLUMN `storage_uri`,
  DROP COLUMN `structure`,
  DROP COLUMN `citations`,
  DROP COLUMN `input_params`,
  DROP COLUMN `meta`;
