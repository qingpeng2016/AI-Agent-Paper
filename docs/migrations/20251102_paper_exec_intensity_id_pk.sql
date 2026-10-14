-- paper_exec_intensity_topic / write_paper：主键改为 id，code 保留为业务唯一键

ALTER TABLE `paper_exec_intensity_topic`
  DROP PRIMARY KEY,
  ADD COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_paper_exec_intensity_topic_code` (`code`);

ALTER TABLE `paper_exec_intensity_write_paper`
  DROP PRIMARY KEY,
  ADD COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_paper_exec_intensity_write_paper_code` (`code`);
