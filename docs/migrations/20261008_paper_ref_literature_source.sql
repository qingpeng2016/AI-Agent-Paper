-- 文献数据源：替代 conf 中 literature.arxiv / openalex / semantic_scholar
-- 执行前需已存在表 paper_ref_literature_source（见 docs/all.sql）

INSERT INTO `paper_ref_literature_source` (`code`, `name`, `api_kind`, `base_url`, `auth_type`, `default_config`, `config_schema`) VALUES
  ('arxiv', 'arXiv', 'rest', 'https://export.arxiv.org/api/query', 'none', NULL, NULL),
  ('openalex', 'OpenAlex', 'rest', 'https://api.openalex.org/works', 'none',
   JSON_OBJECT('mailto', ''),
   JSON_OBJECT('type', 'object', 'properties', JSON_OBJECT('mailto', JSON_OBJECT('type', 'string', 'description', 'User-Agent 礼貌池')))),
  ('semantic_scholar', 'Semantic Scholar', 'rest', 'https://api.semanticscholar.org/graph/v1/paper/search', 'api_key',
   NULL,
   JSON_OBJECT('type', 'object', 'properties', JSON_OBJECT('api_key', JSON_OBJECT('type', 'string', 'description', 'Header x-api-key；可选'))))
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `base_url` = VALUES(`base_url`),
  `auth_type` = VALUES(`auth_type`),
  `default_config` = VALUES(`default_config`),
  `config_schema` = VALUES(`config_schema`);
