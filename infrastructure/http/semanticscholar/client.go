package semanticscholar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/conf"
	httpentity "github.com/qingpeng2016/ai-token-mall/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-token-mall/domain/http/repository"
	httpx "github.com/qingpeng2016/ai-token-mall/infrastructure/http"
)

// Client 实现 domain/http/repository.SemanticScholarRepo
type Client struct {
	http    *httpx.Client
	baseURL string
	apiKey  string
}

func NewClient(http *httpx.Client, cfg *conf.Config) httprepo.SemanticScholarRepo {
	lit := conf.GetLiteratureConf()
	baseURL := conf.DefaultLiteratureSemanticScholarBaseURL
	apiKey := ""
	if lit != nil && lit.SemanticScholar != nil {
		if u := strings.TrimSpace(lit.SemanticScholar.BaseURL); u != "" {
			baseURL = u
		}
		apiKey = strings.TrimSpace(lit.SemanticScholar.APIKey)
	}
	return &Client{http: http, baseURL: baseURL, apiKey: apiKey}
}

func (c *Client) Search(ctx context.Context, q httpentity.SemanticScholarSearchQuery) ([]httpentity.SemanticScholarPaper, error) {
	params := map[string]string{
		"query":  q.Query,
		"limit":  strconv.Itoa(q.MaxResults),
		"fields": "title,authors,abstract,year,externalIds,url,paperId",
	}
	headers := map[string]string{
		"Accept": "application/json",
	}
	if c.apiKey != "" {
		headers["x-api-key"] = c.apiKey
	}
	resp, err := c.http.Get(ctx, c.baseURL, params, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("semantic scholar http %d: %s", resp.StatusCode(), truncate(string(resp.Body()), 200))
	}
	var body searchResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	out := make([]httpentity.SemanticScholarPaper, 0, len(body.Data))
	for _, p := range body.Data {
		out = append(out, paperToEntity(p))
	}
	return out, nil
}

type searchResponse struct {
	Data []paperJSON `json:"data"`
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
