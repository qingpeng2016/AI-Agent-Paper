-- Venue 约束模板：贡献类型、实验标准、写作风格（与 paper_template_document 互补）
-- user_id=0 系统默认，venue_id=0 通用；解析优先级：用户+venue → 用户+0 → 0+venue → 0+0
CREATE TABLE IF NOT EXISTS `paper_template_venue` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板 ID',
  `user_id`             BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=系统；>0=用户自定义（上传/编辑）',
  `venue_id`            BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=默认通用；>0=paper_ref_venue.id',
  `name`                VARCHAR(128) NOT NULL COMMENT '模板展示名',
  `constraint_template` JSON         NOT NULL COMMENT 'venue 约束 JSON（contribution_types、experiment_bar、writing_style 等）',
  `storage_uri`         VARCHAR(1024) DEFAULT NULL COMMENT '用户上传源文件相对 storage/ 路径（可选）',
  `source`              VARCHAR(16)  NOT NULL DEFAULT 'system' COMMENT 'system|upload|edit',
  `status`              VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled|deleted',
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_template_venue_scope` (`user_id`, `venue_id`),
  KEY `idx_paper_template_venue_lookup` (`venue_id`, `user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='目标 venue 约束模板（贡献/实验/写作）';

-- 系统默认 · 通用（venue_id=0）
INSERT INTO `paper_template_venue` (`user_id`, `venue_id`, `name`, `constraint_template`, `source`, `status`) VALUES
(0, 0, '系统默认 · 通用 venue',
 JSON_OBJECT(
   'schema_version', 1,
   'locale', 'zh',
   'description', '未指定或未匹配 paper_ref_venue 时使用；约束选题、综述与实验方案的贡献形态与实验深度。',
   'contribution_types', JSON_OBJECT(
     'allowed', JSON_ARRAY('method', 'theory', 'empirical', 'survey', 'system'),
     'labels_zh', JSON_OBJECT(
       'method', '方法/模型',
       'theory', '理论/分析',
       'empirical', '实证/实验',
       'survey', '综述/梳理',
       'system', '系统/工具'
     ),
     'prompt_hint', '每条 idea 的 contribution 须明确属于 allowed 之一，并与方向一致；勿混合未声明的类型。'
   ),
   'experiment_bar', JSON_OBJECT(
     'code', 'standard',
     'label_zh', '常规范实验',
     'requirements', JSON_ARRAY(
       '至少一个可复现主实验',
       '列出可对照基线（名称具体）',
       '指标与数据集可追踪',
       '建议说明算力/数据规模'
     ),
     'prompt_hint', '实验方案须写清基线、指标与主结论所依赖的证据；勿夸大未设计的实验。'
   ),
   'writing_style', JSON_OBJECT(
     'code', 'scholarly_generic',
     'label_zh', '通用学术表述',
     'prompt_hint', '中文表述准确、克制；避免营销式用语；引用须对应 Corpus ref_id。'
   )
 ), 'system', 'active')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `constraint_template` = VALUES(`constraint_template`),
  `source` = VALUES(`source`),
  `status` = VALUES(`status`);

-- 系统默认 · ML 顶会（与 paper_ref_venue.code=ml_top3 对齐）
INSERT INTO `paper_template_venue` (`user_id`, `venue_id`, `name`, `constraint_template`, `source`, `status`)
SELECT 0, v.`id`, '系统默认 · NeurIPS/ICLR/ICML',
 JSON_OBJECT(
   'schema_version', 1,
   'locale', 'zh',
   'ref_venue_code', 'ml_top3',
   'description', 'ML 顶会：方法/理论/实证贡献；强基线与消融；会议短文风格。',
   'contribution_types', JSON_OBJECT(
     'allowed', JSON_ARRAY('method', 'theory', 'empirical'),
     'labels_zh', JSON_OBJECT(
       'method', '方法',
       'theory', '理论',
       'empirical', '实证'
     ),
     'prompt_hint', '贡献须能在 8 页主会体量内讲清；method/theory/empirical 择主、可辅，勿堆叠无关贡献。'
   ),
   'experiment_bar', JSON_OBJECT(
     'code', 'strong_baselines_ablations',
     'label_zh', '强基线 + 消融',
     'requirements', JSON_ARRAY(
       'SOTA 或公认强基线至少 2 个',
       '主结果建议 ≥3 random seed 或等价稳定性说明',
       '消融须对应核心设计选择',
       '报告算力/训练成本或吞吐等可审稿指标'
     ),
     'prompt_hint', '与 paper_ref_venue.rubric.experiment_bar=strong_baselines_ablations 一致；缺基线须在 audit issues 标 major。'
   ),
   'writing_style', JSON_OBJECT(
     'code', 'ml_conference',
     'label_zh', 'ML 会议短文',
     'prompt_hint', 'Related Work 聚焦差异；Experiments 先主结果后消融；claim 与表格/图一致。'
   )
 ), 'system', 'active'
FROM `paper_ref_venue` v
WHERE v.`code` = 'ml_top3'
LIMIT 1
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `constraint_template` = VALUES(`constraint_template`),
  `source` = VALUES(`source`),
  `status` = VALUES(`status`);
