-- paper_llm_description.scene → prompt_scene
ALTER TABLE `paper_llm_description`
  CHANGE COLUMN `scene` `prompt_scene` VARCHAR(32) NOT NULL DEFAULT 'shared' COMMENT 'topic|write_paper|shared';

-- paper_llm_constraint_gate.scene → prompt_scene（复合主键）
ALTER TABLE `paper_llm_constraint_gate`
  DROP PRIMARY KEY;

ALTER TABLE `paper_llm_constraint_gate`
  CHANGE COLUMN `scene` `prompt_scene` VARCHAR(32) NOT NULL COMMENT 'topic|write_paper';

ALTER TABLE `paper_llm_constraint_gate`
  ADD PRIMARY KEY (`prompt_scene`, `code`);
