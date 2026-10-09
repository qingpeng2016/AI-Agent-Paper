-- 已有库：为 paper_ref_literature_source 增加优先级与默认勾选
ALTER TABLE `paper_ref_literature_source`
  ADD COLUMN `priority` INT NOT NULL DEFAULT 100 COMMENT '检索与 UI 展示顺序，越小越优先' AFTER `rate_limit_hint`,
  ADD COLUMN `default_selected` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '选题表单默认勾选' AFTER `priority`,
  ADD KEY `idx_paper_ref_literature_source_active_priority` (`status`, `priority`, `id`);

UPDATE `paper_ref_literature_source` SET `priority` = 1, `default_selected` = 1 WHERE `code` = 'arxiv';
UPDATE `paper_ref_literature_source` SET `priority` = 2, `default_selected` = 1 WHERE `code` = 'openalex';
UPDATE `paper_ref_literature_source` SET `priority` = 3, `default_selected` = 1 WHERE `code` = 'semantic_scholar';
UPDATE `paper_ref_literature_source` SET `priority` = 4, `default_selected` = 0 WHERE `code` = 'crossref';
UPDATE `paper_ref_literature_source` SET `priority` = 5, `default_selected` = 0 WHERE `code` = 'pubmed';

INSERT INTO `paper_ref_literature_source` (`code`, `name`, `api_kind`, `base_url`, `auth_type`, `priority`, `default_selected`) VALUES
  ('crossref', 'Crossref', 'rest', 'https://api.crossref.org/works', 'none', 4, 0),
  ('pubmed', 'PubMed', 'rest', 'https://eutils.ncbi.nlm.nih.gov/entrez/eutils', 'none', 5, 0)
ON DUPLICATE KEY UPDATE
  `priority` = VALUES(`priority`),
  `default_selected` = VALUES(`default_selected`);
