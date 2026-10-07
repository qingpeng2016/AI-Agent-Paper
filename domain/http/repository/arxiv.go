package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/http/entity"
)

// ArxivRepo arXiv API，由 infrastructure/http/arxiv 实现
type ArxivRepo interface {
	Search(ctx context.Context, q entity.ArxivSearchQuery) ([]entity.ArxivPaper, error)
}
