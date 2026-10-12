ALTER TABLE `paper_user_literature`
  CHANGE COLUMN `title` `literature_title` VARCHAR(512) NOT NULL COMMENT '文献标题';
