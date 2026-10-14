package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperManuscriptSubmissionImpl struct {
	db *gorm.DB
}

func NewPaperManuscriptSubmissionImpl(db *gorm.DB) repository.PaperManuscriptSubmissionRepo {
	return &PaperManuscriptSubmissionImpl{db: db}
}

func (r *PaperManuscriptSubmissionImpl) Create(ctx context.Context, row *entity.PaperManuscriptSubmission) error {
	if row.RecordStatus == "" {
		row.RecordStatus = "active"
	}
	if row.Track == "" {
		row.Track = "main"
	}
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *PaperManuscriptSubmissionImpl) GetByIDForUser(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
) (*entity.PaperManuscriptSubmission, error) {
	var row entity.PaperManuscriptSubmission
	err := r.db.WithContext(ctx).
		Where("id = ? AND manuscript_id = ? AND user_id = ? AND record_status = ?", id, manuscriptID, userID, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperManuscriptSubmissionImpl) listQuery(ctx context.Context, manuscriptID uint64) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&entity.PaperManuscriptSubmission{}).
		Where("manuscript_id = ? AND record_status = ?", manuscriptID, "active").
		Order("COALESCE(submitted_at, created_at) DESC, id DESC")
}

func (r *PaperManuscriptSubmissionImpl) ListByManuscriptPaged(
	ctx context.Context,
	manuscriptID uint64,
	offset, limit int,
) ([]entity.PaperManuscriptSubmission, error) {
	var rows []entity.PaperManuscriptSubmission
	q := r.listQuery(ctx, manuscriptID)
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *PaperManuscriptSubmissionImpl) CountByManuscript(ctx context.Context, manuscriptID uint64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&entity.PaperManuscriptSubmission{}).
		Where("manuscript_id = ? AND record_status = ?", manuscriptID, "active").
		Count(&n).Error
	return n, err
}

func (r *PaperManuscriptSubmissionImpl) GetLatestForManuscript(
	ctx context.Context,
	manuscriptID uint64,
) (*entity.PaperManuscriptSubmission, error) {
	var row entity.PaperManuscriptSubmission
	err := r.listQuery(ctx, manuscriptID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperManuscriptSubmissionImpl) UpdateRevision(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
	status string,
	decision, comments *string,
	revisionDeadline *time.Time,
) (bool, error) {
	updates := map[string]any{
		"status": status,
	}
	if decision != nil {
		updates["decision"] = *decision
	}
	if comments != nil {
		updates["comments_summary"] = *comments
	}
	if revisionDeadline != nil {
		updates["revision_deadline"] = revisionDeadline
	} else {
		updates["revision_deadline"] = nil
	}
	res := r.db.WithContext(ctx).
		Model(&entity.PaperManuscriptSubmission{}).
		Where("id = ? AND manuscript_id = ? AND user_id = ? AND record_status = ?", id, manuscriptID, userID, "active").
		Updates(updates)
	return res.RowsAffected > 0, res.Error
}

func (r *PaperManuscriptSubmissionImpl) MarkResubmitted(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
	newRound int,
	status string,
) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&entity.PaperManuscriptSubmission{}).
		Where("id = ? AND manuscript_id = ? AND user_id = ? AND record_status = ?", id, manuscriptID, userID, "active").
		Updates(map[string]any{
			"round":  newRound,
			"status": status,
		})
	return res.RowsAffected > 0, res.Error
}

func (r *PaperManuscriptSubmissionImpl) UpdateMainPDF(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
	uri, filename string,
) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&entity.PaperManuscriptSubmission{}).
		Where("id = ? AND manuscript_id = ? AND user_id = ? AND record_status = ?", id, manuscriptID, userID, "active").
		Updates(map[string]any{
			"main_pdf_uri":      uri,
			"main_pdf_filename": filename,
		})
	return res.RowsAffected > 0, res.Error
}

func (r *PaperManuscriptSubmissionImpl) InsertEvent(ctx context.Context, ev *entity.PaperManuscriptSubmissionEvent) error {
	return r.db.WithContext(ctx).Create(ev).Error
}

func (r *PaperManuscriptSubmissionImpl) ListEvents(ctx context.Context, submissionID uint64) ([]entity.PaperManuscriptSubmissionEvent, error) {
	var rows []entity.PaperManuscriptSubmissionEvent
	err := r.db.WithContext(ctx).
		Where("submission_id = ?", submissionID).
		Order("created_at ASC, id ASC").
		Find(&rows).Error
	return rows, err
}
