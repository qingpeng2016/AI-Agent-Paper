-- cs_ai / cs_systems 展示名缩短
UPDATE `paper_ref_discipline`
SET
  `name` = '人工智能',
  `name_en` = 'Artificial Intelligence',
  `updated_at` = NOW()
WHERE `code` = 'cs_ai';

UPDATE `paper_ref_discipline`
SET
  `name` = '计算机',
  `name_en` = 'Computer Science',
  `updated_at` = NOW()
WHERE `code` = 'cs_systems';
