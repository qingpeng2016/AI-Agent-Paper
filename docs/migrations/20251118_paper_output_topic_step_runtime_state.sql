-- 选题等步骤：运行态（async_llm、检查点、段内 drain）与业务 result 分离
ALTER TABLE paper_output_topic_step
  ADD COLUMN runtime_state JSON NULL COMMENT '步骤运行态 JSON（bot/检查点/排队）' AFTER result;
