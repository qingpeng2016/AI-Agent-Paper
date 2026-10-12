ALTER TABLE `paper_user_literature`
  CHANGE COLUMN `file_kind` `file_ext` VARCHAR(16) NOT NULL COMMENT 'pdf|bib|txt|other';
