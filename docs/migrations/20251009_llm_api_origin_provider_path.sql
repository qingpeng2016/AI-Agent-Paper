-- api_base_url 统一为 origin（scheme://host）；path 由 provider_code 在代码中决定。
-- Gemini 请用 provider_code=google，勿再用 openai_compatible + 带 /v1beta/openai 的 base。

UPDATE `paper_llm_model_config`
SET
  `provider_code` = 'google',
  `api_base_url` = 'https://generativelanguage.googleapis.com',
  `model_name` = 'gemini-3.8-flash'
WHERE `label` LIKE '%Gemini%' AND `status` = 'active';

UPDATE `paper_llm_model_config`
SET `api_base_url` = 'https://api.anthropic.com'
WHERE `provider_code` = 'anthropic' AND `status` = 'active';

UPDATE `paper_llm_model_config`
SET `api_base_url` = 'https://api.anthropic.com'
WHERE `provider_code` = 'anthropic' AND `api_base_url` LIKE 'https://api.anthropic.com/%';
