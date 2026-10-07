package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/http/entity"
)

// SemanticScholarRepo Semantic Scholar Graph API，由 infrastructure/http/semanticscholar 实现
type SemanticScholarRepo interface {
	Search(ctx context.Context, q entity.SemanticScholarSearchQuery) ([]entity.SemanticScholarPaper, error)
}
