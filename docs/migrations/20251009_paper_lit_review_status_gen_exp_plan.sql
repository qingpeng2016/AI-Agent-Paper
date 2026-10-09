-- 若曾误写入超长 status 可手工修正；新代码使用 gen_exp_plan（≤16 字符）
UPDATE `paper_output_literature_review`
SET `status` = 'gen_exp_plan'
WHERE `status` = 'generating_experiment_plan';
