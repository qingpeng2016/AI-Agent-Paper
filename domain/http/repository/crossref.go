package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
)

type CrossrefRepo interface {
	Count(ctx context.Context, query string) (int, error)
	Search(ctx context.Context, q entity.CrossrefSearchQuery) ([]entity.CrossrefWork, error)
}
