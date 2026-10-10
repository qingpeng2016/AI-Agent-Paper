-- 三表模板改为中/英双字段；已有库会 DROP 模板表后重建（仅系统 seed，用户自定义行请先备份）
DROP TABLE IF EXISTS `paper_template_document`;
DROP TABLE IF EXISTS `paper_template_venue`;

CREATE TABLE `paper_template_document` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板 ID',
  `user_id`             BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=系统；>0=用户自定义',
  `venue_id`            BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=默认通用；>0=paper_ref_venue.id',
  `stage_code`          VARCHAR(64)  NOT NULL COMMENT 'generate_ideas（含 novelty）|audit|experiment_plan|…',
  `name`                VARCHAR(128) NOT NULL COMMENT '模板展示名',
  `outline_template_zh` JSON         NOT NULL COMMENT '中文文档 JSON 提纲',
  `outline_template_en` JSON         NOT NULL COMMENT '英文文档 JSON 提纲',
  `storage_uri`         VARCHAR(1024) DEFAULT NULL COMMENT '用户上传源文件相对 storage/ 路径',
  `source`              VARCHAR(16)  NOT NULL DEFAULT 'system' COMMENT 'system|upload|edit',
  `status`              VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled|deleted',
  `created_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_template_document_scope` (`user_id`, `venue_id`, `stage_code`),
  KEY `idx_paper_template_document_lookup` (`stage_code`, `venue_id`, `user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='各阶段文档 JSON 提纲模板（中英）';

CREATE TABLE `paper_template_venue` (
  `id`                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板 ID',
  `user_id`               BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=系统；>0=用户自定义',
  `venue_id`              BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=默认通用；>0=paper_ref_venue.id',
  `name`                  VARCHAR(128) NOT NULL COMMENT '模板展示名',
  `constraint_template_zh` JSON       NOT NULL COMMENT '中文 venue 约束 JSON',
  `constraint_template_en` JSON       NOT NULL COMMENT '英文 venue 约束 JSON',
  `storage_uri`           VARCHAR(1024) DEFAULT NULL,
  `source`                VARCHAR(16)  NOT NULL DEFAULT 'system',
  `status`                VARCHAR(16)  NOT NULL DEFAULT 'active',
  `created_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_template_venue_scope` (`user_id`, `venue_id`),
  KEY `idx_paper_template_venue_lookup` (`venue_id`, `user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='venue 约束模板（中英）';

-- paper_template_llm_prompt：不存在则整表创建（中英双字段）；保留 stage_code 供 paper_llm_call_log 引用
CREATE TABLE IF NOT EXISTS `paper_template_llm_prompt` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板 ID',
  `stage_code`        VARCHAR(64)  NOT NULL DEFAULT 'default' COMMENT '环节键：default|retrieve|generate_ideas|…',
  `stage_name`        VARCHAR(128) NOT NULL COMMENT '环节中文名（运营/排查）',
  `template_body_zh`  MEDIUMTEXT   NOT NULL COMMENT '中文话术；占位符 {{var_name}}',
  `template_body_en`  MEDIUMTEXT   NOT NULL COMMENT '英文话术；占位符 {{var_name}}',
  `status`            VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT 'active|disabled',
  `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_template_llm_prompt_stage` (`stage_code`),
  KEY `idx_paper_template_llm_prompt_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台话术（按 stage_code，中英）';

-- 若你仍保留旧表且仅有 template_body 列，请先备份后执行：
-- ALTER TABLE `paper_template_llm_prompt` ADD COLUMN `template_body_zh` ... ADD COLUMN `template_body_en` ...;
-- UPDATE ... SET template_body_zh = template_body; ALTER ... DROP COLUMN `template_body`;

-- ----- seed: paper_template_document (user_id=0, venue_id=0) -----
INSERT INTO `paper_template_document` (
  `user_id`, `venue_id`, `stage_code`, `name`,
  `outline_template_zh`, `outline_template_en`, `source`, `status`
) VALUES
(0, 0, 'generate_ideas', '系统默认 · 脑暴选题与新颖性',
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'zh', 'artifact', 'ideas_and_novelty',
   'description', '一次 LLM 输出 ideas 与 novelty；字段键名与顺序须与本模板一致。',
   'output', JSON_OBJECT(
     'root_keys', JSON_ARRAY('ideas', 'novelty'),
     'ideas', JSON_OBJECT('max_items_param', 'max_ideas', 'item_fields', JSON_ARRAY(
       JSON_OBJECT('key', 'title', 'label_zh', '选题标题', 'required', true),
       JSON_OBJECT('key', 'problem', 'label_zh', '问题与动机', 'required', true),
       JSON_OBJECT('key', 'approach', 'label_zh', '技术路线', 'required', true),
       JSON_OBJECT('key', 'contribution', 'label_zh', '预期贡献', 'required', true, 'hint', '对齐目标 venue 的贡献类型'),
       JSON_OBJECT('key', 'reference_keys', 'label_zh', '关键文献', 'required', true, 'type', 'external_key[]')
     )),
     'novelty', JSON_OBJECT('fields', JSON_ARRAY(
       JSON_OBJECT('key', 'lines', 'label_zh', '新颖性要点', 'type', 'string[]'),
       JSON_OBJECT('key', 'risks', 'label_zh', '逐题风险', 'type', 'array', 'item_keys', JSON_ARRAY('idea_title', 'risk', 'note', 'overlap_refs')),
       JSON_OBJECT('key', 'risk', 'enum', JSON_ARRAY('low', 'medium', 'high')),
       JSON_OBJECT('key', 'synthesis', 'label_zh', '综合结论', 'type', 'string')
     ))
   )
 ),
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'en', 'artifact', 'ideas_and_novelty',
   'description', 'Single LLM output for ideas and novelty; keep field keys and order exactly as this template.',
   'output', JSON_OBJECT(
     'root_keys', JSON_ARRAY('ideas', 'novelty'),
     'ideas', JSON_OBJECT('max_items_param', 'max_ideas', 'item_fields', JSON_ARRAY(
       JSON_OBJECT('key', 'title', 'label_en', 'Idea title', 'required', true),
       JSON_OBJECT('key', 'problem', 'label_en', 'Problem & motivation', 'required', true),
       JSON_OBJECT('key', 'approach', 'label_en', 'Approach', 'required', true),
       JSON_OBJECT('key', 'contribution', 'label_en', 'Expected contribution', 'required', true, 'hint', 'Align with target venue contribution types'),
       JSON_OBJECT('key', 'reference_keys', 'label_en', 'Key references', 'required', true, 'type', 'external_key[]')
     )),
     'novelty', JSON_OBJECT('fields', JSON_ARRAY(
       JSON_OBJECT('key', 'lines', 'label_en', 'Novelty bullets', 'type', 'string[]'),
       JSON_OBJECT('key', 'risks', 'label_en', 'Per-idea risks', 'type', 'array', 'item_keys', JSON_ARRAY('idea_title', 'risk', 'note', 'overlap_refs')),
       JSON_OBJECT('key', 'risk', 'enum', JSON_ARRAY('low', 'medium', 'high')),
       JSON_OBJECT('key', 'synthesis', 'label_en', 'Synthesis', 'type', 'string')
     ))
   )
 ), 'system', 'active'),
(0, 0, 'audit', '系统默认 · 审查结论与文献综述',
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'zh', 'artifact', 'audit_and_literature_review',
   'description', 'audit JSON + literature_review；正文 content_medium 须遵循 markdown_outline 一级标题。',
   'output', JSON_OBJECT(
     'root_keys', JSON_ARRAY('audit', 'literature_review'),
     'audit', JSON_OBJECT('fields', JSON_ARRAY('issues', 'summary', 'off_topic'),
       'issue_item_keys', JSON_ARRAY('severity', 'claim', 'fix'),
       'severity_enum', JSON_ARRAY('blocker', 'major', 'minor')),
     'literature_review', JSON_OBJECT(
       'fields', JSON_ARRAY('title', 'summary', 'structure', 'content_medium', 'citations'),
       'structure_enum', JSON_ARRAY('thematic', 'chronological', 'method'),
       'citation_item_keys', JSON_ARRAY('ref_id', 'note'))
   ),
   'markdown_outline', JSON_OBJECT(
     'rules', 'content_medium 使用 Markdown；须按下列 ## 标题顺序输出，不得增删或改写标题',
     'headings', JSON_ARRAY(
       JSON_OBJECT('level', 2, 'id', 'background', 'title', '研究背景与问题定义'),
       JSON_OBJECT('level', 2, 'id', 'methods_landscape', 'title', '核心概念与方法谱系'),
       JSON_OBJECT('level', 2, 'id', 'related_work', 'title', '与本文方向相关的代表工作'),
       JSON_OBJECT('level', 2, 'id', 'gaps', 'title', '研究缺口与开放问题'),
       JSON_OBJECT('level', 2, 'id', 'synthesis', 'title', '与候选选题的关联及综述小结')
     ))
 ),
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'en', 'artifact', 'audit_and_literature_review',
   'description', 'audit JSON + literature_review; content_medium must follow markdown_outline H2 titles.',
   'output', JSON_OBJECT(
     'root_keys', JSON_ARRAY('audit', 'literature_review'),
     'audit', JSON_OBJECT('fields', JSON_ARRAY('issues', 'summary', 'off_topic'),
       'issue_item_keys', JSON_ARRAY('severity', 'claim', 'fix'),
       'severity_enum', JSON_ARRAY('blocker', 'major', 'minor')),
     'literature_review', JSON_OBJECT(
       'fields', JSON_ARRAY('title', 'summary', 'structure', 'content_medium', 'citations'),
       'structure_enum', JSON_ARRAY('thematic', 'chronological', 'method'),
       'citation_item_keys', JSON_ARRAY('ref_id', 'note'))
   ),
   'markdown_outline', JSON_OBJECT(
     'rules', 'content_medium in Markdown; use exactly these ## headings in order; do not rename or omit',
     'headings', JSON_ARRAY(
       JSON_OBJECT('level', 2, 'id', 'background', 'title', 'Background and problem definition'),
       JSON_OBJECT('level', 2, 'id', 'methods_landscape', 'title', 'Concepts and method landscape'),
       JSON_OBJECT('level', 2, 'id', 'related_work', 'title', 'Representative related work'),
       JSON_OBJECT('level', 2, 'id', 'gaps', 'title', 'Research gaps and open problems'),
       JSON_OBJECT('level', 2, 'id', 'synthesis', 'title', 'Link to candidate ideas and summary')
     ))
 ), 'system', 'active'),
(0, 0, 'experiment_plan', '系统默认 · 实验方案',
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'zh', 'artifact', 'experiment_plan',
   'description', 'experiment_plan JSON；content_medium 与 meta 字段须与 markdown_outline 章节一致。',
   'output', JSON_OBJECT('root_key', 'experiment_plan',
     'fields', JSON_ARRAY('title', 'summary', 'content_medium', 'meta'),
     'meta', JSON_OBJECT(
       'hypothesis', JSON_OBJECT('type', 'string', 'label_zh', '核心假设'),
       'baselines', JSON_OBJECT('type', 'string[]', 'label_zh', '对照基线'),
       'metrics', JSON_OBJECT('type', 'string[]', 'label_zh', '评价指标'),
       'ablations', JSON_OBJECT('type', 'string[]', 'label_zh', '消融项'),
       'timeline', JSON_OBJECT('type', 'string[]', 'label_zh', '时间线')
     )),
   'markdown_outline', JSON_OBJECT(
     'rules', 'content_medium 使用 Markdown；须按下列 ## 标题顺序输出，不得增删或改写标题',
     'headings', JSON_ARRAY(
       JSON_OBJECT('level', 2, 'id', 'hypothesis', 'title', '核心假设与研究问题'),
       JSON_OBJECT('level', 2, 'id', 'main_experiment', 'title', '主实验设计'),
       JSON_OBJECT('level', 2, 'id', 'baselines', 'title', '对照基线'),
       JSON_OBJECT('level', 2, 'id', 'metrics', 'title', '评价指标与统计检验'),
       JSON_OBJECT('level', 2, 'id', 'ablations', 'title', '消融与敏感性分析'),
       JSON_OBJECT('level', 2, 'id', 'timeline', 'title', '数据、算力与实施时间线')
     ))
 ),
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'en', 'artifact', 'experiment_plan',
   'description', 'experiment_plan JSON; content_medium and meta must match markdown_outline sections.',
   'output', JSON_OBJECT('root_key', 'experiment_plan',
     'fields', JSON_ARRAY('title', 'summary', 'content_medium', 'meta'),
     'meta', JSON_OBJECT(
       'hypothesis', JSON_OBJECT('type', 'string', 'label_en', 'Core hypothesis'),
       'baselines', JSON_OBJECT('type', 'string[]', 'label_en', 'Baselines'),
       'metrics', JSON_OBJECT('type', 'string[]', 'label_en', 'Metrics'),
       'ablations', JSON_OBJECT('type', 'string[]', 'label_en', 'Ablations'),
       'timeline', JSON_OBJECT('type', 'string[]', 'label_en', 'Timeline')
     )),
   'markdown_outline', JSON_OBJECT(
     'rules', 'content_medium in Markdown; use exactly these ## headings in order',
     'headings', JSON_ARRAY(
       JSON_OBJECT('level', 2, 'id', 'hypothesis', 'title', 'Hypothesis and research questions'),
       JSON_OBJECT('level', 2, 'id', 'main_experiment', 'title', 'Main experiment design'),
       JSON_OBJECT('level', 2, 'id', 'baselines', 'title', 'Baselines'),
       JSON_OBJECT('level', 2, 'id', 'metrics', 'title', 'Metrics and statistical testing'),
       JSON_OBJECT('level', 2, 'id', 'ablations', 'title', 'Ablations and sensitivity'),
       JSON_OBJECT('level', 2, 'id', 'timeline', 'title', 'Data, compute, and schedule')
     ))
 ), 'system', 'active');

INSERT INTO `paper_template_venue` (
  `user_id`, `venue_id`, `name`, `constraint_template_zh`, `constraint_template_en`, `source`, `status`
) VALUES
(0, 0, '系统默认 · 通用 venue',
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'zh',
   'description', '未指定或未匹配 paper_ref_venue 时使用；约束选题、综述与实验方案的贡献形态与实验深度。',
   'contribution_types', JSON_OBJECT(
     'allowed', JSON_ARRAY('method', 'theory', 'empirical', 'survey', 'system'),
     'labels_zh', JSON_OBJECT('method', '方法/模型', 'theory', '理论/分析', 'empirical', '实证/实验', 'survey', '综述/梳理', 'system', '系统/工具'),
     'prompt_hint', '每条 idea 的 contribution 须明确属于 allowed 之一，并与方向一致；勿混合未声明的类型。'),
   'experiment_bar', JSON_OBJECT(
     'code', 'standard', 'label_zh', '常规范实验',
     'requirements', JSON_ARRAY('至少一个可复现主实验', '列出可对照基线（名称具体）', '指标与数据集可追踪', '建议说明算力/数据规模'),
     'prompt_hint', '实验方案须写清基线、指标与主结论所依赖的证据；勿夸大未设计的实验。'),
   'writing_style', JSON_OBJECT('code', 'scholarly_generic', 'label_zh', '通用学术表述',
     'prompt_hint', '中文表述准确、克制；避免营销式用语；引用须对应 Corpus ref_id。')
 ),
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'en',
   'description', 'Fallback when venue is unspecified; constrains contribution shape and experiment depth.',
   'contribution_types', JSON_OBJECT(
     'allowed', JSON_ARRAY('method', 'theory', 'empirical', 'survey', 'system'),
     'labels_en', JSON_OBJECT('method', 'Method/model', 'theory', 'Theory', 'empirical', 'Empirical', 'survey', 'Survey', 'system', 'System/tool'),
     'prompt_hint', 'Each idea contribution must map to one allowed type and match the direction.'),
   'experiment_bar', JSON_OBJECT(
     'code', 'standard', 'label_en', 'Standard experiments',
     'requirements', JSON_ARRAY('At least one reproducible main experiment', 'Name concrete baselines', 'Traceable metrics and datasets', 'State compute/data scale when possible'),
     'prompt_hint', 'Plan must tie claims to evidence; do not invent unplanned experiments.'),
   'writing_style', JSON_OBJECT('code', 'scholarly_generic', 'label_en', 'Generic scholarly prose',
     'prompt_hint', 'Precise, restrained English; citations must use Corpus ref_id.')
 ), 'system', 'active');

INSERT INTO `paper_template_venue` (
  `user_id`, `venue_id`, `name`, `constraint_template_zh`, `constraint_template_en`, `source`, `status`
)
SELECT 0, v.`id`, '系统默认 · NeurIPS/ICLR/ICML',
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'zh', 'ref_venue_code', 'ml_top3',
   'description', 'ML 顶会：方法/理论/实证贡献；强基线与消融；会议短文风格。',
   'contribution_types', JSON_OBJECT(
     'allowed', JSON_ARRAY('method', 'theory', 'empirical'),
     'labels_zh', JSON_OBJECT('method', '方法', 'theory', '理论', 'empirical', '实证'),
     'prompt_hint', '贡献须能在 8 页主会体量内讲清；method/theory/empirical 择主、可辅，勿堆叠无关贡献。'),
   'experiment_bar', JSON_OBJECT(
     'code', 'strong_baselines_ablations', 'label_zh', '强基线 + 消融',
     'requirements', JSON_ARRAY('SOTA 或公认强基线至少 2 个', '主结果建议 ≥3 random seed 或等价稳定性说明', '消融须对应核心设计选择', '报告算力/训练成本或吞吐等可审稿指标'),
     'prompt_hint', '缺基线须在 audit issues 标 major。'),
   'writing_style', JSON_OBJECT('code', 'ml_conference', 'label_zh', 'ML 会议短文',
     'prompt_hint', 'Related Work 聚焦差异；Experiments 先主结果后消融；claim 与表格/图一致。')
 ),
 JSON_OBJECT(
   'schema_version', 1, 'locale', 'en', 'ref_venue_code', 'ml_top3',
   'description', 'Top ML conferences: method/theory/empirical; strong baselines and ablations; short paper style.',
   'contribution_types', JSON_OBJECT(
     'allowed', JSON_ARRAY('method', 'theory', 'empirical'),
     'labels_en', JSON_OBJECT('method', 'Method', 'theory', 'Theory', 'empirical', 'Empirical'),
     'prompt_hint', 'Fit a ~8-page main-track claim; one primary contribution type, avoid unrelated stacking.'),
   'experiment_bar', JSON_OBJECT(
     'code', 'strong_baselines_ablations', 'label_en', 'Strong baselines + ablations',
     'requirements', JSON_ARRAY('≥2 strong/SOTA baselines', '≥3 seeds or equivalent stability', 'Ablations tied to core design', 'Report compute/throughput where relevant'),
     'prompt_hint', 'Missing baselines should be audit major issues.'),
   'writing_style', JSON_OBJECT('code', 'ml_conference', 'label_en', 'ML conference short paper',
     'prompt_hint', 'Related work: contrast; experiments: main results then ablations; claims match tables/figures.')
 ), 'system', 'active'
FROM `paper_ref_venue` v WHERE v.`code` = 'ml_top3' LIMIT 1;

DELETE FROM `paper_template_llm_prompt` WHERE `stage_code` IN ('default', 'generate_ideas', 'audit', 'experiment_plan', 'retrieve', 'novelty');

INSERT INTO `paper_template_llm_prompt` (`stage_code`, `stage_name`, `template_body_zh`, `template_body_en`, `status`) VALUES
('default', '默认（全局兜底）',
 '你是 Paper Agent 平台的学术科研助手。只根据用户消息中给出的文献与前置步骤结果作答，不得编造未出现的来源或实验。{{output_language_rule}} 若要求 JSON，严格遵守用户消息中的 schema。',
 'You are an academic research assistant for the Paper Agent platform. Answer only from literature and prior-step results in the user message; do not fabricate sources or experiments. {{output_language_rule}} When JSON is required, follow the schema in the user message.',
 'active'),
('generate_ideas', '脑暴候选选题',
 '环节：脑暴 + 新颖性（与 novelty 同一次模型调用完成）。最多 {{max_ideas}} 条 idea；仅依据用户消息 Corpus 与 CorpusFiles；输出 JSON 含 ideas 与 novelty（lines/risks/synthesis），以用户消息 schema 为准。{{output_language_rule}}',
 'Stage: brainstorming + novelty in one call. At most {{max_ideas}} ideas; use only Corpus and CorpusFiles from the user message. Output JSON with ideas and novelty (lines/risks/synthesis) per the user message schema. {{output_language_rule}}',
 'active'),
('audit', '审查结论+生成文献综述',
 '环节：审查结论 + 生成文献综述。方向：{{direction}}。目标期刊：{{venue}}。须再次审查第二步 generate_ideas（含 novelty）是否贴题；以严格审稿人视角给出 issues/summary/off_topic；并基于 Corpus 生成文献综述（structure/content_medium/citations）。输出 JSON 以用户消息 schema 为准（含 audit 与 literature_review）。{{output_language_rule}}',
 'Stage: audit conclusion + literature review. Direction: {{direction}}. Target venue: {{venue}}. Re-check generate_ideas (with novelty) for fit; reviewer-style issues/summary/off_topic; literature review from Corpus (structure/content_medium/citations). Output JSON per user schema (audit + literature_review). {{output_language_rule}}',
 'active'),
('experiment_plan', '生成实验方案',
 '环节：生成实验方案（experiment planning）。方向：{{direction}}。目标期刊/会议：{{venue}}。须以用户消息中的 LiteratureReview（文献综述全文/摘要）与选题 Corpus 为依据，不得编造未给出的基线或数据。输出可审稿的实验计划：核心假设、主实验与对照基线、评价指标、消融设计、算力/数据资源与时间线（写入 paper_output_experiment_plan：title、summary、content_medium、meta）。严格 JSON，字段以用户消息 schema 为准（含 experiment_plan）。{{output_language_rule}}',
 'Stage: experiment planning. Direction: {{direction}}. Target venue: {{venue}}. Use LiteratureReview and topic Corpus from the user message only; do not invent baselines or data. Output a review-ready plan (hypothesis, main experiment, baselines, metrics, ablations, resources/timeline) as experiment_plan JSON per user schema. {{output_language_rule}}',
 'active');
