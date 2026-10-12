ALTER TABLE `paper_user_literature`
  CHANGE COLUMN `storage_uri` `storage_url` VARCHAR(1024) DEFAULT NULL COMMENT '对象存储或工作区相对路径';
