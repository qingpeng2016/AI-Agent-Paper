-- paper_llm_description：去掉自增 id，以 description_code 为主键
ALTER TABLE `paper_llm_description`
  DROP INDEX `uk_paper_llm_description_code`;

ALTER TABLE `paper_llm_description`
  DROP PRIMARY KEY,
  DROP COLUMN `id`,
  ADD PRIMARY KEY (`description_code`);
