package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

var topicDiscoveryStageOrder = []string{"idea_discovery", "topic_confirm"}

var topicDiscoveryPhasedStageOrder = []string{
	"research_lit",
	"idea_creator",
	"novelty_check",
	"research_review",
	"research_refine",
	"topic_confirm",
}

type PaperOutputTopicStepImpl struct {
	db *gorm.DB
}

func NewPaperOutputTopicStepImpl(db *gorm.DB) repository.PaperOutputTopicStepRepo {
	return &PaperOutputTopicStepImpl{db: db}
}

func (r *PaperOutputTopicStepImpl) BeginRun(ctx context.Context, userID, manuscriptID uint64, inputParams []byte) (int, error) {
	if manuscriptID == 0 {
		return 0, errors.New("manuscript_id required")
	}
	var runVersion int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxVer *int
		if err := tx.Model(&entity.PaperOutputTopicStep{}).
			Where("user_id = ? AND manuscript_id = ?", userID, manuscriptID).
			Select("COALESCE(MAX(run_version), 0)").
			Scan(&maxVer).Error; err != nil {
			return err
		}
		next := 1
		if maxVer != nil {
			next = *maxVer + 1
		}
		runVersion = next
		now := time.Now()
		for i, stage := range topicDiscoveryStageOrder {
			row := entity.PaperOutputTopicStep{
				ManuscriptID: manuscriptID,
				UserID:       userID,
				RunVersion:   next,
				StageCode:    stage,
				Sort:         i + 1,
				Status:       "pending",
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if i == 0 && len(inputParams) > 0 {
				row.InputParams = inputParams
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return runVersion, err
}

func (r *PaperOutputTopicStepImpl) BeginPhasedRun(ctx context.Context, userID, manuscriptID uint64, inputParams []byte) (int, error) {
	if manuscriptID == 0 {
		return 0, errors.New("manuscript_id required")
	}
	var runVersion int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxVer *int
		if err := tx.Model(&entity.PaperOutputTopicStep{}).
			Where("user_id = ? AND manuscript_id = ?", userID, manuscriptID).
			Select("COALESCE(MAX(run_version), 0)").
			Scan(&maxVer).Error; err != nil {
			return err
		}
		next := 1
		if maxVer != nil {
			next = *maxVer + 1
		}
		runVersion = next
		now := time.Now()
		for i, stage := range topicDiscoveryPhasedStageOrder {
			row := entity.PaperOutputTopicStep{
				ManuscriptID: manuscriptID,
				UserID:       userID,
				RunVersion:   next,
				StageCode:    stage,
				Sort:         i + 1,
				Status:       "pending",
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if stage == "research_lit" && len(inputParams) > 0 {
				row.InputParams = inputParams
			}
			_ = i
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return runVersion, err
}

func (r *PaperOutputTopicStepImpl) ListByUserRun(ctx context.Context, userID uint64, runVersion int) ([]entity.PaperOutputTopicStep, error) {
	var rows []entity.PaperOutputTopicStep
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND run_version = ?", userID, runVersion).
		Order("sort ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperOutputTopicStepImpl) GetLatestRunByUser(ctx context.Context, userID uint64) (int, []entity.PaperOutputTopicStep, error) {
	var probe entity.PaperOutputTopicStep
	err := r.db.WithContext(ctx).
		Select("run_version").
		Where("user_id = ?", userID).
		Order("run_version DESC, id DESC").
		Take(&probe).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil, nil
	}
	if err != nil {
		return 0, nil, err
	}
	steps, err := r.ListByUserRun(ctx, userID, probe.RunVersion)
	return probe.RunVersion, steps, err
}

func (r *PaperOutputTopicStepImpl) GetLatestRunByManuscript(ctx context.Context, userID, manuscriptID uint64) (int, []entity.PaperOutputTopicStep, error) {
	if manuscriptID == 0 {
		return 0, nil, nil
	}
	var probe entity.PaperOutputTopicStep
	err := r.db.WithContext(ctx).
		Select("run_version").
		Where("user_id = ? AND manuscript_id = ?", userID, manuscriptID).
		Order("run_version DESC, id DESC").
		Take(&probe).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil, nil
	}
	if err != nil {
		return 0, nil, err
	}
	steps, err := r.ListByUserRun(ctx, userID, probe.RunVersion)
	return probe.RunVersion, steps, err
}

func (r *PaperOutputTopicStepImpl) GetStep(ctx context.Context, userID uint64, runVersion int, stageCode string) (*entity.PaperOutputTopicStep, error) {
	var row entity.PaperOutputTopicStep
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND run_version = ? AND stage_code = ?", userID, runVersion, stageCode).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperOutputTopicStepImpl) SaveStep(ctx context.Context, step *entity.PaperOutputTopicStep) error {
	return r.db.WithContext(ctx).Save(step).Error
}

func maxRunVersionPerManuscriptWhere(db *gorm.DB) *gorm.DB {
	return db.Where(`(user_id, manuscript_id, run_version) IN (
		SELECT user_id, manuscript_id, MAX(run_version)
		FROM paper_output_topic_step
		GROUP BY user_id, manuscript_id
	)`)
}

func (r *PaperOutputTopicStepImpl) ListByStatusOnMaxRunPerManuscript(ctx context.Context, status string) ([]entity.PaperOutputTopicStep, error) {
	var rows []entity.PaperOutputTopicStep
	q := r.db.WithContext(ctx).Where("status = ?", status)
	err := maxRunVersionPerManuscriptWhere(q).
		Order("updated_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperOutputTopicStepImpl) ListByStatusAndStageOnMaxRunPerManuscript(ctx context.Context, status, stageCode string) ([]entity.PaperOutputTopicStep, error) {
	var rows []entity.PaperOutputTopicStep
	q := r.db.WithContext(ctx).
		Where("status = ? AND stage_code = ?", status, stageCode)
	err := maxRunVersionPerManuscriptWhere(q).
		Order("updated_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperOutputTopicStepImpl) ListByStatusAndStage(ctx context.Context, status, stageCode string, limit int) ([]entity.PaperOutputTopicStep, error) {
	q := r.db.WithContext(ctx).
		Where("status = ? AND stage_code = ?", status, stageCode).
		Order("updated_at ASC")
	if limit > 0 {
		if limit > 200 {
			limit = 200
		}
		q = q.Limit(limit)
	}
	var rows []entity.PaperOutputTopicStep
	err := q.Find(&rows).Error
	return rows, err
}

func (r *PaperOutputTopicStepImpl) CancelRun(ctx context.Context, userID uint64, runVersion int) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.PaperOutputTopicStep{}).
		Where("user_id = ? AND run_version = ?", userID, runVersion).
		Updates(map[string]any{
			"status":       "cancelled",
			"completed_at": now,
		}).Error
}

func (r *PaperOutputTopicStepImpl) BindManuscript(ctx context.Context, userID uint64, runVersion int, manuscriptID uint64) error {
	return r.db.WithContext(ctx).Model(&entity.PaperOutputTopicStep{}).
		Where("user_id = ? AND run_version = ?", userID, runVersion).
		Update("manuscript_id", manuscriptID).Error
}
