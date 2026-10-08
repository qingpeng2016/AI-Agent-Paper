package http

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/datatypes"
)

// ResolveLiteratureBaseURL 优先 paper_ref_literature_source.base_url，否则内置默认。
func ResolveLiteratureBaseURL(
	ctx context.Context,
	repo repository.PaperRefLiteratureSourceRepo,
	code string,
	fallback string,
) string {
	if repo != nil {
		row, err := repo.FindActiveByCode(ctx, code)
		if err == nil && row != nil {
			if u := strings.TrimSpace(row.BaseURL); u != "" {
				return u
			}
		}
	}
	return fallback
}

func ConfigStringFromJSON(raw datatypes.JSON, key string) string {
	key = strings.TrimSpace(key)
	if key == "" || len(raw) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func ResolveOpenAlexMailto(ctx context.Context, repo repository.PaperRefLiteratureSourceRepo) string {
	const code = "openalex"
	if repo != nil {
		row, err := repo.FindActiveByCode(ctx, code)
		if err == nil && row != nil {
			if m := ConfigStringFromJSON(row.DefaultConfig, "mailto"); m != "" {
				return m
			}
		}
	}
	return ""
}

func ResolveSemanticScholarAPIKey(ctx context.Context, repo repository.PaperRefLiteratureSourceRepo) string {
	const code = "semantic_scholar"
	if repo != nil {
		row, err := repo.FindActiveByCode(ctx, code)
		if err == nil && row != nil {
			if k := ConfigStringFromJSON(row.DefaultConfig, "api_key"); k != "" {
				return k
			}
		}
	}
	return strings.TrimSpace(os.Getenv("SEMANTIC_SCHOLAR_API_KEY"))
}

// Literature source codes（与 paper_ref_literature_source.code 一致）
const (
	LiteratureSourceCodeArxiv           = "arxiv"
	LiteratureSourceCodeOpenAlex        = "openalex"
	LiteratureSourceCodeSemanticScholar = "semantic_scholar"
)
