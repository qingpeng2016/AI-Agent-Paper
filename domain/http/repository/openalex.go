package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/http/entity"
)

// OpenAlexRepo OpenAlex Works API，由 infrastructure/http/openalex 实现
type OpenAlexRepo interface {
	Search(ctx context.Context, q entity.OpenAlexSearchQuery) ([]entity.OpenAlexWork, error)
}
