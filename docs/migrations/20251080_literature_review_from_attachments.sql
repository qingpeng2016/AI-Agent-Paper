-- paper_output_literature_review: 来源 + PDF/Word/PPT 存储路径
ALTER TABLE `paper_output_literature_review`
  ADD COLUMN `from` VARCHAR(16) NOT NULL DEFAULT 'generated' COMMENT 'upload|generated' AFTER `status`,
  ADD COLUMN `pdf_uri` VARCHAR(1024) DEFAULT NULL COMMENT 'PDF 相对 storage 路径' AFTER `storage_uri`,
  ADD COLUMN `word_uri` VARCHAR(1024) DEFAULT NULL COMMENT 'Word 相对 storage 路径' AFTER `pdf_uri`,
  ADD COLUMN `ppt_uri` VARCHAR(1024) DEFAULT NULL COMMENT '汇报 PPT 相对 storage 路径' AFTER `word_uri`;

-- 已有数据：默认视为生成，并填充演示附件路径（可按环境覆盖）
UPDATE `paper_output_literature_review` r
SET
  `from` = IF(MOD(r.id, 10) = 0, 'upload', 'generated'),
  `pdf_uri` = CONCAT('literature-review/', r.manuscript_id, '/', r.id, '/v', r.version, '.pdf'),
  `word_uri` = CONCAT('literature-review/', r.manuscript_id, '/', r.id, '/v', r.version, '.docx'),
  `ppt_uri` = CONCAT('literature-review/', r.manuscript_id, '/', r.id, '/report_v', r.version, '.pptx')
WHERE `pdf_uri` IS NULL;
