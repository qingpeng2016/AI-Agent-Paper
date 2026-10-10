-- 论文工作语言：zh | en（LLM 话术与输出字段语言）
ALTER TABLE `paper_manuscript`
  ADD COLUMN `content_language` VARCHAR(8) NOT NULL DEFAULT 'en'
    COMMENT 'zh|en：与 AI 沟通及产出文本语言' AFTER `discipline_id`;

UPDATE `paper_manuscript` SET `content_language` = 'en' WHERE `content_language` = '' OR `content_language` IS NULL;
