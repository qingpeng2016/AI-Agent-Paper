package literature

import (
	"context"
	"fmt"
	"strings"

	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

const (
	sourceArxiv           = "arxiv"
	sourceOpenAlex        = "openalex"
	sourceSemanticScholar = "semantic_scholar"
	sourceCrossref        = "crossref"
	sourcePubMed          = "pubmed"
)

func (s *LiteratureSearchService) CountPlatformTotal(ctx context.Context, sourceCode, query string) (int, error) {
	query, _, err := normalizeLiteratureQuery(query, 1)
	if err != nil {
		return 0, err
	}
	sourceCode = strings.TrimSpace(sourceCode)
	switch sourceCode {
	case sourceArxiv:
		return s.arxiv.Count(ctx, query)
	case sourceOpenAlex:
		return s.openalex.Count(ctx, query)
	case sourceSemanticScholar:
		return s.semanticScholar.Count(ctx, query)
	case sourceCrossref:
		return s.crossref.Count(ctx, query)
	case sourcePubMed:
		return s.pubmed.Count(ctx, query)
	default:
		return 0, fmt.Errorf("unsupported literature source: %s", sourceCode)
	}
}

// SearchPlatformPage slot：OpenAlex 为 1-based 页码；其余平台为 offset/start。
func (s *LiteratureSearchService) SearchPlatformPage(ctx context.Context, sourceCode, query string, limit, slot int) (*response.PaperLiteratureSearchResult, error) {
	query, limit, err := normalizeLiteratureQuery(query, limit)
	if err != nil {
		return nil, err
	}
	if slot < 0 {
		slot = 0
	}
	sourceCode = strings.TrimSpace(sourceCode)
	switch sourceCode {
	case sourceArxiv:
		papers, err := s.arxiv.Search(ctx, httpentity.ArxivSearchQuery{Query: query, MaxResults: limit, Start: slot})
		if err != nil {
			return nil, err
		}
		return &response.PaperLiteratureSearchResult{Query: query, SourceCode: sourceArxiv, Hits: arxivPapersToHits(papers)}, nil
	case sourceOpenAlex:
		page := slot
		if page < 1 {
			page = 1
		}
		works, err := s.openalex.Search(ctx, httpentity.OpenAlexSearchQuery{Query: query, MaxResults: limit, Page: page})
		if err != nil {
			return nil, err
		}
		return &response.PaperLiteratureSearchResult{Query: query, SourceCode: sourceOpenAlex, Hits: openAlexWorksToHits(works)}, nil
	case sourceSemanticScholar:
		papers, err := s.semanticScholar.Search(ctx, httpentity.SemanticScholarSearchQuery{Query: query, MaxResults: limit, Offset: slot})
		if err != nil {
			return nil, err
		}
		return &response.PaperLiteratureSearchResult{Query: query, SourceCode: sourceSemanticScholar, Hits: semanticScholarPapersToHits(papers)}, nil
	case sourceCrossref:
		works, err := s.crossref.Search(ctx, httpentity.CrossrefSearchQuery{Query: query, MaxResults: limit, Offset: slot})
		if err != nil {
			return nil, err
		}
		return &response.PaperLiteratureSearchResult{Query: query, SourceCode: sourceCrossref, Hits: crossrefWorksToHits(works)}, nil
	case sourcePubMed:
		articles, err := s.pubmed.Search(ctx, httpentity.PubMedSearchQuery{Query: query, MaxResults: limit, Offset: slot})
		if err != nil {
			return nil, err
		}
		return &response.PaperLiteratureSearchResult{Query: query, SourceCode: sourcePubMed, Hits: pubmedArticlesToHits(articles)}, nil
	default:
		return nil, fmt.Errorf("unsupported literature source: %s", sourceCode)
	}
}
