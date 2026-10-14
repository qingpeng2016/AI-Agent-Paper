ALTER TABLE `paper_output_experiment_plan`
  ADD COLUMN `pdf_uri` VARCHAR(1024) DEFAULT NULL COMMENT 'PDF 相对 storage 路径' AFTER `storage_uri`,
  ADD COLUMN `word_uri` VARCHAR(1024) DEFAULT NULL COMMENT 'Word 相对 storage 路径' AFTER `pdf_uri`;
