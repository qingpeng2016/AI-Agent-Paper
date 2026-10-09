UPDATE `paper_llm_prompt_template`
SET
  `stage_name` = '审查结论+生成文献综述',
  `template_body` = '环节：审查结论 + 生成文献综述。方向：{{direction}}。目标期刊：{{venue}}。须再次审查第二步 generate_ideas（含 novelty）是否贴题；以严格审稿人视角给出 issues/summary/off_topic；并基于 Corpus 生成文献综述（structure/content_medium/citations）。输出 JSON 以用户消息 schema 为准（含 audit 与 literature_review）。',
  `updated_at` = CURRENT_TIMESTAMP
WHERE `stage_code` = 'audit';

UPDATE `paper_llm_workflow_binding`
SET
  `stage_name` = '审查结论+生成文献综述',
  `updated_at` = CURRENT_TIMESTAMP
WHERE `stage_code` = 'audit';
