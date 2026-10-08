package paper

import (
	"context"
	"regexp"
	"strings"
	"unicode"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-agent-paper/domain/http/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

const (
	sourceArxiv           = "arxiv"
	sourceOpenAlex        = "openalex"
	sourceSemanticScholar = "semantic_scholar"
)

type LiteratureSearchService struct {
	arxiv           httprepo.ArxivRepo
	openalex        httprepo.OpenAlexRepo
	semanticScholar httprepo.SemanticScholarRepo
}

func NewLiteratureSearchService(
	arxiv httprepo.ArxivRepo,
	openalex httprepo.OpenAlexRepo,
	semanticScholar httprepo.SemanticScholarRepo,
) *LiteratureSearchService {
	return &LiteratureSearchService{
		arxiv:           arxiv,
		openalex:        openalex,
		semanticScholar: semanticScholar,
	}
}

func (s *LiteratureSearchService) SearchArxiv(ctx context.Context, query string, limit int) (*response.PaperLiteratureSearchResult, error) {
	query, limit, err := normalizeLiteratureQuery(query, limit)
	if err != nil {
		return nil, err
	}
	papers, err := s.arxiv.Search(ctx, httpentity.ArxivSearchQuery{Query: query, MaxResults: limit})
	if err != nil {
		return nil, err
	}
	return &response.PaperLiteratureSearchResult{
		Query:      query,
		SourceCode: sourceArxiv,
		Hits:       arxivPapersToHits(papers),
	}, nil
}

func (s *LiteratureSearchService) SearchOpenAlex(ctx context.Context, query string, limit int) (*response.PaperLiteratureSearchResult, error) {
	query, limit, err := normalizeLiteratureQuery(query, limit)
	if err != nil {
		return nil, err
	}
	works, err := s.openalex.Search(ctx, httpentity.OpenAlexSearchQuery{Query: query, MaxResults: limit})
	if err != nil {
		return nil, err
	}
	return &response.PaperLiteratureSearchResult{
		Query:      query,
		SourceCode: sourceOpenAlex,
		Hits:       openAlexWorksToHits(works),
	}, nil
}

func (s *LiteratureSearchService) SearchSemanticScholar(ctx context.Context, query string, limit int) (*response.PaperLiteratureSearchResult, error) {
	query, limit, err := normalizeLiteratureQuery(query, limit)
	if err != nil {
		return nil, err
	}
	papers, err := s.semanticScholar.Search(ctx, httpentity.SemanticScholarSearchQuery{Query: query, MaxResults: limit})
	if err != nil {
		return nil, err
	}
	return &response.PaperLiteratureSearchResult{
		Query:      query,
		SourceCode: sourceSemanticScholar,
		Hits:       semanticScholarPapersToHits(papers),
	}, nil
}

var literatureKeywordMarkers = []string{
	"关键词：", "关键词:", "关键字：", "关键字:",
	"Keywords:", "keywords:", "KEYWORDS:",
}

var englishTokenRE = regexp.MustCompile(`[A-Za-z][A-Za-z0-9+\-./]*`)

// ExtractLiteratureSearchQuery turns a long mixed CN/EN research direction into a short
// English query suitable for arXiv / OpenAlex / Semantic Scholar.
func ExtractLiteratureSearchQuery(direction string) string {
	direction = strings.TrimSpace(direction)
	if direction == "" {
		return ""
	}
	lower := strings.ToLower(direction)
	for _, marker := range literatureKeywordMarkers {
		idx := strings.Index(direction, marker)
		if idx < 0 {
			idx = strings.Index(lower, strings.ToLower(marker))
		}
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(direction[idx+len(marker):])
		if q := joinLiteratureKeywordTerms(rest); q != "" {
			return truncateLiteratureQuery(q, 400)
		}
	}
	if q := extractEnglishSearchTerms(direction); q != "" {
		return truncateLiteratureQuery(q, 400)
	}
	return truncateLiteratureQuery(strings.Join(strings.Fields(direction), " "), 400)
}

func joinLiteratureKeywordTerms(block string) string {
	block = strings.NewReplacer("，", ",", "；", ";", "、", ",").Replace(block)
	var terms []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(terms) > 0 && (strings.HasPrefix(line, "研究") || strings.HasPrefix(strings.ToLower(line), "research")) {
			break
		}
		for _, part := range strings.Split(line, ",") {
			part = strings.TrimSpace(part)
			if part == "" || mostlyChineseRunes(part) {
				continue
			}
			terms = append(terms, part)
		}
	}
	return strings.Join(terms, " ")
}

func extractEnglishSearchTerms(text string) string {
	seen := map[string]struct{}{}
	var terms []string
	for _, m := range englishTokenRE.FindAllString(text, -1) {
		t := strings.TrimSpace(m)
		if len(t) < 3 {
			continue
		}
		key := strings.ToLower(t)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		terms = append(terms, t)
	}
	return strings.Join(terms, " ")
}

func mostlyChineseRunes(s string) bool {
	var han, other int
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			han++
		} else if !unicode.IsSpace(r) {
			other++
		}
	}
	return han > 0 && han >= other
}

func truncateLiteratureQuery(q string, maxRunes int) string {
	q = strings.TrimSpace(q)
	if maxRunes <= 0 || len([]rune(q)) <= maxRunes {
		return q
	}
	return string([]rune(q)[:maxRunes])
}

func normalizeLiteratureQuery(query string, limit int) (string, int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", 0, errorx.ErrParamsError
	}
	if len([]rune(query)) > 120 || strings.Contains(query, "关键词") || strings.Contains(strings.ToLower(query), "keywords:") {
		if extracted := ExtractLiteratureSearchQuery(query); extracted != "" {
			query = extracted
		}
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return query, limit, nil
}

func arxivPapersToHits(papers []httpentity.ArxivPaper) []response.PaperLiteratureHit {
	hits := make([]response.PaperLiteratureHit, 0, len(papers))
	for _, p := range papers {
		hits = append(hits, response.PaperLiteratureHit{
			SourceCode:    sourceArxiv,
			ExternalKey:   p.ExternalKey,
			Title:         p.Title,
			Authors:       p.Authors,
			Abstract:      p.Abstract,
			PublishedYear: p.PublishedYear,
			URL:           p.URL,
		})
	}
	return hits
}

func openAlexWorksToHits(works []httpentity.OpenAlexWork) []response.PaperLiteratureHit {
	hits := make([]response.PaperLiteratureHit, 0, len(works))
	for _, w := range works {
		hits = append(hits, response.PaperLiteratureHit{
			SourceCode:    sourceOpenAlex,
			ExternalKey:   w.ExternalKey,
			Title:         w.Title,
			Authors:       w.Authors,
			PublishedYear: w.PublishedYear,
			DOI:           w.DOI,
			URL:           w.URL,
		})
	}
	return hits
}

func semanticScholarPapersToHits(papers []httpentity.SemanticScholarPaper) []response.PaperLiteratureHit {
	hits := make([]response.PaperLiteratureHit, 0, len(papers))
	for _, p := range papers {
		hits = append(hits, response.PaperLiteratureHit{
			SourceCode:    sourceSemanticScholar,
			ExternalKey:   p.ExternalKey,
			Title:         p.Title,
			Authors:       p.Authors,
			Abstract:      p.Abstract,
			PublishedYear: p.PublishedYear,
			DOI:           p.DOI,
			URL:           p.URL,
		})
	}
	return hits
}
