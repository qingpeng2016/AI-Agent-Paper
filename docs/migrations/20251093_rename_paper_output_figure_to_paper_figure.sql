-- 已有库：paper_output_figure -> paper_user_figure
-- 若已是 paper_user_figure（新装 all.sql）可跳过本文件

RENAME TABLE `paper_output_figure` TO `paper_user_figure`;

ALTER TABLE `paper_user_figure`
  RENAME INDEX `idx_paper_output_figure_ms` TO `idx_paper_user_fig_ms`;
