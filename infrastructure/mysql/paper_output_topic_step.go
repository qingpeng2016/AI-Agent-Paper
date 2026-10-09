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

func (r *PaperOutputTopicStepImpl) BeginRun(ctx context.Context, userID uint64, inputParams []byte) (int, error) {
	var runVersion int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxVer *int
		if err := tx.Model(&entity.PaperOutputTopicStep{}).
			Where("user_id = ?", userID).
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
				ManuscriptID: 0,
				UserID:       userID,
				RunVersion:   next,
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

func (r *PaperOutputTopicStepImpl) ListByUserRun(ctx context.Context, userID uint64, runVersion int) ([]entity.PaperOutputTopicStep, error) {
	var rows []entity.PaperOutputTopicStep
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND run_version = ?", userID, runVersion).
		Order("FIELD(stage_code, 'retrieve', 'generate_ideas', 'novelty', 'audit')").
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

func (r *PaperOutputTopicStepImpl) ListByStatusAndStage(ctx context.Context, status, stageCode string, limit int) ([]entity.PaperOutputTopicStep, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	var rows []entity.PaperOutputTopicStep
	err := r.db.WithContext(ctx).
		Where("status = ? AND stage_code = ?", status, stageCode).
		Order("updated_at ASC").
		Limit(limit).
		Find(&rows).Error
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
