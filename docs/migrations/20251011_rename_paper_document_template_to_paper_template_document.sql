-- 已有库若已执行旧 migration 创建 paper_document_template，执行本文件一次即可
RENAME TABLE `paper_document_template` TO `paper_template_document`;
