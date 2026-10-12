package pubmed

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

func NewClient(http *httpx.Client, sources repository.PaperLiteraturePlatformRepo) httprepo.PubMedRepo {
	baseURL := httpx.ResolveLiteratureBaseURL(
		context.Background(),
		sources,
		httpx.LiteratureSourceCodePubMed,
		conf.DefaultLiteraturePubMedEutilsBaseURL,
	)
	return &Client{http: http, baseURL: strings.TrimRight(baseURL, "/")}
}

func (c *Client) Count(ctx context.Context, query string) (int, error) {
	ids, count, err := c.esearch(ctx, query, 0, 0)
	_ = ids
	return count, err
}

func (c *Client) Search(ctx context.Context, q httpentity.PubMedSearchQuery) ([]httpentity.PubMedArticle, error) {
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}
	limit := q.MaxResults
	if limit <= 0 {
		limit = 1
	}
	ids, _, err := c.esearch(ctx, q.Query, offset, limit)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return c.esummary(ctx, ids)
}

func (c *Client) esearch(ctx context.Context, query string, offset, limit int) ([]string, int, error) {
	params := map[string]string{
		"db":       "pubmed",
		"term":     query,
		"retmode":  "json",
		"retstart": strconv.Itoa(offset),
	}
	if limit > 0 {
		params["retmax"] = strconv.Itoa(limit)
	} else {
		params["retmax"] = "0"
	}
	url := c.baseURL + "/esearch.fcgi"
	resp, err := c.http.Get(ctx, url, params, nil)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, 0, fmt.Errorf("pubmed esearch http %d", resp.StatusCode())
	}
	var body esearchResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, 0, err
	}
	return body.ESearchResult.IdList, body.ESearchResult.Count, nil
}

type esearchResponse struct {
	ESearchResult struct {
		Count  int      `json:"count"`
		IdList []string `json:"idlist"`
	} `json:"esearchresult"`
}

func (c *Client) esummary(ctx context.Context, ids []string) ([]httpentity.PubMedArticle, error) {
	params := map[string]string{
		"db":      "pubmed",
		"id":      strings.Join(ids, ","),
		"retmode": "json",
	}
	url := c.baseURL + "/esummary.fcgi"
	resp, err := c.http.Get(ctx, url, params, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("pubmed esummary http %d", resp.StatusCode())
	}
	var body esummaryResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	out := make([]httpentity.PubMedArticle, 0, len(ids))
	for _, id := range ids {
		doc, ok := body.Result[id]
		if !ok || doc.Uid == "" {
			continue
		}
		title := strings.TrimSpace(doc.Title)
		if title == "" {
			continue
		}
		year, _ := strconv.Atoi(strings.TrimSpace(doc.PubDate))
		if year == 0 && len(doc.PubDate) >= 4 {
			year, _ = strconv.Atoi(doc.PubDate[:4])
		}
		doi := ""
		for _, idt := range doc.ArticleIds {
			if strings.EqualFold(idt.IdType, "doi") {
				doi = strings.TrimSpace(idt.Value)
				break
			}
		}
		u := "https://pubmed.ncbi.nlm.nih.gov/" + id + "/"
		out = append(out, httpentity.PubMedArticle{
			ExternalKey:   "pubmed:" + id,
			Title:         title,
			Authors:       doc.Authors,
			PublishedYear: year,
			DOI:           doi,
			URL:           u,
		})
	}
	return out, nil
}

type esummaryResponse struct {
	Result map[string]summaryDocJSON `json:"result"`
}

type summaryDocJSON struct {
	Uid        string   `json:"uid"`
	Title      string   `json:"title"`
	PubDate    string   `json:"pubdate"`
	Authors    []string `json:"authors"`
	ArticleIds []struct {
		IdType string `json:"idtype"`
		Value  string `json:"value"`
	} `json:"articleids"`
}
