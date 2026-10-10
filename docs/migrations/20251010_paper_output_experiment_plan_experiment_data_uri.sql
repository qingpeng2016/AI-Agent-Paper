ALTER TABLE `paper_output_experiment_plan`
  ADD COLUMN `experiment_data_uri` VARCHAR(1024) DEFAULT NULL COMMENT '实验数据文件相对 storage/ 路径' AFTER `storage_uri`;
