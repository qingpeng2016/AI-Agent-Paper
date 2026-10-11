-- paper_llm_constraint_gate：按 scene 分 topic / write_paper
-- paper_llm_description：gate 文案 10000–10008（topic）+ 11000–11008（write_paper）

ALTER TABLE `paper_llm_description`
  ADD COLUMN `scene` VARCHAR(32) NOT NULL DEFAULT 'shared' COMMENT 'topic|write_paper|shared' AFTER `description_code`;

UPDATE `paper_llm_description`
SET `scene` = 'topic'
WHERE `prompt_code` = 'constraint_gate' AND `description_code` BETWEEN 10000 AND 10008;

ALTER TABLE `paper_llm_constraint_gate`
  ADD COLUMN `scene` VARCHAR(32) NOT NULL DEFAULT 'topic' COMMENT 'topic|write_paper' FIRST;

UPDATE `paper_llm_constraint_gate` SET `scene` = 'topic';

ALTER TABLE `paper_llm_constraint_gate` DROP PRIMARY KEY;
ALTER TABLE `paper_llm_constraint_gate` ADD PRIMARY KEY (`scene`, `code`);

INSERT INTO `paper_llm_description` (`description_code`, `scene`, `model_config_id`, `prompt_code`, `prompt_type`, `prompt_name`, `llm_desc_zh`, `llm_desc_en`) VALUES
(11000, 'write_paper', 1, 'constraint_gate', '门禁约束', '引用门禁·基础·写作',
 '引用门禁（基础）：正文 citations/ref_id 须对应 verified Corpus；禁止编造文献；无法溯源则 blocker。',
 'Citation (basic, manuscript): in-text refs must map to verified Corpus; no fabrication; missing traceability is blocker.'),
(11001, 'write_paper', 1, 'constraint_gate', '门禁约束', '引用门禁·标准·写作',
 '引用门禁（标准）：含基础项；Related Work/Method/Experiment 等关键句须有 citations；orphan ref 为 major。',
 'Citation (standard, manuscript): includes basic; key sections need citations; orphan refs major.'),
(11002, 'write_paper', 1, 'constraint_gate', '门禁约束', '引用门禁·严格·写作',
 '引用门禁（严格）：含标准项；逐段抽查全文引用是否充分、是否 over-cite 无关文献。',
 'Citation (strict, manuscript): includes standard; section-level citation adequacy across full draft.'),
(11003, 'write_paper', 1, 'constraint_gate', '门禁约束', '论断门禁·基础·写作',
 '论断门禁（基础）：摘要与正文 claim 须可被 Corpus/实验结果支撑；禁止无证据的 SOTA/首创表述。',
 'Claim (basic, manuscript): abstract and body claims supportable; no unsupported SOTA/first claims.'),
(11004, 'write_paper', 1, 'constraint_gate', '门禁约束', '论断门禁·标准·写作',
 '论断门禁（标准）：含基础；SOTA 须点明基线/指标；贡献类型符合 venue；实验节与摘要/结论 claim 一致。',
 'Claim (standard, manuscript): baselines for SOTA; venue contribution types; experiments align with claims.'),
(11005, 'write_paper', 1, 'constraint_gate', '门禁约束', '论断门禁·严格·写作',
 '论断门禁（严格）：含标准；禁止因果夸大；empirical 须含可复现性要点；缺则为 major。',
 'Claim (strict, manuscript): no causal overreach; reproducibility hints required for empirical claims.'),
(11006, 'write_paper', 1, 'constraint_gate', '门禁约束', '驳论门禁·基础·写作',
 '驳论门禁（基础）：venue 明显不匹配则 blocker；列出拒稿式理由。',
 'Kill (basic, manuscript): clear venue misfit is blocker.'),
(11007, 'write_paper', 1, 'constraint_gate', '门禁约束', '驳论门禁·标准·写作',
 '驳论门禁（标准）：含基础；致命 flaw、与 Corpus/实验矛盾且未讨论、假新颖为 blocker/major。',
 'Kill (standard, manuscript): fatal flaws, contradictions, false novelty.'),
(11008, 'write_paper', 1, 'constraint_gate', '门禁约束', '驳论门禁·严格·写作',
 '驳论门禁（严格）：含标准；实验/数据不可行、不可复现为 major/blocker。',
 'Kill (strict, manuscript): infeasible or non-reproducible experiments.')
ON DUPLICATE KEY UPDATE
  `scene` = VALUES(`scene`),
  `model_config_id` = VALUES(`model_config_id`),
  `prompt_code` = VALUES(`prompt_code`),
  `prompt_type` = VALUES(`prompt_type`),
  `prompt_name` = VALUES(`prompt_name`),
  `llm_desc_zh` = VALUES(`llm_desc_zh`), `llm_desc_en` = VALUES(`llm_desc_en`);

INSERT INTO `paper_llm_constraint_gate` (
  `scene`, `code`, `name`, `llm_description_citation_check_code`, `llm_description_claim_check_code`, `llm_description_kill_check_code`, `is_default`
) VALUES
  ('write_paper', 'polished', '基础', 11000, 11003, 11006, 0),
  ('write_paper', 'standard', '标准', 11001, 11004, 11007, 1),
  ('write_paper', 'strict',   '严格', 11002, 11005, 11008, 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `llm_description_citation_check_code` = VALUES(`llm_description_citation_check_code`),
  `llm_description_claim_check_code` = VALUES(`llm_description_claim_check_code`),
  `llm_description_kill_check_code` = VALUES(`llm_description_kill_check_code`),
  `is_default` = VALUES(`is_default`);
