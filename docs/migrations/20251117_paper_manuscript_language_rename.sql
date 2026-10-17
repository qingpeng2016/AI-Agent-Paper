-- paper_manuscript：content_language → language（zh|en，AI 产出与交付物工作语言）
ALTER TABLE `paper_manuscript`
  CHANGE COLUMN `content_language` `language` VARCHAR(8) NOT NULL DEFAULT 'en'
  COMMENT 'zh|en：与 AI 沟通及产出文本语言';
