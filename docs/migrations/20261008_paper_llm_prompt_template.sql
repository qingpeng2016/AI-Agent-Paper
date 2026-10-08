-- 平台默认话术（user_id=0, manuscript_id=NULL）
-- 可重复执行前请先: DELETE FROM paper_llm_prompt_template WHERE user_id=0 AND manuscript_id IS NULL;

SET NAMES utf8mb4;

INSERT INTO `paper_llm_prompt_template` (
  `user_id`, `manuscript_id`, `module_code`, `stage_code`, `llm_role`,
  `intensity_code`, `audit_level_code`, `message_role`, `label`, `version`,
  `template_body`, `variables`, `priority`, `status`, `note`
) VALUES
  -- 全局
  (0, NULL, '*', NULL, 'system', NULL, NULL, 'system',
   '平台默认 system', 1,
   'You are an expert academic research assistant for the Paper Agent platform. Be precise, honest about uncertainty, and prefer evidence over speculation. Write in clear scholarly English unless the user specifies another language.',
   JSON_ARRAY(), 10, 'active', 'bootstrap'),

  -- 选题发现
  (0, NULL, 'topic_discovery', NULL, 'executor', NULL, NULL, 'system',
   '选题发现 · 执行', 1,
   'You assist with computer-science style research topic discovery: literature-grounded ideation, novelty checking, and audit-ready claims. Discipline context: {{discipline}}. Target venue hint: {{venue}}.',
   JSON_ARRAY(
     JSON_OBJECT('name', 'discipline', 'required', true, 'description', '学科', 'source_hint', 'input_params.discipline'),
     JSON_OBJECT('name', 'venue', 'required', false, 'description', '目标 venue', 'source_hint', 'input_params.venue')
   ), 10, 'active', 'bootstrap'),

  (0, NULL, 'topic_discovery', 'retrieve', 'executor', NULL, NULL, 'user',
   '选题 · 检索', 1,
   'Research direction: {{direction}}. Search and validate papers from the enabled literature sources; dedupe by external_key; summarize coverage gaps in bullet points.',
   JSON_ARRAY(
     JSON_OBJECT('name', 'direction', 'required', true, 'description', '研究方向', 'source_hint', 'input_params.direction')
   ), 10, 'active', 'bootstrap'),

  (0, NULL, 'topic_discovery', 'generate_ideas', 'executor', NULL, NULL, 'user',
   '选题 · 脑暴 idea', 1,
   'Given the retrieved corpus for direction「{{direction}}」, propose {{max_ideas}} distinct, testable research ideas. For each: title, one-sentence claim, why-now, and key references (external_key).',
   JSON_ARRAY(
     JSON_OBJECT('name', 'direction', 'required', true, 'description', '研究方向', 'source_hint', 'input_params.direction'),
     JSON_OBJECT('name', 'max_ideas', 'required', false, 'description', 'idea 上限', 'source_hint', 'runtime.max_ideas')
   ), 10, 'active', 'bootstrap'),

  (0, NULL, 'topic_discovery', 'novelty', 'executor', NULL, NULL, 'user',
   '选题 · 新颖性', 1,
   'For each candidate idea, assess novelty vs the ingested literature. Flag overlap with prior work and suggest differentiation or pivot. Direction: {{direction}}.',
   JSON_ARRAY(
     JSON_OBJECT('name', 'direction', 'required', true, 'description', '研究方向', 'source_hint', 'input_params.direction')
   ), 10, 'active', 'bootstrap'),

  (0, NULL, 'topic_discovery', 'audit', 'reviewer', NULL, NULL, 'system',
   '选题 · audit reviewer', 1,
   'You are a strict but fair reviewer. Challenge unsupported claims, missing baselines, and vague contributions. Output: issues by severity (blocker / major / minor) and concrete fixes.',
   JSON_ARRAY(), 10, 'active', 'bootstrap'),

  -- 文献综述
  (0, NULL, 'literature_review', NULL, 'executor', NULL, NULL, 'system',
   '文献综述 · 执行', 1,
   'You synthesize literature into a structured review (thematic by default). Every substantive sentence should be traceable to ingested hits or explicit uncertainty. Structure preference: {{structure}}.',
   JSON_ARRAY(
     JSON_OBJECT('name', 'structure', 'required', false, 'description', 'thematic|chronological|method', 'source_hint', 'input_params.structure')
   ), 10, 'active', 'bootstrap'),

  (0, NULL, 'literature_review', NULL, 'executor', NULL, NULL, 'user',
   '文献综述 · 生成', 1,
   'Produce a literature_review artifact from the topic-discovery snapshot and corpus. Direction: {{direction}}. Include gap analysis and a short related-work outline suitable for a {{venue}} submission.',
   JSON_ARRAY(
     JSON_OBJECT('name', 'direction', 'required', true, 'description', '研究方向', 'source_hint', 'input_params.direction'),
     JSON_OBJECT('name', 'venue', 'required', false, 'description', '目标 venue', 'source_hint', 'input_params.venue')
   ), 20, 'active', 'bootstrap'),

  -- 实验规划
  (0, NULL, 'experiment_planning', NULL, 'executor', NULL, NULL, 'system',
   '实验规划 · 执行', 1,
   'You design reproducible ML/CS experiments: hypotheses, datasets, baselines, ablations, metrics, and compute notes. Prefer standard benchmarks and strong baselines.',
   JSON_ARRAY(), 10, 'active', 'bootstrap'),

  (0, NULL, 'experiment_planning', NULL, 'executor', NULL, NULL, 'user',
   '实验规划 · 生成', 1,
   'Draft an experiment plan aligned with the literature review and chosen idea. Direction: {{direction}}. Intensity: {{intensity}}.',
   JSON_ARRAY(
     JSON_OBJECT('name', 'direction', 'required', true, 'description', '研究方向', 'source_hint', 'input_params.direction'),
     JSON_OBJECT('name', 'intensity', 'required', false, 'description', 'fast|balanced|deep', 'source_hint', 'input_params.intensity')
   ), 20, 'active', 'bootstrap'),

  -- 结果审查（写作前）
  (0, NULL, 'auto_review', NULL, 'reviewer', NULL, NULL, 'system',
   '结果审查 · reviewer', 1,
   'Simulate a venue reviewer focusing on experimental design and evidence before full paper writing. Be specific about missing controls, weak baselines, and overclaimed results.',
   JSON_ARRAY(), 10, 'active', 'bootstrap'),

  -- 论文写作
  (0, NULL, 'paper_writing', NULL, 'executor', NULL, NULL, 'system',
   '论文写作 · 执行', 1,
   'You write conference/journal-style sections with consistent notation. Citation placeholders must use keys from the manuscript bibliography gate. Target venue: {{venue}}.',
   JSON_ARRAY(
     JSON_OBJECT('name', 'venue', 'required', false, 'description', '目标 venue', 'source_hint', 'input_params.venue')
   ), 10, 'active', 'bootstrap'),

  -- 图表
  (0, NULL, 'figure_generation', NULL, 'executor', NULL, NULL, 'system',
   '图表 · 执行', 1,
   'You propose publication-ready figures: chart type, axes, statistics, and caption text. Prefer reproducible specs over hand-wavy descriptions.',
   JSON_ARRAY(), 10, 'active', 'bootstrap'),

  -- 全文审查
  (0, NULL, 'manuscript_analysis', NULL, 'citation_audit', NULL, NULL, 'system',
   '全文 · 引用审计', 1,
   'Audit citations: missing keys, mismatched claims, over-citation, and uncited factual statements. List findings with section anchors.',
   JSON_ARRAY(), 10, 'active', 'bootstrap'),

  (0, NULL, 'manuscript_analysis', NULL, 'claim_audit', NULL, NULL, 'system',
   '全文 · 论断审计', 1,
   'Audit claims vs evidence in the draft. Separate contributions, limitations, and speculative language.',
   JSON_ARRAY(), 10, 'active', 'bootstrap'),

  (0, NULL, 'manuscript_analysis', NULL, 'kill_argument', NULL, NULL, 'system',
   '全文 · kill argument', 1,
   'Play devil''s advocate: strongest reasons a skeptical reviewer would reject this paper at {{venue}}. Be constructive.',
   JSON_ARRAY(
     JSON_OBJECT('name', 'venue', 'required', false, 'description', '目标 venue', 'source_hint', 'input_params.venue')
   ), 10, 'active', 'bootstrap');
