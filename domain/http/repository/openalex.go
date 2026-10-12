package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
)

// OpenAlexRepo OpenAlex Works API，由 infrastructure/http/openalex 实现
type OpenAlexRepo interface {
	Count(ctx context.Context, query string) (int, error)
	Search(ctx context.Context, q entity.OpenAlexSearchQuery) ([]entity.OpenAlexWork, error)
}
