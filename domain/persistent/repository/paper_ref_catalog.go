package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperRefCatalogRepo interface {
	ListActiveDisciplines(ctx context.Context) ([]entity.PaperRefDiscipline, error)
	ListExecutionIntensities(ctx context.Context) ([]entity.PaperRefExecutionIntensity, error)
	ListAuditLevels(ctx context.Context) ([]entity.PaperRefAuditLevel, error)
}
