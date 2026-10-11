-- 对照《研究生教育学科专业目录（2022年）》补全常见一级学科缺口，并微调已有名称。
-- 口径：学术型一级学科（不含 xx5x 专业学位）；工学 39 项仍归 engineering + cs_*；军事学未收录。

INSERT INTO `paper_ref_discipline` (`code`, `name`, `name_en`, `sort`, `status`) VALUES
  ('atmospheric_science', '大气科学', 'Atmospheric Science', 61, 'active'),
  ('ocean_science', '海洋科学', 'Oceanography', 62, 'active'),
  ('geology', '地质学', 'Geology', 63, 'active'),
  ('geophysics', '地球物理学', 'Geophysics', 64, 'active'),
  ('ecology', '生态学', 'Ecology', 71, 'active'),
  ('basic_medicine', '基础医学', 'Basic Medical Sciences', 79, 'active'),
  ('tcm', '中医学', 'Traditional Chinese Medicine', 81, 'active'),
  ('integrative_medicine', '中西医结合', 'Integrative Chinese & Western Medicine', 82, 'active'),
  ('stomatology', '口腔医学', 'Stomatology', 83, 'active'),
  ('nursing', '护理学', 'Nursing', 86, 'active'),
  ('chinese_pharmacy', '中药学', 'Chinese Pharmacology', 89, 'active'),
  ('forensic_medicine', '法医学', 'Forensic Medicine', 90, 'active'),
  ('ethnology', '民族学', 'Ethnology', 132, 'active'),
  ('marxism_theory', '马克思主义理论', 'Marxist Theory', 133, 'active'),
  ('party_history_governance', '中共党史党建学', 'Party History & Governance', 134, 'active'),
  ('public_administration', '公共管理', 'Public Administration', 126, 'active'),
  ('agro_forestry_economics', '农林经济管理', 'Agricultural & Forestry Economics', 127, 'active'),
  ('information_resource_mgmt', '信息资源管理', 'Information Resource Management', 128, 'active'),
  ('sports_science', '体育学', 'Sport Science', 151, 'active'),
  ('archaeology', '考古学', 'Archaeology', 184, 'active'),
  ('area_studies', '区域国别学', 'Area & Country Studies', 210, 'active')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `name_en` = VALUES(`name_en`),
  `sort` = VALUES(`sort`),
  `status` = VALUES(`status`);

UPDATE `paper_ref_discipline` SET `name` = '临床医学', `name_en` = 'Clinical Medicine', `updated_at` = NOW() WHERE `code` = 'medicine';
UPDATE `paper_ref_discipline` SET `name` = '公共卫生与预防医学', `name_en` = 'Public Health & Preventive Medicine', `updated_at` = NOW() WHERE `code` = 'public_health';
UPDATE `paper_ref_discipline` SET `name` = '新闻传播学', `name_en` = 'Journalism & Communication', `updated_at` = NOW() WHERE `code` = 'communication';
UPDATE `paper_ref_discipline` SET `name` = '外国语言文学', `name_en` = 'Foreign Languages & Literature', `updated_at` = NOW() WHERE `code` = 'linguistics';
UPDATE `paper_ref_discipline` SET `name` = '中国语言文学', `name_en` = 'Chinese Language & Literature', `updated_at` = NOW() WHERE `code` = 'literature';
UPDATE `paper_ref_discipline` SET `name` = '艺术学', `name_en` = 'Arts', `updated_at` = NOW() WHERE `code` = 'arts_design';
UPDATE `paper_ref_discipline` SET `name_en` = 'Neuroscience', `sort` = 91, `updated_at` = NOW() WHERE `code` = 'neuroscience';
