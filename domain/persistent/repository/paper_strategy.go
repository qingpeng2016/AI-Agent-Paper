package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperStrategyRepo interface {
	FindActiveRouteByBehaviorKey(ctx context.Context, behaviorKey string) (*entity.PaperStrategyRoute, error)
	FindPackageByID(ctx context.Context, id uint64) (*entity.PaperStrategyPackage, error)
}
