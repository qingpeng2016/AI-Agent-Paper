-- 新颖性已与 generate_ideas 合并，不再保留单独 stage_code=novelty 的系统模板
DELETE FROM `paper_template_document`
WHERE `user_id` = 0 AND `venue_id` = 0 AND `stage_code` = 'novelty';
