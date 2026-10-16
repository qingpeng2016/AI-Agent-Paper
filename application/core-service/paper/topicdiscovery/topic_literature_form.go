package topicdiscovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
)

// topicLiteratureSourceRunEnabled 是否允许在选题表单勾选并提交（库表 status 须为 active）。
func topicLiteratureSourceRunEnabled(platformStatus, code string, nativeIdeaDiscovery bool) bool {
	if strings.TrimSpace(platformStatus) != "active" {
		return false
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	if nativeIdeaDiscovery {
		return code == sourceArxiv
	}
	_, ok := literatureRetrieveSupportedSources[code]
	return ok
}

func (s *TopicDiscoveryRunService) validateTopicLiteratureSourceSelection(
	ctx context.Context,
	codes []string,
	nativeIdeaDiscovery bool,
) ([]string, error) {
	trimmed := normalizeStringList(codes)
	if len(trimmed) == 0 {
		return nil, errorx.ErrParamsError.WithDetail("请至少选择一个文献来源")
	}
	active, err := s.litSource.ListActiveOrdered(ctx)
	if err != nil {
		return nil, err
	}
	activeByCode := map[string]string{}
	for _, row := range active {
		activeByCode[row.Code] = row.Status
	}
	for _, c := range trimmed {
		st, ok := activeByCode[c]
		if !ok {
			return nil, errorx.ErrParamsError.WithDetail(fmt.Sprintf("文献来源「%s」无效或已停用", c))
		}
		if !topicLiteratureSourceRunEnabled(st, c, nativeIdeaDiscovery) {
			return nil, errorx.ErrParamsError.WithDetail(fmt.Sprintf("文献来源「%s」暂不支持当前选题流程", c))
		}
	}
	return s.orderSourceCodesBySort(ctx, trimmed), nil
}
