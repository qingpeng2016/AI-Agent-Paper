package crossref

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

type Client struct {
	http    *httpx.Client
	baseURL string
}

func NewClient(http *httpx.Client, sources repository.PaperLiteraturePlatformRepo) httprepo.CrossrefRepo {
	baseURL := httpx.ResolveLiteratureBaseURL(
		context.Background(),
		sources,
		httpx.LiteratureSourceCodeCrossref,
		conf.DefaultLiteratureCrossrefBaseURL,
	)
	return &Client{http: http, baseURL: baseURL}
}

func (c *Client) Count(ctx context.Context, query string) (int, error) {
	_, total, err := c.search(ctx, query, 1, 0)
	return total, err
}

func (c *Client) Search(ctx context.Context, q httpentity.CrossrefSearchQuery) ([]httpentity.CrossrefWork, error) {
	works, _, err := c.search(ctx, q.Query, q.MaxResults, q.Offset)
	return works, err
}

func (c *Client) search(ctx context.Context, query string, rows, offset int) ([]httpentity.CrossrefWork, int, error) {
	if rows <= 0 {
		rows = 1
	}
	if offset < 0 {
		offset = 0
	}
	params := map[string]string{
		"query":  query,
		"rows":   strconv.Itoa(rows),
		"offset": strconv.Itoa(offset),
	}
	resp, err := c.http.Get(ctx, c.baseURL, params, map[string]string{"Accept": "application/json"})
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, 0, fmt.Errorf("crossref http %d", resp.StatusCode())
	}
	var body worksResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, 0, err
	}
	total := body.Message.TotalResults
	out := make([]httpentity.CrossrefWork, 0, len(body.Message.Items))
	for _, it := range body.Message.Items {
		w := itemToWork(it)
		if w.Title != "" {
			out = append(out, w)
		}
	}
	return out, total, nil
}

type worksResponse struct {
	Message struct {
		TotalResults int        `json:"total-results"`
		Items        []itemJSON `json:"items"`
	} `json:"message"`
}

type itemJSON struct {
	DOI       string   `json:"DOI"`
	Title     []string `json:"title"`
	URL       string   `json:"URL"`
	Issued    *struct {
		DateParts [][]int `json:"date-parts"`
	} `json:"issued"`
	Author []struct {
		Given  string `json:"given"`
		Family string `json:"family"`
	} `json:"author"`
}

func itemToWork(it itemJSON) httpentity.CrossrefWork {
	title := ""
	if len(it.Title) > 0 {
		title = strings.TrimSpace(it.Title[0])
	}
	doi := strings.TrimSpace(it.DOI)
	ext := "crossref:" + doi
	if doi == "" {
		ext = "crossref:" + strings.TrimSpace(it.URL)
	}
	year := 0
	if it.Issued != nil && len(it.Issued.DateParts) > 0 && len(it.Issued.DateParts[0]) > 0 {
		year = it.Issued.DateParts[0][0]
	}
	var authors []string
	for _, a := range it.Author {
		n := strings.TrimSpace(strings.TrimSpace(a.Given) + " " + strings.TrimSpace(a.Family))
		if n != "" {
			authors = append(authors, n)
		}
	}
	u := strings.TrimSpace(it.URL)
	if u == "" && doi != "" {
		u = "https://doi.org/" + doi
	}
	return httpentity.CrossrefWork{
		ExternalKey:   ext,
		Title:         title,
		Authors:       authors,
		PublishedYear: year,
		DOI:           doi,
		URL:           u,
	}
}
