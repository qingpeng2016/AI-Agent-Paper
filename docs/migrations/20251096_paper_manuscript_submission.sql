-- 论文投稿：对外投递记录 + 操作日志（汇总字段仍在 paper_manuscript.submission_*）
CREATE TABLE IF NOT EXISTS `paper_manuscript_submission` (
  `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '投递记录 ID',
  `manuscript_id`       BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript.id',
  `user_id`             BIGINT UNSIGNED NOT NULL COMMENT 'users.id',
  `venue_code`          VARCHAR(64)  NOT NULL COMMENT '目标 venue 编码，如 neurips_2026',
  `venue_label`         VARCHAR(128) NOT NULL COMMENT '展示名，如 NeurIPS 2026',
  `track`               VARCHAR(16)  NOT NULL DEFAULT 'main' COMMENT 'main|short',
  `round`               INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '修回轮次（首次投递为 0）',
  `status`              VARCHAR(24)  NOT NULL DEFAULT 'submitted' COMMENT 'not_submitted|submitted|under_review|revision|rebuttal|resubmitted|accepted|rejected',
  `draft_version`       INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定的写作稿本版本号',
  `anonymous`           TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '双盲投稿',
  `portal_url`          VARCHAR(1024) DEFAULT NULL COMMENT '投稿系统链接',
  `external_id`         VARCHAR(128)  DEFAULT NULL COMMENT 'Portal submission id',
  `title`               VARCHAR(512)  DEFAULT NULL COMMENT '提交标题（可覆盖论文标题）',
  `abstract`            TEXT          DEFAULT NULL COMMENT '提交摘要',
  `main_pdf_uri`        VARCHAR(1024) DEFAULT NULL COMMENT '主 PDF 相对 storage 路径',
  `main_pdf_filename`   VARCHAR(256)  DEFAULT NULL COMMENT '主 PDF 原始文件名',
  `decision`            VARCHAR(128)  DEFAULT NULL COMMENT '审稿结论',
  `comments_summary`    TEXT          DEFAULT NULL COMMENT '意见摘要与改稿要点',
  `revision_deadline`   DATETIME(3)   DEFAULT NULL COMMENT '修回截止',
  `submitted_at`        DATETIME(3)   DEFAULT NULL COMMENT '标记已在 portal 提交的时间',
  `checklist`           JSON          DEFAULT NULL COMMENT '校验项快照',
  `record_status`       VARCHAR(16)   NOT NULL DEFAULT 'active' COMMENT 'active|deleted',
  `created_at`          DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`          DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_pms_manuscript` (`manuscript_id`, `record_status`, `updated_at`),
  KEY `idx_pms_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文对外投递记录';

CREATE TABLE IF NOT EXISTS `paper_manuscript_submission_event` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `submission_id` BIGINT UNSIGNED NOT NULL COMMENT 'paper_manuscript_submission.id',
  `event_type`    VARCHAR(32)  NOT NULL COMMENT 'created|submitted|revision_saved|resubmitted|pdf_uploaded',
  `message`       VARCHAR(512) NOT NULL DEFAULT '' COMMENT '日志说明',
  `created_at`    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_pmse_submission` (`submission_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='论文投递操作日志';
