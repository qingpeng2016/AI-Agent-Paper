package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

var topicDiscoveryStageOrder = []string{"retrieve", "generate_ideas", "novelty", "audit"}

type PaperOutputTopicStepImpl struct {
	db *gorm.DB
}

func NewPaperOutputTopicStepImpl(db *gorm.DB) repository.PaperOutputTopicStepRepo {
	return &PaperOutputTopicStepImpl{db: db}
}

func (r *PaperOutputTopicStepImpl) BeginRun(ctx context.Context, manuscriptID, userID uint64, inputParams []byte) (int, error) {
	var runVersion int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.PaperOutputTopicStep{}).
			Where("manuscript_id = ?", manuscriptID).
			Update("is_current_run", false).Error; err != nil {
			return err
		}
		var maxVer *int
		if err := tx.Model(&entity.PaperOutputTopicStep{}).
			Where("manuscript_id = ?", manuscriptID).
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
				IsCurrentRun: true,
				StageCode:    stage,
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

func (r *PaperOutputTopicStepImpl) ListByRun(ctx context.Context, manuscriptID uint64, runVersion int) ([]entity.PaperOutputTopicStep, error) {
	var rows []entity.PaperOutputTopicStep
	err := r.db.WithContext(ctx).
		Where("manuscript_id = ? AND run_version = ?", manuscriptID, runVersion).
		Order("FIELD(stage_code, 'retrieve', 'generate_ideas', 'novelty', 'audit')").
		Find(&rows).Error
	return rows, err
}

func (r *PaperOutputTopicStepImpl) GetCurrentRun(ctx context.Context, manuscriptID uint64) (int, []entity.PaperOutputTopicStep, error) {
	var runVersion *int
	err := r.db.WithContext(ctx).Model(&entity.PaperOutputTopicStep{}).
		Where("manuscript_id = ? AND is_current_run = ?", manuscriptID, true).
		Select("MAX(run_version)").
		Scan(&runVersion).Error
	if err != nil {
		return 0, nil, err
	}
	if runVersion == nil || *runVersion == 0 {
		return 0, nil, nil
	}
	steps, err := r.ListByRun(ctx, manuscriptID, *runVersion)
	return *runVersion, steps, err
}

func (r *PaperOutputTopicStepImpl) GetStep(ctx context.Context, manuscriptID uint64, runVersion int, stageCode string) (*entity.PaperOutputTopicStep, error) {
	var row entity.PaperOutputTopicStep
	err := r.db.WithContext(ctx).
		Where("manuscript_id = ? AND run_version = ? AND stage_code = ?", manuscriptID, runVersion, stageCode).
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

func (r *PaperOutputTopicStepImpl) CancelCurrentRun(ctx context.Context, manuscriptID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.PaperOutputTopicStep{}).
		Where("manuscript_id = ? AND is_current_run = ?", manuscriptID, true).
		Updates(map[string]any{
			"status":         "cancelled",
			"is_current_run": false,
			"completed_at":   now,
		}).Error
}
