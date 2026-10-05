package alipay

import (
	"context"

	httpentity "github.com/qingpeng2016/ai-token-mall/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-token-mall/domain/http/repository"
)

type AlipayService struct {
	alipay httprepo.AlipayRepo
}

func NewAlipayService(alipay httprepo.AlipayRepo) *AlipayService {
	return &AlipayService{alipay: alipay}
}

// FetchAccountLogs 获取支付宝账务明细（成交流水）
func (s *AlipayService) FetchAccountLogs(ctx context.Context, startTime, endTime string, pageNo, pageSize int) (*httpentity.AccountLogQueryResult, error) {
	return s.alipay.QueryAccountLogs(ctx, httpentity.AccountLogQuery{
		StartTime: startTime,
		EndTime:   endTime,
		PageNo:    pageNo,
		PageSize:  pageSize,
	})
}
