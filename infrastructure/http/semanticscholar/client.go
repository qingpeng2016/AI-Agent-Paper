package semanticscholar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/conf"
	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-agent-paper/domain/http/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
)

// Client 实现 domain/http/repository.SemanticScholarRepo
type Client struct {
	http    *httpx.Client
	baseURL string
	apiKey  string
}

func NewClient(http *httpx.Client, sources repository.PaperLiteraturePlatformRepo) httprepo.SemanticScholarRepo {
	ctx := context.Background()
	baseURL := httpx.ResolveLiteratureBaseURL(
		ctx,
		sources,
		httpx.LiteratureSourceCodeSemanticScholar,
		conf.DefaultLiteratureSemanticScholarBaseURL,
	)
	apiKey := httpx.ResolveSemanticScholarAPIKey(ctx, sources)
	return &Client{http: http, baseURL: baseURL, apiKey: apiKey}
}

func (c *Client) Count(ctx context.Context, query string) (int, error) {
	_, total, err := c.search(ctx, httpentity.SemanticScholarSearchQuery{Query: query, MaxResults: 1, Offset: 0})
	return total, err
}

func (c *Client) Search(ctx context.Context, q httpentity.SemanticScholarSearchQuery) ([]httpentity.SemanticScholarPaper, error) {
	papers, _, err := c.search(ctx, q)
	return papers, err
}

func (c *Client) search(ctx context.Context, q httpentity.SemanticScholarSearchQuery) ([]httpentity.SemanticScholarPaper, int, error) {
	limit := q.MaxResults
	if limit <= 0 {
		limit = 1
	}
	params := map[string]string{
		"query":  q.Query,
		"limit":  strconv.Itoa(limit),
		"fields": "title,authors,abstract,year,externalIds,url,paperId",
	}
	if q.Offset > 0 {
		params["offset"] = strconv.Itoa(q.Offset)
	}
	headers := map[string]string{
		"Accept": "application/json",
	}
	if c.apiKey != "" {
		headers["x-api-key"] = c.apiKey
	}
	resp, err := c.http.GetLiterature(ctx, c.baseURL, params, headers)
	if err != nil {
		return nil, 0, httpx.LogLiteratureTransportError(ctx, "semantic_scholar", "search", c.baseURL, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, 0, httpx.NewLiteratureHTTPError(ctx, "semantic_scholar", "search", c.baseURL, resp, fmt.Errorf("semantic scholar http %d", resp.StatusCode()))
	}
	var body searchResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, 0, httpx.NewLiteratureHTTPError(ctx, "semantic_scholar", "search", c.baseURL, resp, err)
	}
	out := make([]httpentity.SemanticScholarPaper, 0, len(body.Data))
	for _, p := range body.Data {
		out = append(out, paperToEntity(p))
	}
	total := body.Total
	if total <= 0 && len(out) > 0 {
		total = len(out)
	}
	return out, total, nil
}

type searchResponse struct {
	Total int         `json:"total"`
	Data  []paperJSON `json:"data"`
}

type paperJSON struct {
	PaperID     string `json:"paperId"`
	Title       string `json:"title"`
	Abstract    string `json:"abstract"`
	Year        int    `json:"year"`
	URL         string `json:"url"`
	ExternalIds struct {
		DOI   string `json:"DOI"`
		ArXiv string `json:"ArXiv"`
	} `json:"externalIds"`
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
}

func paperToEntity(p paperJSON) httpentity.SemanticScholarPaper {
	ext := "s2:" + p.PaperID
	var authors []string
	for _, a := range p.Authors {
		n := strings.TrimSpace(a.Name)
		if n != "" {
			authors = append(authors, n)
		}
	}
	doi := strings.TrimSpace(p.ExternalIds.DOI)
	u := strings.TrimSpace(p.URL)
	if u == "" && p.ExternalIds.ArXiv != "" {
		u = "https://arxiv.org/abs/" + p.ExternalIds.ArXiv
	}
	return httpentity.SemanticScholarPaper{
		ExternalKey:   ext,
		Title:         strings.TrimSpace(p.Title),
		Authors:       authors,
		Abstract:      strings.TrimSpace(p.Abstract),
		PublishedYear: p.Year,
		DOI:           doi,
		URL:           u,
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
