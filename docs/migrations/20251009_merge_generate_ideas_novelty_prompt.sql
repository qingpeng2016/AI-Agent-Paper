UPDATE `paper_llm_prompt_template`
SET
  `template_body` = '环节：脑暴 + 新颖性（与 novelty 同一次模型调用完成）。最多 {{max_ideas}} 条 idea；仅依据用户消息 Corpus 与 CorpusFiles；输出 JSON 含 ideas 与 novelty（lines/risks/synthesis），以用户消息 schema 为准。',
  `updated_at` = CURRENT_TIMESTAMP
WHERE `stage_code` = 'generate_ideas' AND `status` = 'active';

UPDATE `paper_llm_prompt_template`
SET
  `template_body` = '（通常与 generate_ideas 合并为一次调用；本模板仅用于旧 run 单独补跑 novelty。）方向 {{direction}}：对照 Corpus 比较每条 idea，输出 lines/risks。',
  `updated_at` = CURRENT_TIMESTAMP
WHERE `stage_code` = 'novelty' AND `status` = 'active';
