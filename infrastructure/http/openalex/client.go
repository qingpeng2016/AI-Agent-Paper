package openalex

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

// Client 实现 domain/http/repository.OpenAlexRepo
type Client struct {
	http    *httpx.Client
	baseURL string
	mailto  string
}

func NewClient(http *httpx.Client, sources repository.PaperLiteraturePlatformRepo) httprepo.OpenAlexRepo {
	ctx := context.Background()
	baseURL := httpx.ResolveLiteratureBaseURL(
		ctx,
		sources,
		httpx.LiteratureSourceCodeOpenAlex,
		conf.DefaultLiteratureOpenAlexBaseURL,
	)
	mailto := httpx.ResolveOpenAlexMailto(ctx, sources)
	return &Client{http: http, baseURL: baseURL, mailto: mailto}
}

func (c *Client) Count(ctx context.Context, query string) (int, error) {
	_, total, err := c.search(ctx, httpentity.OpenAlexSearchQuery{Query: query, MaxResults: 1, Page: 1})
	return total, err
}

func (c *Client) Search(ctx context.Context, q httpentity.OpenAlexSearchQuery) ([]httpentity.OpenAlexWork, error) {
	works, _, err := c.search(ctx, q)
	return works, err
}

func (c *Client) search(ctx context.Context, q httpentity.OpenAlexSearchQuery) ([]httpentity.OpenAlexWork, int, error) {
	perPage := q.MaxResults
	if perPage <= 0 {
		perPage = 1
	}
	params := map[string]string{
		"search":   q.Query,
		"per-page": strconv.Itoa(perPage),
	}
	if q.Page > 1 {
		params["page"] = strconv.Itoa(q.Page)
	}
	headers := map[string]string{
		"Accept": "application/json",
	}
	if c.mailto != "" {
		headers["User-Agent"] = "AI-Agent-Paper/1.0 (mailto:" + c.mailto + ")"
	}
	resp, err := c.http.Get(ctx, c.baseURL, params, headers)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, 0, fmt.Errorf("openalex http %d", resp.StatusCode())
	}
	var body worksResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, 0, err
	}
	out := make([]httpentity.OpenAlexWork, 0, len(body.Results))
	for _, w := range body.Results {
		out = append(out, workToEntity(w))
	}
	return out, body.Meta.Count, nil
}

type worksResponse struct {
	Meta    metaJSON   `json:"meta"`
	Results []workJSON `json:"results"`
}

type metaJSON struct {
	Count int `json:"count"`
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
