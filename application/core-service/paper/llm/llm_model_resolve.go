package llm

import "github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"

func ModelConfigIDFromRoute(route *entity.PaperStrategyRoute) uint {
	if route != nil && route.DefaultModelConfigID > 0 {
		return uint(route.DefaultModelConfigID)
	}
	return 0
}
