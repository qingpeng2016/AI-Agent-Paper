-- 平台 Claude + 基础 ref 种子（ref 段可重复执行）
-- paper_llm_model_config.api_key 填 Claude 明文 Key（勿重复 INSERT）
-- 话术: docs/migrations/20261008_paper_llm_prompt_template.sql

SET NAMES utf8mb4;

-- ---------------------------------------------------------------------------
-- ref（文献源 / 强度 / 审计 / 学科）
-- ---------------------------------------------------------------------------

INSERT INTO `paper_ref_execution_intensity` (`code`, `name`, `multiplier`, `max_papers`, `max_ideas`) VALUES
  ('fast', '更快', 0.60, 30, 6),
  ('balanced', 'Balanced（平衡）', 1.00, 80, 12),
  ('deep', '更深', 1.80, 200, 20)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`), `multiplier` = VALUES(`multiplier`),
  `max_papers` = VALUES(`max_papers`), `max_ideas` = VALUES(`max_ideas`);

INSERT INTO `paper_ref_audit_level` (
  `code`, `name`, `citation_strength`, `claim_strength`, `kill_argument_strength`, `audit_rounds`
) VALUES
  ('standard', 'Standard', 1, 1, 0, 1),
  ('polished', 'Polished（精修）', 2, 2, 1, 2),
  ('strict', 'Strict', 3, 3, 2, 3)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

INSERT INTO `paper_ref_discipline` (`code`, `name`, `name_en`, `sort`, `literature_source_codes`) VALUES
  ('cs_ai', '计算机/人工智能', 'Computer Science & AI', 10,
   JSON_ARRAY('arxiv', 'openalex', 'semantic_scholar')),
  ('general', '跨学科通用', 'General', 0, JSON_ARRAY('openalex', 'crossref'))
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`), `literature_source_codes` = VALUES(`literature_source_codes`);

INSERT INTO `paper_ref_literature_source` (`code`, `name`, `api_kind`, `base_url`, `auth_type`, `default_config`, `config_schema`) VALUES
  ('arxiv', 'arXiv', 'rest', 'https://export.arxiv.org/api/query', 'none', NULL, NULL),
  ('openalex', 'OpenAlex', 'rest', 'https://api.openalex.org/works', 'none',
   JSON_OBJECT('mailto', ''),
   JSON_OBJECT('type', 'object', 'properties', JSON_OBJECT('mailto', JSON_OBJECT('type', 'string')))),
  ('semantic_scholar', 'Semantic Scholar', 'rest', 'https://api.semanticscholar.org/graph/v1/paper/search', 'api_key',
   NULL,
   JSON_OBJECT('type', 'object', 'properties', JSON_OBJECT('api_key', JSON_OBJECT('type', 'string'))))
ON DUPLICATE KEY UPDATE
  `base_url` = VALUES(`base_url`), `default_config` = VALUES(`default_config`);

-- ---------------------------------------------------------------------------
-- LLM：Claude 端点 + 全模块默认 binding + 占位 system 话术
-- ---------------------------------------------------------------------------

INSERT INTO `paper_llm_model_config` (
  `label`, `provider_code`, `model_name`, `api_base_url`,
  `api_key`, `timeout_ms`, `max_retries`,
  `supports_vision`, `context_window_hint`, `status`
) VALUES (
  'Claude Sonnet（平台默认）',
  'anthropic',
  'claude-sonnet-4-20250514',
  'https://api.anthropic.com',
  'PASTE_YOUR_ANTHROPIC_API_KEY_HERE',
  120000,
  2,
  1,
  200000,
  'active'
);

SET @claude_model_id = LAST_INSERT_ID();

INSERT INTO `paper_llm_workflow_binding` (
  `module_code`, `stage_code`, `llm_role`, `model_config_id`, `priority`, `status`
) VALUES
  ('*', NULL, 'executor', @claude_model_id, 10, 'active');

-- 话术模板见: docs/migrations/20261008_paper_llm_prompt_template.sql
