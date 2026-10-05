package product

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

// card_features 占位符：{limit_tokens}、{rpm_limit} 等，出库前按 products 字段替换。
func expandCardFeatures(features []string, p *entity.Products) []string {
	if len(features) == 0 {
		return features
	}
	out := make([]string, len(features))
	for i, line := range features {
		out[i] = expandCardFeatureLine(line, p)
	}
	return out
}

func expandCardFeatureLine(line string, p *entity.Products) string {
	if !strings.Contains(line, "{") {
		return line
	}
	s := line
	s = strings.ReplaceAll(s, "{limit_tokens}", formatTokenCompact(p.LimitTokens))
	s = strings.ReplaceAll(s, "{rpm_limit}", strconv.Itoa(p.RPMLimit))
	return s
}

func formatTokenCompact(n int64) string {
	if n >= 1_000_000 {
		if n%1_000_000 == 0 {
			return fmt.Sprintf("%dM", n/1_000_000)
		}
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		if n%1_000 == 0 {
			return fmt.Sprintf("%dK", n/1_000)
		}
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return strconv.FormatInt(n, 10)
}
