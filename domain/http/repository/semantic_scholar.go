package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
)

// SemanticScholarRepo Semantic Scholar Graph API，由 infrastructure/http/semanticscholar 实现
type SemanticScholarRepo interface {
	Count(ctx context.Context, query string) (int, error)
	Search(ctx context.Context, q entity.SemanticScholarSearchQuery) ([]entity.SemanticScholarPaper, error)
}
