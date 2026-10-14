-- Demo: 论文投稿（paper_manuscript_submission + event），user_id=1
-- 依赖：20251096_paper_manuscript_submission.sql、你的 100 条 paper_manuscript（id 1–100, user_id=1）
-- 当前论文 id=6「复杂脑网络在抑郁症研究中的应用」含 3 条投递 + 完整日志；另 9 篇各 1 条投递便于列表扫状态
--
-- 执行前可选检查：
--   SELECT id, title, is_current FROM paper_manuscript WHERE user_id=1 AND id=6;
-- 推荐整文件执行（含 DELETE）

SET NAMES utf8mb4;

DELETE e FROM `paper_manuscript_submission_event` AS e
INNER JOIN `paper_manuscript_submission` AS s ON s.`id` = e.`submission_id`
WHERE s.`user_id` = 1 AND s.`manuscript_id` BETWEEN 1 AND 100;

DELETE FROM `paper_manuscript_submission`
WHERE `user_id` = 1 AND `manuscript_id` BETWEEN 1 AND 100;

-- 部分 demo 稿：写作版本 / 定稿（未改动的行保持原 seed 值）
UPDATE `paper_manuscript`
SET
  `draft_version` = 3,
  `draft_status` = 'frozen',
  `draft_updated_at` = '2026-10-12 18:00:00'
WHERE `id` = 6 AND `user_id` = 1;

UPDATE `paper_manuscript`
SET
  `draft_version` = 2,
  `draft_status` = 'ready',
  `draft_updated_at` = '2026-10-10 12:00:00'
WHERE `id` IN (10, 20, 30, 40, 50, 60, 70, 80, 90, 100) AND `user_id` = 1;

INSERT INTO `paper_manuscript_submission` (
  `id`, `manuscript_id`, `user_id`, `venue_code`, `venue_label`, `track`, `round`, `status`,
  `draft_version`, `anonymous`, `portal_url`, `external_id`, `title`, `abstract`,
  `main_pdf_uri`, `main_pdf_filename`, `decision`, `comments_summary`, `revision_deadline`,
  `submitted_at`, `checklist`, `record_status`, `created_at`, `updated_at`
) VALUES
-- manuscript 6：历史投递 + 在审 + 当前待改稿（汇总以最新 submitted_at 为准）
(1, 6, 1, 'iclr_2026', 'ICLR 2026', 'main', 0, 'rejected',
  2, 1, 'https://openreview.net/group?id=ICLR.cc/2026/Conference', 'iclr26-demo-6-a',
  '复杂脑网络在抑郁症研究中的应用', 'Graph-theoretic markers of major depressive disorder from resting-state fMRI.',
  NULL, NULL, 'Reject', 'Limited novelty vs prior graph ML depression work; clarify clinical utility.',
  NULL, '2026-09-20 10:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', true)),
  'active', '2026-09-18 09:00:00.000', '2026-09-25 11:00:00.000'),

(2, 6, 1, 'neurips_2026', 'NeurIPS 2026', 'main', 1, 'resubmitted',
  3, 1, 'https://openreview.net/group?id=NeurIPS.cc/2026/Conference', 'neurips26-demo-6-b',
  'Complex Brain Networks in Depression: A Graph Learning Perspective', 'We resubmit with new ablations on Schaefer-100 parcellations and external HCP validation.',
  'submission-data/manuscript_6/submission_2/main_v3.pdf', 'main_v3.pdf', 'Major Revision', 'Address reviewer 2 on confounds; add power analysis appendix.',
  NULL, '2026-10-05 14:30:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', false)),
  'active', '2026-09-28 08:00:00.000', '2026-10-08 16:00:00.000'),

(3, 6, 1, 'ieee_tmi', 'IEEE TMI', 'main', 0, 'revision',
  3, 0, 'https://mc.manuscriptcentral.com/tmi-ieee', 'TMI-2026-0042',
  '复杂脑网络在抑郁症研究中的应用', '中文摘要占位：静息态 fMRI 图论指标与 HAMD 相关及组间差异。',
  'submission-data/manuscript_6/submission_3/main_depression_networks.pdf', 'main_depression_networks.pdf',
  'Major Revision', 'R1: 补充跨数据集泛化；R2: 澄清图构建阈值敏感性；R3: 英文语言润色。',
  '2026-12-15 23:59:59.000', '2026-10-14 09:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', false)),
  'active', '2026-10-12 10:00:00.000', '2026-10-14 09:30:00.000'),

-- 其它稿件：单条投递，覆盖多种状态（我的论文列表扫一眼）
(4, 10, 1, 'neurips_2026', 'NeurIPS 2026', 'main', 0, 'submitted',
  2, 1, 'https://openreview.net/', 'demo-ms10', 'Demo 论文 v10：稀疏注意力 方向研究', NULL,
  NULL, NULL, NULL, NULL, NULL, '2026-10-01 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', false)),
  'active', '2026-10-01 11:00:00.000', '2026-10-01 12:00:00.000'),

(5, 20, 1, 'iclr_2026', 'ICLR 2026', 'main', 0, 'under_review',
  2, 1, 'https://openreview.net/', 'demo-ms20', 'Demo 论文 v20：稀疏注意力 方向研究', NULL,
  NULL, NULL, NULL, NULL, NULL, '2026-10-03 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', true)),
  'active', '2026-10-03 11:00:00.000', '2026-10-05 08:00:00.000'),

(6, 30, 1, 'acl_2026', 'ACL 2026', 'short', 0, 'revision',
  2, 1, 'https://softconf.com/', 'demo-ms30', 'Demo 论文 v30：稀疏注意力 方向研究', NULL,
  NULL, NULL, 'Minor Revision', 'Clarify baseline comparison table.', '2026-11-20 23:59:59.000', '2026-10-06 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', false)),
  'active', '2026-10-06 11:00:00.000', '2026-10-10 09:00:00.000'),

(7, 40, 1, 'ieee_tmi', 'IEEE TMI', 'main', 0, 'rejected',
  2, 0, NULL, NULL, 'Demo 论文 v40：稀疏注意力 方向研究', NULL,
  NULL, NULL, 'Reject', 'Out of scope for TMI.', NULL, '2026-09-15 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', false), JSON_OBJECT('id', 'supp', 'ok', true)),
  'active', '2026-09-14 11:00:00.000', '2026-09-20 10:00:00.000'),

(8, 50, 1, 'neurips_2026', 'NeurIPS 2026', 'main', 0, 'accepted',
  2, 1, 'https://openreview.net/', 'demo-ms50-accept', 'Demo 论文 v50：稀疏注意力 方向研究', NULL,
  NULL, NULL, 'Accept', NULL, NULL, '2026-08-20 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', true)),
  'active', '2026-08-18 11:00:00.000', '2026-09-01 09:00:00.000'),

(9, 60, 1, 'iclr_2026', 'ICLR 2026', 'main', 1, 'resubmitted',
  2, 1, 'https://openreview.net/', 'demo-ms60-r1', 'Demo 论文 v60：稀疏注意力 方向研究', NULL,
  NULL, NULL, 'Major Revision', 'Resubmitted after rebuttal.', NULL, '2026-10-08 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', true)),
  'active', '2026-09-25 11:00:00.000', '2026-10-08 12:00:00.000'),

(10, 70, 1, 'neurips_2026', 'NeurIPS 2026', 'main', 0, 'rebuttal',
  2, 1, 'https://openreview.net/', 'demo-ms70-reb', 'Demo 论文 v70：稀疏注意力 方向研究', NULL,
  NULL, NULL, 'Weak Accept', 'Author rebuttal period.', '2026-10-25 23:59:59.000', '2026-10-09 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', false)),
  'active', '2026-10-09 11:00:00.000', '2026-10-11 08:00:00.000'),

(11, 80, 1, 'acl_2026', 'ACL 2026', 'main', 0, 'submitted',
  2, 1, NULL, NULL, 'Demo 论文 v80：稀疏注意力 方向研究', NULL,
  NULL, NULL, NULL, NULL, NULL, '2026-10-11 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', false)),
  'active', '2026-10-11 11:00:00.000', '2026-10-11 12:00:00.000'),

(12, 100, 1, 'iclr_2026', 'ICLR 2026', 'main', 0, 'under_review',
  2, 1, 'https://openreview.net/', 'demo-ms100', 'Demo 论文 v100：稀疏注意力 方向研究', NULL,
  NULL, NULL, NULL, NULL, NULL, '2026-10-13 12:00:00.000',
  JSON_ARRAY(JSON_OBJECT('id', 'pages', 'ok', true), JSON_OBJECT('id', 'anon', 'ok', true), JSON_OBJECT('id', 'supp', 'ok', true)),
  'active', '2026-10-13 11:00:00.000', '2026-10-13 12:00:00.000');

INSERT INTO `paper_manuscript_submission_event` (`id`, `submission_id`, `event_type`, `message`, `created_at`) VALUES
(1, 1, 'created', '创建投递记录', '2026-09-18 09:00:00.000'),
(2, 1, 'submitted', '标记已在 OpenReview 完成提交', '2026-09-20 10:00:00.000'),
(3, 2, 'created', '创建投递记录', '2026-09-28 08:00:00.000'),
(4, 2, 'submitted', '标记已在 OpenReview 完成提交', '2026-10-01 10:00:00.000'),
(5, 2, 'revision_saved', '保存改稿计划 · Major Revision', '2026-10-03 15:00:00.000'),
(6, 2, 'resubmitted', '标记已修回 · R1', '2026-10-05 14:30:00.000'),
(7, 3, 'created', '创建投递记录', '2026-10-12 10:00:00.000'),
(8, 3, 'submitted', '标记已在 ScholarOne 完成提交', '2026-10-14 09:00:00.000'),
(9, 3, 'pdf_uploaded', '上传主 PDF：main_depression_networks.pdf', '2026-10-14 09:15:00.000'),
(10, 3, 'revision_saved', '保存改稿计划 · Major Revision', '2026-10-14 09:30:00.000'),
(11, 4, 'created', '创建投递记录', '2026-10-01 11:00:00.000'),
(12, 4, 'submitted', '标记已提交', '2026-10-01 12:00:00.000'),
(13, 5, 'created', '创建投递记录', '2026-10-03 11:00:00.000'),
(14, 5, 'submitted', '标记已提交', '2026-10-03 12:00:00.000'),
(15, 6, 'created', '创建投递记录', '2026-10-06 11:00:00.000'),
(16, 6, 'submitted', '标记已提交', '2026-10-06 12:00:00.000'),
(17, 6, 'revision_saved', '保存改稿计划', '2026-10-10 09:00:00.000'),
(18, 8, 'created', '创建投递记录', '2026-08-18 11:00:00.000'),
(19, 8, 'submitted', '标记已提交', '2026-08-20 12:00:00.000'),
(20, 9, 'resubmitted', '标记已修回 · R1', '2026-10-08 12:00:00.000');

-- 与后端 syncManuscriptSummary 一致：每篇取最新 submitted_at 的投递写回汇总列
UPDATE `paper_manuscript` AS m
INNER JOIN (
  SELECT s.`manuscript_id`, s.`status`, s.`venue_label`, s.`track`, s.`revision_deadline`
  FROM `paper_manuscript_submission` AS s
  INNER JOIN (
    SELECT `manuscript_id`, MAX(COALESCE(`submitted_at`, `created_at`)) AS `latest_at`
    FROM `paper_manuscript_submission`
    WHERE `user_id` = 1 AND `record_status` = 'active' AND `manuscript_id` BETWEEN 1 AND 100
    GROUP BY `manuscript_id`
  ) AS t ON t.`manuscript_id` = s.`manuscript_id`
    AND COALESCE(s.`submitted_at`, s.`created_at`) = t.`latest_at`
  WHERE s.`user_id` = 1 AND s.`record_status` = 'active'
) AS x ON x.`manuscript_id` = m.`id`
SET
  m.`submission_status` = x.`status`,
  m.`submission_label` = CONCAT(x.`venue_label`, ' · ', IF(LOWER(x.`track`) = 'short', 'Short', 'Main')),
  m.`submission_deadline` = x.`revision_deadline`
WHERE m.`user_id` = 1 AND m.`id` BETWEEN 1 AND 100;
