package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/http/entity"
)

// AlipayRepo 支付宝 OpenAPI（账务明细等），由 infrastructure/http/alipay 实现
type AlipayRepo interface {
	QueryAccountLogs(ctx context.Context, q entity.AccountLogQuery) (*entity.AccountLogQueryResult, error)
}
