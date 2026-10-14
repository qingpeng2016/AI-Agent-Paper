package figure

import (
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
)

func parseFigureCategoryID(raw string) (uint64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errorx.ErrParamsError
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errorx.ErrParamsError
	}
	return id, nil
}

func parseFigureIDList(ids []string) ([]uint64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	out := make([]uint64, 0, len(ids))
	seen := make(map[uint64]struct{}, len(ids))
	for _, raw := range ids {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			return nil, errorx.ErrParamsError
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}
