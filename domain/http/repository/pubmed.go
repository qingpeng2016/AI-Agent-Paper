package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
)

type PubMedRepo interface {
	Count(ctx context.Context, query string) (int, error)
	Search(ctx context.Context, q entity.PubMedSearchQuery) ([]entity.PubMedArticle, error)
}
