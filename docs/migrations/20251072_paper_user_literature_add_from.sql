ALTER TABLE `paper_user_literature`
  ADD COLUMN `from` VARCHAR(64) NOT NULL DEFAULT 'user_upload' COMMENT '来源：user_upload|arxiv|openalex|semantic_scholar|crossref|pubmed' AFTER `user_id`;

ALTER TABLE `paper_user_literature`
  ADD KEY `idx_paper_user_lit_from` (`user_id`, `from`, `status`);
