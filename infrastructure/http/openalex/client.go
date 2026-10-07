package openalex

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

// Client 实现 domain/http/repository.OpenAlexRepo
type Client struct {
	http    *httpx.Client
	baseURL string
	mailto  string
}

func NewClient(http *httpx.Client, cfg *conf.Config) httprepo.OpenAlexRepo {
	lit := conf.GetLiteratureConf()
	baseURL := conf.DefaultLiteratureOpenAlexBaseURL
	mailto := ""
	if lit != nil && lit.OpenAlex != nil {
		if u := strings.TrimSpace(lit.OpenAlex.BaseURL); u != "" {
			baseURL = u
		}
		mailto = strings.TrimSpace(lit.OpenAlex.Mailto)
	}
	return &Client{http: http, baseURL: baseURL, mailto: mailto}
}

func (c *Client) Search(ctx context.Context, q httpentity.OpenAlexSearchQuery) ([]httpentity.OpenAlexWork, error) {
	params := map[string]string{
		"search":   q.Query,
		"per-page": strconv.Itoa(q.MaxResults),
	}
	headers := map[string]string{
		"Accept": "application/json",
	}
	if c.mailto != "" {
		headers["User-Agent"] = "AI-Token-Mall/1.0 (mailto:" + c.mailto + ")"
	}
	resp, err := c.http.Get(ctx, c.baseURL, params, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("openalex http %d", resp.StatusCode())
	}
	var body worksResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	out := make([]httpentity.OpenAlexWork, 0, len(body.Results))
	for _, w := range body.Results {
		out = append(out, workToEntity(w))
	}
	return out, nil
}

type worksResponse struct {
	Results []workJSON `json:"results"`
}

type workJSON struct {
	ID              string `json:"id"`
	DisplayName     string `json:"display_name"`
	DOI             string `json:"doi"`
	PublicationYear int    `json:"publication_year"`
	Authorships     []struct {
		Author struct {
			DisplayName string `json:"display_name"`
		} `json:"author"`
	} `json:"authorships"`
}

func workToEntity(w workJSON) httpentity.OpenAlexWork {
	ext := strings.TrimPrefix(w.ID, "https://openalex.org/")
	if ext == w.ID {
		ext = w.ID
	}
	ext = "openalex:" + ext
	doi := strings.TrimPrefix(strings.TrimSpace(w.DOI), "https://doi.org/")
	var authors []string
	for _, a := range w.Authorships {
		n := strings.TrimSpace(a.Author.DisplayName)
		if n != "" {
			authors = append(authors, n)
		}
	}
	return httpentity.OpenAlexWork{
		ExternalKey:   ext,
		Title:         strings.TrimSpace(w.DisplayName),
		Authors:       authors,
		PublishedYear: w.PublicationYear,
		DOI:           doi,
		URL:           w.ID,
	}
}
