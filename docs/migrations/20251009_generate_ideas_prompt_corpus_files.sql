UPDATE `paper_llm_prompt_template`
SET
  `template_body` = '环节：脑暴候选选题。最多提出 {{max_ideas}} 条可检验的论文方向。仅依据用户消息 Corpus（文献元数据/摘要/retrieve 简报）与 CorpusFiles（已落盘 PDF 与 external_key 对应关系）；每条 idea 的 reference_keys 须来自上述范围，不得引入库外文献。无 PDF 的篇目仅能用 Corpus 摘要/标题并在 idea 中注明依据级别。具体 JSON 字段以用户消息末尾为准。',
  `updated_at` = CURRENT_TIMESTAMP
WHERE `stage_code` = 'generate_ideas' AND `status` = 'active';
