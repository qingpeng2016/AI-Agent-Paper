-- 取消「新版本顶替旧版本」语义：历史 superseded 恢复为 completed；之后新建记录不再改写旧行 status
UPDATE `paper_output_literature_review`
SET `status` = 'completed'
WHERE `status` = 'superseded';
