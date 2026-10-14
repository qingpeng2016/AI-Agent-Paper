package figure

import (
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
)

func ParseFigureListFilter(keyword, categoryIDRaw, manuscriptIDRaw string) (repository.FigureListFilter, error) {
	f := repository.FigureListFilter{
		Keyword: strings.TrimSpace(keyword),
	}
	catRaw := strings.TrimSpace(categoryIDRaw)
	if catRaw != "" {
		if catRaw == "none" {
			zero := uint64(0)
			f.CategoryID = &zero
		} else {
			id, err := strconv.ParseUint(catRaw, 10, 64)
			if err != nil || id == 0 {
				return f, errorx.ErrParamsError
			}
			f.CategoryID = &id
		}
	}
	msRaw := strings.TrimSpace(manuscriptIDRaw)
	if msRaw != "" {
		if msRaw == "none" {
			zero := uint64(0)
			f.ManuscriptID = &zero
		} else {
			id, err := strconv.ParseUint(msRaw, 10, 64)
			if err != nil || id == 0 {
				return f, errorx.ErrParamsError
			}
			f.ManuscriptID = &id
		}
	}
	return f, nil
}
