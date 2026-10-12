ALTER TABLE `paper_ref_venue`
  ADD COLUMN `sort` INT NOT NULL DEFAULT 100 COMMENT '选题下拉展示顺序，越小越靠前' AFTER `name`;

UPDATE `paper_ref_venue` SET `sort` = `id`;
