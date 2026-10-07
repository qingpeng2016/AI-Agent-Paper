package paper

import (
	"context"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	httpentity "github.com/qingpeng2016/ai-token-mall/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-token-mall/domain/http/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
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

func normalizeLiteratureQuery(query string, limit int) (string, int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", 0, errorx.ErrParamsError
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
