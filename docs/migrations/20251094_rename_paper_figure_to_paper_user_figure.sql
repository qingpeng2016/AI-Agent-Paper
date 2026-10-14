-- 图表管理表统一 paper_user 前缀（与 paper_user_literature 对齐）
-- 新库若 all.sql 已是 paper_user_figure* 可跳过

RENAME TABLE `paper_figure` TO `paper_user_figure`;
RENAME TABLE `paper_figure_category` TO `paper_user_figure_category`;
RENAME TABLE `paper_figure_category_rel` TO `paper_user_figure_category_rel`;
RENAME TABLE `paper_figure_manuscript_rel` TO `paper_user_figure_manuscript_rel`;

ALTER TABLE `paper_user_figure`
  RENAME INDEX `idx_paper_figure_ms` TO `idx_paper_user_fig_ms`;

ALTER TABLE `paper_user_figure_category`
  RENAME INDEX `idx_paper_figure_cat_status_id` TO `idx_paper_user_fig_cat_status_id`;

ALTER TABLE `paper_user_figure_category_rel`
  RENAME INDEX `uk_paper_figure_cat_rel` TO `uk_paper_user_fig_cat_rel`,
  RENAME INDEX `idx_paper_figure_cat_rel_cat` TO `idx_paper_user_fig_cat_rel_cat`;

ALTER TABLE `paper_user_figure_manuscript_rel`
  RENAME INDEX `uk_paper_figure_ms_rel` TO `uk_paper_user_fig_ms_rel`,
  RENAME INDEX `idx_paper_figure_ms_rel_ms` TO `idx_paper_user_fig_ms_rel_ms`;
