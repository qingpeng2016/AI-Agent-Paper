-- paper_user_literature：storage_url → storage_uri（与其它 paper_output_* 一致）
ALTER TABLE `paper_user_literature`
  CHANGE COLUMN `storage_url` `storage_uri` VARCHAR(1024) DEFAULT NULL COMMENT '对象存储或工作区相对路径';
