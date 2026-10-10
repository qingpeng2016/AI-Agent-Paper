-- 文档提纲模板：user_id=0 系统默认，venue_id=0 通用 venue；非 0 为用户/期刊定制
-- 新颖性已并入 generate_ideas 模板（output.novelty），不设单独 stage_code=novelty 系统行
-- 若曾创建旧表 paper_document_template：RENAME TABLE `paper_document_template` TO `paper_template_document`;
CREATE TABLE IF NOT EXISTS `paper_template_document` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板 ID',
  `user_id`           BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=系统；>0=用户自定义（上传/编辑）',
  `venue_id`          BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=默认通用；>0=paper_ref_venue.id',
  `stage_code`        VARCHAR(64)  NOT NULL COMMENT '环节键：generate_ideas（含 novelty）|audit|experiment_plan|…',
  `name`              VARCHAR(128) NOT NULL COMMENT '模板展示名',
  `outline_template`  JSON         NOT NULL COMMENT '阶段文档 JSON 提纲（schema_version、output、markdown_outline 等）',
  `storage_uri`       VARCHAR(1024) DEFAULT NULL COMMENT '用户上传源文件相对 storage/ 路径（可选）',
  `source`            VARCHAR(16)  NOT NULL DEFAULT 'system' COMMENT 'system|upload|edit',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled|deleted',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_template_document_scope` (`user_id`, `venue_id`, `stage_code`),
  KEY `idx_paper_template_document_lookup` (`stage_code`, `venue_id`, `user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='各阶段文档 JSON 提纲模板';

-- 系统默认（user_id=0, venue_id=0）
INSERT INTO `paper_template_document` (`user_id`, `venue_id`, `stage_code`, `name`, `outline_template`, `source`, `status`) VALUES
(0, 0, 'generate_ideas', '系统默认 · 脑暴选题与新颖性',
 JSON_OBJECT(
   'schema_version', 1,
   'artifact', 'ideas_and_novelty',
   'locale', 'zh',
   'description', '一次 LLM 输出 ideas 与 novelty；字段键名与顺序须与本模板一致。',
   'output', JSON_OBJECT(
     'root_keys', JSON_ARRAY('ideas', 'novelty'),
     'ideas', JSON_OBJECT(
       'max_items_param', 'max_ideas',
       'item_fields', JSON_ARRAY(
         JSON_OBJECT('key', 'title', 'label_zh', '选题标题', 'required', true),
         JSON_OBJECT('key', 'problem', 'label_zh', '问题与动机', 'required', true),
         JSON_OBJECT('key', 'approach', 'label_zh', '技术路线', 'required', true),
         JSON_OBJECT('key', 'contribution', 'label_zh', '预期贡献', 'required', true, 'hint', '对齐目标 venue 的贡献类型'),
         JSON_OBJECT('key', 'reference_keys', 'label_zh', '关键文献', 'required', true, 'type', 'external_key[]')
       )
     ),
     'novelty', JSON_OBJECT(
       'fields', JSON_ARRAY(
         JSON_OBJECT('key', 'lines', 'label_zh', '新颖性要点', 'type', 'string[]'),
         JSON_OBJECT('key', 'risks', 'label_zh', '逐题风险', 'type', 'array', 'item_keys', JSON_ARRAY('idea_title', 'risk', 'note', 'overlap_refs')),
         JSON_OBJECT('key', 'risk', 'enum', JSON_ARRAY('low', 'medium', 'high')),
         JSON_OBJECT('key', 'synthesis', 'label_zh', '综合结论', 'type', 'string')
       )
     )
   )
 ), 'system', 'active'),
(0, 0, 'audit', '系统默认 · 审查结论与文献综述',
 JSON_OBJECT(
   'schema_version', 1,
   'artifact', 'audit_and_literature_review',
   'locale', 'zh',
   'description', 'audit JSON + literature_review；正文 content_medium 须遵循 markdown_outline 一级标题。',
   'output', JSON_OBJECT(
     'root_keys', JSON_ARRAY('audit', 'literature_review'),
     'audit', JSON_OBJECT(
       'fields', JSON_ARRAY('issues', 'summary', 'off_topic'),
       'issue_item_keys', JSON_ARRAY('severity', 'claim', 'fix'),
       'severity_enum', JSON_ARRAY('blocker', 'major', 'minor')
     ),
     'literature_review', JSON_OBJECT(
       'fields', JSON_ARRAY('title', 'summary', 'structure', 'content_medium', 'citations'),
       'structure_enum', JSON_ARRAY('thematic', 'chronological', 'method'),
       'citation_item_keys', JSON_ARRAY('ref_id', 'note')
     )
   ),
   'markdown_outline', JSON_OBJECT(
     'rules', 'content_medium 使用 Markdown；须按下列 ## 标题顺序输出，不得增删或改写标题',
     'headings', JSON_ARRAY(
       JSON_OBJECT('level', 2, 'id', 'background', 'title', '研究背景与问题定义'),
       JSON_OBJECT('level', 2, 'id', 'methods_landscape', 'title', '核心概念与方法谱系'),
       JSON_OBJECT('level', 2, 'id', 'related_work', 'title', '与本文方向相关的代表工作'),
       JSON_OBJECT('level', 2, 'id', 'gaps', 'title', '研究缺口与开放问题'),
       JSON_OBJECT('level', 2, 'id', 'synthesis', 'title', '与候选选题的关联及综述小结')
     )
   )
 ), 'system', 'active'),
(0, 0, 'experiment_plan', '系统默认 · 实验方案',
 JSON_OBJECT(
   'schema_version', 1,
   'artifact', 'experiment_plan',
   'locale', 'zh',
   'description', 'experiment_plan JSON；content_medium 与 meta 字段须与 markdown_outline 章节一致。',
   'output', JSON_OBJECT(
     'root_key', 'experiment_plan',
     'fields', JSON_ARRAY('title', 'summary', 'content_medium', 'meta'),
     'meta', JSON_OBJECT(
       'hypothesis', JSON_OBJECT('type', 'string', 'label_zh', '核心假设'),
       'baselines', JSON_OBJECT('type', 'string[]', 'label_zh', '对照基线'),
       'metrics', JSON_OBJECT('type', 'string[]', 'label_zh', '评价指标'),
       'ablations', JSON_OBJECT('type', 'string[]', 'label_zh', '消融项'),
       'timeline', JSON_OBJECT('type', 'string[]', 'label_zh', '时间线')
     )
   ),
   'markdown_outline', JSON_OBJECT(
     'rules', 'content_medium 使用 Markdown；须按下列 ## 标题顺序输出，不得增删或改写标题',
     'headings', JSON_ARRAY(
       JSON_OBJECT('level', 2, 'id', 'hypothesis', 'title', '核心假设与研究问题'),
       JSON_OBJECT('level', 2, 'id', 'main_experiment', 'title', '主实验设计'),
       JSON_OBJECT('level', 2, 'id', 'baselines', 'title', '对照基线'),
       JSON_OBJECT('level', 2, 'id', 'metrics', 'title', '评价指标与统计检验'),
       JSON_OBJECT('level', 2, 'id', 'ablations', 'title', '消融与敏感性分析'),
       JSON_OBJECT('level', 2, 'id', 'timeline', 'title', '数据、算力与实施时间线')
     )
   )
 ), 'system', 'active')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `outline_template` = VALUES(`outline_template`),
  `source` = VALUES(`source`),
  `status` = VALUES(`status`);
