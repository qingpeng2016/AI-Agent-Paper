package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperRefCatalogRepo interface {
	ListActiveDisciplines(ctx context.Context) ([]entity.PaperRefDiscipline, error)
	FindDisciplineByCode(ctx context.Context, code string) (*entity.PaperRefDiscipline, error)
	FindDisciplineByID(ctx context.Context, id uint64) (*entity.PaperRefDiscipline, error)
	ListExecutionIntensities(ctx context.Context) ([]entity.PaperLLMConstraintExecIntensity, error)
	ListAuditLevels(ctx context.Context) ([]entity.PaperLLMConstraintGate, error)
	FindExecutionIntensityByCode(ctx context.Context, code string) (*entity.PaperLLMConstraintExecIntensity, error)
	FindAuditLevelByCode(ctx context.Context, code string) (*entity.PaperLLMConstraintGate, error)
	FindLLMDescriptionByCode(ctx context.Context, code uint64) (*entity.PaperLLMDescription, error)
	ListActiveVenues(ctx context.Context) ([]entity.PaperRefVenue, error)
	FindVenueByCode(ctx context.Context, code string) (*entity.PaperRefVenue, error)
	FindVenueByID(ctx context.Context, id uint64) (*entity.PaperRefVenue, error)
}
