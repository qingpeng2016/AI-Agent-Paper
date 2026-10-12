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
		return nil, 0, httpx.LogLiteratureTransportError(ctx, "pubmed", "esearch", url, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, 0, httpx.NewLiteratureHTTPError(ctx, "pubmed", "esearch", url, resp, fmt.Errorf("pubmed esearch http %d", resp.StatusCode()))
	}
	var body esearchResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, 0, httpx.NewLiteratureHTTPError(ctx, "pubmed", "esearch", url, resp, err)
	}
	return body.ESearchResult.IdList, int(body.ESearchResult.Count), nil
}

type esearchResponse struct {
	ESearchResult struct {
		Count  flexJSONInt `json:"count"`
		IdList []string    `json:"idlist"`
	} `json:"esearchresult"`
}

// flexJSONInt NCBI esearch 有时返回 "count":"123" 而非数字。
type flexJSONInt int

func (f *flexJSONInt) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*f = flexJSONInt(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		*f = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = flexJSONInt(n)
	return nil
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
		return nil, httpx.LogLiteratureTransportError(ctx, "pubmed", "esummary", url, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, httpx.NewLiteratureHTTPError(ctx, "pubmed", "esummary", url, resp, fmt.Errorf("pubmed esummary http %d", resp.StatusCode()))
	}
	docs, err := parseEsummaryDocs(resp.Body())
	if err != nil {
		return nil, httpx.NewLiteratureHTTPError(ctx, "pubmed", "esummary", url, resp, err)
	}
	out := make([]httpentity.PubMedArticle, 0, len(ids))
	for _, id := range ids {
		doc, ok := docs[id]
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
		pmcid := ""
		for _, idt := range doc.ArticleIds {
			switch strings.ToLower(strings.TrimSpace(idt.IdType)) {
			case "doi":
				if doi == "" {
					doi = strings.TrimSpace(idt.Value)
				}
			case "pmc", "pmcid":
				if pmcid == "" {
					pmcid = normalizePubMedPMCID(idt.Value)
				}
			}
		}
		u := "https://pubmed.ncbi.nlm.nih.gov/" + id + "/"
		out = append(out, httpentity.PubMedArticle{
			ExternalKey:   "pubmed:" + id,
			Title:         title,
			Authors:       doc.authorNames(),
			PublishedYear: year,
			DOI:           doi,
			URL:           u,
			PMCID:         pmcid,
		})
	}
	return out, nil
}

func parseEsummaryDocs(body []byte) (map[string]summaryDocJSON, error) {
	var wrap struct {
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, err
	}
	out := make(map[string]summaryDocJSON, len(wrap.Result))
	for key, raw := range wrap.Result {
		if key == "uids" {
			continue
		}
		var doc summaryDocJSON
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("pubmed esummary doc %s: %w", key, err)
		}
		out[key] = doc
	}
	return out, nil
}

type summaryDocJSON struct {
	Uid        string           `json:"uid"`
	Title      string           `json:"title"`
	PubDate    string           `json:"pubdate"`
	Authors    pubmedAuthorList `json:"authors"`
	ArticleIds []struct {
		IdType string `json:"idtype"`
		Value  string `json:"value"`
	} `json:"articleids"`
}

func (d summaryDocJSON) authorNames() []string {
	return []string(d.Authors)
}

func normalizePubMedPMCID(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "pmc-id:")
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ";")
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(strings.ToUpper(s), "PMC") {
		s = "PMC" + s
	}
	return strings.ToUpper(s)
}

// pubmedAuthorList NCBI 返回 ["A","B"] 或 [{"name":"A"},…]。
type pubmedAuthorList []string

func (a *pubmedAuthorList) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*a = nil
		return nil
	}
	var names []string
	if err := json.Unmarshal(b, &names); err == nil {
		*a = names
		return nil
	}
	var objs []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(b, &objs); err != nil {
		return err
	}
	out := make([]string, 0, len(objs))
	for _, o := range objs {
		if n := strings.TrimSpace(o.Name); n != "" {
			out = append(out, n)
		}
	}
	*a = out
	return nil
}
