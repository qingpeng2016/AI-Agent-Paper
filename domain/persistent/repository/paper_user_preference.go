package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperUserPreferenceRepo interface {
	GetByUserID(ctx context.Context, userID uint64) (*entity.PaperUserPreference, error)
	Save(ctx context.Context, row *entity.PaperUserPreference) error
}
