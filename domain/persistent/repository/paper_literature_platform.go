package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperLiteraturePlatformRepo interface {
	FindActiveByCode(ctx context.Context, code string) (*entity.PaperLiteraturePlatform, error)
	ListActiveOrdered(ctx context.Context) ([]entity.PaperLiteraturePlatform, error)
}
