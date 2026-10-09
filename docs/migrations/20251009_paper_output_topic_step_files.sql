ALTER TABLE `paper_output_topic_step`
  ADD COLUMN `files` JSON DEFAULT NULL COMMENT '本步关联文件列表（PDF 路径等，JSON 数组只追加）' AFTER `extra`;
