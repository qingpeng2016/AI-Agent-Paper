package paper

import (
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
)

func ParseUserLiteratureListFilter(keyword, categoryIDRaw string) (repository.UserLiteratureListFilter, error) {
	f := repository.UserLiteratureListFilter{
		Keyword: strings.TrimSpace(keyword),
	}
	raw := strings.TrimSpace(categoryIDRaw)
	if raw == "" {
		return f, nil
	}
	if raw == "none" {
		zero := uint64(0)
		f.CategoryID = &zero
		return f, nil
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return f, errorx.ErrParamsError
	}
	f.CategoryID = &id
	return f, nil
}
